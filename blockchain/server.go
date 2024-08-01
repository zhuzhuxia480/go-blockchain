package blockchain

import (
	"bytes"
	"debug/dwarf"
	"encoding/gob"
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
	Block []byte
}

type getblocks struct {
	AddrFrom string
}

type getdata struct {
	AddrFrom string
	Type string
	ID []byte
}

type inv struct {
	AddrFrom string
	Type string
	Item [][]byte
}

type tx struct {
	AddrFrom string
	Transaction []byte
}

type verzion struct {
	Version int
	BestHeight int
	AddrFrom string
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

func sendVersion(addr string, bc *BlockChain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}

func handleConnection(conn net.Conn, bc *BlockChain)  {
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
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handleVersion(request, bc)
	default:
		log.Println("Unknown command!")
	}
}

func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panicln(err)
	}
	defer ln.Close()

	bc := NewBlockChain()

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























