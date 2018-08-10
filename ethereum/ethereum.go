//go:generate abigen --sol anchorContract.sol --pkg ethereum --out factomAnchor.go
package ethereum

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/FactomProject/EthereumAPI"
	"github.com/FactomProject/anchormaker/config"
	"github.com/FactomProject/anchormaker/database"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"math/big"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

//https://ethereum.github.io/browser-solidity/#version=soljson-latest.js

var WalletAddress string = "0x84964e1FfC60d0ad4DA803678b167c6A783A2E01"
var WalletPassword string = "password"
var ContractAddress string = "0x9e0C6b5f502BD293D7661bE1b2bE0147dcaF0010"
var GasLimit string = "200000"
var EthGasStationAddress string
var IgnoreWrongEntries bool = false
var just_connected_to_net = true

var conn *ethclient.Client
var factomAnchor *FactomAnchor

//"0xbbcc0c80"
var FunctionPrefix string = "0x" + EthereumAPI.StringToMethodID("setAnchor(uint256,uint256)") //TODO: update prefix on final smart contract deployment

func LoadConfig(c *config.AnchorConfig) {
	WalletAddress = strings.ToLower(c.Ethereum.WalletAddress)
	WalletPassword = c.Ethereum.WalletPassword
	ContractAddress = strings.ToLower(c.Ethereum.ContractAddress)
	GasLimit = c.Ethereum.GasLimit
	EthGasStationAddress = c.Ethereum.EthGasStationAddress
	IgnoreWrongEntries = c.Ethereum.IgnoreWrongEntries

	var err error
	// Create IPC based RPC connection to the local node
	conn, err = ethclient.Dial(c.Ethereum.GethIPCURL)
	if err != nil {
		fmt.Printf("Failed to connect to ethereum node over IPC: %v\n", err)
		panic(err)
	}

	// Get an instance of the deployed smart contract
	factomAnchor, err = NewFactomAnchor(common.HexToAddress(ContractAddress), conn)
	if err != nil {
		fmt.Printf("Failed to initialize FactomAnchor contract: %v\n", err)
		panic(err)
	}
}

func SynchronizeEthereumData(dbo *database.AnchorDatabaseOverlay) (int, error) {
	fmt.Println("SynchronizeEthereumData")
	txCount := 0

	synced, err := CheckIfEthSynced()
	if err != nil {
		return 0, err
	} else if synced == false {
		return 0, fmt.Errorf("eth node not synced, waiting")
	}

	ps, err := dbo.FetchProgramState()
	if err != nil {
		return 0, err
	}
	// This mutex could probably be reworked to prevent a short time span of a race here between fetch and lock
	ps.ProgramStateMutex.Lock()
	defer ps.ProgramStateMutex.Unlock()

	var lastBlock int64 = 0

	// Use an event filter to quickly get all anchors set since ps.LastEthereumBlockChecked
	filterOpts := &bind.FilterOpts{}
	filterOpts.Start = uint64(ps.LastEthereumBlockChecked)
	anchorEvents, err := factomAnchor.FilterAnchorMade(filterOpts)
	if err != nil {
		return 0, fmt.Errorf("failed to make anchor filter: %v", err)
	}
	for hasNext := anchorEvents.Next(); hasNext; hasNext = anchorEvents.Next() {
		txCount++
		event := anchorEvents.Event
		if int64(event.Raw.BlockNumber) > lastBlock {
			lastBlock = int64(event.Raw.BlockNumber)
		}

		dbHeight := event.Height
		merkleRoot := fmt.Sprintf("%064x", event.Merkleroot)

		// Check if we have a tx for this dbHeight in the database already
		ad, err := dbo.FetchAnchorData(uint32(dbHeight.Uint64()))
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
		if ad.DBlockKeyMR != merkleRoot {
			fmt.Printf("ad.DBlockKeyMR != keyMR - %v vs %v\n", ad.DBlockKeyMR, merkleRoot)
			continue
		}
		if ad.EthereumRecordHeight > 0 {
			continue
		}

		// We have a tx listed in the database already, but now we know it has been mined.
		// Update the AnchorData to reflect this.
		ad.Ethereum.Address = strings.ToLower(WalletAddress)
		ad.Ethereum.TXID = strings.ToLower(event.Raw.TxHash.String())
		ad.Ethereum.BlockHeight = int64(event.Raw.BlockNumber)
		ad.Ethereum.BlockHash = strings.ToLower(event.Raw.BlockHash.String())
		ad.Ethereum.Offset = int64(event.Raw.TxIndex)

		err = dbo.InsertAnchorData(ad, false)
		if err != nil {
			return 0, err
		}
		fmt.Printf("DBlock %v is already anchored in Ethereum!\n", dbHeight)
	}

	// Update the block to start at for the next synchronization loop
	ps.LastEthereumBlockChecked = lastBlock + 1

	err = dbo.InsertProgramState(ps)
	if err != nil {
		return 0, err
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
	ps, err := dbo.FetchProgramState()
	if err != nil {
		return err
	}
	// This mutex could probably be reworked to prevent a short time span of a race here between fetch and lock
	ps.ProgramStateMutex.Lock()
	defer ps.ProgramStateMutex.Unlock()

	err = dbo.InsertProgramState(ps)
	if err != nil {
		return err
	}

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

	ps, err = dbo.FetchProgramState()
	if err != nil {
		return err
	}

	// We first anchor the newest block before proceeding to anchor older blocks
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

// returns done when we're done anchoring
// returns skip if we can skip anchoring this block
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

	time.Sleep(5 * time.Second)
	tx, err := AnchorBlock(int64(ad.DBlockHeight), ad.DBlockKeyMR)
	if err != nil {
		done = false
		skip = true
		return
	}
	fmt.Printf("Submitted Ethereum Anchor for DBlock %v\n", height)

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

	var gasPrice int64
	if gasPriceEstimates, err := GetGasPriceEstimates(EthGasStationAddress); err != nil {
		fmt.Printf("Failed to get gas price estimates from %v\n", EthGasStationAddress)
		fmt.Println("Defaulting gas price to 40 GWei")
		gasPrice = 40000000000
	} else {
		gasPrice = gasPriceEstimates.Fast.Int64()
	}

	data := "0x"
	data += EthereumAPI.StringToMethodID("setAnchor(uint256,uint256)")
	data += EthereumAPI.IntToData(height)
	data += keyMR

	tx := new(EthereumAPI.TransactionObject)
	tx.From = WalletAddress
	tx.To = ContractAddress

	tx.Gas = EthereumAPI.IntToQuantity(gasInt)
	tx.GasPrice = EthereumAPI.IntToQuantity(gasPrice)
	tx.Data = data

	fmt.Printf("Ethereum tx: %v\n", tx)

	txHash, err := EthereumAPI.PersonalSendTransaction(tx, WalletPassword)

	if err != nil {
		fmt.Printf("failed to submit tx: %v\n", err)
		return "", err
	}
	fmt.Printf("Tx submitted with txHash: %v\n", txHash)

	return txHash, nil
}

func CheckIfEthSynced() (bool, error) {
	// Check if the eth node is connected
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

		// If our local node is still catching up, don't submit any new anchors to Ethereum
		if highestblk > currentblk {
			return false, fmt.Errorf("geth node is not caught up to the blockchain, waiting 10 sec. local height: %v blockchain: %v Delta: %v", currentblk, highestblk, (highestblk - currentblk))
		}
	}

	// We might have gotten here with the eth node having connections, but still having a stale blockchain.
	// So check the timestamp of the latest block to see if it is too far behind
	currentTime := time.Now().Unix()
	highestBlockTimeStr, err := EthereumAPI.EthGetBlockByNumber("latest", true)
	if err != nil {
		return false, fmt.Errorf("Error parsing geth rpc. Expecting a block info, got %v. %v", highestBlockTimeStr, err)
	}
	highestBlockTime, err := strconv.ParseInt(highestBlockTimeStr.Timestamp, 0, 64)
	if err != nil {
		return false, fmt.Errorf("Error parsing geth rpc. Expecting a block time, got %v. %v", highestBlockTimeStr.Timestamp, err)
	}
	// Give a 2 hour tolerance for a block to be 2 hours behind, due to miner vagaries. 2 hr * 60 sec * 60 min
	// If our local node is still catching up, don't submit any new anchors to Ethereum
	maxAge := 2 * 60 * 60
	if int(currentTime) > (maxAge + int(highestBlockTime)) {
		return false, fmt.Errorf("Blockchain tip is more than 2 hours old. timenow %v, blocktime %v, delta: %v ", currentTime, highestBlockTime, (currentTime - highestBlockTime))
	}

	return true, nil
}

func CheckBalance() (int64, error) {
	return EthereumAPI.EthGetBalance(WalletAddress, EthereumAPI.Latest)
}

// GasPriceEstimates holds multiple price estimates (in Wei) and their corresponding wait times (in minutes)
type GasPriceEstimates struct {
	BlockNumber uint64
	BlockTime float64
	Speed float64
	SafeLow *big.Int
	SafeLowWait float64
	Average *big.Int
	AverageWait float64
	Fast *big.Int
	FastWait float64
	Fastest *big.Int
	FastestWait float64
}

// GetGasPriceEstimates polls the ethgasstation API at the given URL and returns its most recent estimates
func GetGasPriceEstimates(url string) (*GasPriceEstimates, error) {
	type rawEstimate struct {
		BlockNumber uint64 `json:"blockNum"`
		BlockTime float64 `json:"block_time"`
		Speed float64 `json:"speed"`
		SafeLow float64 `json:"safeLow"`
		SafeLowWait float64 `json:"safeLowWait"`
		Average float64 `json:"average"`
		AverageWait float64 `json:"avgWait"`
		Fast float64 `json:"fast"`
		FastWait float64 `json:"fastWait"`
		Fastest float64 `json:"fastest"`
		FastestWait float64 `json:"fastestWait"`
	}

	client := http.Client{
		Timeout: time.Second * 2,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	raw := rawEstimate{}
	err = json.Unmarshal(body, &raw)
	if err != nil {
		return nil, err
	}

	// convert weird GWei * 10 units to wei, for usability
	estimates := GasPriceEstimates{
		BlockNumber: raw.BlockNumber,
		BlockTime: raw.BlockTime,
		Speed: raw.Speed,
		SafeLow: big.NewInt(int64(raw.SafeLow * 1e8)),
		SafeLowWait: raw.SafeLowWait,
		Average: big.NewInt(int64(raw.Average * 1e8)),
		AverageWait: raw.AverageWait,
		Fast: big.NewInt(int64(raw.Fast * 1e8)),
		FastWait: raw.FastWait,
		Fastest: big.NewInt(int64(raw.Fastest * 1e8)),
		FastestWait: raw.FastestWait,
	}
	return &estimates, nil
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
