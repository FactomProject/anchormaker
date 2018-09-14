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

var IgnoreWrongEntries bool = true

var AnchorSigPublicKeys []interfaces.Verifier

var ServerECKey *primitives.PrivateKey
var ServerPrivKey *primitives.PrivateKey
var ECAddress *factom.ECAddress

var FactoidBalanceThreshold int64
var ECBalanceThreshold int64

//6e4540d08d5ac6a1a394e982fb6a2ab8b516ee751c37420055141b94fe070bfe
var EthereumAnchorChainID interfaces.IHash

var FirstEthereumAnchorChainEntryHash interfaces.IHash

func init() {
	e := CreateFirstEthereumAnchorEntry()
	EthereumAnchorChainID = e.ChainID
	FirstEthereumAnchorChainEntryHash = e.GetHash()
}

func LoadConfig(c *config.AnchorConfig) {
	for _, v := range c.Anchor.AnchorSigPublicKey {
		pubKey := new(primitives.PublicKey)
		err := pubKey.UnmarshalText([]byte(v))
		if err != nil {
			panic(err)
		}
		AnchorSigPublicKeys = append(AnchorSigPublicKeys, pubKey)
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
	AnchorSigPublicKeys = append(AnchorSigPublicKeys, ServerPrivKey.Pub)

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

// Returns number of blocks synchronized
func SynchronizeFactomData(dbo *database.AnchorDatabaseOverlay) (int, error) {
	fmt.Println("SynchronizeFactomData")
	blockCount := 0
	ps, err := dbo.FetchProgramState()
	if err != nil {
		return 0, err
	}
	//note, this mutex could probably be reworked to prevent a short time span of a race here between fetch and lock
	ps.ProgramStateMutex.Lock()
	defer ps.ProgramStateMutex.Unlock()
	nextHeight := ps.LastFactomDBlockHeightChecked
	if nextHeight > 0 {
		//If it's 0, we don't know if we have ANY blocks. If it's more than 0, we know we have that block, so we skip it
		nextHeight++
	}

	dBlockList := []interfaces.IDirectoryBlock{}

	for {
		dBlock, err := api.GetDBlockByHeight(nextHeight)
		if err != nil {
			return 0, err
		}
		if dBlock == nil {
			break
		}

		dBlockList = append(dBlockList, dBlock)
		fmt.Printf("Fetched dblock %v\n", dBlock.GetDatabaseHeight())
		nextHeight = dBlock.GetDatabaseHeight() + 1
	}

	if len(dBlockList) == 0 {
		return 0, nil
	}
	var currentHeadHeight uint32 = 0

	for _, dBlock := range dBlockList {
		for _, v := range dBlock.GetDBEntries() {
			//Looking for Ethereum anchor record
			if v.GetChainID().String() == EthereumAnchorChainID.String() {
				//fmt.Printf("Entry is being parsed - %v\n", v.GetChainID())
				entryBlock, err := api.GetEBlock(v.GetKeyMR().String())
				if err != nil {
					return 0, err
				}
				for _, eh := range entryBlock.GetEntryHashes() {
					if eh.IsMinuteMarker() == true {
						continue
					}
					//fmt.Printf("\t%v\n", eh.String())
					if eh.String() == FirstEthereumAnchorChainEntryHash.String() {
						continue
					}
					//fmt.Printf("Fetching %v\n", eh.String())
					entry, err := api.GetEntry(eh.String())
					if err != nil {
						return 0, err
					}
					//fmt.Printf("Entry - %v\n", entry)
					ar, valid, err := anchor.UnmarshalAndValidateAnchorEntryAnyVersion(entry, AnchorSigPublicKeys)
					if err != nil {
						return 0, err
					}
					if valid == false {
						fmt.Printf("Invalid anchor - %v\n", entry)
						continue
						//return 0, fmt.Errorf("Invalid anchor - %v\n", entry)
					}
					//fmt.Printf("anchor - %v\n", ar)

					anchorData, err := dbo.FetchAnchorData(ar.DBHeight)
					if err != nil {
						return 0, err
					}
					if anchorData.MerkleRoot == "" {
						// Calculate Merkle root that should be in the anchor record
						hi := ar.DBHeight
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
						anchorData.MerkleRoot = merkleRoot.String()
					}
					if anchorData.MerkleRoot != ar.KeyMR {
						if IgnoreWrongEntries == false {
							fmt.Printf("%v, %v\n", ar.DBHeight, anchorData)
							panic(fmt.Sprintf("%v vs %v", anchorData.MerkleRoot, ar.KeyMR))
							return 0, fmt.Errorf("AnchorData MerkleRoot does not match AnchorRecord MerkleRoot")
						} else {
							fmt.Printf("Bad AR: Height %v has MerkleRoot %v, found anchorrecord %v\n", ar.DBHeight, anchorData.MerkleRoot, ar.KeyMR)
							continue
						}
					}

					if ar.Ethereum != nil {
						fmt.Printf("Found Ethereum anchor record in Factom DBlock %v: %v, %v\n", dBlock.GetDatabaseHeight(), ar.DBHeight, ar.KeyMR)
						anchorData.Ethereum.Address = ar.Ethereum.Address
						anchorData.Ethereum.TXID = ar.Ethereum.TXID
						anchorData.Ethereum.BlockHeight = ar.Ethereum.BlockHeight
						anchorData.Ethereum.BlockHash = ar.Ethereum.BlockHash
						anchorData.Ethereum.Offset = ar.Ethereum.Offset

						anchorData.EthereumRecordHeight = dBlock.GetDatabaseHeight()
						//fmt.Printf("dBlock.GetDatabaseHeight() - %v\n", dBlock.GetDatabaseHeight())
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

		// Updating new directory blocks
		anchorData, err := dbo.FetchAnchorData(dBlock.GetDatabaseHeight())
		if err != nil {
			return 0, err
		}
		if anchorData == nil {
			anchorData := new(database.AnchorData)
			anchorData.DBlockHeight = dBlock.GetDatabaseHeight()
			err = dbo.InsertAnchorData(anchorData, false)
			if err != nil {
				return 0, err
			}
			blockCount++
		}
		currentHeadHeight = dBlock.GetDatabaseHeight()

		ps.LastFactomDBlockHeightChecked = currentHeadHeight

		err = dbo.InsertProgramState(ps)
		if err != nil {
			return 0, err
		}
	}

	err = dbo.UpdateAnchorDataHead()
	if err != nil {
		return 0, err
	}

	return blockCount, nil
}

func SaveAnchorsIntoFactom(dbo *database.AnchorDatabaseOverlay) error {
	fmt.Println("SaveAnchorsIntoFactom")
	ps, err := dbo.FetchProgramState()
	if err != nil {
		fmt.Println("error checking progam state")
		return err
	}
	// This mutex could probably be reworked to prevent a short time span of a race here between fetch and lock
	ps.ProgramStateMutex.Lock()
	defer ps.ProgramStateMutex.Unlock()
	anchorData, err := dbo.FetchAnchorDataHead()
	if err != nil {
		fmt.Println("error fetching anchor data head")
		return err
	}
	if anchorData == nil {
		anchorData, err = dbo.FetchAnchorData(0)
		if err != nil {
			fmt.Println("error FetchAnchorData")
			return err
		}
		if anchorData == nil {
			//nothing found
			return nil
		}
	}

	// Try to submit as many as 10 anchor records/receipts into Factom
	for i := 0; i < 10; {
		// Only anchor records that haven't been anchored before
		// Check that the anchor tx was confirmed on other blockchain, and that we haven't recorded that tx's receipt in Factom yet
		// Factom Entry Hash has to be empty and Ethereum TxID must not be empty
		if anchorData.EthereumRecordEntryHash == "" && anchorData.Ethereum.BlockHash != "" {
			anchorRecord := new(anchor.AnchorRecord)
			anchorRecord.AnchorRecordVer = 1
			anchorRecord.DBHeight = anchorData.DBlockHeight
			// TODO: update AnchorRecord struct in factomd to reflect that this is no longer the KeyMR of a singe DBlock, but for a window
			anchorRecord.KeyMR = anchorData.MerkleRoot
			anchorRecord.RecordHeight = ps.LastFactomDBlockHeightChecked + 1

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
			anchorData.EthereumRecordEntryHash = tx

			//Resetting AnchorRecord
			anchorRecord.Ethereum = nil

			err = dbo.InsertAnchorData(anchorData, false)
			if err != nil {
				fmt.Println("error in InsertAnchorData")
				return err
			}
			i++
		}
		anchorData, err = dbo.FetchAnchorData(anchorData.DBlockHeight + 1)
		if err != nil {
			fmt.Println("error in FetchAnchorData")
			return err
		}
		if anchorData == nil {
			fmt.Println("error anchordata is nil")
			break
		}
	}
	return nil
}

// Takes care of sending the entry to the Factom network, returns txID
func CreateAndSendAnchor(ar *anchor.AnchorRecord) (string, error) {
	fmt.Printf("Sending anchor record to Factom: %v\n", ar)
	if ar.Ethereum != nil {
		txID, err := submitEntryToAnchorChain(ar, EthereumAnchorChainID)
		if err != nil {
			return "", err
		}
		return txID, nil
	}
	return "", nil
}

// Construct the entry and submit it to the server
func submitEntryToAnchorChain(aRecord *anchor.AnchorRecord, chainID interfaces.IHash) (string, error) {
	entry, err := CreateAnchorEntry(aRecord, chainID, ServerPrivKey)
	if err != nil {
		return "", err
	}

	_, txID, err := JustFactomize(entry)
	if err != nil {
		return "", err
	}

	return txID, err
}

func CreateAnchorEntry(aRecord *anchor.AnchorRecord, chainID interfaces.IHash, serverPrivKey *primitives.PrivateKey) (*entryBlock.Entry, error) {
	record, sig, err := aRecord.MarshalAndSignV2(ServerPrivKey)
	if err != nil {
		return nil, err
	}

	entry := new(entryBlock.Entry)
	entry.ChainID = chainID
	entry.Content = primitives.ByteSlice{Bytes: record}
	entry.ExtIDs = []primitives.ByteSlice{primitives.ByteSlice{Bytes: sig}}

	return entry, nil
}

func JustFactomizeChain(entry *entryBlock.Entry) (string, string, error) {
	//Convert entryBlock Entry into factom Entry
	//fmt.Printf("Entry - %v\n", entry)
	j, err := entry.JSONByte()
	if err != nil {
		return "", "", err
	}
	e := new(factom.Entry)
	err = e.UnmarshalJSON(j)
	if err != nil {
		return "", "", err
	}

	chain := factom.NewChain(e)

	//Commit and reveal
	tx1, err := factom.CommitChain(chain, ECAddress)
	if err != nil {
		fmt.Println("Entry commit error : ", err)
		return "", "", err
	}

	time.Sleep(10 * time.Second)
	tx2, err := factom.RevealChain(chain)
	if err != nil {
		fmt.Println("Entry reveal error : ", err)
		return "", "", err
	}

	return tx1, tx2, nil
}

func JustFactomize(entry *entryBlock.Entry) (string, string, error) {
	//Convert entryBlock Entry into factom Entry
	//fmt.Printf("Entry - %v\n", entry)
	j, err := entry.JSONByte()
	if err != nil {
		return "", "", err
	}
	e := new(factom.Entry)
	err = e.UnmarshalJSON(j)
	if err != nil {
		return "", "", err
	}

	//Commit and reveal
	tx1, err := factom.CommitEntry(e, ECAddress)
	if err != nil {
		fmt.Println("Entry commit error : ", err)
		return "", "", err
	}

	time.Sleep(3 * time.Second)
	tx2, err := factom.RevealEntry(e)
	if err != nil {
		fmt.Println("Entry reveal error : ", err)
		return "", "", err
	}

	return tx1, tx2, nil
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
	wsapiIP := fmt.Sprintf("localhost:%d", 8089)
	go wsapi.Start(w, wsapiIP, config.ReadConfig().Walletd)
	defer func() {
		time.Sleep(10 * time.Millisecond)
		wsapi.Stop()
	}()
	factom.SetWalletServer(wsapiIP)

	ecAddress, err := factoid.PublicKeyStringToECAddressString(ServerECKey.PublicKeyString())
	if err != nil {
		return err
	}

	fmt.Printf("TopupECAddress - %v, %v\n", fAddress, ecAddress)

	tx, err := factom.BuyExactEC(fAddress, ecAddress, uint64(ECBalanceThreshold), true)
	if err != nil {
		return err
	}

	fmt.Printf("Topup tx - %v\n", tx)

	for i := 0; ; i++ {
		i = i % 3
		time.Sleep(5 * time.Second)
		ack, err := factom.FactoidACK(tx.TxID, "")
		if err != nil {
			return err
		}

		str, err := primitives.EncodeJSONString(ack)
		if err != nil {
			return err
		}
		fmt.Printf("Topup ack - %v", str)
		for j := 0; j < i+1; j++ {
			fmt.Printf(".")
		}
		fmt.Printf("  \r")

		if ack.Status != "DBlockConfirmed" {
			continue
		}

		fmt.Printf("Topup ack - %v\n", str)

		break
	}

	_, ecBalance, err := CheckFactomBalance()
	if err != nil {
		return err
	}
	if ecBalance < ECBalanceThreshold {
		return fmt.Errorf("Balance was not increased!")
	}

	return nil
}

func CreateFirstEthereumAnchorEntry() *entryBlock.Entry {
	answer := new(entryBlock.Entry)

	answer.Version = 0
	answer.ExtIDs = []primitives.ByteSlice{primitives.ByteSlice{Bytes: []byte("FactomEthereumAnchorChain")}}
	answer.Content = primitives.ByteSlice{Bytes: []byte("This is the Factom Ethereum anchor chain, which records the anchors Factom puts on the Ethereum network.\n")}
	answer.ChainID = entryBlock.NewChainID(answer)

	return answer
}
