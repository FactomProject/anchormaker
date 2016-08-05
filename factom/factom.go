package factom

import (
	"fmt"
	"time"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factom/wallet"
	"github.com/FactomProject/factom/wallet/wsapi"

	"github.com/FactomProject/anchormaker/api"
	"github.com/FactomProject/anchormaker/config"
	"github.com/FactomProject/anchormaker/database"

	"github.com/FactomProject/factomd/anchor"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var AnchorSigPublicKey *primitives.PublicKey

var ServerECKey *primitives.PrivateKey
var ServerPrivKey *primitives.PrivateKey
var ECAddress *factom.ECAddress

var FactoidBalanceThreshold int64
var ECBalanceThreshold int64

func LoadConfig(c *config.AnchorConfig) {
	AnchorSigPublicKey = new(primitives.PublicKey)
	err := AnchorSigPublicKey.UnmarshalText([]byte(c.Anchor.AnchorSigPublicKey))
	if err != nil {
		panic(err)
	}

	key, err := primitives.NewPrivateKeyFromHex(c.Anchor.ServerECKey)
	if err != nil {
		panic(err)
	}
	ServerECKey = key

	ecAddress, err := factom.MakeECAddress(key.Key[:32])
	if err != nil {
		panic(err)
	}
	ECAddress = ecAddress

	key, err = primitives.NewPrivateKeyFromHex(c.App.ServerPrivKey)
	if err != nil {
		panic(err)
	}
	ServerPrivKey = key

	FactoidBalanceThreshold = c.Factom.FactoidBalanceThreshold
	ECBalanceThreshold = c.Factom.ECBalanceThreshold
}

func CheckFactomBalance() (int64, int64, error) {
	ecBalance, err := api.GetECBalance(ServerECKey.PublicKeyString())
	if err != nil {
		return 0, 0, err
	}

	fBalance, err := api.GetFactoidBalance(ServerPrivKey.PublicKeyString())
	if err != nil {
		return 0, 0, err
	}
	return fBalance, ecBalance, nil
}

//Returns number of blocks synchronized
func SynchronizeFactomData(dbo *database.AnchorDatabaseOverlay) (int, error) {
	blockCount := 0
	anchorData, err := dbo.FetchAnchorDataHead()
	if err != nil {
		panic(err)
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
		panic(err)
		return 0, err
	}
	if ps.LastFactomDBlockChecked != "" {
		endKeyMR = ps.LastFactomDBlockChecked
	}

	dBlockHead, err := api.GetDBlockHead()
	if err != nil {
		panic(err)
		return 0, err
	}

	//already fully synchronized
	if endKeyMR == dBlockHead {
		return 0, nil
	}

	dBlock, err := api.GetDBlock(dBlockHead)
	if err != nil {
		panic(err)
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
		if keymr == "0000000000000000000000000000000000000000000000000000000000000000" {
			startHeight = int(dBlock.GetDatabaseHeight())
			break
		}
		dBlock, err = api.GetDBlock(keymr)
		if err != nil {
			panic(err)
			return 0, err
		}
		if dBlock == nil {
			panic(err)
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
						panic(err)
						return 0, err
					}
					//fmt.Printf("Entry - %v\n", entry)
					//TODO: update existing anchor entries
					ar, valid, err := anchor.UnmarshalAndvalidateAnchorRecord(entry.GetContent(), AnchorSigPublicKey)
					if err != nil {
						panic(err)
						return 0, err
					}
					if valid == false {
						panic(err)
						return 0, fmt.Errorf("Invalid anchor - %v\n", entry)
					}
					//fmt.Printf("anchor - %v\n", ar)

					anchorData, err = dbo.FetchAnchorData(ar.DBHeight)
					if err != nil {
						panic(err)
						return 0, err
					}
					if anchorData.DBlockKeyMR != ar.KeyMR {
						panic(err)
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
						panic(err)
						return 0, err
					}
					blockCount++
				}
			}
		}

		//Updating new directory blocks
		anchorData, err = dbo.FetchAnchorData(uint32(i))
		if err != nil {
			panic(err)
			return 0, err
		}
		if anchorData == nil {
			anchorData := new(database.AnchorData)
			anchorData.DBlockHeight = dBlock.GetDatabaseHeight()
			anchorData.DBlockKeyMR = dBlock.GetKeyMR().String()
			err = dbo.InsertAnchorData(anchorData, false)
			if err != nil {
				panic(err)
				return 0, err
			}
			blockCount++
		}
	}

	err = dbo.UpdateAnchorDataHead()
	if err != nil {
		panic(err)
		return 0, err
	}

	ps.LastFactomDBlockChecked = dBlockHead
	ps.LastFactomDBlockHeightChecked = currentHeadHeight

	err = dbo.InsertProgramState(ps)
	if err != nil {
		panic(err)
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
			if anchorData.BitcoinRecordEntryHash == "" && anchorData.Bitcoin.BlockHash != "" {
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
			if anchorData.EthereumRecordEntryHash == "" && anchorData.Ethereum.BlockHash != "" {
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

/*
//Construct the entry and submit it to the server
func submitEntryToAnchorChain(aRecord *anchor.AnchorRecord, chainID interfaces.IHash) error {
	//Marshal aRecord into json
	jsonARecord, err := json.Marshal(aRecord)
	anchorLog.Debug("submitEntryToAnchorChain - jsonARecord: ", string(jsonARecord))
	if err != nil {
		return err
	}
	bufARecord := new(bytes.Buffer)
	bufARecord.Write(jsonARecord)
	//Sign the json aRecord with the server key
	aRecordSig := serverPrivKey.Sign(jsonARecord)
	//Encode sig into Hex string
	bufARecord.Write([]byte(hex.EncodeToString(aRecordSig.Sig[:])))

	entry := factom.NewEntry()
	entry.ChainID = cfg.Anchor.AnchorChainID
	entry.Content = bufARecord.Bytes()
	err = primitives.JustFactomize(entry, cfg.Anchor.ServerECKey)

	return err
}*/

func JustFactomize(entry *entryBlock.Entry) error {
	//Convert entryBlock Entry into factom Entry
	//fmt.Printf("Entry - %v\n", entry)
	j, err := entry.JSONByte()
	if err != nil {
		return err
	}
	e := new(factom.Entry)
	err = e.UnmarshalJSON(j)
	if err != nil {
		return err
	}

	//Commit and reveal
	if _, err := factom.CommitEntry(e, ECAddress); err != nil {
		fmt.Println("Entry commit error : ", err)
		return err
	}

	time.Sleep(3 * time.Second)
	if _, err := factom.RevealEntry(e); err != nil {
		fmt.Println("Entry reveal error : ", err)
		return err
	}

	return nil
}

func TopupECAddress() error {
	fmt.Printf("TopupECAddress\n")
	w, err := wallet.NewMapDBWallet()
	if err != nil {
		return err
	}
	defer w.Close()
	priv, err := primitives.PrivateKeyStringToHumanReadableFactoidPrivateKey(ServerPrivKey.PrivateKeyString())
	if err != nil {
		return err
	}
	fa, err := factom.GetFactoidAddress(priv)
	err = w.InsertFCTAddress(fa)
	if err != nil {
		return err
	}

	fAddress, err := factoid.PublicKeyStringToFactoidAddressString(ServerPrivKey.PublicKeyString())
	if err != nil {
		return err
	}
	go wsapi.Start(w, fmt.Sprintf(":%d", 8089))
	defer func() {
		time.Sleep(10 * time.Millisecond)
		wsapi.Stop()
	}()

	ecAddress, err := factoid.PublicKeyStringToECAddressString(ServerECKey.PublicKeyString())
	if err != nil {
		return err
	}

	fmt.Printf("TopupECAddress - %v, %v\n", fAddress, ecAddress)
	/*
		tx, err := factom.BuyEC(fAddress, ecAddress, uint64(ECBalanceThreshold))
		if err != nil {
			return err
		}

	fmt.Printf("tx - %v\n", tx)*/

	return nil
}

/*

func saveToAnchorChain(dirBlockInfo *common.DirBlockInfo) {
	anchorLog.Debug("in saveToAnchorChain")
	anchorRec := new(AnchorRecord)
	anchorRec.AnchorRecordVer = 1
	anchorRec.DBHeight = dirBlockInfo.DBHeight
	anchorRec.KeyMR = dirBlockInfo.DBMerkleRoot.String()
	_, recordHeight, _ := db.FetchBlockHeightCache()
	anchorRec.RecordHeight = uint32(recordHeight + 1) // need the next block height
	if defaultAddress != nil {
		anchorRec.Bitcoin.Address = defaultAddress.String()
	}
	anchorRec.Bitcoin.TXID = dirBlockInfo.BTCTxHash.BTCString()
	anchorRec.Bitcoin.BlockHeight = dirBlockInfo.BTCBlockHeight
	anchorRec.Bitcoin.BlockHash = dirBlockInfo.BTCBlockHash.BTCString()
	anchorRec.Bitcoin.Offset = dirBlockInfo.BTCTxOffset
	anchorLog.Info("before submitting Entry To AnchorChain. anchor.record: " + spew.Sdump(anchorRec))

	err := submitEntryToAnchorChain(anchorRec)
	if err != nil {
		anchorLog.Error("Error in writing anchor into anchor chain: ", err.Error())
	}
}




*/
