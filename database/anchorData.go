package database

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var AnchorDataStr []byte = []byte("AnchorData")

type AnchorDataBase struct {
	DBlockKeyMR  string
	DBlockHeight uint32

	BitcoinRecordHeight     uint32
	BitcoinRecordEntryHash  string
	EthereumRecordHeight    uint32
	EthereumRecordEntryHash string

	Bitcoin struct {
		Address     string //"1HLoD9E4SDFFPDiYfNYnkBLQ85Y51J3Zb1",
		TXID        string //"9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5", BTC Hash - in reverse byte order
		BlockHeight int64  //345678,
		BlockHash   string //"00000000000000000cc14eacfc7057300aea87bed6fee904fd8e1c1f3dc008d4", BTC Hash - in reverse byte order
		Offset      int64  //87
	}

	Ethereum struct {
		Address     string //0x30aa981f6d2fce81083e584c8ee2f822b548752f
		TXID        string //0x50ea0effc383542811a58704a6d6842ed6d76439a2d942d941896ad097c06a78
		BlockHeight int64  //293003
		BlockHash   string //0x3b504616495fc9cf7be9b5b776692a9abbfb95491fa62abf62dcdf4d53ff5979
		Offset      int64  //0
		//Input       string //0x085f451f0000000000000000000000000000000000000000000000000000000000000000
	}
}

type AnchorData struct {
	AnchorDataBase
}

func (e *AnchorData) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AnchorData) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *AnchorData) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *AnchorData) String() string {
	str, _ := e.JSONString()
	return str
}

var _ interfaces.DatabaseBatchable = (*AnchorData)(nil)

func (c *AnchorData) IsComplete() bool {
	return (c.BitcoinRecordHeight > 0)
	//return (c.EthereumRecordHeight > 0)
	//return (c.BitcoinRecordHeight > 0) && (c.EthereumRecordHeight > 0)
}

func (c *AnchorData) New() interfaces.BinaryMarshallableAndCopyable {
	return new(AnchorData)
}

func (e *AnchorData) GetDatabaseHeight() uint32 {
	return e.DBlockHeight
}

func (e *AnchorData) DatabasePrimaryIndex() interfaces.IHash {
	return UintToHash(e.DBlockHeight)
}

func (e *AnchorData) DatabaseSecondaryIndex() interfaces.IHash {
	h, err := primitives.NewShaHashFromStr(e.DBlockKeyMR)
	if err != nil {
		panic(err)
	}
	return h
}

func UintToHash(i uint32) interfaces.IHash {
	h, err := primitives.NewShaHashFromStr(fmt.Sprintf("%032x", i))
	if err != nil {
		panic(err)
	}
	return h
}

func (e *AnchorData) GetChainID() interfaces.IHash {
	h, err := primitives.NewShaHashFromStr(fmt.Sprintf("%032x", AnchorDataStr))
	if err != nil {
		panic(err)
	}
	return h
}

func (e *AnchorData) MarshalBinary() ([]byte, error) {
	var data primitives.Buffer

	enc := gob.NewEncoder(&data)

	err := enc.Encode(e.AnchorDataBase)
	if err != nil {
		return nil, err
	}
	return data.DeepCopyBytes(), nil
}

func (e *AnchorData) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	dec := gob.NewDecoder(primitives.NewBuffer(data))
	adb := AnchorDataBase{}
	err = dec.Decode(&adb)
	if err != nil {
		return nil, err
	}
	e.AnchorDataBase = adb
	return nil, nil
}

func (e *AnchorData) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}
