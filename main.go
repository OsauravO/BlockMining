package main

import (
	"crypto/sha256"
	"encoding/hex"

)

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

func proofOfWork(blockHeader *BlockHeader) bool {
	
}

func serTx(tx *Transaction) []byte {
	
}

func srlzBhead(bh *BlockHeader) []byte {
	
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