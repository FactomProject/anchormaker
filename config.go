package main

import (
	"log"
	"os"
	"os/user"
	"strings"
	"sync"

	"github.com/FactomProject/FactomCode/common"

	"gopkg.in/gcfg.v1"
)

type anchorConfig struct {
	App struct {
		HomeDir                 string
		LdbPath                 string
		BoltDBPath              string
		DataStorePath           string
		DirectoryBlockInSeconds int
		NodeMode                string
		NodeID                  string
		Passphrase              string
		InitLeader              bool
		ServerPrivKey           string
		ExchangeRate            uint64
	}
	Anchor struct {
		ServerECKey         string
		AnchorChainID       string
		ConfirmationsNeeded int
	}
	Btc struct {
		BTCPubAddr         string
		SendToBTCinSeconds int
		WalletPassphrase   string
		CertHomePath       string
		RpcClientHost      string
		RpcClientEndpoint  string
		RpcClientUser      string
		RpcClientPass      string
		BtcTransFee        float64
		CertHomePathBtcd   string
		RpcBtcdHost        string
		RpcUser            string
		RpcPass            string
	}
	Rpc struct {
		PortNumber       int
		ApplicationName  string
		RefreshInSeconds int
	}
	Wsapi struct {
		PortNumber      int
		ApplicationName string
	}
	Log struct {
		LogPath  string
		LogLevel string
	}
	Wallet struct {
		Address          string
		Port             int
		DataFile         string
		RefreshInSeconds string
		BoltDBPath       string
		FactomdAddress   string
		FactomdPort      int
	}
	Controlpanel struct {
		Port string
	}

	//	AddPeers     []string `short:"a" long:"addpeer" description:"Add a peer to connect with at startup"`
	//	ConnectPeers []string `long:"connect" description:"Connect only to the specified peers at startup"`

	Proxy          string `long:"proxy" description:"Connect via SOCKS5 proxy (eg. 127.0.0.1:9050)"`
	DisableListen  bool   `long:"nolisten" description:"Disable listening for incoming connections -- NOTE: Listening is automatically disabled if the --connect or --proxy options are used without also specifying listen interfaces via --listen"`
	DisableRPC     bool   `long:"norpc" description:"Disable built-in RPC server -- NOTE: The RPC server is disabled by default if no rpcuser/rpcpass is specified"`
	DisableTLS     bool   `long:"notls" description:"Disable TLS for the RPC server -- NOTE: This is only allowed if the RPC server is bound to localhost"`
	DisableDNSSeed bool   `long:"nodnsseed" description:"Disable DNS seeding for peers"`
}

// defaultConfig
const defaultConfig = `
; ------------------------------------------------------------------------------
; App settings
; ------------------------------------------------------------------------------
[app]
HomeDir								= ""
LdbPath					        	= "ldb"
BoltDBPath							= ""
DataStorePath			      		= "data/export/"
DirectoryBlockInSeconds				= 60
; --------------- NodeMode: FULL | SERVER | LIGHT ----------------
NodeMode				        	= FULL
; NodeID is a hash hex string uniquely identifying this server and MUST be set for a federate server (NodeMode is SERVER)
NodeID								= "SERVER_DEFAULT"
Passphrase 							= "please change me"
; This server will start as the ONLY leader initially among federate servers if InitLead is true, and all other servers have to be set as false.
InitLeader							= "false"
ServerPrivKey			      		= 07c0d52cb74f4ca3106d80c4a70488426886bccc6ebc10c6bafb37bf8a65f4c38cee85c62a9e48039d4ac294da97943c2001be1539809ea5f54721f0c5477a0a
ExchangeRate                        = 00666600
[anchor]
ServerECKey							= 397c49e182caa97737c6b394591c614156fbe7998d7bf5d76273961e9fa1edd406ed9e69bfdf85db8aa69820f348d096985bc0b11cc9fc9dcee3b8c68b41dfd5
AnchorChainID						= df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604
ConfirmationsNeeded					= 20
[btc]
WalletPassphrase 	  				= "testNetPa55"
CertHomePath			  			= "btcwallet"
RpcClientHost			  			= "localhost:18332"
RpcClientEndpoint					= "ws"
RpcClientUser			  			= "testuser"
RpcClientPass 						= "notarychain"
BtcTransFee				  			= 0.0001
CertHomePathBtcd					= "btcd"
RpcBtcdHost 			  			= "localhost:18334"
RpcUser								=testuser
RpcPass								=notarychain
[wsapi]
ApplicationName						= "Factom/wsapi"
PortNumber				  			= 8088
; ------------------------------------------------------------------------------
; logLevel - allowed values are: debug, info, notice, warning, error, critical, alert, emergency and none
; ------------------------------------------------------------------------------
[log]
logLevel 							= debug
LogPath								= "factom-d.log"
; ------------------------------------------------------------------------------
; Configurations for fctwallet
; ------------------------------------------------------------------------------
[Wallet]
Address          					= localhost
Port             					= 8089
DataFile         					= fctwallet.dat
RefreshInSeconds 					= 60
BoltDBPath 							= ""
FactomdAddress                      = localhost
FactomdPort                         = 8088
; ------------------------------------------------------------------------------
; Configurations for controlpanel
; ------------------------------------------------------------------------------
[Controlpanel]
Port             					= 8090
`

//var acfg *anchorConfig
var once sync.Once
var filename = getHomeDir() + "/.factom/anchormaker.conf"

func SetConfigFile(f string) {
	filename = f
}

// GetConfig reads the default anchormaker.conf file and returns an anchorConfig
// object corresponding to the state of the file.
func ReadConfig() *anchorConfig {
	once.Do(func() {
		cfg = readAnchorConfig()
	})
	//debug.PrintStack()
	return cfg
}

func ReReadConfig() *anchorConfig {
	cfg = readAnchorConfig()

	return cfg
}

func readAnchorConfig() *anchorConfig {
	if len(os.Args) > 1 { //&& strings.Contains(strings.ToLower(os.Args[1]), "anchormaker.conf") {
		filename = os.Args[1]
	}
	if strings.HasPrefix(filename, "~") {
		filename = getHomeDir() + filename
	}
	cfg := new(anchorConfig)
	//log.Println("read factom config file: ", filename)

	err := gcfg.ReadFileInto(cfg, filename)
	if err != nil {
		log.Println("ERROR Reading config file!\nServer starting with default settings...\n", err)
		gcfg.ReadStringInto(cfg, defaultConfig)
	}

	//log.Println("nodeMode=", cfg.App.NodeMode, ", NodeID=", cfg.App.NodeID)
	// should have a version check here
	if cfg.App.NodeMode == common.SERVER_NODE && len(cfg.App.NodeID) == 0 {
		log.Println("ERROR!!! When running as a federate server (NodeMode is SERVER) in milestone II, NodeID must be set")
	}

	// Default to home directory if not set
	if len(cfg.App.HomeDir) < 1 {
		cfg.App.HomeDir = getHomeDir() + "/.factom/"
	}

	// TODO: improve the paths after milestone 1
	cfg.App.LdbPath = cfg.App.HomeDir + cfg.App.LdbPath
	cfg.App.BoltDBPath = cfg.App.HomeDir + cfg.App.BoltDBPath
	cfg.App.DataStorePath = cfg.App.HomeDir + cfg.App.DataStorePath
	cfg.Log.LogPath = cfg.App.HomeDir + cfg.Log.LogPath
	cfg.Wallet.BoltDBPath = cfg.App.HomeDir + cfg.Wallet.BoltDBPath

	return cfg
}

func getHomeDir() string {
	// Get the OS specific home directory via the Go standard lib.
	var homeDir string
	usr, err := user.Current()
	if err == nil {
		homeDir = usr.HomeDir
	}

	// Fall back to standard HOME environment variable that works
	// for most POSIX OSes if the directory from the Go standard
	// lib failed.
	if err != nil || homeDir == "" {
		homeDir = os.Getenv("HOME")
	}
	return homeDir
}
