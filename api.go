// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/FactomProject/FactomCode/anchor"
	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/FactomCode/database/ldb"
	"github.com/FactomProject/FactomCode/factomlog"
	"github.com/FactomProject/FactomCode/util"
	"github.com/FactomProject/btcd/wire"
	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/web"
)

const (
	httpOK  = 200
	httpBad = 400
)

var (
	portNumber       = 8103
	inMsgQueue       = make(chan wire.FtmInternalMsg, 1000) //incoming message queue for factom application messages
	appConfig        = util.ReadConfig().App
	serverPrivKeyHex = "3931696a6e7768764233584b5179377059374e3844317844764862557145614d41485745334b4e376232765a6957376f514641"
	serverPrivKey    common.PrivateKey
	serverPubKey     common.PublicKey
)

var _ = fmt.Println

var server = web.NewServer()

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func Start() {
	logcfg := util.ReadConfig().Log
	logPath := UserHomeDir() + "/.factom/anchor.log"
	logLevel := logcfg.LogLevel

	logfile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		fmt.Println("ERR:", err)
		logfile, err = os.Create(logPath)
		if err != nil {
			fmt.Println("NOW ERR:", err)
		}
	}

	anchorLog := factomlog.New(logfile, logLevel, "ANCHOR")

	anchorLog.Debug("Setting Handlers")
	server.Post("/init", handleInit)
	server.Post("/anchor/([^/]+)", handleAnchor)

	anchorLog.Info("Starting server")
	go server.Run(fmt.Sprintf(":%d", portNumber))
}

func Stop() {
	server.Close()
}

func handleInit(ctx *web.Context) {
	myHex := hex.EncodeToString([]byte("91ijnwhvB3XKQy7pY7N8D1xDvHbUqEaMAHWE3KN7b2vZiW7oQFA"))
	ctx.WriteString(fmt.Sprintf("%v\n", myHex))
	startAnchoring()
}

func handleAnchor(ctx *web.Context, anchorNum string) {
	ctx.WriteString(fmt.Sprintf("NUM: %v\n", anchorNum))
	testBlock := directoryBlock.NewDirectoryBlock(0, nil)
	fmt.Println(testBlock)

	head, err := factom.GetDBlockHead()
	if err != nil {
		fmt.Println(err)
		return
	}
	dblock, err := factom.GetDBlock(head.KeyMR)
	if err != nil {
		fmt.Println(err)
		return
	}
	myNum := uint32(dblock.Header.SequenceNumber)
	fmt.Println("DBlock:", head.KeyMR)
	fmt.Println(dblock)

	placeAnchor(convertStringToCommonHash(head.KeyMR), myNum)
}

func main() {
	fmt.Println("+=======================+")
	fmt.Println("|  anchormaker-bitcoin  |")
	fmt.Println("+=======================+")

	startAnchoring()
	Start()
	for {
		time.Sleep(time.Second)
	}
}

func startAnchoring() { //(ldb database.Db, inQ chan wire.FtmInternalMsg) {
	serverPrivKeyHex = appConfig.ServerPrivKey

	ldbpath := appConfig.LdbPath
	db, err := ldb.OpenLevelDB(ldbpath, false)

	if err != nil {
		log.Println("err opening db: %v", err)
	}

	if db == nil {
		log.Println("Creating new db ...")
		db, err = ldb.OpenLevelDB(ldbpath, true)

		if err != nil {
			panic(err)
		}
	}

	if appConfig.NodeMode == common.SERVER_NODE || appConfig.NodeMode == "FULL" {
		serverPrivKey, err = common.NewPrivateKeyFromHex(serverPrivKeyHex)
		if err != nil {
			panic("Cannot parse Server Private Key from configuration file: " + err.Error())
		}
		//Set server's public key
		serverPubKey = serverPrivKey.Pub

		anchor.InitAnchor(db, inMsgQueue, serverPrivKey)
	}
}

func convertIHashToCommonHash(ihash interfaces.IHash) *common.Hash {
	chash := common.NewHash()
	chash.SetBytes(ihash.Bytes())
	return chash
}

func convertStringToCommonHash(stringHash string) *common.Hash {
	chash := common.NewHash()
	chash.SetBytes([]byte(stringHash))
	return chash
}

// Place an anchor into btc
func placeAnchor(keyMR *common.Hash, dbHeight uint32) error {
	fmt.Println("PLACE")
	fmt.Printf("%+v\n", keyMR)
	// Only Servers can write the anchor to Bitcoin network
	if (appConfig.NodeMode == common.SERVER_NODE || appConfig.NodeMode == "FULL") && len(keyMR.Bytes()) > 0 {
		// todo: need to make anchor as a go routine, independent of factomd
		// same as blockmanager to btcd
		//go anchor.SendRawTransactionToBTC(keyMR, dbHeight)
		fmt.Println("PLACE2")
		go tryAnchor(keyMR, dbHeight)

	}
	return nil
}

func tryAnchor(keyMR *common.Hash, dbHeight uint32) {
	a, err := anchor.SendRawTransactionToBTC(keyMR, dbHeight)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("A:", a)
	}
}
