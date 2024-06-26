package main

import (

	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

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

type TxInfo struct {
	TxID   string
	WTxID  string
	Fee    uint64
	Weight uint64
}

type MerkleNode struct {
	Left  *MerkleNode
	Data  []byte
	Right *MerkleNode
}

var blockHeader = BlockHeader{
	Version:     7,
	PrblockHash: "0000000000000000000000000000000000000000000000000000000000000000",
	MerkleRoot:  "",
	Time:        time.Now().Unix(),
	Bits:        0x1f00ffff,
	Nonce:       0,
}


func doubleHash(data []byte) []byte {
	return sha256h(sha256h(data))
}

func sha256h(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func rb(data []byte) []byte {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
	return data
}

const target string = "0000ffff00000000000000000000000000000000000000000000000000000000"

func checkByteArray(a, b []byte) int {
	mini := len(a)
	if len(b) < mini {
		mini = len(b)
	}
	for i := 0; i < mini; i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	} else if len(a) > len(b) {
		return 1
	}
	return 0
}

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

func CreateCoinbase(netReward uint64) *Transaction {
	witComm := witnMerkle()
	return &Transaction{
		Version: 1,
		Vin: []Input{
			{
				TxID:       "0000000000000000000000000000000000000000000000000000000000000000",
				Vout:       0xffffffff,
				Prevout:    Prevout{Scriptpubkey: "0014df4bf9f3621073202be59ae590f55f42879a21a0", ScriptpubkeyAsm: "0014df4bf9f3621073202be59ae590f55f42879a21a0", ScriptpubkeyType: "p2pkh", ScriptpubkeyAddress: "bc1qma9lnumzzpejq2l9ntjepa2lg2re5gdqn3nf0c", Value: uint64(netReward)},
				IsCoinbase: true,
				Sequence:   0xffffffff,
				Scriptsig:  "03951a0604f15ccf5609013803062b9b5a0100072f425443432f20",
				Witness:    []string{"0000000000000000000000000000000000000000000000000000000000000000"},
			},
		},
		Vout: []Prevout{
			{Scriptpubkey: "0014df4bf9f3621073202be59ae590f55f42879a21a0", ScriptpubkeyAsm: "0014df4bf9f3621073202be59ae590f55f42879a21a0", ScriptpubkeyType: "p2pkh", ScriptpubkeyAddress: "bc1qma9lnumzzpejq2l9ntjepa2lg2re5gdqn3nf0c", Value: uint64(netReward)},
			{Scriptpubkey: "6a24" + "aa21a9ed" + witComm, ScriptpubkeyAsm: "OP_RETURN" + "OP_PUSHBYTES_36" + "aa21a9ed" + witComm, ScriptpubkeyType: "op_return", ScriptpubkeyAddress: "bc1qma9lnumzzpejq2l9ntjepa2lg2re5gdqn3nf0c", Value: uint64(0)},
		},
		Locktime: 0,
	}
}

func calculateTxID(serializedTx []byte) string {
	hash := doubleHash(serializedTx)
	reversedHash := rb(hash)
	return hex.EncodeToString(reversedHash)
}

func calculateMerkleRoot(txIDs []string) string {
	merkleTree := merkTree(txIDs)
	return hex.EncodeToString(merkleTree.Data)
}

func merkNode(lnode, rnode *MerkleNode, data []byte) *MerkleNode {
	node := &MerkleNode{}
	if lnode == nil && rnode == nil {
		node.Data = rb(data)
	} else {
		node.Data = doubleHash(append(lnode.Data, rnode.Data...))
	}
	node.Left, node.Right = lnode, rnode
	return node
}

func Comp(a, b TxInfo) bool {
	return float64(a.Fee)/float64(a.Weight) > float64(b.Fee)/float64(b.Weight)
}

func merkTree(leaves []string) *MerkleNode {
	var nodes []MerkleNode
	for _, leaf := range leaves {
		data, _ := hex.DecodeString(leaf)
		var node MerkleNode = *merkNode(nil, nil, data)
		nodes = append(nodes, node)
	}
	for len(nodes) > 1 {
		var newLevel []MerkleNode
		for i := 0; i < len(nodes); i += 2 {
			if len(nodes)%2 != 0 {
				nodes = append(nodes, nodes[len(nodes)-1])
			}
			node := *merkNode(&nodes[i], &nodes[i+1], nil)
			newLevel = append(newLevel, node)
		}
		nodes = newLevel
	}
	return &nodes[0]
}

func witnMerkle() string {
	_, _, wTxIDs := Ordering()
	wTxIDs = append([]string{"0000000000000000000000000000000000000000000000000000000000000000"}, wTxIDs...)
	merkleRoot := merkTree(wTxIDs)
	cm_str := hex.EncodeToString(merkleRoot.Data) + "0000000000000000000000000000000000000000000000000000000000000000"
	witnComm, _ := hex.DecodeString(cm_str)
	witnComm = sha256h(sha256h(witnComm))
	fmt.Println("Witness Commitment: ", hex.EncodeToString(witnComm))
	return hex.EncodeToString(witnComm)
}

func Ordering() (uint64, []string, []string) {
	var allowedTxIDs, notAllowedTxIDs []string
	Directory := "./mempool"
	files, _ := os.ReadDir(Directory)
	var txInfo []TxInfo
	for _, file := range files {
		txData, _ := os.ReadFile(Directory + "/" + file.Name())
		var tx Transaction
		json.Unmarshal(txData, &tx)
		var fee uint64
		for _, vin := range tx.Vin {
			fee += vin.Prevout.Value
		}
		for _, vout := range tx.Vout {
			fee -= vout.Value
		}
		serlzd := serTx(&tx)
		segserlzd := SegWitSerialize(&tx)
		txID := rb(doubleHash(serlzd))
		wtxID := rb(doubleHash(segserlzd))
		txInfo = append(txInfo, TxInfo{TxID: hex.EncodeToString(txID), WTxID: hex.EncodeToString(wtxID), Fee: fee, Weight: uint64(calWitSize(&tx) + calBaseSize(&tx)*4)})
	}
	sort.Slice(txInfo, func(i, j int) bool {
		return Comp(txInfo[i], txInfo[j])
	})
	var reward uint64
	var maxWeight uint64 = 3900000
	for _, tx := range txInfo {
		if maxWeight >= tx.Weight {
			maxWeight -= tx.Weight
			allowedTxIDs = append(allowedTxIDs, tx.TxID)
			notAllowedTxIDs = append(notAllowedTxIDs, tx.WTxID)
			reward += tx.Fee
		}
	}
	return reward, allowedTxIDs, notAllowedTxIDs
}

func calBaseSize(tx *Transaction) int {
	return len(serTx(tx))
}

func calWitSize(tx *Transaction) int {
	if !isSegWitTransaction(tx) {
		return 0
	}
	var serlzd []byte
	if isSegWitTransaction(tx) {
		serlzd = append(serlzd, []byte{0x00, 0x01}...)
		for _, vin := range tx.Vin {
			serlzd = append(serlzd, SerializeVarInt(uint64(len(vin.Witness)))...)
			for _, witness := range vin.Witness {
				witnessBytes, _ := hex.DecodeString(witness)
				serlzd = append(serlzd, SerializeVarInt(uint64(len(witnessBytes)))...)
				serlzd = append(serlzd, witnessBytes...)
			}
		}
	}
	return len(serlzd)
}

func SerializeVarInt(n uint64) []byte {
	if n < 0xfd {
		return []byte{byte(n)}
	} else if n <= 0xffff {
		return append([]byte{0xfd}, u16ToB(uint16(n))...)
	} else if n <= 0xffffffff {
		return append([]byte{0xfe}, u32ToB(uint32(n))...)
	} else {
		return append([]byte{0xff}, u64ToB(n)...)
	}
}

func isSegWitTransaction(tx *Transaction) bool {
	for _, vin := range tx.Vin {
		if len(vin.Witness) > 0 {
			return true
		}
	}
	return false
}

func serTx(tx *Transaction) []byte {
	var serlzd []byte
	serlzd = append(serlzd, u32ToB(tx.Version)...)
	serlzd = append(serlzd, SerializeVarInt(uint64(len(tx.Vin)))...)
	for _, vin := range tx.Vin {
		txidBytes, _ := hex.DecodeString(vin.TxID)
		serlzd = append(serlzd, rb(txidBytes)...)
		serlzd = append(serlzd, u32ToB(vin.Vout)...)
		Scriptsig_bytes, _ := hex.DecodeString(vin.Scriptsig)
		serlzd = append(serlzd, SerializeVarInt(uint64(len(Scriptsig_bytes)))...)
		serlzd = append(serlzd, Scriptsig_bytes...)
		serlzd = append(serlzd, u32ToB(vin.Sequence)...)
	}
	serlzd = append(serlzd, SerializeVarInt(uint64(len(tx.Vout)))...)
	for _, vout := range tx.Vout {
		serlzd = append(serlzd, u64ToB(vout.Value)...)
		scriptPubKeyBytes, _ := hex.DecodeString(vout.Scriptpubkey)
		serlzd = append(serlzd, SerializeVarInt(uint64(len(scriptPubKeyBytes)))...)
		serlzd = append(serlzd, scriptPubKeyBytes...)
	}
	serlzd = append(serlzd, u32ToB(tx.Locktime)...)
	return serlzd
}

func srlzBhead(bh *BlockHeader) []byte {
	var serlzd []byte
	serlzd = append(serlzd, u32ToB(bh.Version)...)
	prblockHashbytes, _ := hex.DecodeString(bh.PrblockHash)
	serlzd = append(serlzd, prblockHashbytes...)
	merkRootB, _ := hex.DecodeString(bh.MerkleRoot)
	serlzd = append(serlzd, merkRootB...)
	bh.Time = time.Now().Unix()
	serlzd = append(serlzd, u32ToB(uint32(bh.Time))...)
	serlzd = append(serlzd, u32ToB(bh.Bits)...)
	serlzd = append(serlzd, u32ToB(bh.Nonce)...)
	return serlzd
}

func SegWitSerialize(tx *Transaction) []byte {
	var serlzd []byte
	serlzd = append(serlzd, u32ToB(tx.Version)...)
	if isSegWitTransaction(tx) {
		serlzd = append(serlzd, []byte{0x00, 0x01}...)
	}
	serlzd = append(serlzd, SerializeVarInt(uint64(len(tx.Vin)))...)
	for _, vin := range tx.Vin {
		txidBytes, _ := hex.DecodeString(vin.TxID)
		serlzd = append(serlzd, rb(txidBytes)...)
		serlzd = append(serlzd, u32ToB(vin.Vout)...)
		Scriptsig_bytes, _ := hex.DecodeString(vin.Scriptsig)
		serlzd = append(serlzd, SerializeVarInt(uint64(len(Scriptsig_bytes)))...)
		serlzd = append(serlzd, Scriptsig_bytes...)
		serlzd = append(serlzd, u32ToB(vin.Sequence)...)
	}
	serlzd = append(serlzd, SerializeVarInt(uint64(len(tx.Vout)))...)
	for _, vout := range tx.Vout {
		serlzd = append(serlzd, u64ToB(vout.Value)...)
		scriptPubKeyBytes, _ := hex.DecodeString(vout.Scriptpubkey)
		serlzd = append(serlzd, SerializeVarInt(uint64(len(scriptPubKeyBytes)))...)
		serlzd = append(serlzd, scriptPubKeyBytes...)
	}
	if isSegWitTransaction(tx) {
		for _, vin := range tx.Vin {
			serlzd = append(serlzd, SerializeVarInt(uint64(len(vin.Witness)))...)
			for _, witness := range vin.Witness {
				witnessBytes, _ := hex.DecodeString(witness)
				serlzd = append(serlzd, SerializeVarInt(uint64(len(witnessBytes)))...)
				serlzd = append(serlzd, witnessBytes...)
			}
		}
	}
	serlzd = append(serlzd, u32ToB(tx.Locktime)...)
	return serlzd
}

func mineBlock(header *BlockHeader) bool {
	return proofOfWork(header)
}

func writeBlockData(header BlockHeader, coinbaseTx *Transaction, txIDs []string) {
	outputF, _ := os.Create("output.txt")
	defer outputF.Close()
	serializedBlockHeader := srlzBhead(&header)
	segWitCoinbaseTx := SegWitSerialize(coinbaseTx)
	outputF.WriteString(hex.EncodeToString(serializedBlockHeader) + "\n")
	outputF.WriteString(hex.EncodeToString(segWitCoinbaseTx) + "\n")
	for _, txID := range txIDs {
		outputF.WriteString(txID + "\n")
	}
}

func main() {
	networkReward, transactionIDs, _ := Ordering()
	coinbaseTx := CreateCoinbase(networkReward)
	serializedCoinbaseTx := serTx(coinbaseTx)
	coinbaseTxID := calculateTxID(serializedCoinbaseTx)
	transactionIDs = append([]string{coinbaseTxID}, transactionIDs...)
	merkleRoot := calculateMerkleRoot(transactionIDs)
	blockHeader.MerkleRoot = merkleRoot
	if mineBlock(&blockHeader) {
		writeBlockData(blockHeader, coinbaseTx, transactionIDs)
	}
}
