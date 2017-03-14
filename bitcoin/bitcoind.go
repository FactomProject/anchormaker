package bitcoin

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	//"github.com/FactomProject/anchormaker/anchorLog"
	"github.com/FactomProject/anchormaker/bitcoin/bitcoind"
	"github.com/FactomProject/anchormaker/config"
	//"github.com/FactomProject/factomd/common/interfaces"
	//"github.com/FactomProject/factomd/common/primitives"
	//"github.com/FactomProject/go-bip32"
)

var BTCAddress string = "mxnf2a9MfEjvkjS4zL7efoWSgbZe5rMn1m"

//var BTCPrivKey = "cRhC7gEZMJdZ35SrBbcRX19R1sM3f5F1tHsmjPvsbfLSds81FxQp"
var BTCFee float64 = 0.0002
var MinConfirmations int64 = 1
var WalletPassphrase string = "password"
var RPCAddress string = "http://localhost:18332/"
var RPCUser string = "user"
var RPCPass string = "pass"

func init() {
	bitcoind.SetAddress(RPCAddress, RPCUser, RPCPass)
}

func InitRPCClient(cfg *config.AnchorConfig) error {
	BTCAddress = cfg.Bitcoin.BTCAddress
	BTCFee = cfg.Bitcoin.BTCFee

	MinConfirmations = cfg.Bitcoin.MinConfirmations
	WalletPassphrase = cfg.Bitcoin.WalletPassphrase

	RPCAddress = cfg.Bitcoin.RPCAddress
	RPCUser = cfg.Bitcoin.RPCUser
	RPCPass = cfg.Bitcoin.RPCPass

	bitcoind.SetAddress(RPCAddress, RPCUser, RPCPass)
	return nil
}

func UpdateFee() {
	fee, _, err := bitcoind.EstimateFee(1)
	if err != nil {
		//If we have an error, revert to default
		BTCFee = 0.001
		return
	}
	if fee > 0 {
		//If bitcoind gives us an estimate, use it
		//Our transactions are ~243bytes, estimatefee lists price per 1kB
		//So we divide by ~4
		BTCFee = fee * 0.243
		return
	}
	//If bitcoind can't estimate the fee, revert to default
	BTCFee = 0.001
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
	var inputCount int = 0
	var BTCFeeForThisTx float64
	BTCFeeForThisTx = BTCFee
	for _, v := range inputs {
		totalInputs += v.Amount
		usedList = append(usedList, bitcoind.RawTransactionInput{TxID: v.TXId, VOut: v.VOut})
		inputCount += 1
		if totalInputs > BTCFee*4 {
			break
		}
	}
	//additional inputs beyond one each take 146 bytes.
	//add additional fee for the extra inputs
	if inputCount > 1 {
		if inputCount > 5 {
			fmt.Printf("Trying to make a tx with > 5 inputs. only paying fees for 5 of the %v inputs\n", inputCount)
			inputCount = 5 //dont let spam dust burn a lot of fees.
		}
		var feeMultiplier float64
		var extraBytes float64
		extraBytes = float64(inputCount-1) * 146 //each additional input takes about 146 bytes
		feeMultiplier = (243 + extraBytes) / 243 //the normal transaction takes 243 bytes
		if feeMultiplier > 5 {
			fmt.Printf("Trying use a fee multiplier > 5x. Reducing to 5 from %v\n", feeMultiplier)
			feeMultiplier = 5
		}
		BTCFeeForThisTx = BTCFee * feeMultiplier
		fmt.Printf("More than 1 input for a BTC tx, increasing fee from %v to %v\n", BTCFee, BTCFeeForThisTx)
	}
	if totalInputs < BTCFeeForThisTx {
		return "", nil //fmt.Errorf("Not enough money to cover fees")
	}

	outputs := map[string]interface{}{}
	outputs[address] = trimBTCFloat(totalInputs - BTCFeeForThisTx)
	outputs["data"] = data

	raw, resp, err := bitcoind.CreateRawTransaction(usedList, outputs)
	if err != nil {
		fmt.Printf("Problem with tx. Inputs:\n%v\n Outputs:\n%v\n", usedList, outputs)
		return "", err
	}
	if resp.Error != nil {
		fmt.Printf("Problem with tx. Inputs:\n%v\n Outputs:\n%v\n", usedList, outputs)
		return "", fmt.Errorf("%v", resp.Error)
	}
	bitcoind.WalletPassPhrase(WalletPassphrase, 10)
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

func trimBTCFloat(f float64) float64 {
	var tmp string = "%.8f"
	tmp = fmt.Sprintf(tmp, f)
	answer, _ := strconv.ParseFloat(tmp, 64)
	return answer
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
	fmt.Printf("ListBitcoinTransactionsSinceBlockReturned\n")
	if err != nil {
		fmt.Printf("error in bitcoind.ListSinceBlock %v %v\n", block, MinConfirmations)
		return nil, "", err
	}
	if resp == nil || txs == nil {
		if resp == nil {
			fmt.Printf("nil response in bitcoind.ListSinceBlock %v %v\n", block, MinConfirmations)
		}
		if txs == nil {
			fmt.Printf("no transactions returned in bitcoind.ListSinceBlock %v %v\n", block, MinConfirmations)
		}

		return nil, "", fmt.Errorf("Function returned nothing - should not happen!")
	}
	if resp.Error != nil {
		fmt.Printf("error returned from resp.Error bitcoind.ListSinceBlock %v %v\n", block, MinConfirmations)
		return nil, "", fmt.Errorf("%v", resp.Error)
	}
	fmt.Printf("sending ToTransactions %v\n", len(txs.Transactions))
	ts, err := ToTransactions(txs.Transactions)
	if err != nil {
		fmt.Printf("error in ToTransactions\n")
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

		gotTx, r, err := bitcoind.GetTransaction(v.TxID)
		if err != nil {
			fmt.Printf("GetTransaction Error for Tx - %v\n", v.String())
			return nil, err
		}
		if r.Error != nil {
			fmt.Printf("GetTransaction Error for Tx - %v\n", v.String())
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

		if gotTx.BlockIndex == 0 {
			//bitcoin API call couldn't get the index.  something is fishy, so error
			//the default is zero, but the anchor cannot be in the coinbase, which should be index zero.
			fmt.Printf("bitcoin Error couldn't get block index for Tx - %v\n", v.String())
			return nil, fmt.Errorf("bitcoin Error couldn't get block index for Tx - %v\n", v.String())
		}
		tx.TransactionBlockIndex = gotTx.BlockIndex

		block, resp, err := bitcoind.GetBlock(tx.BlockHash)
		if err != nil {
			return nil, err
		}
		if resp.Error != nil {
			return nil, fmt.Errorf("%v", resp.Error)
		}
		tx.BlockNumber = block.Height

		time.Sleep(50 * time.Millisecond) //slow down accessing the bitcoind RPC when pulling lots of transactions

		answer = append(answer, tx)
	}

	return answer, nil
}
