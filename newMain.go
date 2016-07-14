package main

import (
	"fmt"

	"github.com/FactomProject/anchormaker/api"
	//"github.com/FactomProject/anchormaker/database"
)

func main() {
	//dbo := database.NewMapDB()

	dBlockHead, err := api.GetDBlockHead()
	if err != nil {
		panic(err)
	}

	dBlock, err := api.GetDBlock(dBlockHead)
	if err != nil {
		panic(err)
	}
	fmt.Printf("dBlock - %v\n", dBlock)
}
