package ethereum

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/FactomProject/EthereumAPI"
	"github.com/FactomProject/anchormaker/config"
	"github.com/FactomProject/anchormaker/database"
)

//https://ethereum.github.io/browser-solidity/#version=soljson-latest.js

var WalletAddress string = "0x4da6BAe6689f60e30B575Ca7D3B075605135ee86"
var WalletPassword string = "pass"
var ContractAddress string = "0x7e79c06E18Af0464382c2cd089A20dc49F2EBf86"
var GasPrice string = "0x10FFFF"
var IgnoreWrongEntries bool = false

//"0xbbcc0c80"
var FunctionPrefix string = "0x" + EthereumAPI.StringToMethodID("setAnchor(uint256,uint256)") //TODO: update prefix on final smart contract deployment

func LoadConfig(c *config.AnchorConfig) {
	WalletAddress = strings.ToLower(c.Ethereum.WalletAddress)
	WalletPassword = c.Ethereum.WalletPassword
	ContractAddress = strings.ToLower(c.Ethereum.ContractAddress)
	GasPrice = c.Ethereum.GasPrice
	IgnoreWrongEntries = c.Ethereum.IgnoreWrongEntries

	EthereumAPI.EtherscanTestNet = c.Ethereum.TestNet

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
			if strings.ToLower(tx.From) != WalletAddress {
				fmt.Printf("Not from our address - %v vs %v\n", tx.From, WalletAddress)
				//ignoring transactions that are not ours
				continue
			}
			//makign sure the input is of correct length
			if len(tx.Input) == 138 {
				//making sure the right function is called
				if tx.Input[:10] == FunctionPrefix {
					dbHeight, keyMR := ParseInput(tx.Input)

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

					ad.Ethereum.Address = strings.ToLower(tx.From)
					ad.Ethereum.TXID = strings.ToLower(tx.Hash)
					ad.Ethereum.BlockHeight = Atoi(tx.BlockNumber)
					ad.Ethereum.BlockHash = strings.ToLower(tx.BlockHash)
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

func ParseInput(input string) (dBlockHeight uint32, keyMR string) {
	if len(input) == 138 {
		if input[:10] == FunctionPrefix {
			input = input[10:]
			dBlockHeight, input = uint32(AtoiHex(input[:64])), input[64:]
			keyMR = input[:64]
			return
		}
	}
	return 0, ""
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

	ps, err := dbo.FetchProgramState()
	if err != nil {
		return err
	}
	//We first anchor the newest block before proceeding to anchor older blocks
	_, _, err = AnchorBlockByHeight(dbo, ps.LastFactomDBlockHeightChecked)
	if err != nil {
		return err
	}

	for i := 0; i < 10; {
		done, skip, err := AnchorBlockByHeight(height)
		if err != nil {
			return err
		}
		height++
		if done == true {
			return nil
		}
		if skip == true {
			continue
		}
		i++
	}

	return nil
}

//returns done when we're done anchoring
//returns skip if we can skip anchoring this block
func AnchorBlockByHeight(dbo *database.AnchorDatabaseOverlay, height uint32) (done bool, skip bool, err error) {
	ad, err = dbo.FetchAnchorData(height)
	if err != nil {
		done = true
		skip = false
		return
	}
	if ad == nil {
		done = true
		skip = false
		return
	}
	if ad.Ethereum.TXID != "" {
		done = false
		skip = true
		return
	}

	//fmt.Printf("Anchoring %v\n", height)
	time.Sleep(5 * time.Second)
	tx, err := AnchorBlock(int64(ad.DBlockHeight), ad.DBlockKeyMR)
	if err != nil {
		done = false
		skip = true
		return
	}
	fmt.Printf("Anchored %v\n", height)

	ad.Ethereum.TXID = tx
	err = dbo.InsertAnchorData(ad, false)
	if err != nil {
		done = false
		skip = true
		return
	}

	done = false
	skip = false
	return
}

func AnchorBlock(height int64, keyMR string) (string, error) {
	data := "0x"
	data += EthereumAPI.StringToMethodID("setAnchor(uint256,uint256)")
	data += EthereumAPI.IntToData(height)
	data += keyMR

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
