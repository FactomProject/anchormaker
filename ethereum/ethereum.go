package main

import (
	"fmt"

	"github.com/FactomProject/EthereumAPI"
)

//https://ethereum.github.io/browser-solidity/#version=soljson-latest.js

func main() {
	ver, err := EthereumAPI.EthProtocolVersion()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Ver - %v\n", ver)

	block, err := EthereumAPI.EthGetBlockByNumber("0x1", true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Block - %v\n", block)

	peers, err := EthereumAPI.NetPeerCount()
	if err != nil {
		panic(err)
	}
	fmt.Printf("peers - %v\n", peers)
}

func AnchorBlock(height int64, keyMR string, hash string) (string, error) {
	data := "0x"
	data += EthereumAPI.StringToMethodID("setAnchor(uint256,uint256,uint256)")
	data += EthereumAPI.IntToData(height)
	data += keyMR
	data += hash

	tx := new(EthereumAPI.TransactionObject)
	tx.From = "0x838F9b4d8EA3ff2F1bD87B13684f59c4C57A618b"
	tx.To = "0x8A8FbaBBec1E99148083E9314dFfd82395dd8F18"
	tx.Gas = "0x10FFFF"
	tx.Data = data

	fmt.Printf("tx - %v\n", tx)

	txHash, err := EthereumAPI.EthSendTransaction(tx)

	fmt.Printf("txHash - %v\n", txHash)
	if err != nil {
		fmt.Printf("err - %v", err)
		return "", err
	}

	return txHash, nil
}

func GetAnchorData(height int64) (string, string, error) {
	data := "0x"
	data += EthereumAPI.StringToMethodID("anchors(uint256)")
	data += EthereumAPI.IntToData(height)

	tx := new(EthereumAPI.TransactionObject)
	tx.From = "0x838F9b4d8EA3ff2F1bD87B13684f59c4C57A618b"
	tx.To = "0x8A8FbaBBec1E99148083E9314dFfd82395dd8F18"
	tx.Gas = "0x10FFFF"
	tx.Data = data

	txData, err := EthereumAPI.EthCall(tx, "latest")
	if err != nil {
		fmt.Printf("err - %v", err)
		return "", "", err
	}
	fmt.Printf("txData - %v\n", txData)
	return txData[2:66], txData[66:], nil
}

func GetTransaction(txHash string) {
	data, err := EthereumAPI.EthGetTransactionByHash(txHash)
	if err != nil {
		fmt.Printf("err - %v", err)
		return // "", "", err
	}
	fmt.Printf("data - %v\n", data)
}

//https://testnet.etherscan.io/apis
//https://testnet.etherscan.io/api?module=account&action=txlist&address=0x8A8FbaBBec1E99148083E9314dFfd82395dd8F18&startblock=0&endblock=99999999&sort=asc&apikey=YourApiKeyToken
func GetContractTransactions() {
	data, err := EthereumAPI.EtherscanTxList("0x8A8FbaBBec1E99148083E9314dFfd82395dd8F18")
	if err != nil {
		fmt.Printf("err - %v", err)
		return // "", "", err
	}
	fmt.Printf("data - %v\n", data)
}
