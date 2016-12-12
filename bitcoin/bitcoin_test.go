package bitcoin_test

import (
	"testing"

	. "github.com/FactomProject/anchormaker/bitcoin"
)

func TestTest(t *testing.T) {
	tx, err := GetOurUnspentOutputs(BTCAddress)
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Errorf("%v", tx)

	t.FailNow()
}
