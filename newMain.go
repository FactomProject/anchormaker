package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/FactomProject/anchormaker/bitcoin"
	"github.com/FactomProject/anchormaker/config"
	"github.com/FactomProject/anchormaker/database"
	"github.com/FactomProject/anchormaker/ethereum"
	"github.com/FactomProject/anchormaker/factom"
)

func main() {
	dbo := database.NewMapDB()

	c := config.ReadConfig()
	bitcoin.LoadConfig(c)
	ethereum.LoadConfig(c)
	factom.LoadConfig(c)

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
				panic(err)
			}

			err = AnchorLoop(dbo)
			if err != nil {
				panic(err)
			}
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

		txCount, err := ethereum.SynchronizeEthereumData(dbo)
		if err != nil {
			return err
		}
		fmt.Printf("txCount - %v\n", txCount)

		btcCount, err := bitcoin.SynchronizeBitcoinData(dbo)
		if err != nil {
			return err
		}
		fmt.Printf("btcCount - %v\n", btcCount)

		if (blockCount + txCount + btcCount) == 0 {
			break
		}
		i++
	}
	return nil
}

func AnchorLoop(dbo *database.AnchorDatabaseOverlay) error {
	err := ethereum.AnchorBlocksIntoEthereum(dbo)
	if err != nil {
		return err
	}

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
