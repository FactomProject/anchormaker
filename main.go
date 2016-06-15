package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/FactomCode/database"
	"github.com/FactomProject/FactomCode/database/ldb"
	"github.com/FactomProject/FactomCode/factomlog"
	"github.com/FactomProject/factom"

	"github.com/btcsuitereleases/btcd/btcjson"
	"github.com/btcsuitereleases/btcd/chaincfg"
	"github.com/btcsuitereleases/btcd/txscript"
	"github.com/btcsuitereleases/btcd/wire"
	"github.com/btcsuitereleases/btcrpcclient"
	"github.com/btcsuitereleases/btcutil"
	"github.com/davecgh/go-spew/spew"

	factomwire "github.com/FactomProject/FactomCode/wire"
)

var (
	zeroID           = "0000000000000000000000000000000000000000000000000000000000000000"
	myMonitor        = Monitor{Height: 0}
	balances         []balance // unspent balance & address & its WIF
	cfg              *anchorConfig
	dclient, wclient *btcrpcclient.Client
	dirBlockInfoMap  = make(map[string]*common.DirBlockInfo) //DBMerkleRoot string as key
	db               database.Db
	dblocks          []DBlock
	walletLocked     = true
	reAnchorAfter    = 4  // hours. For anchors that do not get bitcoin callback info for over 10 hours, then re-anchor them.
	tenMinutes       = 10 // 10 minute mark
	defaultAddress   btcutil.Address
	minBalance       btcutil.Amount

	fee                 btcutil.Amount // tx fee for written into btc
	confirmationsNeeded int

	//Server Private key for milestone 1
	serverPrivKey    common.PrivateKey
	serverPrivKeyHex string
	serverPubKey     common.PublicKey

	//Server Entry Credit private key
	serverECKey common.PrivateKey
	//Anchor chain ID
	anchorChainID *common.Hash
	//Logger
	anchorLog *factomlog.FLogger
	//InmsgQ for submitting the entry to server
	inMsgQ = make(chan factomwire.FtmInternalMsg, 100) //incoming message queue for factom application messages
)

type Monitor struct {
	Height       int
	BlockHead    string
	DeepestBlock string
}

type DBlock struct {
	PrevBlockKeyMR string
	Timestamp      uint64
	SequenceNumber int

	//BlockTimeStr string
	KeyMR string
}

type balance struct {
	unspentResult btcjson.ListUnspentResult
	address       btcutil.Address
	wif           *btcutil.WIF
}

//AnchorRecord is used to construct anchor chain
type AnchorRecord struct {
	AnchorRecordVer int
	DBHeight        uint32
	KeyMR           string
	RecordHeight    uint32

	Bitcoin struct {
		Address     string //"1HLoD9E4SDFFPDiYfNYnkBLQ85Y51J3Zb1",
		TXID        string //"9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5", BTC Hash - in reverse byte order
		BlockHeight int32  //345678,
		BlockHash   string //"00000000000000000cc14eacfc7057300aea87bed6fee904fd8e1c1f3dc008d4", BTC Hash - in reverse byte order
		Offset      int32  //87
	}
}

/*func main() {
	fmt.Println("=======================")
	fmt.Println("grabbing chain:", AnchorChainID)
	fmt.Println("=======================")
	firstEntry, err := factom.GetFirstEntry(AnchorChainID)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(firstEntry)
	//factom.DB
	myChain, err := factom.GetAllChainEntries(AnchorChainID)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("GO")

	for i, entry := range myChain {
		if i > 100 {
			break
		}
		fmt.Println(entry.String())

	}
}*/

func initDB(ldbpath string) {
	//init db
	var err error
	db, err = ldb.OpenLevelDB(ldbpath, false)

	if err != nil {
		fmt.Printf("err opening db: %v\n", err)

	}

	if db == nil {
		fmt.Println("Creating new db ...")
		db, err = ldb.OpenLevelDB(ldbpath, true)

		if err != nil {
			panic(err)
		}
	}
	fmt.Println("Database started from: " + ldbpath)

}

func initServerKeys() {
	var err error
	serverPrivKey, err = common.NewPrivateKeyFromHex(serverPrivKeyHex)
	if err != nil {
		panic("Cannot parse Server Private Key from configuration file: " + err.Error())
	}
	//Set server's public key
	serverPubKey = serverPrivKey.Pub
}

func main() {
	readConfig()
	anchorLog.Info("Anchormaker: Initializing db...\n")
	initServerKeys()
	initDB(cfg.App.LdbPath)
	fmt.Printf("Anchormaker: Starting synchWithFactomState...\n")

	//go synchWithFactomState()
	InitAnchor(db, inMsgQ, serverPrivKey)
	synchWithFactomState()
}

func synchWithFactomState() {
	for {
		err := loadRemainingBlocks()
		if err != nil {
			fmt.Println(err)
			if strings.HasSuffix(err.Error(), "getsockopt: connection refused") {
				fmt.Printf("\nIn order for anchormaker to synchronize successfully, factomd must be running.\n")
				fmt.Print("Once factomd is running, press 'Enter' to continue...")
				bufio.NewReader(os.Stdin).ReadBytes('\n')
				continue
			} else {
				os.Exit(1)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func loadRemainingBlocks() error {
	factomHeight, err := factom.GetDBlockHeight()
	if err != nil {
		return err
	}
	//processRemainingAnchors()
	if len(myMonitor.DeepestBlock) == 0 || myMonitor.DeepestBlock == zeroID {
		if myMonitor.Height < factomHeight {
			factomHead, err := factom.GetDBlockHead()
			if err != nil {
				return err
			}

			loadBlocksStartingAt(factomHead.KeyMR)

			myMonitor.Height = factomHeight
			myMonitor.BlockHead = factomHead.KeyMR
			myMonitor.DeepestBlock = zeroID
		}
	} else {
		loadBlocksStartingAt(myMonitor.DeepestBlock)
	}

	//processRemainingAnchors()
	return nil
}

func loadBlocksStartingAt(startBlockKeyMR string) {
	if len(dblocks) > 5 {
		checkMissingDirBlockInfo()
	}
	saveKeyMR := ""
	nextKeyMR := startBlockKeyMR
	sleepCounter := 0
	for nextKeyMR != myMonitor.BlockHead && nextKeyMR != zeroID {
		saveKeyMR = (nextKeyMR + " ")[:len(nextKeyMR)]
		nextKeyMR = processBlock(saveKeyMR)
		sleepCounter++
		if sleepCounter%1000 == 0 {
			time.Sleep(10 * time.Second)
		}

		myMonitor.DeepestBlock = saveKeyMR
	}
	myMonitor.DeepestBlock = zeroID
}

func processBlock(keyMR string) string {
	dblock, err := GetDBlockFromFactom(keyMR)
	if err != nil {
		fmt.Printf("Error getting top dblock from factom: %+v\n", err)
		return dblock.PrevBlockKeyMR
	}

	dblocks = append(dblocks, *dblock)

	return dblock.PrevBlockKeyMR
}

func GetDBlockFromFactom(keyMR string) (*DBlock, error) {
	answer := new(DBlock)

	body, err := factom.GetDBlock(keyMR)
	if err != nil {
		return answer, err
	}

	answer = new(DBlock)
	answer.PrevBlockKeyMR = body.Header.PrevBlockKeyMR
	answer.Timestamp = body.Header.Timestamp
	answer.SequenceNumber = body.Header.SequenceNumber

	answer.KeyMR = keyMR

	//fmt.Println("GOT BLOCK:", answer)

	return answer, nil
}

func checkMissingDirBlockInfo() {
	//dblocks, _ := db.FetchAllDBlocks()

	dirBlockInfoMap2, err := db.FetchAllDirBlockInfo()
	if err != nil {
		fmt.Println("FETCH DBIM:", err)
	}
	fmt.Println(len(dirBlockInfoMap2))
	for _, dblock := range dblocks {
		if _, ok := dirBlockInfoMap2[dblock.KeyMR]; ok {
			// anchorLog.Debug("Existing dirBlock.KeyMR", dblock.KeyMR.String())
			continue
		} else {
			dirBlockInfo := &common.DirBlockInfo{}
			dbmr, _ := common.HexToHash(dblock.KeyMR)
			dirBlockInfo.DBHash = dbmr
			dirBlockInfo.DBHeight = uint32(dblock.SequenceNumber)
			dirBlockInfo.Timestamp = int64(dblock.Timestamp * 60) //time.Now().Unix()
			dirBlockInfo.DBMerkleRoot = dbmr
			dirBlockInfo.BTCConfirmed = false
			dirBlockInfo.BTCTxHash = common.NewHash()
			dirBlockInfo.BTCBlockHash = common.NewHash()

			/*dirBlockInfo := common.NewDirBlockInfoFromDBlock(&dblock)
			dirBlockInfo.Timestamp = time.Now().Unix()
			db.InsertDirBlockInfo(dirBlockInfo)*/
			dirBlockInfoMap[dirBlockInfo.DBMerkleRoot.String()] = dirBlockInfo
			db.InsertDirBlockInfo(dirBlockInfo)

		}
	}
	//fmt.Println(dirBlockInfoMap)
}

// SendRawTransactionToBTC is the main function used to anchor factom
// dir block hash to bitcoin blockchain
func SendRawTransactionToBTC(hash *common.Hash, blockHeight uint32) (*wire.ShaHash, error) {
	anchorLog.Debug("SendRawTransactionToBTC: hash=", hash.String(), ", dir block height=", blockHeight) //strconv.FormatUint(blockHeight, 10))
	dirBlockInfo, err := sanityCheck(hash)
	if err != nil {
		return nil, err
	}
	return doTransaction(hash, blockHeight, dirBlockInfo)
}

func doTransaction(hash *common.Hash, blockHeight uint32, dirBlockInfo *common.DirBlockInfo) (*wire.ShaHash, error) {
	b := balances[0]
	balances = balances[1:]
	anchorLog.Info("new balances.len=", len(balances))

	msgtx, err := createRawTransaction(b, hash.Bytes(), blockHeight)
	if err != nil {
		return nil, fmt.Errorf("cannot create Raw Transaction: %s", err)
	}

	shaHash, err := sendRawTransaction(msgtx)
	if err != nil {
		return nil, fmt.Errorf("cannot send Raw Transaction: %s", err)
	}
	anchorLog.Info("btc.tx.sha=", shaHash.String())

	if dirBlockInfo != nil {
		dirBlockInfo.BTCTxHash = toHash(shaHash)
		dirBlockInfo.Timestamp = time.Now().Unix()
		db.InsertDirBlockInfo(dirBlockInfo)
	}

	return shaHash, nil
}

func sanityCheck(hash *common.Hash) (*common.DirBlockInfo, error) {
	dirBlockInfo := dirBlockInfoMap[hash.String()]
	if dirBlockInfo == nil {
		s := fmt.Sprintf("Anchor Error: hash %s does not exist in dirBlockInfoMap.\n", hash.String())
		anchorLog.Error(s)
		return nil, errors.New(s)
	}
	if dirBlockInfo.BTCConfirmed {
		s := fmt.Sprintf("Anchor Warning: hash %s has already been confirmed in btc block chain.\n", hash.String())
		anchorLog.Error(s)
		return nil, errors.New(s)
	}
	if dclient == nil || wclient == nil {
		s := fmt.Sprintf("\n\n$$$ WARNING: rpc clients and/or wallet are not initiated successfully. No anchoring for now.\n")
		anchorLog.Warning(s)
		return nil, errors.New(s)
	}
	if len(balances) == 0 {
		anchorLog.Warning("len(balances) == 0, start rescan UTXO *** ")
		updateUTXO(minBalance)
	}
	if len(balances) == 0 {
		anchorLog.Warning("len(balances) == 0, start rescan UTXO *** ")
		updateUTXO(fee)
	}
	if len(balances) == 0 {
		s := fmt.Sprintf("\n\n$$$ WARNING: No balance in your wallet. No anchoring for now.\n")
		anchorLog.Warning(s)
		return nil, errors.New(s)
	}
	return dirBlockInfo, nil
}

func createRawTransaction(b balance, hash []byte, blockHeight uint32) (*wire.MsgTx, error) {
	msgtx := wire.NewMsgTx()

	if err := addTxOuts(msgtx, b, hash, blockHeight); err != nil {
		return nil, fmt.Errorf("cannot addTxOuts: %s", err)
	}

	if err := addTxIn(msgtx, b); err != nil {
		return nil, fmt.Errorf("cannot addTxIn: %s", err)
	}

	if err := validateMsgTx(msgtx, []btcjson.ListUnspentResult{b.unspentResult}); err != nil {
		return nil, fmt.Errorf("cannot validateMsgTx: %s", err)
	}

	return msgtx, nil
}

func addTxIn(msgtx *wire.MsgTx, b balance) error {
	output := b.unspentResult
	//anchorLog.Infof("unspentResult: %s\n", spew.Sdump(output))
	prevTxHash, err := wire.NewShaHashFromStr(output.TxID)
	if err != nil {
		return fmt.Errorf("cannot get sha hash from str: %s", err)
	}
	if prevTxHash == nil {
		anchorLog.Error("prevTxHash == nil")
	}

	outPoint := wire.NewOutPoint(prevTxHash, output.Vout)
	msgtx.AddTxIn(wire.NewTxIn(outPoint, nil))
	if outPoint == nil {
		anchorLog.Error("outPoint == nil")
	}

	// OnRedeemingTx
	err = dclient.NotifySpent([]*wire.OutPoint{outPoint})
	if err != nil {
		anchorLog.Error("NotifySpent err: ", err.Error())
	}

	subscript, err := hex.DecodeString(output.ScriptPubKey)
	if err != nil {
		return fmt.Errorf("cannot decode scriptPubKey: %s", err)
	}
	if subscript == nil {
		anchorLog.Error("subscript == nil")
	}

	sigScript, err := txscript.SignatureScript(msgtx, 0, subscript, txscript.SigHashAll, b.wif.PrivKey, true)
	if err != nil {
		return fmt.Errorf("cannot create scriptSig: %s", err)
	}
	if sigScript == nil {
		anchorLog.Error("sigScript == nil")
	}

	msgtx.TxIn[0].SignatureScript = sigScript
	return nil
}

func addTxOuts(msgtx *wire.MsgTx, b balance, hash []byte, blockHeight uint32) error {
	anchorHash, err := prependBlockHeight(blockHeight, hash)
	if err != nil {
		anchorLog.Errorf("ScriptBuilder error: %v\n", err)
	}

	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_RETURN)
	builder.AddData(anchorHash)

	// latest routine from Conformal btcsuite returns 2 parameters, not 1... not sure what to do for people with the old conformal libraries :(
	opReturn, err := builder.Script()
	msgtx.AddTxOut(wire.NewTxOut(0, opReturn))
	if err != nil {
		anchorLog.Errorf("ScriptBuilder error: %v\n", err)
	}

	amount, _ := btcutil.NewAmount(b.unspentResult.Amount)
	change := amount - fee

	// Check if there are leftover unspent outputs, and return coins back to
	// a new address we own.
	if change > 0 {
		// Spend change.
		pkScript, err := txscript.PayToAddrScript(b.address)
		if err != nil {
			return fmt.Errorf("cannot create txout script: %s", err)
		}
		msgtx.AddTxOut(wire.NewTxOut(int64(change), pkScript))
	}
	return nil
}

func validateMsgTx(msgtx *wire.MsgTx, inputs []btcjson.ListUnspentResult) error {
	flags := txscript.ScriptBip16 | txscript.ScriptStrictMultiSig //ScriptCanonicalSignatures
	bip16 := time.Now().After(txscript.Bip16Activation)
	if bip16 {
		flags |= txscript.ScriptBip16
	}

	for i := range msgtx.TxIn {
		scriptPubKey, err := hex.DecodeString(inputs[i].ScriptPubKey)
		if err != nil {
			return fmt.Errorf("cannot decode scriptPubKey: %s", err)
		}
		engine, err := txscript.NewEngine(scriptPubKey, msgtx, i, flags)
		//engine, err := txscript.NewEngine(scriptPubKey, msgtx, i, flags, nil)
		if err != nil {
			anchorLog.Errorf("cannot create script engine: %s\n", err)
			return fmt.Errorf("cannot create script engine: %s", err)
		}
		if err = engine.Execute(); err != nil {
			anchorLog.Errorf("cannot execute script engine: %s\n  === UnspentResult: %s", err, spew.Sdump(inputs[i]))
			return fmt.Errorf("cannot execute script engine: %s", err)
		}
	}
	return nil
}

func sendRawTransaction(msgtx *wire.MsgTx) (*wire.ShaHash, error) {
	//anchorLog.Debug("sendRawTransaction: msgTx=", spew.Sdump(msgtx))
	buf := bytes.Buffer{}
	buf.Grow(msgtx.SerializeSize())
	if err := msgtx.BtcEncode(&buf, wire.ProtocolVersion); err != nil {
		return nil, err
	}

	// use rpc client for btcd here for better callback info
	// this should not require wallet to be unlocked
	shaHash, err := dclient.SendRawTransaction(msgtx, false)
	if err != nil {
		return nil, fmt.Errorf("failed in rpcclient.SendRawTransaction: %s", err)
	}
	anchorLog.Info("btc txHash returned: ", shaHash) // new tx hash
	return shaHash, nil
}

func createBtcwalletNotificationHandlers() btcrpcclient.NotificationHandlers {
	ntfnHandlers := btcrpcclient.NotificationHandlers{
		OnWalletLockState: func(locked bool) {
			anchorLog.Info("wclient: OnWalletLockState, locked=", locked)
			walletLocked = locked
		},
	}
	return ntfnHandlers
}

func createBtcdNotificationHandlers() btcrpcclient.NotificationHandlers {
	ntfnHandlers := btcrpcclient.NotificationHandlers{
		OnRedeemingTx: func(transaction *btcutil.Tx, details *btcjson.BlockDetails) {
			if details != nil {
				// do not block OnRedeemingTx callback
				//anchorLog.Info(" saveDirBlockInfo.")
				go saveDirBlockInfo(transaction, details)
			}
		},
	}
	return ntfnHandlers
}

// InitAnchor inits rpc clients for factom
// and load up unconfirmed DirBlockInfo from leveldb
func InitAnchor(ldb database.Db, q chan factomwire.FtmInternalMsg, serverKey common.PrivateKey) {
	anchorLog.Debug("InitAnchor")
	fmt.Println("INITANCHOR")
	db = ldb
	inMsgQ = q
	serverPrivKey = serverKey
	minBalance, _ = btcutil.NewAmount(0.01)

	var err error
	dirBlockInfoMap, err = db.FetchAllUnconfirmedDirBlockInfo()
	if err != nil {
		anchorLog.Error("InitAnchor error - " + err.Error())
		return
	}
	anchorLog.Debug("init dirBlockInfoMap.len=", len(dirBlockInfoMap))

	// this might take a while to check missing DirBlockInfo for existing DirBlocks in database
	go checkMissingDirBlockInfo()

	if err = InitRPCClient(); err != nil {
		anchorLog.Error(err.Error())
	} else {
		updateUTXO(minBalance)
	}

	ticker0 := time.NewTicker(time.Second * time.Duration(4))
	go func() {
		for _ = range ticker0.C {
			if wclient != nil && dclient != nil {
				checkForAnchor()
			} else {
				anchorLog.Warning("\n\n$$$ WARNING: rpc clients and/or wallet are not initiated successfully. No anchoring for now.\n")
			}
		}
	}()

	ticker := time.NewTicker(time.Minute * time.Duration(tenMinutes))
	go func() {
		for _ = range ticker.C {
			anchorLog.Info("In 10 minutes ticker...")
			readConfig()
			if dclient == nil || wclient == nil {
				if err = InitRPCClient(); err != nil {
					anchorLog.Error(err.Error())
				}
			}
			if wclient != nil {
				checkTxConfirmations()
			}
		}
	}()
}

func readConfig() {
	cfg = readAnchorConfig()

	// setup logger
	logfile, _ := os.OpenFile(cfg.Log.LogPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	anchorLog = factomlog.New(logfile, cfg.Log.LogLevel, "ANCH")
	anchorLog.Info("readConfig")

	confirmationsNeeded = cfg.Anchor.ConfirmationsNeeded
	fee, _ = btcutil.NewAmount(cfg.Btc.BtcTransFee)

	//Added anchor parameters
	var err error
	serverECKey, err = common.NewPrivateKeyFromHex(cfg.Anchor.ServerECKey)
	if err != nil {
		panic("Cannot parse Server EC Key from configuration file: " + err.Error())
	}
	anchorChainID, err = common.HexToHash(cfg.Anchor.AnchorChainID)
	// anchorLog.Debug("anchorChainID: ", anchorChainID)
	if err != nil || anchorChainID == nil {
		panic("Cannot parse Server AnchorChainID from configuration file: " + err.Error())
	}
	serverPrivKeyHex = cfg.App.ServerPrivKey
}

// InitRPCClient is used to create rpc client for btcd and btcwallet
// and it can be used to test connecting to btcd / btcwallet servers
// running in different machine.
func InitRPCClient() error {
	anchorLog.Debug("init RPC client")
	if cfg == nil {
		readConfig()
	}
	certHomePath := cfg.Btc.CertHomePath
	rpcClientHost := cfg.Btc.RpcClientHost
	rpcClientEndpoint := cfg.Btc.RpcClientEndpoint
	rpcClientUser := cfg.Btc.RpcClientUser
	rpcClientPass := cfg.Btc.RpcClientPass
	certHomePathBtcd := cfg.Btc.CertHomePathBtcd
	rpcBtcdHost := cfg.Btc.RpcBtcdHost

	// Connect to local btcwallet RPC server using websockets.
	ntfnHandlers := createBtcwalletNotificationHandlers()
	certHomeDir := btcutil.AppDataDir(certHomePath, false)
	anchorLog.Debug("btcwallet.cert.home=", certHomeDir)
	certs, err := ioutil.ReadFile(filepath.Join(certHomeDir, "rpc.cert"))
	if err != nil {
		return fmt.Errorf("cannot read rpc.cert file: %s\n", err)
	}
	connCfg := &btcrpcclient.ConnConfig{
		Host:         rpcClientHost,
		Endpoint:     rpcClientEndpoint,
		User:         rpcClientUser,
		Pass:         rpcClientPass,
		Certificates: certs,
	}
	wclient, err = btcrpcclient.New(connCfg, &ntfnHandlers)
	if err != nil {
		return fmt.Errorf("cannot create rpc client for btcwallet: %s\n", err)
	}
	anchorLog.Debug("successfully created rpc client for btcwallet")

	// Connect to local btcd RPC server using websockets.
	dntfnHandlers := createBtcdNotificationHandlers()
	certHomeDir = btcutil.AppDataDir(certHomePathBtcd, false)
	anchorLog.Debug("btcd.cert.home=", certHomeDir)
	certs, err = ioutil.ReadFile(filepath.Join(certHomeDir, "rpc.cert"))
	if err != nil {
		return fmt.Errorf("cannot read rpc.cert file for btcd rpc server: %s\n", err)
	}
	dconnCfg := &btcrpcclient.ConnConfig{
		Host:         rpcBtcdHost,
		Endpoint:     rpcClientEndpoint,
		User:         rpcClientUser,
		Pass:         rpcClientPass,
		Certificates: certs,
	}
	dclient, err = btcrpcclient.New(dconnCfg, &dntfnHandlers)
	if err != nil {
		return fmt.Errorf("cannot create rpc client for btcd: %s\n", err)
	}
	anchorLog.Debug("successfully created rpc client for btcd")

	return nil
}

func unlockWallet(timeoutSecs int64) error {
	err := wclient.WalletPassphrase(cfg.Btc.WalletPassphrase, int64(timeoutSecs))
	if err != nil {
		return fmt.Errorf("cannot unlock wallet with passphrase: %s", err)
	}
	walletLocked = false
	return nil
}

// ByAmount defines the methods needed to satisify sort.Interface to
// sort a slice of UTXOs by their amount.
type ByAmount []balance

func (u ByAmount) Len() int           { return len(u) }
func (u ByAmount) Less(i, j int) bool { return u[i].unspentResult.Amount < u[j].unspentResult.Amount }
func (u ByAmount) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }

func updateUTXO(base btcutil.Amount) error {
	anchorLog.Info("updateUTXO: base=", base.ToBTC())
	if wclient == nil {
		anchorLog.Info("updateUTXO: wclient is nil")
		return nil
	}
	err := unlockWallet(int64(6)) //600
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	unspentResults, err := wclient.ListUnspentMin(confirmationsNeeded) //minConf=1
	if err != nil {
		return fmt.Errorf("cannot list unspent. %s", err)
	}
	anchorLog.Info("updateUTXO: unspentResults.len=", len(unspentResults))

	if len(unspentResults) > 0 {
		balances = make([]balance, 0, len(unspentResults))
		// var i int
		for _, b := range unspentResults {
			if b.Amount > base.ToBTC() && b.Account != "imported" { //fee.ToBTC()
				balances = append(balances, balance{unspentResult: b})
				// i++
			}
		}
	}
	anchorLog.Info("updateUTXO: balances.len=", len(balances))

	// Sort eligible balances so that we first pick the ones with highest one
	sort.Sort(sort.Reverse(ByAmount(balances)))
	for i, b := range balances {
		addr, err := btcutil.DecodeAddress(b.unspentResult.Address, &chaincfg.TestNet3Params)
		if err != nil {
			return fmt.Errorf("cannot decode address: %s", err)
		}
		balances[i].address = addr

		wif, err := wclient.DumpPrivKey(addr)
		if err != nil {
			return fmt.Errorf("cannot get WIF: %s", err)
		}
		balances[i].wif = wif
		//anchorLog.Infof("balance[%d]=%s \n", i, spew.Sdump(balances[i]))
	}

	if len(balances) > 0 {
		defaultAddress = balances[0].address
	}
	return nil
}

func prependBlockHeight(height uint32, hash []byte) ([]byte, error) {
	// dir block genesis block height starts with 0, for now
	// similar to bitcoin genesis block
	h := uint64(height)
	if 0xFFFFFFFFFFFF&h != h {
		return nil, errors.New("bad block height")
	}
	header := []byte{'F', 'a'}
	big := make([]byte, 8)
	binary.BigEndian.PutUint64(big, h) //height)
	newdata := append(big[2:8], hash...)
	newdata = append(header, newdata...)
	return newdata, nil
}

func saveDirBlockInfo(transaction *btcutil.Tx, details *btcjson.BlockDetails) {
	anchorLog.Debug("in saveDirBlockInfo")
	var saved = false
	for _, dirBlockInfo := range dirBlockInfoMap {
		if bytes.Compare(dirBlockInfo.BTCTxHash.Bytes(), transaction.Sha().Bytes()) == 0 {
			doSaveDirBlockInfo(transaction, details, dirBlockInfo, false)
			saved = true
			break
		}
	}
	// This happends when there's a double spending or tx malleated(for dir block 122 and its btc tx)
	// Original: https://www.blocktrail.com/BTC/tx/ac82f4173259494b22f4987f1e18608f38f1ff756fb4a3c637dfb5565aa5e6cf
	// malleated: https://www.blocktrail.com/BTC/tx/a9b2d6b5d320c7f0f384a49b167524aca9c412af36ed7b15ca7ea392bccb2538
	// re-anchored: https://www.blocktrail.com/BTC/tx/ac82f4173259494b22f4987f1e18608f38f1ff756fb4a3c637dfb5565aa5e6cf
	// In this case, if tx malleation is detected, then use the malleated tx to replace the original tx;
	// Otherwise, it will end up being re-anchored.
	if !saved {
		anchorLog.Infof("Not saved to db, (maybe btc tx malleated): btc.tx=%s\n blockDetails=%s\n", spew.Sdump(transaction), spew.Sdump(details))
		checkTxMalleation(transaction, details)
	}
}

func doSaveDirBlockInfo(transaction *btcutil.Tx, details *btcjson.BlockDetails, dirBlockInfo *common.DirBlockInfo, replace bool) {
	if replace {
		dirBlockInfo.BTCTxHash = toHash(transaction.Sha()) // in case of tx being malleated
	}
	dirBlockInfo.BTCTxOffset = int32(details.Index)
	dirBlockInfo.BTCBlockHeight = details.Height
	btcBlockHash, _ := wire.NewShaHashFromStr(details.Hash)
	dirBlockInfo.BTCBlockHash = toHash(btcBlockHash)
	dirBlockInfo.Timestamp = time.Now().Unix()
	db.InsertDirBlockInfo(dirBlockInfo)
	anchorLog.Infof("In doSaveDirBlockInfo, dirBlockInfo:%s saved to db\n", spew.Sdump(dirBlockInfo))

	// to make factom / explorer more user friendly, instead of waiting for
	// over 2 hours to know if it's anchored, we can create the anchor chain instantly
	// then change it when the btc main chain re-org happens.
	saveToAnchorChain(dirBlockInfo)
}

func saveToAnchorChain(dirBlockInfo *common.DirBlockInfo) {
	anchorLog.Debug("in saveToAnchorChain")
	anchorRec := new(AnchorRecord)
	anchorRec.AnchorRecordVer = 1
	anchorRec.DBHeight = dirBlockInfo.DBHeight
	anchorRec.KeyMR = dirBlockInfo.DBMerkleRoot.String()
	_, recordHeight, _ := db.FetchBlockHeightCache()
	anchorRec.RecordHeight = uint32(recordHeight + 1) // need the next block height
	anchorRec.Bitcoin.Address = defaultAddress.String()
	anchorRec.Bitcoin.TXID = dirBlockInfo.BTCTxHash.BTCString()
	anchorRec.Bitcoin.BlockHeight = dirBlockInfo.BTCBlockHeight
	anchorRec.Bitcoin.BlockHash = dirBlockInfo.BTCBlockHash.BTCString()
	anchorRec.Bitcoin.Offset = dirBlockInfo.BTCTxOffset
	anchorLog.Info("before submitting Entry To AnchorChain. anchor.record: " + spew.Sdump(anchorRec))

	err := submitEntryToAnchorChain(anchorRec)
	if err != nil {
		anchorLog.Error("Error in writing anchor into anchor chain: ", err.Error())
	}
}

func toHash(txHash *wire.ShaHash) *common.Hash {
	h := new(common.Hash)
	h.SetBytes(txHash.Bytes())
	return h
}

func toShaHash(hash *common.Hash) *wire.ShaHash {
	h, _ := wire.NewShaHash(hash.Bytes())
	return h
}

// UpdateDirBlockInfoMap allows factom processor to update DirBlockInfo
// when a new Directory Block is saved to db
func UpdateDirBlockInfoMap(dirBlockInfo *common.DirBlockInfo) {
	anchorLog.Debug("UpdateDirBlockInfoMap: ", spew.Sdump(dirBlockInfo))
	dirBlockInfoMap[dirBlockInfo.DBMerkleRoot.String()] = dirBlockInfo
}

func checkForAnchor() {
	anchorLog.Debug("checkForAnchor: map.len=", len(dirBlockInfoMap))
	timeNow := time.Now().Unix()
	time0 := 60 * 60 * reAnchorAfter
	dirBlockInfos := make([]*common.DirBlockInfo, 0, len(dirBlockInfoMap))
	for _, v := range dirBlockInfoMap {
		dirBlockInfos = append(dirBlockInfos, v)
	}
	// anchor the latest dir block first
	sort.Sort(sort.Reverse(ByTimestamp(dirBlockInfos)))
	for _, dirBlockInfo := range dirBlockInfos {
		if bytes.Compare(dirBlockInfo.BTCTxHash.Bytes(), common.NewHash().Bytes()) == 0 {
			anchorLog.Debug("first time anchor: ", spew.Sdump(dirBlockInfo))
			SendRawTransactionToBTC(dirBlockInfo.DBMerkleRoot, dirBlockInfo.DBHeight)
		} else {
			// This is the re-anchor case for the missed callback or malleated tx,
			// that is, it has BTCTxHash and/or BTCBlockHash, BTCBlockHeight etc but
			// BTCConfirm is false.
			lapse := timeNow - dirBlockInfo.Timestamp
			if lapse > int64(time0) {
				anchorLog.Debugf("re-anchor: time lapse=%d, %s\n", lapse, spew.Sdump(dirBlockInfo))
				SendRawTransactionToBTC(dirBlockInfo.DBMerkleRoot, dirBlockInfo.DBHeight)
			}
		}
	}
}

func checkTxConfirmations() {
	timeNow := time.Now().Unix()
	time1 := 60 * 5 * confirmationsNeeded
	dirBlockInfos := make([]*common.DirBlockInfo, 0, len(dirBlockInfoMap))
	for _, v := range dirBlockInfoMap {
		if bytes.Compare(v.BTCTxHash.Bytes(), common.NewHash().Bytes()) != 0 {
			dirBlockInfos = append(dirBlockInfos, v)
		}
	}
	sort.Sort(ByTimestamp(dirBlockInfos))
	for _, dirBlockInfo := range dirBlockInfos {
		lapse := timeNow - dirBlockInfo.Timestamp
		if lapse > int64(time1) {
			anchorLog.Debugf("checkTxConfirmations: time lapse=%d", lapse)
			checkConfirmations(dirBlockInfo)
		}
	}
}

func checkConfirmations(dirBlockInfo *common.DirBlockInfo) error {
	anchorLog.Debug("check Confirmations for btc tx: ", toShaHash(dirBlockInfo.BTCTxHash).String())
	txResult, err := wclient.GetTransaction(toShaHash(dirBlockInfo.BTCTxHash))
	if err != nil {
		anchorLog.Debugf(err.Error())
		return err
	}
	anchorLog.Debugf("GetTransactionResult: %s\n", spew.Sdump(txResult))
	if txResult.Confirmations >= int64(confirmationsNeeded) {
		btcBlockHash, _ := wire.NewShaHashFromStr(txResult.BlockHash)
		var rewrite = false
		// Either the call back is not recorded in case of BTCBlockHash is zero hash,
		// or bad things like re-organization of btc main chain happened
		if bytes.Compare(dirBlockInfo.BTCBlockHash.Bytes(), btcBlockHash.Bytes()) != 0 {
			anchorLog.Debugf("BTCBlockHash changed: original BTCBlockHeight=%d, original BTCBlockHash=%s, original tx offset=%d\n", dirBlockInfo.BTCBlockHeight, toShaHash(dirBlockInfo.BTCBlockHash).String(), dirBlockInfo.BTCTxOffset)
			dirBlockInfo.BTCBlockHash = toHash(btcBlockHash)
			btcBlock, err := wclient.GetBlockVerbose(btcBlockHash, true)
			if err != nil {
				anchorLog.Debugf(err.Error())
			}
			if btcBlock.Height > 0 {
				dirBlockInfo.BTCBlockHeight = int32(btcBlock.Height)
			}
			anchorLog.Debugf("BTCBlockHash changed: new BTCBlockHeight=%d, new BTCBlockHash=%s, btcBlockVerbose.Height=%d\n", dirBlockInfo.BTCBlockHeight, btcBlockHash.String(), btcBlock.Height)
			rewrite = true
		}
		dirBlockInfo.BTCConfirmed = true // needs confirmationsNeeded (20) to be confirmed.
		dirBlockInfo.Timestamp = time.Now().Unix()
		db.InsertDirBlockInfo(dirBlockInfo)
		delete(dirBlockInfoMap, dirBlockInfo.DBMerkleRoot.String()) // delete it after confirmationsNeeded (20)
		anchorLog.Debugf("Fully confirmed %d times. txid=%s, dirblockInfo=%s\n", txResult.Confirmations, txResult.TxID, spew.Sdump(dirBlockInfo))
		if rewrite {
			anchorLog.Debug("rewrite to anchor chain: ", spew.Sdump(dirBlockInfo))
			saveToAnchorChain(dirBlockInfo)
		}
	}
	return nil
}

// ByTimestamp defines the methods needed to satisify sort.Interface to
// sort a slice of DirBlockInfo by their Timestamp.
type ByTimestamp []*common.DirBlockInfo

func (u ByTimestamp) Len() int { return len(u) }
func (u ByTimestamp) Less(i, j int) bool {
	if u[i].Timestamp == u[j].Timestamp {
		return u[i].DBHeight < u[j].DBHeight
	}
	return u[i].Timestamp < u[j].Timestamp
}
func (u ByTimestamp) Swap(i, j int) { u[i], u[j] = u[j], u[i] }

func checkTxMalleation(transaction *btcutil.Tx, details *btcjson.BlockDetails) {
	anchorLog.Debug("in checkTxMalleation")
	dirBlockInfos := make([]*common.DirBlockInfo, 0, len(dirBlockInfoMap))
	for _, v := range dirBlockInfoMap {
		// find those already anchored but no call back yet
		if v.BTCBlockHeight == 0 && bytes.Compare(v.BTCTxHash.Bytes(), common.NewHash().Bytes()) != 0 {
			dirBlockInfos = append(dirBlockInfos, v)
		}
	}
	sort.Sort(ByTimestamp(dirBlockInfos))
	anchorLog.Debugf("malleated tx candidate count=%d, dirBlockInfo list=%s\n", len(dirBlockInfos), spew.Sdump(dirBlockInfos))
	for _, dirBlockInfo := range dirBlockInfos {
		tx, err := wclient.GetRawTransaction(toShaHash(dirBlockInfo.BTCTxHash))
		if err != nil {
			anchorLog.Debugf(err.Error())
			continue
		}
		anchorLog.Debugf("GetRawTransaction=%s, dirBlockInfo=%s\n", spew.Sdump(tx), spew.Sdump(dirBlockInfo))
		// compare OP_RETURN
		if reflect.DeepEqual(transaction.MsgTx().TxOut[0], tx.MsgTx().TxOut[0]) {
			anchorLog.Debugf("Tx Malleated: original.txid=%s, malleated.txid=%s\n", dirBlockInfo.BTCTxHash.BTCString(), transaction.Sha().String())
			doSaveDirBlockInfo(transaction, details, dirBlockInfo, true)
			break
		}
	}
}
