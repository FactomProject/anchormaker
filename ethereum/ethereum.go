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
	"github.com/FactomProject/anchormaker/api"
)

//https://ethereum.github.io/browser-solidity/#version=soljson-latest.js

var WalletAddress string
var WalletPassword string
var ContractAddress string
var GasLimit string
var EthGasStationAddress string
var IgnoreWrongEntries bool
var justConnectedToNet = true

var conn *ethclient.Client
var factomAnchor *FactomAnchor

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
		merkleRoot := fmt.Sprintf("%064x", event.MerkleRoot)

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
		if ad.MerkleRoot == "" {
			// Calculate Merkle root that should be in the anchor record
			hi := ad.DBlockHeight
			var lo uint32
			if hi < 999 {
				lo = 0
			} else {
				lo = hi - 999
			}
			merkleRoot, err := api.GetMerkleRootOfDBlockWindow(lo, hi)
			if err != nil {
				return 0, err
			}
			ad.MerkleRoot = merkleRoot.String()
		}
		if ad.MerkleRoot != merkleRoot {
			fmt.Printf("Merkle Root for DBlock %d from database != one found in Ethereum contract - %v vs %v\n", ad.DBlockHeight, ad.MerkleRoot, merkleRoot)
			continue
		}
		if ad.EthereumRecordHeight > 0 {
			continue
		}

		// We have a tx listed in the database already, but now we know it has been mined.
		// Update the AnchorData to reflect this.
		ad.Ethereum.Address = strings.ToLower(event.Raw.Address.String())
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

	ps, err = dbo.FetchProgramState()
	if err != nil {
		return err
	}

	// First, anchor a merkle root of the newest 1000 blocks (change windowSize for testing purposes)
	windowSize := uint32(1000)
	dblockhi := ps.LastFactomDBlockHeightChecked
	var dblocklo uint32
	if dblockhi < windowSize - 1 {
		dblocklo = 0
	} else {
		dblocklo = dblockhi - windowSize + 1
	}
	_, _, err = AnchorBlockWindow(dbo, dblocklo, dblockhi)
	if err != nil {
		return err
	}

	// Get the AnchorData head (the highest complete AnchorData record, no anchoring holes before this point)
	ad, err := dbo.FetchAnchorDataHead()
	if err != nil {
		return err
	}
	if ad == nil {
		// We haven't anchored any of the backlog, start with the first 1000 block window
		dblocklo = 0
		dblockhi = windowSize - 1
	} else {
		// We've anchored some of the backlog, so move to the next 1000 block window
		dblocklo = ad.DBlockHeight + 1
		dblockhi = dblocklo + windowSize - 1
	}

	// Try to submit up to 10 anchors in the backlog
	for i := 0; i < 10; {
		// Find the next highest AnchorData we have submitted to Ethereum
		// and check if we need to submit another anchor to close the gap more
		adNext, err := dbo.FetchNextHighestAnchorDataSubmitted(dblocklo)
		if err != nil {
			return err
		}
		if adNext == nil {
			// There was no AnchorData submitted for a block height higher than dblocklo
			return nil
		}
		if dblocklo >= adNext.DBlockHeight - windowSize + 1 {
			// There is no gap between the AnchorData head and the next AnchorData submitted
			// We must be all caught up on this section
			return nil
		}

		done, skip, err := AnchorBlockWindow(dbo, dblocklo, dblockhi)
		if err != nil {
			return err
		}
		dblocklo += windowSize
		dblockhi += windowSize

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

// AnchorBlockWindow creates a Merkle root of all Directory Blocks from height lo to hi, and then submits that MR to Ethereum.
// returns done when we're done anchoring
// returns skip if we can skip anchoring this block
func AnchorBlockWindow(dbo *database.AnchorDatabaseOverlay, lo, hi uint32) (done bool, skip bool, err error) {
	ad, err := dbo.FetchAnchorData(hi)
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

	merkleRoot, err := api.GetMerkleRootOfDBlockWindow(lo, hi)
	if err != nil {
		return true, false, err
	}

	tx, err := SendAnchor(int64(hi), merkleRoot.String())
	if err != nil {
		done = false
		skip = true
		return
	}
	fmt.Printf("Submitted Ethereum Anchor for DBlocks %v to %v\n", lo, hi)

	ad.MerkleRoot = merkleRoot.String()
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

func SendAnchor(height int64, merkleRoot string) (string, error) {
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
	data += merkleRoot

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

/*
// GetKeymrAtHeight returns the merkle root at a given DBlock height as a hex string
func GetKeymrAtHeight(height int64) (string, error) {
	opts := bind.CallOpts{}
	opts.Pending = false

	keymr, err := factomAnchor.GetAnchor(&opts, big.NewInt(height))
	if err != nil {
		fmt.Printf("error getting keymr: %v", err)
		return "", err
	}
	return fmt.Sprintf("%064x", keymr), nil
}
*/

func CheckIfEthSynced() (bool, error) {
	// Check if the eth node is connected
	peerCount, err := EthereumAPI.NetPeerCount()
	if err != nil {
		fmt.Println("Is geth run with --rpcapi \"*,net,*\"")
		return false, err
	}
	if int(*peerCount) == 0 { //if our local node is not connected to any nodes, don't make any anchors in ethereum
		justConnectedToNet = true
		return false, fmt.Errorf("geth node is not connected to any peers, waiting 10 sec.")
	}

	if justConnectedToNet == true {
		fmt.Println("Geth has just connected to the first peer. Waiting 30s to discover new blocks")
		time.Sleep(30 * time.Second)
		justConnectedToNet = false
	}

	syncResponse, err := EthereumAPI.EthSyncing()
	if err != nil {
		fmt.Println("Is geth run with --rpcapi \"*,eth,*\"")
		return false, err
	}
	if syncResponse.HighestBlock != "" {
		highestBlk, err := strconv.ParseInt(syncResponse.HighestBlock, 0, 64)
		if err != nil {
			return false, fmt.Errorf("Error parsing geth rpc. Expecting a hex number for highestblock, got %v", syncResponse.HighestBlock)
		}

		currentBlk, err := strconv.ParseInt(syncResponse.CurrentBlock, 0, 64)
		if err != nil {
			return false, fmt.Errorf("Error parsing geth rpc. Expecting a hex number for currentblock, got %v", syncResponse.CurrentBlock)
		}

		// If our local node is still catching up, don't submit any new anchors to Ethereum
		if highestBlk > currentBlk {
			return false, fmt.Errorf("geth node is not caught up to the blockchain, waiting 10 sec. local height: %v blockchain: %v Delta: %v", currentBlk, highestBlk, (highestBlk - currentBlk))
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
