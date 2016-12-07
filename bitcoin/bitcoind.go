package bitcoin

import (
	"fmt"

	//"github.com/FactomProject/anchormaker/anchorLog"
	"github.com/FactomProject/anchormaker/bitcoin/bitcoind"
	"github.com/FactomProject/anchormaker/config"

	"github.com/FactomProject/factomd/common/interfaces"
	//"github.com/FactomProject/factomd/common/primitives"
	//"github.com/FactomProject/go-bip32"
)

var BTCAddress = ""

func init() {
	bitcoind.SetAddress("http://localhost:18332/", "user", "pass")
}

func InitRPCClient(cfg *config.AnchorConfig) error {
	return nil
}

// SendRawTransactionToBTC is the main function used to anchor factom
// dir block hash to bitcoin blockchain
func SendRawTransactionToBTC(hash interfaces.IHash, blockHeight uint32) (string, error) {
	return "", nil
}

func ListBitcoinTransactionsSinceBlock(block int64) ([]Transaction, error) {
	bitcoind.GetInfo()
	a, b, err := bitcoind.ListTransactions(nil)
	fmt.Printf("%v, %v, %v\n", a, b, err)
	// "a23f357ffea840509f2c29fa2263244730a3717242996abab50ebe3ae14dff4e"
	tx, err := bitcoind.GetRawTransaction("a849a1ecbf7a1ce59b531e699eaa00d6c8e01bcf6eaab080144d0c70c138dd56")
	fmt.Printf("%v, %v\n", tx, err)
	return nil, nil
}

func ToTransactions(txs []interface{}) []Transaction {
	answer := make([]Transaction, len(txs))
	/*	for i, v := range txs {
		answer[i].TX = v
	}*/
	return answer
}
