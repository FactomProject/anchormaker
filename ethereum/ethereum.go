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

var WalletAddress string = "0x84964e1FfC60d0ad4DA803678b167c6A783A2E01"
var WalletPassword string = "password"
var ContractAddress string = "0x9e0C6b5f502BD293D7661bE1b2bE0147dcaF0010"
var GasLimit string = "200000"
var GasPrice string = "10000000000" //10 gwei
var IgnoreWrongEntries bool = false
var just_connected_to_net = true

//"0xbbcc0c80"
var FunctionPrefix string = "0x" + EthereumAPI.StringToMethodID("setAnchor(uint256,uint256)") //TODO: update prefix on final smart contract deployment

func LoadConfig(c *config.AnchorConfig) {
	WalletAddress = strings.ToLower(c.Ethereum.WalletAddress)
	WalletPassword = c.Ethereum.WalletPassword
	ContractAddress = strings.ToLower(c.Ethereum.ContractAddress)
	GasLimit = c.Ethereum.GasLimit
	GasPrice = c.Ethereum.GasPrice
	IgnoreWrongEntries = c.Ethereum.IgnoreWrongEntries

	EthereumAPI.EtherscanTestNet = c.Ethereum.TestNet
	EthereumAPI.EtherscanTestNetName = c.Ethereum.TestNetName
	EthereumAPI.EtherscanAPIKeyToken = c.Ethereum.EtherscanAPIKey

	//TODO: load ServerAddress into EthereumAPI
}

func SynchronizeEthereumData(dbo *database.AnchorDatabaseOverlay) (int, error) {
	txCount := 0
	fmt.Println("SynchronizeEthereumData")

	synced, err := CheckIfEthSynced()

	if err != nil {
		return 0, err
	} else if synced == false {
		return 0, fmt.Errorf("eth node not synced, waiting.")
	}

	for {
		ps, err := dbo.FetchProgramState()
		if err != nil {
			return 0, err
		}
		//note, this mutex could probably be reworked to prevent a short time span of a race here between fetch and lock
		ps.ProgramStateMutex.Lock()
		defer ps.ProgramStateMutex.Unlock()

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

	ps, err := dbo.FetchProgramState()
	if err != nil {
		return err
	}
	//note, this mutex could probably be reworked to prevent a short time span of a race here between fetch and lock
	ps.ProgramStateMutex.Lock()
	defer ps.ProgramStateMutex.Unlock()

	err = dbo.InsertProgramState(ps)
	if err != nil {
		return err
	}

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
		done, skip, err := AnchorBlockByHeight(dbo, height)
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

func ReadKeymrAtHeight(height int64) (string, error) {
	data := "0x"
	data += EthereumAPI.StringToMethodID("getAnchor(uint256)")
	data += EthereumAPI.IntToData(height)

	callinfo := new(EthereumAPI.TransactionObject)
	callinfo.To = ContractAddress
	callinfo.Data = data

	keymr, err := EthereumAPI.EthCall(callinfo, "latest")
	if err != nil {
		fmt.Printf("err - %v", err)
		return "", err
	}
	return keymr, nil
}

//returns done when we're done anchoring
//returns skip if we can skip anchoring this block
func AnchorBlockByHeight(dbo *database.AnchorDatabaseOverlay, height uint32) (done bool, skip bool, err error) {
	ad, err := dbo.FetchAnchorData(height)
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
	gasInt, err := strconv.ParseInt(GasLimit, 10, 0)
	if err != nil {
		fmt.Printf("error parsing GasLimit in config file - %v", err)
		return "", err
	}
	gasPriceInt, err := strconv.ParseInt(GasPrice, 10, 0)
	if err != nil {
		fmt.Printf("error parsing GasPrice in config file - %v", err)
		return "", err
	}

	data := "0x"
	data += EthereumAPI.StringToMethodID("setAnchor(uint256,uint256)")
	data += EthereumAPI.IntToData(height)
	data += keyMR

	tx := new(EthereumAPI.TransactionObject)
	tx.From = WalletAddress
	tx.To = ContractAddress

	tx.Gas = EthereumAPI.IntToQuantity(gasInt)
	tx.GasPrice = EthereumAPI.IntToQuantity(gasPriceInt)
	tx.Data = data

	fmt.Printf("tx - %v\n", tx)

	txHash, err := EthereumAPI.PersonalSendTransaction(tx, WalletPassword)

	if err != nil {
		fmt.Printf("err - %v", err)
		return "", err
	}
	fmt.Printf("txHash - %v\n", txHash)

	return txHash, nil
}

func CheckIfEthSynced() (bool, error) {
	//check if the eth node is connected
	peercount, err := EthereumAPI.NetPeerCount()
	if err != nil {
		fmt.Println("Is geth run with --rpcapi \"*,net,*\"")
		return false, err
	}
	if int(*peercount) == 0 { //if our local node is not connected to any nodes, don't make any anchors in ethereum
		just_connected_to_net = true
		return false, fmt.Errorf("geth node is not connected to any peers, waiting 10 sec.")
	}

	if just_connected_to_net == true {
		fmt.Println("Geth has just connected to the first peer. Waiting 30s to discover new blocks")
		time.Sleep(30 * time.Second)
		just_connected_to_net = false
	}

	syncresponse, err := EthereumAPI.EthSyncing()
	if err != nil {
		fmt.Println("Is geth run with --rpcapi \"*,eth,*\"")
		return false, err
	}
	if syncresponse.HighestBlock != "" {

		highestblk, err := strconv.ParseInt(syncresponse.HighestBlock, 0, 64)
		if err != nil {
			return false, fmt.Errorf("Error parsing geth rpc. Expecting a hex number for highestblock, got %v", syncresponse.HighestBlock)
		}

		currentblk, err := strconv.ParseInt(syncresponse.CurrentBlock, 0, 64)
		if err != nil {
			return false, fmt.Errorf("Error parsing geth rpc. Expecting a hex number for currentblock, got %v", syncresponse.CurrentBlock)
		}

		if highestblk > currentblk { //if our local node is still catching up, don't make any anchors in ethereum
			return false, fmt.Errorf("geth node is not caught up to the blockchain, waiting 10 sec. local height: %v blockchain: %v Delta: %v", currentblk, highestblk, (highestblk - currentblk))
		}
	}

	//we might have gotten here with the eth node having connections, but still having a stale blockchain.
	//we check the timestamp of the latest block to see if it is too far behind

	currentTime := time.Now().Unix()
	highestBlockTimeStr, err := EthereumAPI.EthGetBlockByNumber("latest", true)
	if err != nil {
		return false, fmt.Errorf("Error parsing geth rpc. Expecting a block info, got %v. %v", highestBlockTimeStr, err)
	}
	highestBlockTime, err := strconv.ParseInt(highestBlockTimeStr.Timestamp, 0, 64)
	if err != nil {
		return false, fmt.Errorf("Error parsing geth rpc. Expecting a block time, got %v. %v", highestBlockTimeStr.Timestamp, err)
	}
	maxAge := 2 * 60 * 60                                    //give a 2 hour tolerance for a block to be 2 hours behind, due to miner vagaries. 60 * 60 = 60 sec * 60 min
	if int(currentTime) > (maxAge + int(highestBlockTime)) { //if our local node is still catching up, don't make any anchors in ethereum
		return false, fmt.Errorf("Blockchain tip is more than 2 hours old. timenow %v, blocktime %v, delta: %v ", currentTime, highestBlockTime, (currentTime - highestBlockTime))
	}

	return true, nil
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
