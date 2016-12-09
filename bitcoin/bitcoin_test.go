package bitcoin_test

import (
	"testing"

	. "github.com/FactomProject/anchormaker/bitcoin"
)

func TestTest(t *testing.T) {
	tx, hash, err := ListBitcoinTransactionsSinceBlock("00000000ac859aa8572cb4ade286e081f6d0ec662a1043693be637e4aeae6b4f")
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Errorf("%v, %v", tx, hash)

	t.FailNow()
}
