package bitcoin

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

type Transaction struct {
	InputAddresses []string
	OPReturn       string
	TxHash         string
	BlockNumber    int64
	BlockHash      string
	OpReturnIndex  int64
}

func (t *Transaction) IsOurs(ourAddress string) bool {
	for _, w := range t.InputAddresses {
		if w == ourAddress {
			return true
		}
	}
	return false
}

func (t *Transaction) GetAnchorData() (dBlockHeight uint32, keyMR string) {
	return opReturnScriptToParts(t.OPReturn)
}

func (t *Transaction) GetHash() string {
	return t.TxHash
}

func (t *Transaction) GetBlockNumber() int64 {
	return t.BlockNumber
}

func (t *Transaction) GetBlockHash() string {
	return t.BlockHash
}

func (t *Transaction) GetTransactionIndex() int64 {
	return t.OpReturnIndex
}

func prependBlockHeight(height uint32, hash []byte) ([]byte, error) {
	// dir block genesis block height starts with 0, for now
	// similar to bitcoin genesis block
	h := uint64(height)
	if 0xFFFFFFFFFFFF&h != h {
		return nil, fmt.Errorf("bad block height")
	}
	header := []byte{'F', 'a'}
	big := make([]byte, 8)
	binary.BigEndian.PutUint64(big, h) //height)
	newdata := append(big[2:8], hash...)
	newdata = append(header, newdata...)
	return newdata, nil
}

func opReturnScriptToParts(script string) (dBlockHeight uint32, keyMR string) {
	//466100000001031c8c948a27d82fbc2e30383f62a3bb499997c36eed9999252908ed6a865bd746aa
	if len(script) != 80 {
		return 0, ""
	}
	if script[:4] != "4661" {
		return 0, ""
	}
	script = script[4:]
	i, err := strconv.ParseInt(script[:12], 16, 64)
	if err != nil {
		return 0, ""
	}
	dBlockHeight = uint32(i)
	keyMR = script[12:]
	return
}
