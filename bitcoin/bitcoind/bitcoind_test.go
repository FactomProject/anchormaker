package bitcoind_test

import (
	"testing"

	"github.com/FactomProject/anchormaker/bitcoin"
	. "github.com/FactomProject/anchormaker/bitcoin/bitcoind"
)

func TestTopup(t *testing.T) {
	WalletPassPhrase("password", 100)
	for i := 0; i < 20; i++ {
		SendToAddress(bitcoin.BTCAddress, 1)
	}
	WalletLock()
}
