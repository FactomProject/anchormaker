package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/FactomProject/anchormaker/api"
	"github.com/FactomProject/anchormaker/config"
	"github.com/FactomProject/anchormaker/database"
	"github.com/FactomProject/anchormaker/ethereum"
	"github.com/FactomProject/anchormaker/factom"
	"github.com/FactomProject/anchormaker/setup"
)

func main() {
	c := config.ReadConfig()

	ethereum.LoadConfig(c)
	factom.LoadConfig(c)
	api.SetServer(c.Factom.FactomdAddress)

	err := setup.Setup(c)
	if err != nil {
		panic(err)
	}

	dbo := database.NewMapDB()
	if c.App.DBType == "Map" {
		fmt.Printf("Starting Map database\n")
		dbo = database.NewMapDB()
	}
	if c.App.DBType == "LDB" {
		fmt.Printf("Starting Level database\n")
		dbo, err = database.NewLevelDB(c.App.LdbPath)
		if err != nil {
			panic(err)
		}
	}
	if c.App.DBType == "Bolt" {
		fmt.Printf("Starting Bolt database\n")
		dbo, err = database.NewBoltDB(c.App.BoltPath)
		if err != nil {
			panic(err)
		}
	}

	var interruptChannel chan os.Signal
	interruptChannel = make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)

	go interruptLoop()

	for {
		// ensuring safe interruption
		select {
		case <-interruptChannel:
			return
		default:
			err := SynchronizationLoop(dbo)
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
				time.Sleep(10 * time.Second)
				continue
			}

			err = AnchorLoop(dbo, c)
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
				time.Sleep(10 * time.Second)
				continue
			}
			fmt.Printf("\n\n\n")
			time.Sleep(10 * time.Second)
		}
	}
}

// Function for quickly shutting down the function, disregarding safety
func interruptLoop() {
	var interruptChannel chan os.Signal
	interruptChannel = make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)
	for i := 0; i < 5; i++ {
		<-interruptChannel
		if i < 4 {
			fmt.Printf("Received interrupt signal %v times. The program will shut down safely after a full loop.\nFor emergency shutdown, interrupt %v more times.\n", i+1, 4-i)
		}
	}
	fmt.Printf("Emergency shutdown!\n")
	os.Exit(1)
}

// SynchronizationLoop ensures AnchorMaker is up to date with all relevant events that
// have occurred on both the Factom and Ethereum networks
func SynchronizationLoop(dbo *database.AnchorDatabaseOverlay) error {
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println("SynchronizationLoop():")
	i := 0
	for {
		// Repeat iteration until we are fully in sync with all both Factom and Ethereum
		// to make sure all of the networks are in sync at the same time
		// (nothing has drifted apart while we were busy with other systems)
		fmt.Printf("Loop %v started\n", i)
		blockCount, err := factom.SynchronizeFactomData(dbo)
		if err != nil {
			return err
		}
		fmt.Printf("New Factom blocks found = %v\n", blockCount)

		txCount, err := ethereum.SynchronizeEthereumData(dbo)
		if err != nil {
			return err
		}
		fmt.Printf("New Ethereum contract txs found = %v\n", txCount)

		if (blockCount + txCount) == 0 {
			break
		}
		fmt.Printf("Loop %v ended with %v new DBlocks and %v new Ethereum txs found\n", i, blockCount, txCount)
		i++
	}
	return nil
}

// AnchorLoop submits Ethereum anchors for all new directory blocks found during the SynchronizationLoop,
// and then submits Factom entries (anchor records) for all newly confirmed contract txs found during the
// SynchronizationLoop
func AnchorLoop(dbo *database.AnchorDatabaseOverlay, c *config.AnchorConfig) error {
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println("AnchorLoop():")

	err := setup.CheckAndTopupBalances(c.Factom.ECBalanceThreshold, c.Factom.FactoidBalanceThreshold, 100)
	if err != nil {
		return err
	}

	err = ethereum.AnchorBlocksIntoEthereum(dbo)
	if err != nil {
		return err
	}

	err = factom.SaveAnchorsIntoFactom(dbo)
	if err != nil {
		return err
	}

	return nil
}
