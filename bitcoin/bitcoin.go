package bitcoin

import (
	"fmt"
	"time"

	"github.com/FactomProject/anchormaker/config"
	"github.com/FactomProject/anchormaker/database"

	"github.com/FactomProject/factomd/common/primitives"
)

var IgnoreWrongEntries bool = true

func LoadConfig(c *config.AnchorConfig) {
	err := InitRPCClient(c)
	if err != nil {
		panic(err)
	}
}

func SynchronizeBitcoinData(dbo *database.AnchorDatabaseOverlay) (int, error) {
	/*

		// ListSinceBlockMinConfAsync returns an instance of a type that can be used to
		// get the result of the RPC at some future time by invoking the Receive
		// function on the returned instance.
		//
		// See ListSinceBlockMinConf for the blocking version and more details.
		func (c *Client) ListSinceBlockMinConfAsync(blockHash *wire.ShaHash, minConfirms int) FutureListSinceBlockResult {
	*/

	txCount := 0
	fmt.Println("SynchronizeBitcoinData")
	for {
		ps, err := dbo.FetchProgramState()
		if err != nil {
			return 0, err
		}
		fmt.Printf("LastBitcoinBlockChecked - %v\n", ps.LastBitcoinBlockChecked)

		txs, newBlock, err := ListBitcoinTransactionsSinceBlock(ps.LastBitcoinBlockChecked)
		if err != nil {
			return 0, err
		}

		if len(txs) == 0 {
			break
		}

		fmt.Printf("Bitcoin Tx count - %v\n", len(txs))
		for _, tx := range txs {
			txCount++
			if tx.IsOurs(BTCAddress) == false {
				fmt.Printf("Not from our address - %v\n", tx.String())
				//ignoring transactions that are not ours
				continue
			}

			dbHeight, keyMR := tx.GetAnchorData()
			fmt.Printf("height, key - %v, %v\n", dbHeight, keyMR)
			if keyMR == "" {
				//ignoring transactions that don't have data
				continue
			}

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
			if ad.BitcoinRecordHeight > 0 {
				continue
			}
			if ad.Bitcoin.TXID != "" {
				continue
			}

			ad.Bitcoin.Address = BTCAddress
			ad.Bitcoin.TXID = tx.GetHash()
			ad.Bitcoin.BlockHeight = tx.GetBlockNumber()
			ad.Bitcoin.BlockHash = tx.GetBlockHash()
			ad.Bitcoin.Offset = tx.GetTransactionIndex()

			fmt.Printf("Saving anchored - %v, %v\n", ad.DBlockHeight, ad.DBlockKeyMR)

			err = dbo.InsertAnchorData(ad, false)
			if err != nil {
				return 0, err
			}
		}
		fmt.Printf("LastBlock - %v\n", newBlock)

		ps.LastBitcoinBlockChecked = newBlock

		err = dbo.InsertProgramState(ps)
		if err != nil {
			return 0, err
		}
	}

	return txCount, nil
}

func AnchorBlocksIntoBitcoin(dbo *database.AnchorDatabaseOverlay) error {
	fmt.Println("AnchorBlocksIntoBitcoin")
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

	for {
		ad, err = dbo.FetchAnchorData(height)
		if err != nil {
			return err
		}
		if ad == nil {
			return nil
		}
		if ad.Bitcoin.TXID != "" {
			height++
			continue
		}

		fmt.Printf("Anchoring %v\n", height)
		time.Sleep(5 * time.Second)
		h, err := primitives.NewShaHashFromStr(ad.DBlockKeyMR)
		if err != nil {
			return err
		}

		tx, err := SendRawTransactionToBTC(h.String(), ad.DBlockHeight)
		if err != nil {
			return err
		}
		fmt.Printf("Anchored %v\n", height)

		ad.Bitcoin.TXID = tx
		err = dbo.InsertAnchorData(ad, false)
		if err != nil {
			return err
		}
		height++
	}

	return nil
}
