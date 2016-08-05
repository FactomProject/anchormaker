package test_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factom/wallet"
	"github.com/FactomProject/factom/wallet/wsapi"

	"github.com/FactomProject/anchormaker/api"
	"github.com/FactomProject/anchormaker/config"
	anchorFactom "github.com/FactomProject/anchormaker/factom"

	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/primitives"
)

func TestTopupECAddress(t *testing.T) {
	c := config.ReadConfig()
	anchorFactom.LoadConfig(c)
	api.SetServer(c.Factom.FactomdAddress)

	fBalance, ecBalance, err := anchorFactom.CheckFactomBalance()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Balances - %v, %v\n", fBalance, ecBalance)

	fmt.Printf("TopupECAddress\n")
	w, err := wallet.NewMapDBWallet()
	if err != nil {
		panic(err)
	}
	defer w.Close()

	serverPrivKey, err := primitives.NewPrivateKeyFromHex("ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973")
	if err != nil {
		panic(err)
	}
	serverECKey, err := primitives.NewPrivateKeyFromHex("ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973")
	if err != nil {
		panic(err)
	}

	priv, err := primitives.PrivateKeyStringToHumanReadableFactoidPrivateKey(serverPrivKey.PrivateKeyString())
	if err != nil {
		panic(err)
	}
	fa, err := factom.GetFactoidAddress(priv)
	err = w.InsertFCTAddress(fa)
	if err != nil {
		panic(err)
	}

	fAddress, err := factoid.PublicKeyStringToFactoidAddressString(serverPrivKey.PublicKeyString())
	if err != nil {
		panic(err)
	}
	go wsapi.Start(w, fmt.Sprintf(":%d", 8089))
	defer func() {
		time.Sleep(10 * time.Millisecond)
		wsapi.Stop()
	}()

	ecAddress, err := factoid.PublicKeyStringToECAddressString(serverECKey.PublicKeyString())
	if err != nil {
		panic(err)
	}

	fmt.Printf("TopupECAddress - %v, %v\n", fAddress, ecAddress)

	tx, err := factom.BuyExactEC(fAddress, ecAddress, uint64(c.Factom.ECBalanceThreshold))
	if err != nil {
		panic(err)
	}

	fmt.Printf("tx - %v\n", tx)

	fmt.Printf("Waiting on ACK...\n")

	for {
		time.Sleep(5 * time.Second)
		ack, err := factom.FactoidACK(tx, "")
		if err != nil {
			panic(err)
		}

		if ack.Status != "DBlockConfirmed" {
			continue
		}

		str, err := primitives.EncodeJSONString(ack)
		if err != nil {
			panic(err)
		}
		fmt.Printf("ack - %v\n", str)

		break
	}

	txResp, err := factom.GetTransaction(tx)
	if err != nil {
		panic(err)
	}
	str, err := primitives.EncodeJSONString(txResp)
	if err != nil {
		panic(err)
	}
	fmt.Printf("txResp - %v\n", str)

	fBalance, ecBalance, err = anchorFactom.CheckFactomBalance()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Updated Balances - %v, %v\n", fBalance, ecBalance)
}
