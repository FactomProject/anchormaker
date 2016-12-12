package bitcoin

import (
	"encoding/hex"
	"fmt"
	"strings"

	//"github.com/FactomProject/anchormaker/anchorLog"
	"github.com/FactomProject/anchormaker/bitcoin/bitcoind"
	"github.com/FactomProject/anchormaker/config"
	//"github.com/FactomProject/factomd/common/interfaces"
	//"github.com/FactomProject/factomd/common/primitives"
	//"github.com/FactomProject/go-bip32"
)

var BTCAddress = "mxnf2a9MfEjvkjS4zL7efoWSgbZe5rMn1m"
var BTCPrivKey = "cRhC7gEZMJdZ35SrBbcRX19R1sM3f5F1tHsmjPvsbfLSds81FxQp"
var BTCFee float64 = 0.001
var MinConfirmations int64 = 1

func init() {
	bitcoind.SetAddress("http://localhost:18332/", "user", "pass")
}

func InitRPCClient(cfg *config.AnchorConfig) error {
	return nil
}

// SendRawTransactionToBTC is the main function used to anchor factom
// dir block hash to bitcoin blockchain
func SendRawTransactionToBTC(hash string, blockHeight uint32) (string, error) {
	fmt.Printf("SendRawTransactionToBTC - %v, %v\n", blockHeight, hash)
	b, err := hex.DecodeString(hash)
	if err != nil {
		return "", err
	}
	data, err := prependBlockHeight(blockHeight, b)
	if err != nil {
		return "", err
	}
	list, err := GetSpendableOutputs(BTCAddress)
	if err != nil {
		return "", err
	}
	return SendTransaction(list, BTCAddress, fmt.Sprintf("%x", data))
}

func SendTransaction(inputs []bitcoind.UnspentOutput, address, data string) (string, error) {
	var totalInputs float64 = 0
	usedList := []bitcoind.RawTransactionInput{}
	for _, v := range inputs {
		totalInputs += v.Amount
		usedList = append(usedList, bitcoind.RawTransactionInput{TxID: v.TXId, VOut: v.VOut})
		if totalInputs > BTCFee {
			break
		}
	}
	if totalInputs < BTCFee {
		return "", fmt.Errorf("Not enough money to cover fees")
	}

	outputs := map[string]interface{}{}
	outputs[address] = totalInputs - BTCFee
	outputs["data"] = data

	raw, resp, err := bitcoind.CreateRawTransaction(usedList, outputs)
	if err != nil {
		return "", err
	}
	if resp.Error != nil {
		return "", fmt.Errorf("%v", resp.Error)
	}
	bitcoind.WalletPassPhrase("password", 10)
	signed, resp, err := bitcoind.SignRawTransaction(raw)
	bitcoind.WalletLock()
	if err != nil {
		return "", err
	}
	if resp.Error != nil {
		return "", fmt.Errorf("%v", resp.Error)
	}

	txID, resp, err := bitcoind.SendRawTransaction(signed.Hex)
	if err != nil {
		return "", err
	}
	if resp.Error != nil {
		return "", fmt.Errorf("%v", resp.Error)
	}

	return txID, nil
}

func GetOurUnspentOutputs(address string) ([]bitcoind.UnspentOutput, error) {
	list, resp, err := bitcoind.ListUnspent()
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("%v", resp.Error)
	}
	if len(list) == 0 {
		return nil, nil
	}
	for i := len(list) - 1; i >= 0; i-- {
		if list[i].Address != address {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list, nil
}

func GetSpendableOutputs(address string) ([]bitcoind.UnspentOutput, error) {
	list, resp, err := bitcoind.ListUnspent()
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("%v", resp.Error)
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list, nil
}

func ListBitcoinTransactionsSinceBlock(block string) ([]Transaction, string, error) {
	fmt.Printf("ListBitcoinTransactionsSinceBlock\n")
	txs, resp, err := bitcoind.ListSinceBlock(block, MinConfirmations)
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
		if v.BlockHash == "" {
			//Ignore unconfirmed transactions
			continue
		}
		if v.Address == "" {
			//Ignore transactions that are just OP_returns
			continue
		}

		fullTx, r, err := bitcoind.GetRawTransactionWithVerbose(v.TxID)
		if err != nil {
			fmt.Printf("Error for Tx - %v\n", v.String())
			return nil, err
		}
		if r.Error != nil {
			fmt.Printf("Error for Tx - %v\n", v.String())
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
			continue
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
