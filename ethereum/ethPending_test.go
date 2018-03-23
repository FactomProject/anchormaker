package ethereum_test

import (
	"testing"

	"github.com/FactomProject/anchormaker/database"
	. "github.com/FactomProject/anchormaker/ethereum"
)

func TestDatabaseOfPendingTxs(t *testing.T) {

	dbo := database.NewMapDB()

	ps, err := dbo.FetchProgramState()
	if err != nil {
		t.Errorf("%v", err)
	}

	var testNonce int64
	var testEthTxID string
	var testEthTxGasPrice int64
	var testFactomDBheight int64
	var testFactomDBkeyMR string
	var testTxTime int64

	testNonce = 0

	err = InitializePendingDB(ps)
	if err != nil {
		t.Errorf("%v", err)
	}

	numPending, err := NumPendingAtNonce(ps, testNonce)
	if err != nil {
		t.Errorf("%v", err)
	}
	if numPending != 0 {
		t.Errorf("incorrect pending number")
	}

	var testTX database.ProgramStatePendingTxInfo

	testEthTxID = "128a2db4c625cd8386a3212ebbf5575975a21df9777b0809f982472df8ccdadd"
	testEthTxGasPrice = 10000000000 //10 gwei
	testFactomDBheight = 2961
	testFactomDBkeyMR = "620cdaf755eccd14f939d965920dea04cba824c73f4573f581f673ef694b5936"
	testTxTime = 1257894000

	testTX.EthTxID = testEthTxID
	testTX.EthTxGasPrice = testEthTxGasPrice
	testTX.FactomDBheight = testFactomDBheight
	testTX.FactomDBkeyMR = testFactomDBkeyMR
	testTX.TxTime = testTxTime

	SaveTransactionAtNonce(ps, testNonce, testTX)

	testEthTxID = "339360c4ab7a12136dd12e0cbcf4bcd05208794bc8b0771b1ecf44b917c6b202"
	testEthTxGasPrice = 20000000000 //20 gwei
	testFactomDBheight = 3352
	testFactomDBkeyMR = "e9df36f28c23d8ac55508667ef7a720bd1d03b44a5f0815064ccfa29ddc91d42d"
	testTxTime = 1521765468

	testTX.EthTxID = testEthTxID
	testTX.EthTxGasPrice = testEthTxGasPrice
	testTX.FactomDBheight = testFactomDBheight
	testTX.FactomDBkeyMR = testFactomDBkeyMR
	testTX.TxTime = testTxTime

	SaveTransactionAtNonce(ps, testNonce, testTX)

	numPending, err = NumPendingAtNonce(ps, testNonce)
	if err != nil {
		t.Errorf("%v", err)
	}
	if numPending != 2 {
		t.Errorf("incorrect pending number")
	}

	numTx, gasprice, err := HighestPendingGasPriceAtNonce(ps, testNonce)
	if err != nil {
		t.Errorf("%v", err)
	}
	if numTx != 2 {
		t.Errorf("incorrect pending number")
	}
	if gasprice != 20000000000 {
		t.Errorf("incorrect gas amount")
	}

	testEthTxID = "964542c206103f7c6b5059d042d11835f8c253e27f64c65f28acb1ed77b60fee"
	testEthTxGasPrice = 15000000000 //15 gwei
	testFactomDBheight = 3351
	testFactomDBkeyMR = "dcce78b53e754aefe1e31a11ef024e25143b57d088df4725c95e90c1fcfabfb2"
	testTxTime = 1521765657

	testTX.EthTxID = testEthTxID
	testTX.EthTxGasPrice = testEthTxGasPrice
	testTX.FactomDBheight = testFactomDBheight
	testTX.FactomDBkeyMR = testFactomDBkeyMR
	testTX.TxTime = testTxTime

	SaveTransactionAtNonce(ps, testNonce, testTX)

	numTx, gasprice, err = HighestPendingGasPriceAtNonce(ps, testNonce)
	if err != nil {
		t.Errorf("%v", err)
	}
	if numTx != 3 {
		t.Errorf("incorrect pending number")
	}
	if gasprice != 20000000000 {
		t.Errorf("incorrect gas amount")
	}
}
