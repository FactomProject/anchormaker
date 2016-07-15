package ethereum

import (
	"fmt"
	"strconv"

	"github.com/FactomProject/EthereumAPI"
	"github.com/FactomProject/anchormaker/database"
)

//https://ethereum.github.io/browser-solidity/#version=soljson-latest.js

var WalletAddress string = "0x838F9b4d8EA3ff2F1bD87B13684f59c4C57A618b"
var ContractAddress string = "0x8A8FbaBBec1E99148083E9314dFfd82395dd8F18"

func SynchronizeEthereumData(dbo *database.AnchorDatabaseOverlay) error {
	ps, err := dbo.FetchProgramState()
	if err != nil {
		return err
	}

	txs, err := EthereumAPI.EtherscanTxListWithStartBlock(ContractAddress, ps.LastEthereumBlockChecked)
	if err != nil {
		return err
	}

	var lastBlock int64 = 0

	for _, tx := range txs {
		if Atoi(tx.BlockNumber) > lastBlock {
			lastBlock = Atoi(tx.BlockNumber)
		}
		if tx.From != WalletAddress {
			//ignoring transactions that are not ours
			continue
		}
		//makign sure the input is of correct length
		if len(tx.Input) == 202 {
			//making sure the right function is called
			if tx.Input[:10] == "0xd36b1da5" {
				dbHeight, keyMR, _ := ParseInput(tx.Input)

				ad, err := dbo.FetchAnchorData(dbHeight)
				if err != nil {
					return err
				}
				if ad == nil {
					return fmt.Errorf("We have anchored block from outside of our DB")
				}
				if ad.DBlockKeyMR != keyMR {
					//return fmt.Errorf("We have anchored invalid KeyMR")
					continue
				}
				if ad.EthereumRecordHeight > 0 {
					continue
				}
				if ad.Ethereum.TxID != "" {
					continue
				}

				ad.Ethereum.Address = tx.From
				ad.Ethereum.TxID = tx.Hash
				ad.Ethereum.BlockHeight = Atoi(tx.BlockNumber)
				ad.Ethereum.BlockHash = tx.BlockHash
				ad.Ethereum.Offset = Atoi(tx.TransactionIndex)

				err = dbo.InsertAnchorData(ad, false)
				if err != nil {
					return err
				}
			}
		}
	}

	ps.LastEthereumBlockChecked = lastBlock
	err = dbo.InsertProgramState(ps)
	if err != nil {
		return err
	}

	return nil
}

func Atoi(s string) int64 {
	i, _ := strconv.ParseInt(s, 0, 64)
	return i
}

func ParseInput(input string) (dBlockHeight uint32, keyMR string, hash string) {
	if len(tx.Input) == 202 {
		if tx[:10] == "0xd36b1da5" {
			tx = tx[10:]
			dBlockHeight, tx = uint32(Atoi(tx[:64])), tx[64:]
			keyMR, tx = tx[:64], tx[64:]
			hash = tx[:64]
			return
		}
	}
	return 0, "", ""
}

func AnchorBlocksIntoEthereum(dbo *database.AnchorDatabaseOverlay) error {
	ad, err := dbo.FetchAnchorDataHead()
	if err != nil {
		return err
	}

	if ad == nil {
		return nil
	}

	for {
		height := ad.DBlockHeight + 1
		ad, err = dbo.FetchAnchorData(height)
		if err != nil {
			return err
		}
		if ad == nil {
			return nil
		}
		if ad.Ethereum.TxID != "" {
			continue
		}

		fmt.Printf("Anchoring %v\n", height)
		tx, err := AnchorBlock(ad.DBlockHeight, ad.DBlockKeyMR, ad.DBlockKeyMR)
		if err != nil {
			return err
		}
		fmt.Printf("Anchored %v\n", height)
	}

	return nil
}

func AnchorBlock(height int64, keyMR string, hash string) (string, error) {
	data := "0x"
	data += EthereumAPI.StringToMethodID("setAnchor(uint256,uint256,uint256)")
	data += EthereumAPI.IntToData(height)
	data += keyMR
	data += hash

	tx := new(EthereumAPI.TransactionObject)
	tx.From = WalletAddress
	tx.To = ContractAddress
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

/*
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


func GetAnchorData(height int64) (string, string, error) {
	data := "0x"
	data += EthereumAPI.StringToMethodID("anchors(uint256)")
	data += EthereumAPI.IntToData(height)

	tx := new(EthereumAPI.TransactionObject)
	tx.From = WalletAddress
	tx.To = ContractAddress
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

func GetContractTransactions() {
	data, err := EthereumAPI.EtherscanTxList(ContractAddress)
	if err != nil {
		fmt.Printf("err - %v", err)
		return // "", "", err
	}
	fmt.Printf("data - %v\n", data)
}
*/
