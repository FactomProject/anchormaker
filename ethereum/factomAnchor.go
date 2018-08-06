// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ethereum

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// FactomAnchorABI is the input ABI used to generate the binding from.
const FactomAnchorABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"creator\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"frozen\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getHeight\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"anchors\",\"outputs\":[{\"name\":\"KeyMR\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"getAnchor\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"freeze\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"maxHeight\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newHeight\",\"type\":\"uint256\"}],\"name\":\"setHeight\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"checkFrozen\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"name\":\"keyMR\",\"type\":\"uint256\"}],\"name\":\"setAnchor\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"height\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"merkleroot\",\"type\":\"uint256\"}],\"name\":\"AnchorMade\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"height\",\"type\":\"uint256\"}],\"name\":\"AnchorHeightSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"height\",\"type\":\"uint256\"}],\"name\":\"AnchoringFrozen\",\"type\":\"event\"}]"

// FactomAnchorBin is the compiled bytecode used for deploying new contracts.
const FactomAnchorBin = `0x608060405234801561001057600080fd5b5060008054600160a060020a031916331781556002556003805460ff191690556103fb8061003f6000396000f3006080604052600436106100a35763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166302d05d3f81146100a8578063054f7d9c146100e657806319efb11d1461010f578063368b733e146101365780634c7df18f1461014e57806362a5af3b1461016657806363aa5ce21461017d578063741a39441461019257806398fb22fd146101aa578063bbcc0c80146101bf575b600080fd5b3480156100b457600080fd5b506100bd6101da565b6040805173ffffffffffffffffffffffffffffffffffffffff9092168252519081900360200190f35b3480156100f257600080fd5b506100fb6101f6565b604080519115158252519081900360200190f35b34801561011b57600080fd5b506101246101ff565b60408051918252519081900360200190f35b34801561014257600080fd5b50610124600435610205565b34801561015a57600080fd5b50610124600435610217565b34801561017257600080fd5b5061017b610229565b005b34801561018957600080fd5b50610124610291565b34801561019e57600080fd5b5061017b600435610297565b3480156101b657600080fd5b506100fb610303565b3480156101cb57600080fd5b5061017b60043560243561030c565b60005473ffffffffffffffffffffffffffffffffffffffff1681565b60035460ff1681565b60025490565b60016020526000908152604090205481565b60009081526001602052604090205490565b60005473ffffffffffffffffffffffffffffffffffffffff16331461024d57600080fd5b6003805460ff1916600117905560025460408051918252517f02392dea61af8262e6609d1b99522854b729caa208dbccef7fd70f9508293aa79181900360200190a1565b60025481565b60005473ffffffffffffffffffffffffffffffffffffffff1633146102bb57600080fd5b60035460ff1615156103005760028190556040805182815290517f650ab55358f2f485c020da127018db860ee6b798f45e30ba3fc96fb1b8b8edbd9181900360200190a15b50565b60035460ff1690565b60005473ffffffffffffffffffffffffffffffffffffffff16331461033057600080fd5b60035460ff1615156103cb576000828152600160205260409020819055600254821061038f5760028290556040805183815290517f650ab55358f2f485c020da127018db860ee6b798f45e30ba3fc96fb1b8b8edbd9181900360200190a15b604080518381526020810183905281517f1c6a33c0de150a46e5647f4482e93d30f3a966487eb86e762d69319ab9a6e6b6929181900390910190a15b50505600a165627a7a72305820f3859c76c55784cdeaedb43aba7f656d1c5107a19dc6c02325ca5cc6396383ae0029`

// DeployFactomAnchor deploys a new Ethereum contract, binding an instance of FactomAnchor to it.
func DeployFactomAnchor(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *FactomAnchor, error) {
	parsed, err := abi.JSON(strings.NewReader(FactomAnchorABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(FactomAnchorBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &FactomAnchor{FactomAnchorCaller: FactomAnchorCaller{contract: contract}, FactomAnchorTransactor: FactomAnchorTransactor{contract: contract}, FactomAnchorFilterer: FactomAnchorFilterer{contract: contract}}, nil
}

// FactomAnchor is an auto generated Go binding around an Ethereum contract.
type FactomAnchor struct {
	FactomAnchorCaller     // Read-only binding to the contract
	FactomAnchorTransactor // Write-only binding to the contract
	FactomAnchorFilterer   // Log filterer for contract events
}

// FactomAnchorCaller is an auto generated read-only Go binding around an Ethereum contract.
type FactomAnchorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FactomAnchorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FactomAnchorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FactomAnchorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FactomAnchorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FactomAnchorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FactomAnchorSession struct {
	Contract     *FactomAnchor     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FactomAnchorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FactomAnchorCallerSession struct {
	Contract *FactomAnchorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// FactomAnchorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FactomAnchorTransactorSession struct {
	Contract     *FactomAnchorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// FactomAnchorRaw is an auto generated low-level Go binding around an Ethereum contract.
type FactomAnchorRaw struct {
	Contract *FactomAnchor // Generic contract binding to access the raw methods on
}

// FactomAnchorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FactomAnchorCallerRaw struct {
	Contract *FactomAnchorCaller // Generic read-only contract binding to access the raw methods on
}

// FactomAnchorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FactomAnchorTransactorRaw struct {
	Contract *FactomAnchorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFactomAnchor creates a new instance of FactomAnchor, bound to a specific deployed contract.
func NewFactomAnchor(address common.Address, backend bind.ContractBackend) (*FactomAnchor, error) {
	contract, err := bindFactomAnchor(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &FactomAnchor{FactomAnchorCaller: FactomAnchorCaller{contract: contract}, FactomAnchorTransactor: FactomAnchorTransactor{contract: contract}, FactomAnchorFilterer: FactomAnchorFilterer{contract: contract}}, nil
}

// NewFactomAnchorCaller creates a new read-only instance of FactomAnchor, bound to a specific deployed contract.
func NewFactomAnchorCaller(address common.Address, caller bind.ContractCaller) (*FactomAnchorCaller, error) {
	contract, err := bindFactomAnchor(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FactomAnchorCaller{contract: contract}, nil
}

// NewFactomAnchorTransactor creates a new write-only instance of FactomAnchor, bound to a specific deployed contract.
func NewFactomAnchorTransactor(address common.Address, transactor bind.ContractTransactor) (*FactomAnchorTransactor, error) {
	contract, err := bindFactomAnchor(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FactomAnchorTransactor{contract: contract}, nil
}

// NewFactomAnchorFilterer creates a new log filterer instance of FactomAnchor, bound to a specific deployed contract.
func NewFactomAnchorFilterer(address common.Address, filterer bind.ContractFilterer) (*FactomAnchorFilterer, error) {
	contract, err := bindFactomAnchor(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FactomAnchorFilterer{contract: contract}, nil
}

// bindFactomAnchor binds a generic wrapper to an already deployed contract.
func bindFactomAnchor(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(FactomAnchorABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FactomAnchor *FactomAnchorRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _FactomAnchor.Contract.FactomAnchorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FactomAnchor *FactomAnchorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FactomAnchor.Contract.FactomAnchorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FactomAnchor *FactomAnchorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FactomAnchor.Contract.FactomAnchorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FactomAnchor *FactomAnchorCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _FactomAnchor.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FactomAnchor *FactomAnchorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FactomAnchor.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FactomAnchor *FactomAnchorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FactomAnchor.Contract.contract.Transact(opts, method, params...)
}

// Anchors is a free data retrieval call binding the contract method 0x368b733e.
//
// Solidity: function anchors( uint256) constant returns(KeyMR uint256)
func (_FactomAnchor *FactomAnchorCaller) Anchors(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _FactomAnchor.contract.Call(opts, out, "anchors", arg0)
	return *ret0, err
}

// Anchors is a free data retrieval call binding the contract method 0x368b733e.
//
// Solidity: function anchors( uint256) constant returns(KeyMR uint256)
func (_FactomAnchor *FactomAnchorSession) Anchors(arg0 *big.Int) (*big.Int, error) {
	return _FactomAnchor.Contract.Anchors(&_FactomAnchor.CallOpts, arg0)
}

// Anchors is a free data retrieval call binding the contract method 0x368b733e.
//
// Solidity: function anchors( uint256) constant returns(KeyMR uint256)
func (_FactomAnchor *FactomAnchorCallerSession) Anchors(arg0 *big.Int) (*big.Int, error) {
	return _FactomAnchor.Contract.Anchors(&_FactomAnchor.CallOpts, arg0)
}

// CheckFrozen is a free data retrieval call binding the contract method 0x98fb22fd.
//
// Solidity: function checkFrozen() constant returns(bool)
func (_FactomAnchor *FactomAnchorCaller) CheckFrozen(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _FactomAnchor.contract.Call(opts, out, "checkFrozen")
	return *ret0, err
}

// CheckFrozen is a free data retrieval call binding the contract method 0x98fb22fd.
//
// Solidity: function checkFrozen() constant returns(bool)
func (_FactomAnchor *FactomAnchorSession) CheckFrozen() (bool, error) {
	return _FactomAnchor.Contract.CheckFrozen(&_FactomAnchor.CallOpts)
}

// CheckFrozen is a free data retrieval call binding the contract method 0x98fb22fd.
//
// Solidity: function checkFrozen() constant returns(bool)
func (_FactomAnchor *FactomAnchorCallerSession) CheckFrozen() (bool, error) {
	return _FactomAnchor.Contract.CheckFrozen(&_FactomAnchor.CallOpts)
}

// Creator is a free data retrieval call binding the contract method 0x02d05d3f.
//
// Solidity: function creator() constant returns(address)
func (_FactomAnchor *FactomAnchorCaller) Creator(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _FactomAnchor.contract.Call(opts, out, "creator")
	return *ret0, err
}

// Creator is a free data retrieval call binding the contract method 0x02d05d3f.
//
// Solidity: function creator() constant returns(address)
func (_FactomAnchor *FactomAnchorSession) Creator() (common.Address, error) {
	return _FactomAnchor.Contract.Creator(&_FactomAnchor.CallOpts)
}

// Creator is a free data retrieval call binding the contract method 0x02d05d3f.
//
// Solidity: function creator() constant returns(address)
func (_FactomAnchor *FactomAnchorCallerSession) Creator() (common.Address, error) {
	return _FactomAnchor.Contract.Creator(&_FactomAnchor.CallOpts)
}

// Frozen is a free data retrieval call binding the contract method 0x054f7d9c.
//
// Solidity: function frozen() constant returns(bool)
func (_FactomAnchor *FactomAnchorCaller) Frozen(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _FactomAnchor.contract.Call(opts, out, "frozen")
	return *ret0, err
}

// Frozen is a free data retrieval call binding the contract method 0x054f7d9c.
//
// Solidity: function frozen() constant returns(bool)
func (_FactomAnchor *FactomAnchorSession) Frozen() (bool, error) {
	return _FactomAnchor.Contract.Frozen(&_FactomAnchor.CallOpts)
}

// Frozen is a free data retrieval call binding the contract method 0x054f7d9c.
//
// Solidity: function frozen() constant returns(bool)
func (_FactomAnchor *FactomAnchorCallerSession) Frozen() (bool, error) {
	return _FactomAnchor.Contract.Frozen(&_FactomAnchor.CallOpts)
}

// GetAnchor is a free data retrieval call binding the contract method 0x4c7df18f.
//
// Solidity: function getAnchor(blockNumber uint256) constant returns(uint256)
func (_FactomAnchor *FactomAnchorCaller) GetAnchor(opts *bind.CallOpts, blockNumber *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _FactomAnchor.contract.Call(opts, out, "getAnchor", blockNumber)
	return *ret0, err
}

// GetAnchor is a free data retrieval call binding the contract method 0x4c7df18f.
//
// Solidity: function getAnchor(blockNumber uint256) constant returns(uint256)
func (_FactomAnchor *FactomAnchorSession) GetAnchor(blockNumber *big.Int) (*big.Int, error) {
	return _FactomAnchor.Contract.GetAnchor(&_FactomAnchor.CallOpts, blockNumber)
}

// GetAnchor is a free data retrieval call binding the contract method 0x4c7df18f.
//
// Solidity: function getAnchor(blockNumber uint256) constant returns(uint256)
func (_FactomAnchor *FactomAnchorCallerSession) GetAnchor(blockNumber *big.Int) (*big.Int, error) {
	return _FactomAnchor.Contract.GetAnchor(&_FactomAnchor.CallOpts, blockNumber)
}

// GetHeight is a free data retrieval call binding the contract method 0x19efb11d.
//
// Solidity: function getHeight() constant returns(uint256)
func (_FactomAnchor *FactomAnchorCaller) GetHeight(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _FactomAnchor.contract.Call(opts, out, "getHeight")
	return *ret0, err
}

// GetHeight is a free data retrieval call binding the contract method 0x19efb11d.
//
// Solidity: function getHeight() constant returns(uint256)
func (_FactomAnchor *FactomAnchorSession) GetHeight() (*big.Int, error) {
	return _FactomAnchor.Contract.GetHeight(&_FactomAnchor.CallOpts)
}

// GetHeight is a free data retrieval call binding the contract method 0x19efb11d.
//
// Solidity: function getHeight() constant returns(uint256)
func (_FactomAnchor *FactomAnchorCallerSession) GetHeight() (*big.Int, error) {
	return _FactomAnchor.Contract.GetHeight(&_FactomAnchor.CallOpts)
}

// MaxHeight is a free data retrieval call binding the contract method 0x63aa5ce2.
//
// Solidity: function maxHeight() constant returns(uint256)
func (_FactomAnchor *FactomAnchorCaller) MaxHeight(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _FactomAnchor.contract.Call(opts, out, "maxHeight")
	return *ret0, err
}

// MaxHeight is a free data retrieval call binding the contract method 0x63aa5ce2.
//
// Solidity: function maxHeight() constant returns(uint256)
func (_FactomAnchor *FactomAnchorSession) MaxHeight() (*big.Int, error) {
	return _FactomAnchor.Contract.MaxHeight(&_FactomAnchor.CallOpts)
}

// MaxHeight is a free data retrieval call binding the contract method 0x63aa5ce2.
//
// Solidity: function maxHeight() constant returns(uint256)
func (_FactomAnchor *FactomAnchorCallerSession) MaxHeight() (*big.Int, error) {
	return _FactomAnchor.Contract.MaxHeight(&_FactomAnchor.CallOpts)
}

// Freeze is a paid mutator transaction binding the contract method 0x62a5af3b.
//
// Solidity: function freeze() returns()
func (_FactomAnchor *FactomAnchorTransactor) Freeze(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FactomAnchor.contract.Transact(opts, "freeze")
}

// Freeze is a paid mutator transaction binding the contract method 0x62a5af3b.
//
// Solidity: function freeze() returns()
func (_FactomAnchor *FactomAnchorSession) Freeze() (*types.Transaction, error) {
	return _FactomAnchor.Contract.Freeze(&_FactomAnchor.TransactOpts)
}

// Freeze is a paid mutator transaction binding the contract method 0x62a5af3b.
//
// Solidity: function freeze() returns()
func (_FactomAnchor *FactomAnchorTransactorSession) Freeze() (*types.Transaction, error) {
	return _FactomAnchor.Contract.Freeze(&_FactomAnchor.TransactOpts)
}

// SetAnchor is a paid mutator transaction binding the contract method 0xbbcc0c80.
//
// Solidity: function setAnchor(blockNumber uint256, keyMR uint256) returns()
func (_FactomAnchor *FactomAnchorTransactor) SetAnchor(opts *bind.TransactOpts, blockNumber *big.Int, keyMR *big.Int) (*types.Transaction, error) {
	return _FactomAnchor.contract.Transact(opts, "setAnchor", blockNumber, keyMR)
}

// SetAnchor is a paid mutator transaction binding the contract method 0xbbcc0c80.
//
// Solidity: function setAnchor(blockNumber uint256, keyMR uint256) returns()
func (_FactomAnchor *FactomAnchorSession) SetAnchor(blockNumber *big.Int, keyMR *big.Int) (*types.Transaction, error) {
	return _FactomAnchor.Contract.SetAnchor(&_FactomAnchor.TransactOpts, blockNumber, keyMR)
}

// SetAnchor is a paid mutator transaction binding the contract method 0xbbcc0c80.
//
// Solidity: function setAnchor(blockNumber uint256, keyMR uint256) returns()
func (_FactomAnchor *FactomAnchorTransactorSession) SetAnchor(blockNumber *big.Int, keyMR *big.Int) (*types.Transaction, error) {
	return _FactomAnchor.Contract.SetAnchor(&_FactomAnchor.TransactOpts, blockNumber, keyMR)
}

// SetHeight is a paid mutator transaction binding the contract method 0x741a3944.
//
// Solidity: function setHeight(newHeight uint256) returns()
func (_FactomAnchor *FactomAnchorTransactor) SetHeight(opts *bind.TransactOpts, newHeight *big.Int) (*types.Transaction, error) {
	return _FactomAnchor.contract.Transact(opts, "setHeight", newHeight)
}

// SetHeight is a paid mutator transaction binding the contract method 0x741a3944.
//
// Solidity: function setHeight(newHeight uint256) returns()
func (_FactomAnchor *FactomAnchorSession) SetHeight(newHeight *big.Int) (*types.Transaction, error) {
	return _FactomAnchor.Contract.SetHeight(&_FactomAnchor.TransactOpts, newHeight)
}

// SetHeight is a paid mutator transaction binding the contract method 0x741a3944.
//
// Solidity: function setHeight(newHeight uint256) returns()
func (_FactomAnchor *FactomAnchorTransactorSession) SetHeight(newHeight *big.Int) (*types.Transaction, error) {
	return _FactomAnchor.Contract.SetHeight(&_FactomAnchor.TransactOpts, newHeight)
}

// FactomAnchorAnchorHeightSetIterator is returned from FilterAnchorHeightSet and is used to iterate over the raw logs and unpacked data for AnchorHeightSet events raised by the FactomAnchor contract.
type FactomAnchorAnchorHeightSetIterator struct {
	Event *FactomAnchorAnchorHeightSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FactomAnchorAnchorHeightSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FactomAnchorAnchorHeightSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FactomAnchorAnchorHeightSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FactomAnchorAnchorHeightSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FactomAnchorAnchorHeightSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FactomAnchorAnchorHeightSet represents a AnchorHeightSet event raised by the FactomAnchor contract.
type FactomAnchorAnchorHeightSet struct {
	Height *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterAnchorHeightSet is a free log retrieval operation binding the contract event 0x650ab55358f2f485c020da127018db860ee6b798f45e30ba3fc96fb1b8b8edbd.
//
// Solidity: e AnchorHeightSet(height uint256)
func (_FactomAnchor *FactomAnchorFilterer) FilterAnchorHeightSet(opts *bind.FilterOpts) (*FactomAnchorAnchorHeightSetIterator, error) {

	logs, sub, err := _FactomAnchor.contract.FilterLogs(opts, "AnchorHeightSet")
	if err != nil {
		return nil, err
	}
	return &FactomAnchorAnchorHeightSetIterator{contract: _FactomAnchor.contract, event: "AnchorHeightSet", logs: logs, sub: sub}, nil
}

// WatchAnchorHeightSet is a free log subscription operation binding the contract event 0x650ab55358f2f485c020da127018db860ee6b798f45e30ba3fc96fb1b8b8edbd.
//
// Solidity: e AnchorHeightSet(height uint256)
func (_FactomAnchor *FactomAnchorFilterer) WatchAnchorHeightSet(opts *bind.WatchOpts, sink chan<- *FactomAnchorAnchorHeightSet) (event.Subscription, error) {

	logs, sub, err := _FactomAnchor.contract.WatchLogs(opts, "AnchorHeightSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FactomAnchorAnchorHeightSet)
				if err := _FactomAnchor.contract.UnpackLog(event, "AnchorHeightSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// FactomAnchorAnchorMadeIterator is returned from FilterAnchorMade and is used to iterate over the raw logs and unpacked data for AnchorMade events raised by the FactomAnchor contract.
type FactomAnchorAnchorMadeIterator struct {
	Event *FactomAnchorAnchorMade // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FactomAnchorAnchorMadeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FactomAnchorAnchorMade)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FactomAnchorAnchorMade)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FactomAnchorAnchorMadeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FactomAnchorAnchorMadeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FactomAnchorAnchorMade represents a AnchorMade event raised by the FactomAnchor contract.
type FactomAnchorAnchorMade struct {
	Height     *big.Int
	Merkleroot *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterAnchorMade is a free log retrieval operation binding the contract event 0x1c6a33c0de150a46e5647f4482e93d30f3a966487eb86e762d69319ab9a6e6b6.
//
// Solidity: e AnchorMade(height uint256, merkleroot uint256)
func (_FactomAnchor *FactomAnchorFilterer) FilterAnchorMade(opts *bind.FilterOpts) (*FactomAnchorAnchorMadeIterator, error) {

	logs, sub, err := _FactomAnchor.contract.FilterLogs(opts, "AnchorMade")
	if err != nil {
		return nil, err
	}
	return &FactomAnchorAnchorMadeIterator{contract: _FactomAnchor.contract, event: "AnchorMade", logs: logs, sub: sub}, nil
}

// WatchAnchorMade is a free log subscription operation binding the contract event 0x1c6a33c0de150a46e5647f4482e93d30f3a966487eb86e762d69319ab9a6e6b6.
//
// Solidity: e AnchorMade(height uint256, merkleroot uint256)
func (_FactomAnchor *FactomAnchorFilterer) WatchAnchorMade(opts *bind.WatchOpts, sink chan<- *FactomAnchorAnchorMade) (event.Subscription, error) {

	logs, sub, err := _FactomAnchor.contract.WatchLogs(opts, "AnchorMade")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FactomAnchorAnchorMade)
				if err := _FactomAnchor.contract.UnpackLog(event, "AnchorMade", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// FactomAnchorAnchoringFrozenIterator is returned from FilterAnchoringFrozen and is used to iterate over the raw logs and unpacked data for AnchoringFrozen events raised by the FactomAnchor contract.
type FactomAnchorAnchoringFrozenIterator struct {
	Event *FactomAnchorAnchoringFrozen // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FactomAnchorAnchoringFrozenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FactomAnchorAnchoringFrozen)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FactomAnchorAnchoringFrozen)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FactomAnchorAnchoringFrozenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FactomAnchorAnchoringFrozenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FactomAnchorAnchoringFrozen represents a AnchoringFrozen event raised by the FactomAnchor contract.
type FactomAnchorAnchoringFrozen struct {
	Height *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterAnchoringFrozen is a free log retrieval operation binding the contract event 0x02392dea61af8262e6609d1b99522854b729caa208dbccef7fd70f9508293aa7.
//
// Solidity: e AnchoringFrozen(height uint256)
func (_FactomAnchor *FactomAnchorFilterer) FilterAnchoringFrozen(opts *bind.FilterOpts) (*FactomAnchorAnchoringFrozenIterator, error) {

	logs, sub, err := _FactomAnchor.contract.FilterLogs(opts, "AnchoringFrozen")
	if err != nil {
		return nil, err
	}
	return &FactomAnchorAnchoringFrozenIterator{contract: _FactomAnchor.contract, event: "AnchoringFrozen", logs: logs, sub: sub}, nil
}

// WatchAnchoringFrozen is a free log subscription operation binding the contract event 0x02392dea61af8262e6609d1b99522854b729caa208dbccef7fd70f9508293aa7.
//
// Solidity: e AnchoringFrozen(height uint256)
func (_FactomAnchor *FactomAnchorFilterer) WatchAnchoringFrozen(opts *bind.WatchOpts, sink chan<- *FactomAnchorAnchoringFrozen) (event.Subscription, error) {

	logs, sub, err := _FactomAnchor.contract.WatchLogs(opts, "AnchoringFrozen")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FactomAnchorAnchoringFrozen)
				if err := _FactomAnchor.contract.UnpackLog(event, "AnchoringFrozen", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}
