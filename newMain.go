package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/FactomProject/anchormaker/api"
	"github.com/FactomProject/anchormaker/bitcoin"
	"github.com/FactomProject/anchormaker/config"
	"github.com/FactomProject/anchormaker/database"
	//"github.com/FactomProject/anchormaker/ethereum"
	"github.com/FactomProject/anchormaker/factom"
	"github.com/FactomProject/anchormaker/setup"
)

func main() {
	c := config.ReadConfig()

	bitcoin.LoadConfig(c)
	//ethereum.LoadConfig(c)
	factom.LoadConfig(c)
	api.SetServer(c.Factom.FactomdAddress)

	var err error
	err = setup.Setup(c)
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
		//ensuring safe interruption
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

//Function for quickly shutting down the function, disregarding safety
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

//The loop that synchronizes AnchorMaker with all networks
func SynchronizationLoop(dbo *database.AnchorDatabaseOverlay) error {
	fmt.Printf("SynchronizationLoop\n")
	i := 0
	for {
		//Iterate until we are fully in synch with all of the networks
		//Repeat iteration until there is nothing left to synch
		//to make sure all of the networks are in synch at the same time
		//(nothing has drifted apart while we were busy with other systems)
		fmt.Printf("Loop %v\n", i)
		blockCount, err := factom.SynchronizeFactomData(dbo)
		if err != nil {
			return err
		}
		fmt.Printf("blockCount - %v\n", blockCount)
		/*
			txCount, err := ethereum.SynchronizeEthereumData(dbo)
			if err != nil {
				return err
			}
			fmt.Printf("txCount - %v\n", txCount)
		*/
		btcCount, err := bitcoin.SynchronizeBitcoinData(dbo)
		if err != nil {
			return err
		}
		fmt.Printf("btcCount - %v\n", btcCount)

		//if (blockCount + txCount + btcCount) == 0 {
		if (blockCount + btcCount) == 0 {
			//if (blockCount + txCount) == 0 {
			break
		}
		i++
	}
	return nil
}

func AnchorLoop(dbo *database.AnchorDatabaseOverlay, c *config.AnchorConfig) error {
	var err error

	err = setup.CheckAndTopupBalances(c.Factom.ECBalanceThreshold, c.Factom.FactoidBalanceThreshold, 100)
	if err != nil {
		return err
	}

	/*
		err = ethereum.AnchorBlocksIntoEthereum(dbo)
		if err != nil {
			return err
		}
	*/
	err = bitcoin.AnchorBlocksIntoBitcoin(dbo)
	if err != nil {
		return err
	}

	err = factom.SaveAnchorsIntoFactom(dbo)
	if err != nil {
		return err
	}

	return nil
}
