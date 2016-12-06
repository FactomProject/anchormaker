package bitcoin_test

import (
	"testing"

	. "github.com/FactomProject/anchormaker/bitcoin"
)

func TestTest(t *testing.T) {
	tx, err := ListBitcoinTransactionsSinceBlock(0)
	if err != nil {
		t.Errorf("%v", err)
	}
	for _, v := range tx {
		a, b := v.GetAnchorData()
		t.Errorf("%v - %v, %v", v.GetBlockNumber(), a, b)
	}
	t.FailNow()
}
