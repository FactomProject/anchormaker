package ethereum

import (
	"fmt"

	"github.com/FactomProject/anchormaker/database"
)

//This file holds the code related to the tracking and handling of pending eth transactions.

// this functions sees if the database object has beet setup with non-nil
// datastructures to hold pending eth transactions
func checkIfPendingDBInitialized(ps *database.ProgramState) (bool, error) {
	var err error
	if ps.PendingTxs == nil {
		return false, err
	}
	return true, err
}

// starts the database with
// it needs to be flushed to disk at a higher layer
func InitializePendingDB(ps *database.ProgramState) error {
	inited, err := checkIfPendingDBInitialized(ps)
	if err != nil {
		return err
	}
	if false == inited {
		fmt.Printf("First run of anchormaker, priming database for pending eth transactions\n")
		outermap := make(map[int64]map[int64]*database.ProgramStatePendingTxInfo)
		ps.PendingTxs = outermap
	}
	return nil
}

// This function reads the database and returns how many pending transactions
// that we have a record of at a particular eth nonce height.
// eth transactions can only confirm at the next available nonce.
func NumPendingAtNonce(ps *database.ProgramState, nonceHeight int64) (int64, error) {
	inited, err := checkIfPendingDBInitialized(ps)
	if err != nil {
		return 0, err
	}
	if false == inited {
		return 0, nil
	}
	if ps.PendingTxs[nonceHeight] == nil {
		return 0, nil
	}
	numTrans := int64(len(ps.PendingTxs[nonceHeight]))

	return numTrans, nil
}

func SaveTransactionAtNonce(ps *database.ProgramState, nonceHeight int64, newTx database.ProgramStatePendingTxInfo) error {
	var txsAtNonceHeight map[int64]*database.ProgramStatePendingTxInfo
	err := InitializePendingDB(ps) //initialize the database, if it isn't already initialized
	if err != nil {
		return err
	}
	numOlderTxs, err := NumPendingAtNonce(ps, nonceHeight)
	if err != nil {
		return err
	}
	if numOlderTxs == 0 {
		txsAtNonceHeight = make(map[int64]*database.ProgramStatePendingTxInfo)
	} else {
		txsAtNonceHeight = ps.PendingTxs[nonceHeight]
	}
	// put this latest transaction at the end of the list of known pending eth transactions for this height
	txsAtNonceHeight[numOlderTxs] = &newTx
	ps.PendingTxs[nonceHeight] = txsAtNonceHeight
	return nil
}

// This function returns the number of pending transactions, and the highest gas price that has been offered
func HighestPendingGasPriceAtNonce(ps *database.ProgramState, nonceHeight int64) (int64, int64, error) {
	numOlderTxs, err := NumPendingAtNonce(ps, nonceHeight)
	var highestGasPrice int64
	highestGasPrice = 0
	if err != nil {
		return 0, 0, err
	}
	if numOlderTxs == 0 {
		return 0, 0, nil
	}
	for _, v := range ps.PendingTxs[nonceHeight] {
		if v == nil {
			return 0, 0, fmt.Errorf("nil pointer in pending eth tx database eth nonce height %v.", nonceHeight)
		}
		thisGasPrice := v.EthTxGasPrice
		if thisGasPrice > highestGasPrice {
			highestGasPrice = thisGasPrice
		}
	}
	return numOlderTxs, highestGasPrice, nil
}

//the number of transactions that a particular address has made that are confirmed
//determines the nonce of the next confirmable transaction.
func getConfirmedCount() (int64, error) {
	data := "0x"
	data += EthereumAPI.StringToMethodID("getAnchor(uint256)")
	data += EthereumAPI.IntToData(height)

	callinfo := new(EthereumAPI.TransactionObject)
	callinfo.To = ContractAddress
	callinfo.Data = data

	keymr, err := EthereumAPI.EthCall(callinfo, "latest")
	if err != nil {
		fmt.Printf("err - %v", err)
		return "", err
	}
	return keymr, nil
}
