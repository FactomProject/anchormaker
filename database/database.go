package database

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
	"github.com/FactomProject/factomd/database/mapdb"
)

var CHAIN_HEAD = []byte("ChainHead")

type AnchorDatabaseOverlay struct {
	databaseOverlay.Overlay
}

func NewAnchorOverlay(db interfaces.IDatabase) *AnchorDatabaseOverlay {
	answer := new(AnchorDatabaseOverlay)
	answer.DB = db
	return answer
}

func NewMapDB() *AnchorDatabaseOverlay {
	return NewAnchorOverlay(new(mapdb.MapDB))
}

func NewLevelDB(ldbpath string) (*AnchorDatabaseOverlay, error) {
	db, err := hybridDB.NewLevelMapHybridDB(ldbpath, false)
	if err != nil {
		fmt.Printf("err opening db: %v\n", err)
	}

	if db == nil {
		fmt.Println("Creating new db ...")
		db, err = hybridDB.NewLevelMapHybridDB(ldbpath, true)

		if err != nil {
			return nil, err
		}
	}
	fmt.Println("Database started from: " + ldbpath)
	return NewAnchorOverlay(db), nil
}

func NewBoltDB(boltPath string) (*AnchorDatabaseOverlay, error) {
	db := hybridDB.NewBoltMapHybridDB(nil, boltPath)
	/*if err != nil {
		fmt.Printf("err opening db: %v\n", err)
	}*/

	fmt.Println("Database started from: " + boltPath)
	return NewAnchorOverlay(db), nil
}

func (db *AnchorDatabaseOverlay) InsertAnchorData(data *AnchorData, isHead bool) error {
	if data == nil {
		return nil
	}

	height := data.DatabasePrimaryIndex()

	batch := []interfaces.Record{}
	batch = append(batch, interfaces.Record{AnchorDataStr, height.Bytes(), data})
	if isHead {
		//Chain head consists only of records anchored in both Bitcoin and Ethereum
		batch = append(batch, interfaces.Record{CHAIN_HEAD, data.GetChainID().Bytes(), height})
	}

	return db.PutInBatch(batch)
}

func (db *AnchorDatabaseOverlay) FetchAnchorData(dbHeight uint32) (*AnchorData, error) {
	height := UintToHash(dbHeight)

	data, err := db.DB.Get(AnchorDataStr, height.Bytes(), new(AnchorData))
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	return data.(*AnchorData), nil
}

func (db *AnchorDatabaseOverlay) FetchAnchorDataHead() (*AnchorData, error) {
	ad := new(AnchorData)
	block, err := db.FetchChainHeadByChainID(AnchorDataStr, ad.GetChainID(), ad)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(*AnchorData), nil
}

func (db *AnchorDatabaseOverlay) UpdateAnchorDataHead() error {
	fmt.Printf("UpdateAnchorDataHead\n")
	ad, err := db.FetchAnchorDataHead()
	if err != nil {
		return err
	}
	var nextCheck uint32
	if ad == nil {
		nextCheck = 0
	} else {
		nextCheck = ad.DBlockHeight + 1
	}
	fmt.Printf("\nnextCheck - %v\n", nextCheck)
	head := ad
	for {
		ad, err = db.FetchAnchorData(nextCheck)
		if err != nil {
			return err
		}
		if ad.IsComplete() {
			fmt.Printf("%v is complete\n", nextCheck)
			head = ad
		} else {
			fmt.Printf("%v is not complete\n", nextCheck)
			//First AnchorData that is not complete ends the chain
			break
		}
		nextCheck++
	}
	fmt.Printf("\n")
	if head != nil {
		err = db.InsertAnchorData(head, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *AnchorDatabaseOverlay) InsertProgramState(data *ProgramState) error {
	if data == nil {
		return nil
	}

	batch := []interfaces.Record{}
	batch = append(batch, interfaces.Record{ProgramStateStr, ProgramStateStr, data})

	return db.PutInBatch(batch)
}

func (db *AnchorDatabaseOverlay) FetchProgramState() (*ProgramState, error) {
	data, err := db.DB.Get(ProgramStateStr, ProgramStateStr, new(ProgramState))
	if err != nil {
		return nil, err
	}
	if data == nil {
		return new(ProgramState), nil
	}
	return data.(*ProgramState), nil
}
