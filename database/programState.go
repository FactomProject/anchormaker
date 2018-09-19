package database

import ()

import (
	"bytes"
	"encoding/gob"
	"sync"

	"github.com/FactomProject/factomd/common/primitives"
)

var ProgramStateStr []byte = []byte("ProgramState")

type ProgramStateBase struct {
	LastEthereumBlockChecked      int64
	LastFactomDBlockHeightChecked uint32
	LastConfirmedAnchorDBlockHeight uint32
	PendingTx *PendingTxInfo
}
type PendingTxInfo struct {
	Nonce          uint64
	IsMandatory    bool   // whether or not this transaction needs to go through for this Directory Block height
	EthTxID        string // the transactionid as it can be found in the eth blockchain
	EthTxGasPrice  int64  // the eth/gas that this transaction is offering
	FactomDBheight uint32 // the factom directory block height that this transaction updates
	FactomDBkeyMR  string // the factom directory block key merkle root that this transaction sets
	TxTime         int64  // the unix time that this transaction was created and broadcast into the eth p2p network
}

type ProgramState struct {
	ProgramStateBase
	ProgramStateMutex sync.Mutex
}

func (e *ProgramState) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *ProgramState) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *ProgramState) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *ProgramState) String() string {
	str, _ := e.JSONString()
	return str
}

func (e *ProgramState) MarshalBinary() ([]byte, error) {
	var data primitives.Buffer

	enc := gob.NewEncoder(&data)

	err := enc.Encode(e.ProgramStateBase)
	if err != nil {
		return nil, err
	}
	return data.DeepCopyBytes(), nil
}

func (e *ProgramState) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	dec := gob.NewDecoder(primitives.NewBuffer(data))
	adb := ProgramStateBase{}
	err = dec.Decode(&adb)
	if err != nil {
		return nil, err
	}
	e.ProgramStateBase = adb
	return nil, nil
}

func (e *ProgramState) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}
