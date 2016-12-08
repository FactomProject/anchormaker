package bitcoin

import (
	//"fmt"

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
	//bitcoind.GetInfo()
	//a, b, err := bitcoind.ListTransactions(nil)
	//fmt.Printf("%v, %v, %v\n", a, b, err)
	// "a23f357ffea840509f2c29fa2263244730a3717242996abab50ebe3ae14dff4e"
	//bitcoind.GetRawTransaction("a849a1ecbf7a1ce59b531e699eaa00d6c8e01bcf6eaab080144d0c70c138dd56")
	bitcoind.GetRawTransactionWithVerbose("a849a1ecbf7a1ce59b531e699eaa00d6c8e01bcf6eaab080144d0c70c138dd56")
	//bitcoind.GetTransaction("a849a1ecbf7a1ce59b531e699eaa00d6c8e01bcf6eaab080144d0c70c138dd56")
	//bitcoind.DecodeRawTransaction("010000000286e9712cebfc724290adb73f4fc32c85284c4611d560231ca2d6b86daecef1bb010000006c493046022100d64669005cd806ade8b537cdd5bf7beb762ba59af44ee8ada5dc49db77360688022100d7fbc3514f1d52ca36b018bbd076f5dd90b1454c595306f67a9a3f346b1f53bd0121029613d89b62157962494fa2ac239e937728b38e1f4c7f9f930f0ac5971fe161f7ffffffff27e421887f6531b4fa2d5b02aeec74aa2b5851b3df768b96c0a6e27625b667d4000000006b483045022100d4f4b4978ebaf9cd6ec2f59726e7bcaff61c87432794fa265c74e52c04ed500e02203c0f5be60526c3e6c21bd5d41404abf35b137afb870d6acfbfb92dc0c2d8d455012103e60c1b579d63529d4fe43f1aa1b040e205c18a7140a16c6fd199df540651e578ffffffff020d871300000000001976a914c8fb779ed9f8516adbb51e26babad4ccf14723fe88ac00c2eb0b000000001976a9146ef3f48c1ba7dd8cbde9d5a29dc9e4aeda28d3f388ac00000000")
	return nil, nil
}

func ToTransactions(txs []interface{}) []Transaction {
	answer := make([]Transaction, len(txs))
	/*	for i, v := range txs {
		answer[i].TX = v
	}*/
	return answer
}
