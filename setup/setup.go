package setup

import (
	"fmt"
	"time"

	"github.com/FactomProject/factom"

	"github.com/FactomProject/anchormaker/config"
	anchorFactom "github.com/FactomProject/anchormaker/factom"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/primitives"
)

func CheckAndTopupBalances(ecBalanceThreshold, fBalanceThreshold, minimumECBalance int64) error {
	fBalance, ecBalance, err := anchorFactom.CheckFactomBalance()
	if err != nil {
		return err
	}

	fmt.Printf("Balances - %v, %v\n", fBalance, ecBalance)

	if ecBalance < ecBalanceThreshold {
		if fBalance < fBalanceThreshold {
			if ecBalance < minimumECBalance {
				return fmt.Errorf("EC and F Balances are too low, can't do anything!\n")
			} else {
				return nil
			}
		}
		err = anchorFactom.TopupECAddress()
		if err != nil {
			return err
		}
	}

	return nil
}

func Setup(c *config.AnchorConfig) error {
	fmt.Printf("Setting the server up...\n")

	err := CheckAndTopupBalances(c.Factom.ECBalanceThreshold, c.Factom.FactoidBalanceThreshold, c.Factom.ECBalanceThreshold/10)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return err
	}

	err = CheckAndCreateEthereumAnchorchain()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return err
	}

	fmt.Printf("Setup complete!\n")
	return nil
}

func CheckAndCreateEthereumAnchorchain() error {
	anchor := anchorFactom.CreateFirstEthereumAnchorEntry()
	chainID := anchor.GetChainID()

	head, err := factom.GetChainHead(chainID.String())
	if err != nil {
		if err.Error() != "Missing Chain Head" {
			return err
		}
	}
	if head != "" {
		//Chain already exists, nothing to create!
		return nil
	}

	err = CreateChain(anchor)
	if err != nil {
		return err
	}

	return nil
}

func CreateChain(e *entryBlock.Entry) error {
	tx1, tx2, err := anchorFactom.JustFactomizeChain(e)
	if err != nil {
		return err
	}

	fmt.Printf("Created chain %v - %v, %v\n", e.GetChainID(), tx1, tx2)

	for i := 0; ; i++ {
		i = i % 3
		time.Sleep(5 * time.Second)
		ack, err := factom.FactoidACK(tx1, "")
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			return err
		}
		str, err := primitives.EncodeJSONString(ack)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			return err
		}
		fmt.Printf("ack1 - %v", str)
		for j := 0; j < i+1; j++ {
			fmt.Printf(".")
		}
		fmt.Printf("  \r")

		if ack.Status != "DBlockConfirmed" {
			continue
		}
		fmt.Printf("ack1 - %v\n", str)

		break
	}

	for i := 0; ; i++ {
		i = i % 3
		time.Sleep(5 * time.Second)
		ack, err := factom.FactoidACK(tx2, "")
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			return err
		}

		str, err := primitives.EncodeJSONString(ack)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			return err
		}
		fmt.Printf("ack2 - %v", str)
		for j := 0; j < i+1; j++ {
			fmt.Printf(".")
		}
		fmt.Printf("  \r")

		if ack.Status != "DBlockConfirmed" {
			continue
		}

		fmt.Printf("ack2 - %v\n", str)

		break
	}

	return nil
}
