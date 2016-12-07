package bitcoind

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var Address string
var Username string
var Password string
var AllowInvalidServerCertificate bool

func SetAddress(newAddress string, newUsername string, newPassword string) {
	Address = newAddress
	Username = newUsername
	Password = newPassword
}

var ID int64

func GetID() int64 {
	ID++
	return ID
}

type Result struct {
	Result json.RawMessage `json:"result"`
	Error  interface{}     `json:"error"`
	ID     interface{}     `json:"id"`
}

func (r *Result) String() string {
	s, _ := json.MarshalIndent(r, "", "\t")
	return string(s)
}

func (r *Result) ParseResult(dst interface{}) error {
	return json.Unmarshal(r.Result, dst)
}

func CallWithBasicAuth(method string, params []interface{}) (*Result, error) {
	fmt.Printf("CallWithBasicAuthSingleParam\n")
	data, err := json.Marshal(map[string]interface{}{
		"method": method,
		"id":     GetID(),
		"params": params,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", Address, strings.NewReader(string(data)))
	if err != nil {
		panic(err)
		return nil, err
	}
	req.SetBasicAuth(Username, Password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: AllowInvalidServerCertificate},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := new(Result)
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%v\n", result)
	return result, nil
}

func CallWithBasicAuthSingleParam(method string, params interface{}) (*Result, error) {
	fmt.Printf("CallWithBasicAuthSingleParam\n")
	data, err := json.Marshal(map[string]interface{}{
		"method": method,
		"id":     GetID(),
		"params": params,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", Address, strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(Username, Password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: AllowInvalidServerCertificate},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := new(Result)
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%v\n", result)
	return result, nil
}

/*
//https://en.bitcoin.it/wiki/Api#Full_list\
*/

func BackupWallet(destination []interface{}) (*Result, error) {
	//Safely copies wallet.dat to destination, which can be a directory or a path with filename.
	resp, err := CallWithBasicAuth("backupwallet", destination)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func CreateRawTransaction(data []interface{}) (*Result, error) {
	//version 0.7 Creates a raw transaction spending given inputs.
	resp, err := CallWithBasicAuth("createrawtransaction", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func EncryptWallet(passphrase []interface{}) (*Result, error) {
	//Encrypts the wallet with <passphrase>.
	resp, err := CallWithBasicAuth("encryptwallet", passphrase)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func GetAccount(bitcoinaddress []interface{}) (*Result, error) {
	//Returns the account associated with the given address.
	resp, err := CallWithBasicAuth("getaccount", bitcoinaddress)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func GetAccountAddress(account []interface{}) (*Result, error) {
	//Returns the current bitcoin address for receiving payments to this account.
	resp, err := CallWithBasicAuth("getaccountaddress", account)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}
func GetAddressesByAccount(account []interface{}) (*Result, error) {
	//Returns the list of addresses for the given account.
	resp, err := CallWithBasicAuth("getaddressesbyaccount", account)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func GetBalance(data []interface{}) (*Result, error) {
	//If [account] is not specified, returns the server's total available balance.
	//If [account] is specified, returns the balance in the account.
	resp, err := CallWithBasicAuth("getbalance", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

/*func GetBlockByCount(height []interface{})(map[string]interface{}, error){
	//Dumps the block existing at specified height. Note: this is not available in the official release
	resp, err:=CallWithBasicAuth("getblockbycount", height)
	if err!=nil{
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}*/

func GetBlockCount() (*Result, error) {
	//Returns the number of blocks in the longest block chain.
	resp, err := CallWithBasicAuth("getblockcount", nil)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func GetBlockHash(index []interface{}) (*Result, error) {
	//Returns hash of block in best-block-chain at <index>; index 0 is the genesis block
	resp, err := CallWithBasicAuth("getblockhash", index)
	if err != nil {
		return resp, err
	}
	return resp, err
}

func GetBlockNumber() (*Result, error) {
	//Returns the block number of the latest block in the longest block chain.
	resp, err := CallWithBasicAuth("getblocknumber", nil)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}
func GetConnectionCount() (*Result, error) {
	//Returns the number of connections to other nodes.
	resp, err := CallWithBasicAuth("getconnectioncount", nil)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}
func GetDifficulty() (*Result, error) {
	//Returns the proof-of-work difficulty as a multiple of the minimum difficulty.
	resp, err := CallWithBasicAuth("getdifficulty", nil)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}
func GetGenerate() (*Result, error) {
	//Returns true or false whether bitcoind is currently generating hashes
	resp, err := CallWithBasicAuth("getgenerate", nil)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}
func GetHashesPerSec() (*Result, error) {
	//Returns a recent hashes per second performance measurement while generating.
	resp, err := CallWithBasicAuth("gethashespersec", nil)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func GetMemoryPool(data []interface{}) (*Result, error) {
	//If [data] is not specified, returns data needed to construct a block to work on:
	//"version" : block version
	//"previousblockhash" : hash of current highest block
	//"transactions" : contents of non-coinbase transactions that should be included in the next block
	//"coinbasevalue" : maximum allowable input to coinbase transaction, including the generation award and transaction fees
	//"time" : timestamp appropriate for next block
	//"bits" : compressed target of next block
	//If [data] is specified, tries to solve the block and returns true if it was successful.
	resp, err := CallWithBasicAuth("getmemorypool", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

type GetInfoResult struct {
	Version         int64   `json:"version"`
	ProtocolVersion int64   `json:"protocolversion"`
	WalletVersion   int64   `json:"walletversion"`
	Balance         float64 `json:"balance"`
	Blocks          int64   `json:"blocks"`
	TimeOffset      int64   `json:"timeoffset"`
	Connections     int64   `json:"connections"`
	Proxy           string  `json:"proxy"`
	Difficulty      float64 `json:"difficulty"`
	Testnet         bool    `json:"testnet"`
	KeyPoolOldest   int64   `json:"keypoololdest"`
	KeyPoolSize     int64   `json:"keypoolsize"`
	UnlockedUntil   int64   `json:"unlocked_until"`
	PayTxFee        float64 `json:"paytxfee"`
	RelayFee        float64 `json:"relayfee"`
	Errors          string  `json:"errors"`
}

func GetInfo() (*GetInfoResult, *Result, error) {
	//Returns an object containing various state info.
	resp, err := CallWithBasicAuth("getinfo", nil)
	if err != nil {
		return nil, nil, err
	}
	//result:=resp["result"]
	//c.Infof(result)
	if resp.Error != nil {
		return nil, nil, err
	}
	answer := new(GetInfoResult)
	err = resp.ParseResult(answer)
	if err != nil {
		return nil, nil, err
	}

	return answer, resp, err
}

func GetNewAddress(account []interface{}) (*Result, error) {
	//Returns a new bitcoin address for receiving payments. If [account] is specified (recommended), it is added to the address book so payments received with the address will be credited to [account].
	resp, err := CallWithBasicAuth("getnewaddress", account)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func GetReceivedByAccount(data []interface{}) (*Result, error) {
	//Returns the total amount received by addresses with [account] in transactions with at least [minconf] confirmations. If [account] not provided return will include all transactions to all accounts. (version 0.3.24-beta)
	resp, err := CallWithBasicAuth("getreceivedbyaccount", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func GetReceivedByAddress(data []interface{}) (*Result, error) {
	//Returns the total amount received by <bitcoinaddress> in transactions with at least [minconf] confirmations. While some might consider this obvious, value reported by this only considers *receiving* transactions. It does not check payments that have been made *from* this address. In other words, this is not "getaddressbalance".
	resp, err := CallWithBasicAuth("getreceivedbyaddress", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func GetBlock(txid []interface{}) (*Result, error) {
	//Returns information about the block with the given hash.
	resp, err := CallWithBasicAuth("getblock", txid)
	if err != nil {
		return resp, err
	}
	return resp, err
}

func GetTransaction(txid []interface{}) (*Result, error) {
	//Returns an object about the given transaction containing:
	//"amount" : total amount of the transaction
	//"confirmations" : number of confirmations of the transaction
	//"txid" : the transaction ID
	//"time" : time the transaction occurred
	//"details" - An array of objects containing:
	//"account"
	//"address"
	//"category"
	//"amount"
	resp, err := CallWithBasicAuth("gettransaction", txid)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func GetRawTransaction(txid string) (*Result, error) {
	resp, err := CallWithBasicAuth("getrawtransaction", []interface{}{txid})
	if err != nil {
		return resp, err
	}
	return resp, err
}

func GetWork(data []interface{}) (*Result, error) {
	//If [data] is not specified, returns formatted hash data to work on:
	//"midstate" : precomputed hash state after hashing the first half of the data
	//"data" : block data
	//"hash1" : formatted hash buffer for second hash
	//"target" : little endian hash target
	//If [data] is specified, tries to solve the block and returns true if it was successful.
	resp, err := CallWithBasicAuth("getwork", data)
	if err != nil {
		return resp, err
	}
	////result:=resp["result"]
	////c.Infof(result)

	return resp, err
}

func Help(command string) (*Result, error) {
	//List commands, or get help for a command.
	resp, err := CallWithBasicAuth("help", []interface{}{command})
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func KeyPoolRefill() (*Result, error) {
	//Fills the keypool, requires wallet passphrase to be set.
	resp, err := CallWithBasicAuth("keypoolrefill", nil)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func ListAccounts(minconf interface{}) (*Result, error) {
	//Returns Object that has account names as keys, account balances as values.
	resp, err := CallWithBasicAuthSingleParam("listaccounts", minconf)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func ListSinceBlock(blockid, targetconfirmations interface{}) (*Result, error) {
	//Get all transactions in blocks since block [blockid], or all transactions if omitted.
	resp, err := CallWithBasicAuth("listsinceblock", []interface{}{blockid, targetconfirmations})
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func ListReceivedByAccount(data []interface{}) (*Result, error) {
	//Returns an array of objects containing:
	//"account" : the account of the receiving addresses
	//"amount" : total amount received by addresses with this account
	//"confirmations" : number of confirmations of the most recent transaction included
	resp, err := CallWithBasicAuth("listreceivedbyaccount", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func ListReceivedByAddress(data []interface{}) (*Result, error) {
	//Returns an array of objects containing:
	//"address" : receiving address
	//"account" : the account of the receiving address
	//"amount" : total amount received by the address
	//"confirmations" : number of confirmations of the most recent transaction included
	//To get a list of accounts on the system, execute bitcoind listreceivedbyaddress 0 true
	resp, err := CallWithBasicAuth("listreceivedbyaddress", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

type Transaction struct {
	Account           string        `json:"account"`
	Address           string        `json:"address"`
	Category          string        `json:"category"`
	Amount            float64       `json:"amount"`
	Label             string        `json:"label"`
	Vout              int64         `json:"vout"`
	Fee               float64       `json:"fee"`
	Confirmations     int64         `json:"confirmations"`
	Blockhash         string        `json:"blockhash"`
	Blockindex        int64         `json:"blockindex"`
	Blocktime         int64         `json:"blocktime"`
	Txid              string        `json:"txid"`
	Walletconflicts   []interface{} `json:"walletconflicts"`
	Time              int64         `json:"time"`
	Timereceived      int64         `json:"timereceived"`
	BIP125Replaceable string        `json:"bip125-replaceable"`
	Abandoned         bool          `json:"abandoned"`
}

func ListTransactions(data []interface{}) ([]Transaction, *Result, error) {
	//Returns up to [count] most recent transactions skipping the first [from] transactions for account [account]. If [account] not provided will return recent transaction from all accounts.
	resp, err := CallWithBasicAuth("listtransactions", data)
	if err != nil {
		return nil, nil, err
	}
	//result:=resp["result"]
	//c.Infof(result)
	if resp.Error != nil {
		return nil, nil, err
	}
	answer := []Transaction{}
	err = resp.ParseResult(&answer)
	if err != nil {
		return nil, nil, err
	}

	return answer, resp, err
}

func ListUnspent(data []interface{}) (*Result, error) {
	//Returns up to [count] most recent transactions skipping the first [from] transactions for account [account]. If [account] not provided will return recent transaction from all accounts.
	resp, err := CallWithBasicAuth("listunspent", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func Move(data []interface{}) (*Result, error) {
	//Move from one account in your wallet to another
	resp, err := CallWithBasicAuth("move", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func SendFrom(data []interface{}) (*Result, error) {
	//<amount> is a real and is rounded to 8 decimal places. Will send the given amount to the given address, ensuring the account has a valid balance using [minconf] confirmations. Returns the transaction ID if successful (not in JSON object).
	resp, err := CallWithBasicAuth("sendfrom", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func SendMany(data []interface{}) (*Result, error) {
	//amounts are double-precision floating point numbers
	resp, err := CallWithBasicAuth("sendmany", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func SendToAddress(data []interface{}) (*Result, error) {
	//<amount> is a real and is rounded to 8 decimal places. Returns the transaction ID <txid> if successful.
	resp, err := CallWithBasicAuth("sendtoaddress", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func SetAccount(data []interface{}) (*Result, error) {
	//Sets the account associated with the given address. Assigning address that is already assigned to the same account will create a new address associated with that account.
	resp, err := CallWithBasicAuth("setaccount", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func SetGenerate(data []interface{}) (*Result, error) {
	//<generate> is true or false to turn generation on or off.
	//Generation is limited to [genproclimit] processors, -1 is unlimited.
	resp, err := CallWithBasicAuth("setgenerate", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func SetTxFee(amount []interface{}) (*Result, error) {
	//<amount> is a real and is rounded to the nearest 0.00000001
	resp, err := CallWithBasicAuth("settxfee", amount)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func SignMessage(bitcoinaddress, message interface{}) (*Result, error) {
	//Sign a message with the private key of an address.
	resp, err := CallWithBasicAuth("signmessage", []interface{}{bitcoinaddress, message})
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}
func Stop() (*Result, error) {
	//Stop bitcoin server.
	resp, err := CallWithBasicAuth("stop", nil)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func ValidateAddress(bitcoinaddress interface{}) (*Result, error) {
	//Return information about <bitcoinaddress>.
	resp, err := CallWithBasicAuth("validateaddress", []interface{}{bitcoinaddress})
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func VerifyMessage(bitcoinaddress, signature, message interface{}) (*Result, error) {
	//Verify a signed message.
	resp, err := CallWithBasicAuth("verifymessage", []interface{}{bitcoinaddress, signature, message})
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func WalletLock() (*Result, error) {
	//Removes the wallet encryption key from memory, locking the wallet. After calling this method, you will need to call walletpassphrase again before being able to call any methods which require the wallet to be unlocked.
	resp, err := CallWithBasicAuth("walletlock", nil)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func WalletPassPhrase(passphrase, timeout interface{}) (*Result, error) {
	//Stores the wallet decryption key in memory for <timeout> seconds.
	resp, err := CallWithBasicAuth("walletpassphrase", []interface{}{passphrase, timeout})
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}

func WalletPassPhraseChange(data []interface{}) (*Result, error) {
	//Changes the wallet passphrase from <oldpassphrase> to <newpassphrase>.
	resp, err := CallWithBasicAuth("walletpassphrasechange", data)
	if err != nil {
		return resp, err
	}
	//result:=resp["result"]
	//c.Infof(result)

	return resp, err
}
