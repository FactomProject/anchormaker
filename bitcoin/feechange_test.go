package bitcoin_test

import (
	"fmt"
	. "github.com/FactomProject/anchormaker/bitcoin"
	"testing"
)

func TestFeeChanging(t *testing.T) {
	fmt.Printf("%v\n", BTCFee)
	UpdateFee()
	fmt.Printf("%v\n", BTCFee)
}
