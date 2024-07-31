package blockchain

import "crypto/sha256"

type MerKleTree struct {
	RootNode *MerKleNode
}

type MerKleNode struct {
	Left  *MerKleNode
	Right *MerKleNode
	Data  []byte
}

func NewMerKleTree(data [][]byte) *MerKleTree {
	var nodes []MerKleNode
	if len(data) %2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(data)/2; i++ {
		var newLevel []MerKleNode
		for j := 0; j < len(nodes); j+=2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, node)
		}
		nodes = newLevel
	}
	return &MerKleTree{&nodes[0]}
}

func NewMerkleNode(left, right *MerKleNode, data[]byte) * MerKleNode {
	node := MerKleNode{}
	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else {
		preData := append(left.Data, right.Data...)
		hash := sha256.Sum256(preData)
		node.Data = hash[:]
	}
	node.Left = left
	node.Right = right
	return &node
}





















