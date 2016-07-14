package database

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/database/databaseOverlay"
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

func (db *AnchorDatabaseOverlay) InsertAnchorData(data *AnchorData) error {
	if data == nil {
		return nil
	}

	height := data.DatabasePrimaryIndex()

	batch := []interfaces.Record{}
	batch = append(batch, interfaces.Record{AnchorDataStr, height.Bytes(), data})
	if data.IsComplete() {
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
