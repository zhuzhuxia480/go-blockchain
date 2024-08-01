package blockchain

import "log"

func (cli *CLI) startNode(nodeID, minerAddress string) {
	log.Println("Start Node node:", nodeID)
	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			log.Println("Minging is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panicln("wrong miner address")
		}
	}
	StartServer(nodeID, minerAddress)
}
