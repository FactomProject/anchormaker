package main

import (
	"fmt"

	"github.com/FactomProject/factom"

	"github.com/FactomProject/anchormaker/api"
	"github.com/FactomProject/anchormaker/config"
	anchorFactom "github.com/FactomProject/anchormaker/factom"

	"github.com/FactomProject/factomd/common/entryBlock"
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
		/*err = anchorFactom.TopupECAddress()
		if err != nil {
			panic(err)
		}*/
	}

	err = CheckAndCreateBitcoinAnchorChain()
	if err != nil {
		panic(err)
	}

	err = CheckAndCreateEthereumAnchorchain()
	if err != nil {
		panic(err)
	}
	/*
		tx, err := factom.FactoidACK("579082d3fd55f3edb110de1e15e835b9d9f9c9d5349500360fcd135e56ee9425", "")
		fmt.Printf("tx, err - %v, %v", tx, err)*/
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
	fmt.Printf("Created chain %v\n", e.GetChainID())
	return nil
}

func CreateFirstBitcoinAnchorEntry() *entryBlock.Entry {
	answer := new(entryBlock.Entry)

	answer.Version = 0
	answer.ExtIDs = [][]byte{[]byte("FactomAnchorChain")}
	answer.Content = []byte("This is the Factom anchor chain, which records the anchors Factom puts on Bitcoin and other networks.\n")
	answer.ChainID = entryBlock.NewChainID(answer)

	return answer
}

func CreateFirstEthereumAnchorEntry() *entryBlock.Entry {
	answer := new(entryBlock.Entry)

	answer.Version = 0
	answer.ExtIDs = [][]byte{[]byte("FactomEthereumAnchorChain")}
	answer.Content = []byte("This is the Factom Ethereum anchor chain, which records the anchors Factom puts on the Ethereum network.\n")
	answer.ChainID = entryBlock.NewChainID(answer)

	return answer
}
