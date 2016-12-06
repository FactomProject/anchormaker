package bitcoin

import (
	"fmt"

	//"github.com/FactomProject/anchormaker/anchorLog"
	"github.com/FactomProject/anchormaker/config"

	"github.com/FactomProject/factomd/common/interfaces"
	//"github.com/FactomProject/factomd/common/primitives"

	//"github.com/FactomProject/go-bip32"
	"github.com/blockcypher/gobcy"
)

//Token: 37ebed87408e441c9a48485e5a7b5fed
const GobcyToken = "37ebed87408e441c9a48485e5a7b5fed"
const GobcyCurrency = "btc"
const GobcyNetwork = "main"
const PrivateKey = ""
const BTCAddress = "1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF"
const FullConfirmations = 6

/*
const GobcyCurrency = "btc"
const GobcyNetwork = "test3"
const PrivateKey = "83CB9A117010B853756AAB744E59797F309F2BE88A56F0BF3007181BBB656575"
const BTCAddress = "mfaGqLAwnv8D1cdoxoiKFs8MXfLc3woV8x"
const FullConfirmations = 1
*/

var BTC gobcy.API

func init() {
	BTC = gobcy.API{GobcyToken, GobcyCurrency, GobcyNetwork}
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
	adr, err := BTC.GetAddrFull(BTCAddress, map[string]string{"after": fmt.Sprintf("%v", block), "before": "372809", "limit": "50"})
	if err != nil {
		return nil, err
	}
	answer := ToTransactions(adr.TXs)
	for adr.HasMore {
		adr, err = BTC.GetAddrFullNext(adr, map[string]string{"after": fmt.Sprintf("%v", block), "limit": "50"})
		if err != nil {
			return nil, err
		}
		answer = append(answer, ToTransactions(adr.TXs)...)
		fmt.Printf("len - %v\n", len(adr.TXs))
	}
	return answer, nil
}

func ToTransactions(txs []gobcy.TX) []Transaction {
	answer := make([]Transaction, len(txs))
	for i, v := range txs {
		answer[i].TX = v
	}
	return answer
}
