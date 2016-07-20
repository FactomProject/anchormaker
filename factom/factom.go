package factom

import (
	"fmt"

	"github.com/FactomProject/anchormaker/api"
	"github.com/FactomProject/anchormaker/config"
	"github.com/FactomProject/anchormaker/database"

	"github.com/FactomProject/factomd/anchor"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var AnchorSigPublicKey *primitives.PublicKey

func LoadConfig(c *config.AnchorConfig) {
	AnchorSigPublicKey = new(primitives.PublicKey)
	err := AnchorSigPublicKey.UnmarshalText([]byte(c.Anchor.AnchorSigPublicKey))
	if err != nil {
		panic(err)
	}
}

//Returns number of blocks synchronized
func SynchronizeFactomData(dbo *database.AnchorDatabaseOverlay) (int, error) {
	blockCount := 0
	anchorData, err := dbo.FetchAnchorDataHead()
	if err != nil {
		return 0, err
	}
	startHeight := 0
	endKeyMR := "0000000000000000000000000000000000000000000000000000000000000000"
	if anchorData != nil {
		endKeyMR = anchorData.DBlockKeyMR
		startHeight = int(anchorData.DBlockHeight)
	}
	ps, err := dbo.FetchProgramState()
	if err != nil {
		return 0, err
	}
	if ps.LastFactomDBlockChecked != "" {
		endKeyMR = ps.LastFactomDBlockChecked
	}

	dBlockHead, err := api.GetDBlockHead()
	if err != nil {
		return 0, err
	}

	//already fully synchronized
	if endKeyMR == dBlockHead {
		return 0, nil
	}

	dBlock, err := api.GetDBlock(dBlockHead)
	if err != nil {
		return 0, err
	}
	fmt.Printf("dBlock - %v\n", dBlock)
	currentHeadHeight := dBlock.GetDatabaseHeight()

	dBlockList := make([]interfaces.IDirectoryBlock, int(dBlock.GetDatabaseHeight())+1)
	dBlockList[int(dBlock.GetDatabaseHeight())] = dBlock

	for {
		keymr := dBlock.GetHeader().GetPrevKeyMR().String()
		if keymr == endKeyMR {
			startHeight = int(dBlock.GetDatabaseHeight())
			break
		}
		dBlock, err = api.GetDBlock(keymr)
		if err != nil {
			return 0, err
		}
		if dBlock == nil {
			return 0, fmt.Errorf("dblock " + keymr + " not found")
		}

		dBlockList[int(dBlock.GetDatabaseHeight())] = dBlock
		fmt.Printf("Fetched dblock %v\n", dBlock.GetDatabaseHeight())
	}

	for i := startHeight; i < len(dBlockList); i++ {
		dBlock = dBlockList[i]
		for _, v := range dBlock.GetDBEntries() {
			//Looking for Bitcoin and Ethereum anchors
			if v.GetChainID().String() == "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604" {
				entryBlock, err := api.GetEBlock(v.GetKeyMR().String())
				if err != nil {
					return 0, err
				}
				for _, eh := range entryBlock.GetEntryHashes() {
					if eh.IsMinuteMarker() == true {
						continue
					}
					if eh.String() == "24674e6bc3094eb773297de955ee095a05830e431da13a37382dcdc89d73c7d7" {
						continue
					}
					fmt.Printf("Fetching %v\n", eh.String())
					entry, err := api.GetEntry(eh.String())
					if err != nil {
						return 0, err
					}
					//fmt.Printf("Entry - %v\n", entry)
					//TODO: update existing anchor entries
					ar, valid, err := anchor.UnmarshalAndvalidateAnchorRecord(entry.GetContent(), AnchorSigPublicKey)
					if err != nil {
						return 0, err
					}
					if valid == false {
						return 0, fmt.Errorf("Invalid anchor - %v\n", entry)
					}
					//fmt.Printf("anchor - %v\n", ar)

					anchorData, err = dbo.FetchAnchorData(ar.DBHeight)
					if err != nil {
						return 0, err
					}
					if anchorData.DBlockKeyMR != ar.KeyMR {
						return 0, fmt.Errorf("AnchorData KeyMR does not match AnchorRecord KeyMR")
					}

					if ar.Bitcoin != nil {
						anchorData.Bitcoin.Address = ar.Bitcoin.Address
						anchorData.Bitcoin.TXID = ar.Bitcoin.TXID
						anchorData.Bitcoin.BlockHeight = ar.Bitcoin.BlockHeight
						anchorData.Bitcoin.BlockHash = ar.Bitcoin.BlockHash
						anchorData.Bitcoin.Offset = ar.Bitcoin.Offset

						anchorData.BitcoinRecordHeight = dBlock.GetDatabaseHeight()
						anchorData.BitcoinRecordEntryHash = eh.String()
					}
					if ar.Ethereum != nil {
						anchorData.Ethereum.Address = ar.Ethereum.Address
						anchorData.Ethereum.TXID = ar.Ethereum.TXID
						anchorData.Ethereum.BlockHeight = ar.Ethereum.BlockHeight
						anchorData.Ethereum.BlockHash = ar.Ethereum.BlockHash
						anchorData.Ethereum.Offset = ar.Ethereum.Offset

						anchorData.EthereumRecordHeight = dBlock.GetDatabaseHeight()
						anchorData.EthereumRecordEntryHash = eh.String()
					}

					err = dbo.InsertAnchorData(anchorData, false)
					if err != nil {
						return 0, err
					}
					blockCount++
				}
			}
		}

		//Updating new directory blocks
		anchorData, err = dbo.FetchAnchorData(uint32(i))
		if err != nil {
			return 0, err
		}
		if anchorData == nil {
			anchorData := new(database.AnchorData)
			anchorData.DBlockHeight = dBlock.GetDatabaseHeight()
			anchorData.DBlockKeyMR = dBlock.GetKeyMR().String()
			err = dbo.InsertAnchorData(anchorData, false)
			if err != nil {
				return 0, err
			}
			blockCount++
		}
	}

	err = dbo.UpdateAnchorDataHead()
	if err != nil {
		return 0, err
	}

	ps.LastFactomDBlockChecked = dBlockHead
	ps.LastFactomDBlockHeightChecked = currentHeadHeight

	err = dbo.InsertProgramState(ps)
	if err != nil {
		return 0, err
	}

	return blockCount, nil
}

func SaveAnchorsIntoFactom(dbo *database.AnchorDatabaseOverlay) error {
	ps, err := dbo.FetchProgramState()
	if err != nil {
		return err
	}
	anchorData, err := dbo.FetchAnchorDataHead()
	if err != nil {
		return err
	}
	if anchorData == nil {
		anchorData, err = dbo.FetchAnchorData(0)
		if err != nil {
			return err
		}
		if anchorData == nil {
			//nothing found
			return nil
		}
	}
	for {
		//Only anchor records that haven't been anchored before
		if (anchorData.BitcoinRecordEntryHash == "" && anchorData.Bitcoin.TXID != "") || (anchorData.EthereumRecordEntryHash == "" && anchorData.Ethereum.TXID != "") {
			anchorRecord := new(anchor.AnchorRecord)
			anchorRecord.AnchorRecordVer = 1
			anchorRecord.DBHeight = anchorData.DBlockHeight
			anchorRecord.KeyMR = anchorData.DBlockKeyMR
			anchorRecord.RecordHeight = ps.LastFactomDBlockHeightChecked + 1

			//Bitcoin anchor
			//Factom Entry Hash has to be empty and Bitcoin TxID must not be empty
			if anchorData.BitcoinRecordEntryHash == "" && anchorData.Bitcoin.TXID != "" {
				anchorRecord.Bitcoin = new(anchor.BitcoinStruct)

				anchorRecord.Bitcoin.Address = anchorData.Bitcoin.Address
				anchorRecord.Bitcoin.TXID = anchorData.Bitcoin.TXID
				anchorRecord.Bitcoin.BlockHeight = anchorData.Bitcoin.BlockHeight
				anchorRecord.Bitcoin.BlockHash = anchorData.Bitcoin.BlockHash
				anchorRecord.Bitcoin.Offset = anchorData.Bitcoin.Offset

				tx, err := CreateAndSendAnchor(anchorRecord)
				if err != nil {
					return err
				}
				anchorData.BitcoinRecordEntryHash = tx

				//Resetting AnchorRecord
				anchorRecord.Bitcoin = nil
			}

			//Ethereum anchor
			//Factom Entry Hash has to be empty and Ethereum TxID must not be empty
			if anchorData.EthereumRecordEntryHash == "" && anchorData.Ethereum.TXID != "" {
				anchorRecord.Ethereum = new(anchor.EthereumStruct)

				anchorRecord.Ethereum.Address = anchorData.Ethereum.Address
				anchorRecord.Ethereum.TXID = anchorData.Ethereum.TXID
				anchorRecord.Ethereum.BlockHeight = anchorData.Ethereum.BlockHeight
				anchorRecord.Ethereum.BlockHash = anchorData.Ethereum.BlockHash
				anchorRecord.Ethereum.Offset = anchorData.Ethereum.Offset

				tx, err := CreateAndSendAnchor(anchorRecord)
				if err != nil {
					return err
				}
				anchorData.BitcoinRecordEntryHash = tx

				anchorData.EthereumRecordEntryHash = tx

				//Resetting AnchorRecord
				anchorRecord.Ethereum = nil
			}

			err = dbo.InsertAnchorData(anchorData, false)
			if err != nil {
				return err
			}
		}
		anchorData, err = dbo.FetchAnchorData(anchorData.DBlockHeight + 1)
		if err != nil {
			return err
		}
		if anchorData == nil {
			break
		}
	}
	return nil
}

//Takes care of sending the entry to the Factom network, returns txID
func CreateAndSendAnchor(ar *anchor.AnchorRecord) (string, error) {
	fmt.Printf("Anchoring %v\n", ar)
	if ar.Bitcoin != nil {

	}
	if ar.Ethereum != nil {

	}
	return "", nil
}
