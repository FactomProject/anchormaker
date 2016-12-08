package bitcoin

import (
	"fmt"
	"strings"

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

func ListBitcoinTransactionsSinceBlock(block string) ([]Transaction, string, error) {
	txs, resp, err := bitcoind.ListSinceBlock(block, 6)
	if err != nil {
		return nil, "", err
	}
	if resp.Error != nil {
		return nil, "", fmt.Errorf("%v", resp.Error)
	}
	ts, err := ToTransactions(txs.Transactions)
	if err != nil {
		return nil, "", err
	}
	return ts, txs.LastBlock, nil
}

func ToTransactions(txs []bitcoind.Transaction) ([]Transaction, error) {
	var answer []Transaction

	for _, v := range txs {
		if v.Category != "send" {
			//Ignore transactions that we don't send ourselves
			continue
		}

		fullTx, r, err := bitcoind.GetRawTransactionWithVerbose(v.TxID)
		if err != nil {
			return nil, err
		}
		if r.Error != nil {
			return nil, fmt.Errorf("%v", r.Error)
		}
		var tx Transaction
		for i, out := range fullTx.VOut {
			if strings.Contains(out.ScriptPubKey.ASM, "OP_RETURN") {
				tx.OPReturn = out.ScriptPubKey.Hex
				tx.OpReturnIndex = int64(i)
				break
			}
		}
		if tx.OPReturn == "" {
			//Transaction doesn't have OP_RETURN, can ignore
			//continue
		}

		tx.TxHash = fullTx.TxID
		tx.BlockHash = fullTx.BlockHash
		tx.InputAddresses = []string{v.Address}

		block, resp, err := bitcoind.GetBlock(tx.BlockHash)
		if err != nil {
			return nil, err
		}
		if resp.Error != nil {
			return nil, fmt.Errorf("%v", resp.Error)
		}
		tx.BlockNumber = block.Height

		answer = append(answer, tx)
	}

	return answer, nil
}
