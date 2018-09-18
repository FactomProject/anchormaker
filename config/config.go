package config

import (
	"log"
	"os"
	"os/user"
	"strings"
	"sync"

	"github.com/FactomProject/factom"
	"gopkg.in/gcfg.v1"
)

var cfg *AnchorConfig

type AnchorConfig struct {
	App struct {
		HomeDir       string
		DBType        string
		LdbPath       string
		BoltPath      string
		ServerPrivKey string
	}
	Factom struct {
		FactomdAddress          string
		WalletAddress           string
		FactoidBalanceThreshold int64
		ECBalanceThreshold      int64
	}
	Anchor struct {
		ServerECKey         string
		AnchorSigPublicKey  []string
		ConfirmationsNeeded int
		WindowSize          uint32
	}
	Ethereum struct {
		WalletAddress        string
		WalletKeyPath        string
		WalletPassword       string
		ContractAddress      string
		GasLimit             string
		ServerAddress        string
		GethIPCURL           string
		EthGasStationAddress string
		IgnoreWrongEntries   bool
		TestNet              bool
		TestNetName          string
	}
	Log struct {
		LogPath  string
		LogLevel string
	}
	Walletd factom.RPCConfig

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
; --------------- DBType: LDB | Bolt | Map
DBType								= "Map"
LdbPath								= "AnchormakerLDB"
BoltPath							= "AnchormakerBolt.db"
;ServerPrivKey						= ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973
ServerPrivKey						= 2d9afb9b073394863786d660b8960aa827a3d713e0a400e116d373874429276a
; ServerPrivKey						= 75c67eb4637d8d0a7dba0ba8152bf1b96cba551f888878c7a5b7b8a34ac584e8f06f190d3307f52ff56e2ea6874250cb8ce0332dcc809b80100493b1ff064c59
; ServerPrivKey						= 07c0d52cb74f4ca3106d80c4a70488426886bccc6ebc10c6bafb37bf8a65f4c38cee85c62a9e48039d4ac294da97943c2001be1539809ea5f54721f0c5477a0a
[anchor]
;ServerECKey							= ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973
ServerECKey							= 2d9afb9b073394863786d660b8960aa827a3d713e0a400e116d373874429276a
; ServerECKey 						= 5c0eb59f5d311a1c80ba0302b53433457bdb9e271fc22f064e6981ac8965bc2f1f0a6c2bf854a0994562bf36606345aaa6a1dfee3073fb3276b878751238f762
; ServerECKey						= 397c49e182caa97737c6b394591c614156fbe7998d7bf5d76273961e9fa1edd406ed9e69bfdf85db8aa69820f348d096985bc0b11cc9fc9dcee3b8c68b41dfd5
AnchorSigPublicKey					= 0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a
ConfirmationsNeeded					= 20
WindowSize                          = 1000

; ------------------------------------------------------------------------------
; Factom settings
; ------------------------------------------------------------------------------
[factom]
;FactomdAddress						= "qatest.factom.org:8088"
FactomdAddress						= "localhost:8088"
WalletAddress						= "localhost:8089"
FactoidBalanceThreshold				= 100
ECBalanceThreshold					= 10000

; ------------------------------------------------------------------------------
; Ethereum settings
; ------------------------------------------------------------------------------
[ethereum]
WalletAddress						= "0x84964e1FfC60d0ad4DA803678b167c6A783A2E01"
WalletKeyPath						= ""
WalletPassword						= "password"
ContractAddress 					= "0xd1932fe27273e0dc1a2fa5257c75811fd5555a1d"
GasLimit							= "200000"
ServerAddress						= "localhost:8545"
GethIPCURL							= "/home/$USER/.ethereum/testnet/geth.ipc"
EthGasStationAddress				= "https://ethgasstation.info/json/ethgasAPI.json"
IgnoreWrongEntries					= true
TestNet								= true
TestNetName							= "ropsten"

; ------------------------------------------------------------------------------
; logLevel - allowed values are: debug, info, notice, warning, error, critical, alert, emergency and none
; ------------------------------------------------------------------------------
[log]
logLevel 							= debug
LogPath								= "anchormaker.log"

; ------------------------------------------------------------------------------
; Configurations for factom-walletd
; ------------------------------------------------------------------------------
[Walletd]
; These are the username and password that factom-walletd requires
; This file is also used by factom-cli to determine what login to use
WalletRPCUser                          = ""
WalletRPCPassword                      = ""

; These define if the connection to the wallet should be encrypted, and if it is, what files
; are the secret key and the public certificate.  factom-cli uses the certificate specified here if TLS is enabled.
; To use default files and paths leave /full/path/to/... in place.
WalletTLSEnable                      = false
WalletTLSKeyFile                     = "/full/path/to/walletAPIpriv.key"
WalletTLSCertFile                    = "/full/path/to/walletAPIpub.cert"

; This is where factom-walletd and factom-cli will find factomd to interact with the blockchain
; This value can also be updated to authorize an external ip or domain name when factomd creates a TLS cert
FactomdServer                        = "localhost:8088"

; This is where factom-cli will find factom-walletd to create Factoid and Entry Credit transactions
; This value can also be updated to authorize an external ip or domain name when factom-walletd creates a TLS cert
WalletServer                         = "localhost:8089"
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
	cfg.App.BoltPath = cfg.App.HomeDir + cfg.App.BoltPath
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
