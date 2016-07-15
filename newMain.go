package main

import (
	"fmt"

	"github.com/FactomProject/anchormaker/api"
	"github.com/FactomProject/anchormaker/database"
	"github.com/FactomProject/anchormaker/ethereum"

	"github.com/FactomProject/factomd/anchor"
	"github.com/FactomProject/factomd/common/interfaces"
)

func main() {

	dbo := database.NewMapDB()

	anchorData, err := dbo.FetchAnchorDataHead()
	if err != nil {
		panic(err)
	}
	fmt.Printf("anchorData - %v\n", anchorData)

	SynchronizeFactomData(dbo)

	anchorData, err = dbo.FetchAnchorDataHead()
	if err != nil {
		panic(err)
	}
	fmt.Printf("anchorDataHead - %v\n", anchorData)
	ps, err := dbo.FetchProgramState()
	if err != nil {
		panic(err)
	}
	fmt.Printf("ps - %v\n", ps)

	err = ethereum.SynchronizeEthereumData(dbo)
	if err != nil {
		panic(err)
	}

	err = ethereum.AnchorBlocksIntoEthereum(dbo)
	if err != nil {
		panic(err)
	}

}

func SynchronizeFactomData(dbo *database.AnchorDatabaseOverlay) {
	anchorData, err := dbo.FetchAnchorDataHead()
	if err != nil {
		panic(err)
	}
	startHeight := 0
	endKeyMR := "0000000000000000000000000000000000000000000000000000000000000000"
	if anchorData != nil {
		endKeyMR = anchorData.DBlockKeyMR
		startHeight = int(anchorData.DBlockHeight)
	}
	ps, err := dbo.FetchProgramState()
	if err != nil {
		panic(err)
	}
	if ps.LastFactomDBlockChecked != "" {
		endKeyMR = ps.LastFactomDBlockChecked
	}

	dBlockHead, err := api.GetDBlockHead()
	if err != nil {
		panic(err)
	}

	dBlock, err := api.GetDBlock(dBlockHead)
	if err != nil {
		panic(err)
	}
	fmt.Printf("dBlock - %v\n", dBlock)

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
			panic(err)
		}
		if dBlock == nil {
			panic("dblock " + keymr + " not found")
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
					panic(err)
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
						panic(err)
					}
					//fmt.Printf("Entry - %v\n", entry)
					//TODO: update existing anchor entries
					ar, err := anchor.UnmarshalAnchorRecord(entry.GetContent())
					if err != nil {
						panic(err)
					}
					//fmt.Printf("anchor - %v\n", ar)

					anchorData, err = dbo.FetchAnchorData(ar.DBHeight)
					if err != nil {
						panic(err)
					}
					if anchorData.DBlockKeyMR != ar.KeyMR {
						panic("AnchorData KeyMR does not match AnchorRecord KeyMR")
					}

					anchorData.Bitcoin.Address = ar.Bitcoin.Address
					anchorData.Bitcoin.TXID = ar.Bitcoin.TXID
					anchorData.Bitcoin.BlockHeight = ar.Bitcoin.BlockHeight
					anchorData.Bitcoin.BlockHash = ar.Bitcoin.BlockHash
					anchorData.Bitcoin.Offset = ar.Bitcoin.Offset

					anchorData.BitcoinRecordHeight = dBlock.GetDatabaseHeight()
					anchorData.BitcoinRecordEntryHash = eh.String()
					err = dbo.InsertAnchorData(anchorData, false)
					if err != nil {
						panic(err)
					}
				}
			}

			//TODO: look for Ethereum anchors
		}

		//Updating new directory blocks
		anchorData, err = dbo.FetchAnchorData(uint32(i))
		if err != nil {
			panic(err)
		}
		if anchorData == nil {
			anchorData := new(database.AnchorData)
			anchorData.DBlockHeight = dBlock.GetDatabaseHeight()
			anchorData.DBlockKeyMR = dBlock.GetKeyMR().String()
			err = dbo.InsertAnchorData(anchorData, false)
			if err != nil {
				panic(err)
			}
		}
	}

	err = dbo.UpdateAnchorDataHead()
	if err != nil {
		panic(err)
	}

	ps.LastFactomDBlockChecked = dBlockHead

	err = dbo.InsertProgramState(ps)
	if err != nil {
		panic(err)
	}
}
