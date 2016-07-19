package config

import (
	"log"
	"os"
	"os/user"
	"strings"
	"sync"

	"gopkg.in/gcfg.v1"
)

var cfg *AnchorConfig

type AnchorConfig struct {
	App struct {
		HomeDir       string
		LdbPath       string
		ServerPrivKey string
	}
	Anchor struct {
		ServerECKey         string
		AnchorChainID       string
		AnchorSigPublicKey  string
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
	Ethereum struct {
		WalletAddress   string
		ContractAddress string
		GasPrice        string
		ServerAddress   string
	}
	Log struct {
		LogPath  string
		LogLevel string
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
LdbPath								= "AnchormakerLDB"
ServerPrivKey						= 07c0d52cb74f4ca3106d80c4a70488426886bccc6ebc10c6bafb37bf8a65f4c38cee85c62a9e48039d4ac294da97943c2001be1539809ea5f54721f0c5477a0a
[anchor]
ServerECKey							= 397c49e182caa97737c6b394591c614156fbe7998d7bf5d76273961e9fa1edd406ed9e69bfdf85db8aa69820f348d096985bc0b11cc9fc9dcee3b8c68b41dfd5
AnchorChainID						= df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604
AnchorSigPublicKey					= 0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a
ConfirmationsNeeded					= 20
; ------------------------------------------------------------------------------
; Bitcoin settings
; ------------------------------------------------------------------------------
[btc]
WalletPassphrase 					= "testNetPa55"
CertHomePath						= "btcwallet"
RpcClientHost						= "localhost:18332"
RpcClientEndpoint					= "ws"
RpcClientUser						= "testuser"
RpcClientPass 						= "SecurePassHere"
BtcTransFee							= 0.0001
CertHomePathBtcd					= "btcd"
RpcBtcdHost 						= "localhost:18334"
RpcUser								= "testuser"
RpcPass								= "SecurePassHere"
; ------------------------------------------------------------------------------
; Ethereum settings
; ------------------------------------------------------------------------------
[ethereum]
WalletAddress						= "0x838f9b4d8ea3ff2f1bd87b13684f59c4c57a618b"
ContractAddress 					= "0x8a8fbabbec1e99148083e9314dffd82395dd8f18"
GasPrice							= "0x10FFFF"
ServerAddress						= "localhost:8545"
; ------------------------------------------------------------------------------
; logLevel - allowed values are: debug, info, notice, warning, error, critical, alert, emergency and none
; ------------------------------------------------------------------------------
[log]
logLevel 							= debug
LogPath								= "anchormaker.log"
`

//var acfg *AnchorConfig
var once sync.Once
var filename = getHomeDir() + "/.factom/anchormaker.conf"

func SetConfigFile(f string) {
	filename = f
}

// GetConfig reads the default anchormaker.conf file and returns an AnchorConfig
// object corresponding to the state of the file.
func ReadConfig() *AnchorConfig {
	once.Do(func() {
		cfg = readAnchorConfig()
	})
	//debug.PrintStack()
	return cfg
}

func ReReadConfig() *AnchorConfig {
	cfg = readAnchorConfig()

	return cfg
}

func readAnchorConfig() *AnchorConfig {
	if len(os.Args) > 1 { //&& strings.Contains(strings.ToLower(os.Args[1]), "anchormaker.conf") {
		filename = os.Args[1]
	}
	if strings.HasPrefix(filename, "~") {
		filename = getHomeDir() + filename
	}
	cfg := new(AnchorConfig)
	//log.Println("read anchormaker config file: ", filename)

	err := gcfg.ReadFileInto(cfg, filename)
	if err != nil {
		log.Println("ERROR Reading config file!\nServer starting with default settings...\n", err)
		gcfg.ReadStringInto(cfg, defaultConfig)
	}

	// Default to home directory if not set
	if len(cfg.App.HomeDir) < 1 {
		cfg.App.HomeDir = getHomeDir() + "/.factom/"
	}

	// TODO: improve the paths after milestone 1
	cfg.App.LdbPath = cfg.App.HomeDir + cfg.App.LdbPath
	cfg.Log.LogPath = cfg.App.HomeDir + cfg.Log.LogPath

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
