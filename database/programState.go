package database

import ()

import (
	"bytes"
	"encoding/gob"

	"github.com/FactomProject/factomd/common/primitives"
)

var ProgramStateStr []byte = []byte("ProgramState")

type ProgramStateBase struct {
	LastBitcoinBlockChecked       string
	LastEthereumBlockChecked      int64
	LastFactomDBlockHeightChecked uint32
}

type ProgramState struct {
	ProgramStateBase
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
