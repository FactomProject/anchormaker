package main

import (
	"fmt"
	"time"

	"github.com/FactomProject/factom"

	"github.com/FactomProject/anchormaker/api"
	"github.com/FactomProject/anchormaker/config"
	anchorFactom "github.com/FactomProject/anchormaker/factom"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/primitives"
)

func main() {
	//TODO: setup API, etc.

	c := config.ReadConfig()
	anchorFactom.LoadConfig(c)
	api.SetServer(c.Factom.FactomdAddress)

	fBalance, ecBalance, err := anchorFactom.CheckFactomBalance()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Balances - %v, %v\n", fBalance, ecBalance)

	if ecBalance < c.Factom.ECBalanceThreshold {
		if fBalance < c.Factom.FactoidBalanceThreshold {
			fmt.Printf("EC and F Balances are too low, can't do anything!\n")
			return
		}
		err = anchorFactom.TopupECAddress()
		if err != nil {
			panic(err)
		}
	}

	err = CheckAndCreateBitcoinAnchorChain()
	if err != nil {
		panic(err)
	}

	err = CheckAndCreateEthereumAnchorchain()
	if err != nil {
		panic(err)
	}
}

func CheckAndCreateBitcoinAnchorChain() error {
	anchor := CreateFirstBitcoinAnchorEntry()
	chainID := anchor.GetChainID()

	head, err := factom.GetChainHead(chainID.String())
	if err != nil {
		if err.Error() != "Missing Chain Head" {
			return err
		}
	}
	if head != "" {
		//Chain already exists, nothing to create!
		return nil
	}

	err = CreateChain(anchor)
	if err != nil {
		return err
	}

	return nil
}

func CheckAndCreateEthereumAnchorchain() error {
	anchor := CreateFirstEthereumAnchorEntry()
	chainID := anchor.GetChainID()

	head, err := factom.GetChainHead(chainID.String())
	if err != nil {
		if err.Error() != "Missing Chain Head" {
			return err
		}
	}
	if head != "" {
		//Chain already exists, nothing to create!
		return nil
	}

	err = CreateChain(anchor)
	if err != nil {
		return err
	}

	return nil
}

func CreateChain(e *entryBlock.Entry) error {
	tx1, tx2, err := anchorFactom.JustFactomizeChain(e)
	if err != nil {
		return err
	}

	fmt.Printf("Created chain %v - %v, %v\n", e.GetChainID(), tx1, tx2)

	for i := 0; ; i++ {
		i = i % 3
		time.Sleep(5 * time.Second)
		ack, err := factom.FactoidACK(tx1, "")
		if err != nil {
			panic(err)
		}
		str, err := primitives.EncodeJSONString(ack)
		if err != nil {
			panic(err)
		}
		fmt.Printf("ack1 - %v", str)
		for j := 0; j < i+1; j++ {
			fmt.Printf(".")
		}
		fmt.Printf("  \r")

		if ack.Status != "DBlockConfirmed" {
			continue
		}
		fmt.Printf("ack1 - %v\n", str)

		break
	}

	for i := 0; ; i++ {
		i = i % 3
		time.Sleep(5 * time.Second)
		ack, err := factom.FactoidACK(tx2, "")
		if err != nil {
			panic(err)
		}

		str, err := primitives.EncodeJSONString(ack)
		if err != nil {
			panic(err)
		}
		fmt.Printf("ack2 - %v", str)
		for j := 0; j < i+1; j++ {
			fmt.Printf(".")
		}
		fmt.Printf("  \r")

		if ack.Status != "DBlockConfirmed" {
			continue
		}

		fmt.Printf("ack2 - %v\n", str)

		break
	}

	return nil
}

func CreateFirstBitcoinAnchorEntry() *entryBlock.Entry {
	answer := new(entryBlock.Entry)

	answer.Version = 0
	answer.ExtIDs = []primitives.ByteSlice{primitives.ByteSlice{Bytes: []byte("FactomAnchorChain")}}
	answer.Content = primitives.ByteSlice{Bytes: []byte("This is the Factom anchor chain, which records the anchors Factom puts on Bitcoin and other networks.\n")}
	answer.ChainID = entryBlock.NewChainID(answer)

	return answer
}

func CreateFirstEthereumAnchorEntry() *entryBlock.Entry {
	answer := new(entryBlock.Entry)

	answer.Version = 0
	answer.ExtIDs = []primitives.ByteSlice{primitives.ByteSlice{Bytes: []byte("FactomEthereumAnchorChain")}}
	answer.Content = primitives.ByteSlice{Bytes: []byte("This is the Factom Ethereum anchor chain, which records the anchors Factom puts on the Ethereum network.\n")}
	answer.ChainID = entryBlock.NewChainID(answer)

	return answer
}
