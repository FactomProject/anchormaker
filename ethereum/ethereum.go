package ethereum

import (
	"fmt"
	"strconv"
	"time"

	"github.com/FactomProject/EthereumAPI"
	"github.com/FactomProject/anchormaker/config"
	"github.com/FactomProject/anchormaker/database"
)

//https://ethereum.github.io/browser-solidity/#version=soljson-latest.js

var WalletAddress string = "0x838f9b4d8ea3ff2f1bd87b13684f59c4c57a618b"
var WalletPassword string = "pass"
var ContractAddress string = "0x8a8fbabbec1e99148083e9314dffd82395dd8f18"
var GasPrice string = "0x10FFFF"
var IgnoreWrongEntries bool = false

//"0xd36b1da5"
var FunctionPrefix string = "0x" + EthereumAPI.StringToMethodID("setAnchor(uint256,uint256,uint256)") //TODO: update prefix on final smart contract deployment

func LoadConfig(c *config.AnchorConfig) {
	WalletAddress = c.Ethereum.WalletAddress
	WalletPassword = c.Ethereum.WalletPassword
	ContractAddress = c.Ethereum.ContractAddress
	GasPrice = c.Ethereum.GasPrice
	IgnoreWrongEntries = c.Ethereum.IgnoreWrongEntries

	//TODO: load ServerAddress into EthereumAPI
}

func SynchronizeEthereumData(dbo *database.AnchorDatabaseOverlay) (int, error) {
	txCount := 0
	fmt.Println("SynchronizeEthereumData")
	for {
		ps, err := dbo.FetchProgramState()
		if err != nil {
			return 0, err
		}

		txs, err := EthereumAPI.EtherscanTxListWithStartBlock(ContractAddress, ps.LastEthereumBlockChecked)
		if err != nil {
			return 0, err
		}

		if len(txs) == 0 {
			break
		}

		var lastBlock int64 = 0

		fmt.Printf("Ethereum Tx count - %v\n", len(txs))

		for _, tx := range txs {
			txCount++
			if Atoi(tx.BlockNumber) > lastBlock {
				lastBlock = Atoi(tx.BlockNumber)
			}
			if tx.From != WalletAddress {
				fmt.Printf("Not from our address - %v\n", tx.From)
				//ignoring transactions that are not ours
				continue
			}
			//makign sure the input is of correct length
			if len(tx.Input) == 202 {
				//making sure the right function is called
				if tx.Input[:10] == FunctionPrefix {
					dbHeight, keyMR, _ := ParseInput(tx.Input)

					ad, err := dbo.FetchAnchorData(dbHeight)
					if err != nil {
						return 0, err
					}
					if ad == nil {
						if IgnoreWrongEntries == false {
							return 0, fmt.Errorf("We have anchored block from outside of our DB")
						} else {
							continue
						}
					}
					if ad.DBlockKeyMR != keyMR {
						fmt.Printf("ad.DBlockKeyMR != keyMR - %v vs %v\n", ad.DBlockKeyMR, keyMR)
						//return fmt.Errorf("We have anchored invalid KeyMR")
						continue
					}
					if ad.EthereumRecordHeight > 0 {
						continue
					}
					if ad.Ethereum.TXID != "" {
						continue
					}

					ad.Ethereum.Address = tx.From
					ad.Ethereum.TXID = tx.Hash
					ad.Ethereum.BlockHeight = Atoi(tx.BlockNumber)
					ad.Ethereum.BlockHash = tx.BlockHash
					ad.Ethereum.Offset = Atoi(tx.TransactionIndex)

					err = dbo.InsertAnchorData(ad, false)
					if err != nil {
						return 0, err
					}
					fmt.Printf("Block %v is already anchored!\n", dbHeight)
				} else {
					fmt.Printf("Wrong prefix - %v\n", tx.Input[:10])
				}
			} else {
				fmt.Printf("Wrong len - %v\n", len(tx.Input))
			}
		}
		fmt.Printf("LastBlock - %v\n", lastBlock)

		if len(txs) < 1000 {
			//we have checked all transactions from the last block, so we can safely step over it
			ps.LastEthereumBlockChecked = lastBlock + 1
		} else {
			ps.LastEthereumBlockChecked = lastBlock
		}
		err = dbo.InsertProgramState(ps)
		if err != nil {
			return 0, err
		}
	}

	return txCount, nil
}

func Atoi(s string) int64 {
	i, _ := strconv.ParseInt(s, 0, 64)
	return i
}

func AtoiHex(s string) int64 {
	i, _ := strconv.ParseInt(s, 16, 64)
	return i
}

func ParseInput(input string) (dBlockHeight uint32, keyMR string, hash string) {
	if len(input) == 202 {
		if input[:10] == FunctionPrefix {
			input = input[10:]
			dBlockHeight, input = uint32(AtoiHex(input[:64])), input[64:]
			keyMR, input = input[:64], input[64:]
			hash = input[:64]
			return
		}
	}
	return 0, "", ""
}

func AnchorBlocksIntoEthereum(dbo *database.AnchorDatabaseOverlay) error {
	fmt.Println("AnchorBlocksIntoEthereum")
	ad, err := dbo.FetchAnchorDataHead()
	if err != nil {
		return err
	}

	var height uint32
	if ad == nil {
		height = 0
	} else {
		height = ad.DBlockHeight + 1
	}

	for i := 0; i < 10; {
		ad, err = dbo.FetchAnchorData(height)
		if err != nil {
			return err
		}
		if ad == nil {
			return nil
		}
		if ad.Ethereum.TXID != "" {
			height = ad.DBlockHeight + 1
			continue
		}

		//fmt.Printf("Anchoring %v\n", height)
		time.Sleep(5 * time.Second)
		tx, err := AnchorBlock(int64(ad.DBlockHeight), ad.DBlockKeyMR, ad.DBlockKeyMR)
		if err != nil {
			return err
		}
		fmt.Printf("Anchored %v\n", height)

		ad.Ethereum.TXID = tx
		err = dbo.InsertAnchorData(ad, false)
		if err != nil {
			return err
		}
		height = ad.DBlockHeight + 1

		i++
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
	tx.Gas = GasPrice
	tx.Data = data

	fmt.Printf("tx - %v\n", tx)

	txHash, err := EthereumAPI.PersonalSignAndSendTransaction(tx, WalletPassword)

	if err != nil {
		fmt.Printf("err - %v", err)
		return "", err
	}
	fmt.Printf("txHash - %v\n", txHash)

	return txHash, nil
}

func CheckBalance() (int64, error) {
	return EthereumAPI.EthGetBalance(WalletAddress, EthereumAPI.Latest)
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
