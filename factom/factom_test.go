package factom_test

import (
	"testing"

	"github.com/FactomProject/anchormaker/config"
	. "github.com/FactomProject/anchormaker/factom"
	"github.com/FactomProject/factomd/anchor"
)

func TestCreateValidateEntryRecord(t *testing.T) {
	InitEverything()

	anchorRecord := new(anchor.AnchorRecord)
	anchorRecord.AnchorRecordVer = 1
	anchorRecord.DBHeight = 5
	anchorRecord.KeyMR = "980ab6d50d9fad574ad4df6dba06a8c02b1c67288ee5beab3fbfde2723f73ef6"
	anchorRecord.RecordHeight = 46226

	anchorRecord.Bitcoin = new(anchor.BitcoinStruct)

	anchorRecord.Bitcoin.Address = "1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF"
	anchorRecord.Bitcoin.TXID = "e2ac71c9c0fd8edc0be8c0ba7098b77fb7d90dcca755d5b9348116f3f9d9f951"
	anchorRecord.Bitcoin.BlockHeight = 372576
	anchorRecord.Bitcoin.BlockHash = "000000000000000003059382ed4dd82b2086e99ec78d1b6e811ebb9d53d8656d"
	anchorRecord.Bitcoin.Offset = 1144

	entry, err := CreateAnchorEntry(anchorRecord, BitcoinAnchorChainID, ServerPrivKey)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}

	_, valid, err := anchor.UnmarshalAndValidateAnchorRecordV2(entry.GetContent(), entry.ExternalIDs(), AnchorSigPublicKey)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	if valid == false {
		t.Errorf("Record is invalid!")
		t.FailNow()
	}
}

func InitEverything() {
	c := config.ReadConfig()
	LoadConfig(c)
}
