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
	MerkleRoot   string // Merkle Root of a 1000 block window of Directory Blocks
	DBlockHeight uint32 // Maximum height that is in the 1000 block window

	EthereumRecordHeight    uint32 // Directory Block height for this Ethereum anchor's record within Factom
	EthereumRecordEntryHash string // Entry Hash for this Ethereum anchor's record within Factom

	Ethereum struct {
		ContractAddress string // Contract Address that the anchor was put into
		TxID            string //0x50ea0effc383542811a58704a6d6842ed6d76439a2d942d941896ad097c06a78
		BlockHeight     int64  //293003
		BlockHash       string //0x3b504616495fc9cf7be9b5b776692a9abbfb95491fa62abf62dcdf4d53ff5979
		TxIndex           int64  // Transaction index within its block
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

// IsSubmitted returns whether or not an Ethereum transaction has been submitted for this anchor
func (c *AnchorData) IsSubmitted() bool {
	return c.Ethereum.TxID != ""
}

// IsComplete returns whether or not a given anchor has been recorded back into Factom
func (c *AnchorData) IsComplete() bool {
	return c.EthereumRecordHeight > 0
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
	h, err := primitives.NewShaHashFromStr(e.MerkleRoot)
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
