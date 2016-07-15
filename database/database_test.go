package database_test

import (
	"testing"

	. "github.com/FactomProject/anchormaker/database"
)

func TestAnchorDatabaseOverlay(t *testing.T) {
	ad := new(AnchorData)
	ad.DBlockHeight = 1
	ad.BitcoinRecordHeight = 2
	ad.EthereumRecordHeight = 3

	dbo := NewMapDB()
	err := dbo.InsertAnchorData(ad, true)
	if err != nil {
		t.Errorf("%v", err)
	}

	ad2 := new(AnchorData)
	ad2.DBlockHeight = 3
	ad2.BitcoinRecordHeight = 4

	err = dbo.InsertAnchorData(ad2, false)
	if err != nil {
		t.Errorf("%v", err)
	}

	ad3, err := dbo.FetchAnchorDataHead()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !(ad3.DBlockHeight == 1) || !(ad3.BitcoinRecordHeight == 2) || !(ad3.EthereumRecordHeight == 3) {
		t.Errorf("Invalid AData fetched - %v", ad3)
	}
}
