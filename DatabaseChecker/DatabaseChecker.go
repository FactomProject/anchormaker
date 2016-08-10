package main

import (
	"fmt"

	"github.com/FactomProject/anchormaker/config"
	"github.com/FactomProject/anchormaker/database"
)

func main() {
	c := config.ReadConfig()

	dbo := database.NewMapDB()
	var err error

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

	state, err := dbo.FetchProgramState()
	if err != nil {
		panic(err)
	}
	fmt.Printf("State - %v\n", state.String())

	head, err := dbo.FetchAnchorDataHead()
	if err != nil {
		panic(err)
	}
	fmt.Printf("AnchorDataHead - %v\n", head.String())
	if head == nil {
		head, err = dbo.FetchAnchorData(0)
		if err != nil {
			panic(err)
		}
		fmt.Printf("AnchorData[0] - %v\n", head.String())
	}
}
