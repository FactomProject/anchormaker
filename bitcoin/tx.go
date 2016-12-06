package bitcoin

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/blockcypher/gobcy"
)

type Transaction struct {
	TX gobcy.TX
}

func (t *Transaction) IsOurs(ourAddress string) bool {
	for _, v := range t.TX.Inputs {
		for _, w := range v.Addresses {
			if w == BTCAddress {
				return true
			}
		}
	}
	return false
}

func (t *Transaction) GetAnchorData() (dBlockHeight uint32, keyMR string) {
	for _, v := range t.TX.Outputs {
		if v.DataHex != "" {
			return opReturnScriptToParts(v.DataHex)
		}
	}
	return 0, ""
}

func (t *Transaction) GetHash() string {
	return t.TX.Hash
}

func (t *Transaction) GetBlockNumber() int64 {
	return int64(t.TX.BlockHeight)
}

func (t *Transaction) GetBlockHash() string {
	return t.TX.BlockHash
}

func (t *Transaction) GetTransactionIndex() int64 {
	for i, v := range t.TX.Outputs {
		if v.DataHex != "" {
			return int64(i)
		}
	}
	return 0
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
