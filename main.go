package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

)
func u16ToB(num uint16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, num)
	return buf
}

func u32ToB(num uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, num)
	return buf
}

func u64ToB(num uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, num)
	return buf
}

type BlockHeader struct {
	Version     uint32
	PrblockHash string
	MerkleRoot  string
	Time        int64
	Bits        uint32
	Nonce       uint32
}

type Input struct {
	TxID         string   `json:"txid"`
	Vout         uint32   `json:"vout"`
	Prevout      Prevout  `json:"prevout"`
	Scriptsig    string   `json:"scriptsig"`
	ScriptsigAsm string   `json:"scriptsig_asm"`
	Witness      []string `json:"witness"`
	IsCoinbase   bool     `json:"is_coinbase"`
	Sequence     uint32   `json:"sequence"`
}

type Prevout struct {
	Scriptpubkey        string `json:"scriptpubkey"`
	ScriptpubkeyAsm     string `json:"scriptpubkey_asm"`
	ScriptpubkeyType    string `json:"scriptpubkey_type"`
	ScriptpubkeyAddress string `json:"scriptpubkey_address"`
	Value               uint64 `json:"value"`
}

type Transaction struct {
	Version  uint32    `json:"version"`
	Locktime uint32    `json:"locktime"`
	Vin      []Input   `json:"vin"`
	Vout     []Prevout `json:"vout"`
}

type MerkleNode struct {
	Left  *MerkleNode
	Data  []byte
	Right *MerkleNode
}
func rb(data []byte) []byte {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
	return data
}

const target string = "0000ffff00000000000000000000000000000000000000000000000000000000"

func proofOfWork(blockHeader *BlockHeader) bool {
	targetBytes, _ := hex.DecodeString(target)
	for blockHeader.Nonce <= 0xffffffff {
		sh := srlzBhead(blockHeader)
		hash := rb(doubleHash(sh))
		if checkByteArray(hash, targetBytes) == -1 {
			return true
		}
		blockHeader.Nonce++
	}
	return false
}

func serTx(tx *Transaction) []byte {
	var serlzd []byte
	serlzd = append(serlzd, u32ToB(tx.Version)...)
	for _, vin := range tx.Vin {
		txidBytes, _ := hex.DecodeString(vin.TxID)
		serlzd = append(serlzd, rb(txidBytes)...)
		serlzd = append(serlzd, u32ToB(vin.Vout)...)
		serlzd = append(serlzd, u32ToB(vin.Sequence)...)
	}
	for _, vout := range tx.Vout {
		serlzd = append(serlzd, u64ToB(vout.Value)...)
		serlzd = append(serlzd, scriptpubkey...)
	}
	return serlzd
}

func srlzBhead(bh *BlockHeader) []byte {
	var serlzd []byte
	serlzd = append(serlzd, u32ToB(bh.Version)...)
	prblockHashbytes, _ := hex.DecodeString(bh.PrblockHash)
	serlzd = append(serlzd, u32ToB(bh.Bits)...)
	serlzd = append(serlzd, u32ToB(bh.Nonce)...)
	
	return serlzd
}

func calculateTxID(serializedTx []byte) string {
	hash := doubleHash(serializedTx)
	reversedHash := rb(hash)
	return hex.EncodeToString(reversedHash)
}

func doubleHash(data []byte) []byte {
	return sha256h(sha256h(data))
}

func sha256h(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func merkNode(lnode, rnode *MerkleNode, data []byte) *MerkleNode {
	var node MerkleNode
	node.Left, node.Right = lnode, rnode
	return node
}

func merkTree(leaves []string) *MerkleNode {
	var nodes []MerkleNode
	for _, leaf := range leaves {
		data, _ := hex.DecodeString(leaf)
		var node MerkleNode = *merkNode(nil, nil, data)
		nodes = append(nodes, node)
	}
	return &nodes[0]
}