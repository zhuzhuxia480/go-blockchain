package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}
var blocksInTransit = [][]byte{}
var mempoll = make(map[string]Transaction)

type addr struct {
	AddrList []string
}

type block struct {
	AddrFrom string
	Block    []byte
}

type getblocks struct {
	AddrFrom string
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type inv struct {
	AddrFrom string
	Type     string
	Item     [][]byte
}

type tx struct {
	AddrFrom    string
	Transaction []byte
}

type verzion struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte
	for i, c := range command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return string(command)
}

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}

func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}

func sendAddr(address string) {
	nodes := addr{knownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)
	sendData(address, request)
}

func sendBlock(addr string, b *Block) {
	data := block{
		AddrFrom: addr,
		Block:    b.Serialize(),
	}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)
	sendData(addr, request)
}

func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		log.Println(addr, " is not avaliable!")
		var updateNodes []string
		for _, node := range knownNodes {
			if node != addr {
				updateNodes = append(updateNodes, node)
			}
		}
		knownNodes = updateNodes
		return
	}
	defer conn.Close()
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panicln(err)
	}
}

func sendInv(address, kind string, items [][]byte) {
	inventory := inv{
		AddrFrom: address,
		Type:     kind,
		Item:     items,
	}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)
	sendData(address, request)
}

func sendGetBlocks(address string) {
	payload := gobEncode(getblocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)
	sendData(address, request)
}

func sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{address, kind, id})
	request := append(commandToBytes("getdata"), payload...)
	sendData(address, request)
}

func sendTx(address string, tnx *Transaction) {
	data := tx{
		AddrFrom:    address,
		Transaction: tnx.Serialize(),
	}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)
	sendData(address, request)
}

func sendVersion(addr string, bc *BlockChain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}

func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr
	buff.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panicln(err)
	}

	knownNodes = append(knownNodes, payload.AddrList...)
	log.Printf("There are %d known nodes now~\n", len(knownNodes))
	requestBlocks()
}

func handleBlock(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload block
	buff.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panicln(err)
	}

	block := DeSerializeBlock(payload.Block)
	log.Println("Received a new block")
	bc.AddBlock(block)
	log.Printf("Added block, hash:%x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)
		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()
	}
}

func handleGetBlock(request []byte, bc *BlockChain) {
	var buf bytes.Buffer
	var payload block

	buf.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buf)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panicln(err)
	}

	blockData := payload.Block
	block := DeSerialize(blockData)

	log.Println("Received a new block")
	bc.AddBlock(block)
	log.Printf("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)
		blocksInTransit = blocksInTransit[1:]
	} else { // TODO
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()
	}
}

func handleInv(request []byte, bc *BlockChain) {
	var buf bytes.Buffer
	var payload inv

	buf.Write(request)
	decoder := gob.NewDecoder(&buf)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panicln(err)
	}

	log.Printf("Received inventory with %d %s", len(payload.Item), payload.Type)

	if payload.Type == "block" {
		blocksInTransit := payload.Item
		blockHash := payload.Item[0]
		sendGetData(payload.AddrFrom, "block", blockHash)
		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}
	if payload.Type == "tx" {
		txID := payload.Item[0]
		if mempoll[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func handleGetBlocks(request []byte, bc *BlockChain) {
	var buf bytes.Buffer
	var payload getblocks
	buf.Write(request)
	decoder := gob.NewDecoder(&buf)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panicln(err)
	}
	hashes := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", hashes)
}

func handleGetData(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload getdata
	buff.Write(request)
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panicln(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}
		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempoll[txID]
		sendTx(payload.AddrFrom, &tx)
	}
}

func handleTx(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request)
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panicln(err)
	}
	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	mempoll[hex.EncodeToString(tx.ID)] = tx

	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddrFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(mempoll) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*Transaction
			for _, tx := range mempoll {
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				log.Println("All transactions are invalid, waiting for new ones")
				return
			}

			cbTx := NewCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)

			newblock := bc.MineBLock(txs)
			UTXOSet := UTXOSet{bc}
			UTXOSet.Reindex()

			log.Println("New block is mined")
			for _, tx := range txs {
				delete(mempoll, hex.EncodeToString(tx.ID))
			}

			for _, node := range knownNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newblock.Hash})
				}
			}

			if len(mempoll) > 0 {
				goto MineTransactions
			}
		}
	}
}

func handleVersion(request []byte, bc *BlockChain) {
	var buff bytes.Buffer
	var payload verzion

	buff.Write(request[commandLength:])
	decoder := gob.NewDecoder(&buff)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panicln(err)
	}
	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight
	if myBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}
}

func handleConnection(conn net.Conn, bc *BlockChain) {
	request, err := io.ReadAll(conn)
	if err != nil {
		log.Panicln(err)
	}

	command := bytesToCommand(request[:commandLength])
	log.Println("Received command:", command)

	switch command {
	case "addr":
		handleAddr(request)
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getblocks":
		handleGetBlock(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handleVersion(request, bc)
	default:
		log.Println("Unknown command!")
	}
	conn.Close()
}

func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panicln(err)
	}
	defer ln.Close()

	bc := NewBlockChain(nodeID)

	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panicln(err)
		}
		go handleConnection(conn, bc)
	}
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panicln(err)
	}
	return buff.Bytes()
}

func nodeIsKnown(add string) bool {
	for _, node := range knownNodes {
		if node == add {
			return true
		}
	}
	return false
}
