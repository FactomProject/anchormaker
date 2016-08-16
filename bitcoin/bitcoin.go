package bitcoin

import (
	"fmt"
	"time"

	"github.com/FactomProject/anchormaker/config"
	"github.com/FactomProject/anchormaker/database"

	"github.com/FactomProject/factomd/common/primitives"
)

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
	return 0, nil
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
			height = ad.DBlockHeight + 1
			continue
		}

		//fmt.Printf("Anchoring %v\n", height)
		time.Sleep(5 * time.Second)
		h, err := primitives.NewShaHashFromStr(ad.DBlockKeyMR)
		if err != nil {
			return err
		}

		tx, err := SendRawTransactionToBTC(h, ad.DBlockHeight)
		if err != nil {
			return err
		}
		fmt.Printf("Anchored %v\n", height)

		ad.Bitcoin.TXID = tx
		err = dbo.InsertAnchorData(ad, false)
		if err != nil {
			return err
		}
		height = ad.DBlockHeight + 1
	}

	return nil
}
