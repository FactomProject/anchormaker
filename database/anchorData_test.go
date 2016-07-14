package database_test

import (
	"testing"

	. "github.com/FactomProject/anchormaker/database"
)

func TestAnchorData(t *testing.T) {
	ad := new(AnchorData)
	bin, err := ad.MarshalBinary()
	if err != nil {
		t.Errorf("%v", err)
	}
	ad2 := new(AnchorData)
	err = ad2.UnmarshalBinary(bin)
	if err != nil {
		t.Errorf("%v", err)
	}
}
