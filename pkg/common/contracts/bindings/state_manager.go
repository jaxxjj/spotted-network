// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// IStateManagerHistory is an auto generated low-level Go binding around an user-defined struct.
type IStateManagerHistory struct {
	Value       *big.Int
	BlockNumber uint64
	Timestamp   *big.Int
}

// IStateManagerSetValueParams is an auto generated low-level Go binding around an user-defined struct.
type IStateManagerSetValueParams struct {
	Key   *big.Int
	Value *big.Int
}

// StateManagerMetaData contains all meta data concerning the StateManager contract.
var StateManagerMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"batchSetValues\",\"inputs\":[{\"name\":\"params\",\"type\":\"tuple[]\",\"internalType\":\"structIStateManager.SetValueParams[]\",\"components\":[{\"name\":\"key\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getCurrentValue\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"key\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCurrentValues\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"keys\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"outputs\":[{\"name\":\"values\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getHistory\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"key\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIStateManager.History[]\",\"components\":[{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getHistoryAfterBlockNumber\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"key\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIStateManager.History[]\",\"components\":[{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getHistoryAfterTimestamp\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"key\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"timestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIStateManager.History[]\",\"components\":[{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getHistoryAt\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"key\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"index\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIStateManager.History\",\"components\":[{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getHistoryAtBlock\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"key\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIStateManager.History\",\"components\":[{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getHistoryAtTimestamp\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"key\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"timestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIStateManager.History\",\"components\":[{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getHistoryBeforeBlockNumber\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"key\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIStateManager.History[]\",\"components\":[{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getHistoryBeforeTimestamp\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"key\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"timestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIStateManager.History[]\",\"components\":[{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getHistoryBetweenBlockNumbers\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"key\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"fromBlock\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"toBlock\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIStateManager.History[]\",\"components\":[{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getHistoryBetweenTimestamps\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"key\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"fromTimestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"toTimestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIStateManager.History[]\",\"components\":[{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getHistoryCount\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"key\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getLatestHistory\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"n\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structIStateManager.History[]\",\"components\":[{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"timestamp\",\"type\":\"uint48\",\"internalType\":\"uint48\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getUsedKeys\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setValue\",\"inputs\":[{\"name\":\"key\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"HistoryCommitted\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"key\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"value\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"timestamp\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"blockNumber\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"StateManager__BatchTooLarge\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StateManager__BlockNotFound\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StateManager__ImmutableStateCannotBeModified\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StateManager__IndexOutOfBounds\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StateManager__InvalidBlockRange\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StateManager__InvalidStateType\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StateManager__InvalidTimeRange\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StateManager__KeyNotFound\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StateManager__NoHistoryFound\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StateManager__TimestampNotFound\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StateManager__ValueNotMonotonicDecreasing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"StateManager__ValueNotMonotonicIncreasing\",\"inputs\":[]}]",
}

// StateManagerABI is the input ABI used to generate the binding from.
// Deprecated: Use StateManagerMetaData.ABI instead.
var StateManagerABI = StateManagerMetaData.ABI

// StateManager is an auto generated Go binding around an Ethereum contract.
type StateManager struct {
	StateManagerCaller     // Read-only binding to the contract
	StateManagerTransactor // Write-only binding to the contract
	StateManagerFilterer   // Log filterer for contract events
}

// StateManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type StateManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StateManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type StateManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StateManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type StateManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StateManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type StateManagerSession struct {
	Contract     *StateManager     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// StateManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type StateManagerCallerSession struct {
	Contract *StateManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// StateManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type StateManagerTransactorSession struct {
	Contract     *StateManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// StateManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type StateManagerRaw struct {
	Contract *StateManager // Generic contract binding to access the raw methods on
}

// StateManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type StateManagerCallerRaw struct {
	Contract *StateManagerCaller // Generic read-only contract binding to access the raw methods on
}

// StateManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type StateManagerTransactorRaw struct {
	Contract *StateManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewStateManager creates a new instance of StateManager, bound to a specific deployed contract.
func NewStateManager(address common.Address, backend bind.ContractBackend) (*StateManager, error) {
	contract, err := bindStateManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &StateManager{StateManagerCaller: StateManagerCaller{contract: contract}, StateManagerTransactor: StateManagerTransactor{contract: contract}, StateManagerFilterer: StateManagerFilterer{contract: contract}}, nil
}

// NewStateManagerCaller creates a new read-only instance of StateManager, bound to a specific deployed contract.
func NewStateManagerCaller(address common.Address, caller bind.ContractCaller) (*StateManagerCaller, error) {
	contract, err := bindStateManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &StateManagerCaller{contract: contract}, nil
}

// NewStateManagerTransactor creates a new write-only instance of StateManager, bound to a specific deployed contract.
func NewStateManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*StateManagerTransactor, error) {
	contract, err := bindStateManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &StateManagerTransactor{contract: contract}, nil
}

// NewStateManagerFilterer creates a new log filterer instance of StateManager, bound to a specific deployed contract.
func NewStateManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*StateManagerFilterer, error) {
	contract, err := bindStateManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &StateManagerFilterer{contract: contract}, nil
}

// bindStateManager binds a generic wrapper to an already deployed contract.
func bindStateManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := StateManagerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_StateManager *StateManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _StateManager.Contract.StateManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_StateManager *StateManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _StateManager.Contract.StateManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_StateManager *StateManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _StateManager.Contract.StateManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_StateManager *StateManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _StateManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_StateManager *StateManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _StateManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_StateManager *StateManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _StateManager.Contract.contract.Transact(opts, method, params...)
}

// GetCurrentValue is a free data retrieval call binding the contract method 0xe78372bb.
//
// Solidity: function getCurrentValue(address user, uint256 key) view returns(uint256)
func (_StateManager *StateManagerCaller) GetCurrentValue(opts *bind.CallOpts, user common.Address, key *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getCurrentValue", user, key)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCurrentValue is a free data retrieval call binding the contract method 0xe78372bb.
//
// Solidity: function getCurrentValue(address user, uint256 key) view returns(uint256)
func (_StateManager *StateManagerSession) GetCurrentValue(user common.Address, key *big.Int) (*big.Int, error) {
	return _StateManager.Contract.GetCurrentValue(&_StateManager.CallOpts, user, key)
}

// GetCurrentValue is a free data retrieval call binding the contract method 0xe78372bb.
//
// Solidity: function getCurrentValue(address user, uint256 key) view returns(uint256)
func (_StateManager *StateManagerCallerSession) GetCurrentValue(user common.Address, key *big.Int) (*big.Int, error) {
	return _StateManager.Contract.GetCurrentValue(&_StateManager.CallOpts, user, key)
}

// GetCurrentValues is a free data retrieval call binding the contract method 0xb8929153.
//
// Solidity: function getCurrentValues(address user, uint256[] keys) view returns(uint256[] values)
func (_StateManager *StateManagerCaller) GetCurrentValues(opts *bind.CallOpts, user common.Address, keys []*big.Int) ([]*big.Int, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getCurrentValues", user, keys)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// GetCurrentValues is a free data retrieval call binding the contract method 0xb8929153.
//
// Solidity: function getCurrentValues(address user, uint256[] keys) view returns(uint256[] values)
func (_StateManager *StateManagerSession) GetCurrentValues(user common.Address, keys []*big.Int) ([]*big.Int, error) {
	return _StateManager.Contract.GetCurrentValues(&_StateManager.CallOpts, user, keys)
}

// GetCurrentValues is a free data retrieval call binding the contract method 0xb8929153.
//
// Solidity: function getCurrentValues(address user, uint256[] keys) view returns(uint256[] values)
func (_StateManager *StateManagerCallerSession) GetCurrentValues(user common.Address, keys []*big.Int) ([]*big.Int, error) {
	return _StateManager.Contract.GetCurrentValues(&_StateManager.CallOpts, user, keys)
}

// GetHistory is a free data retrieval call binding the contract method 0x8bdf5803.
//
// Solidity: function getHistory(address user, uint256 key) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCaller) GetHistory(opts *bind.CallOpts, user common.Address, key *big.Int) ([]IStateManagerHistory, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getHistory", user, key)

	if err != nil {
		return *new([]IStateManagerHistory), err
	}

	out0 := *abi.ConvertType(out[0], new([]IStateManagerHistory)).(*[]IStateManagerHistory)

	return out0, err

}

// GetHistory is a free data retrieval call binding the contract method 0x8bdf5803.
//
// Solidity: function getHistory(address user, uint256 key) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerSession) GetHistory(user common.Address, key *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistory(&_StateManager.CallOpts, user, key)
}

// GetHistory is a free data retrieval call binding the contract method 0x8bdf5803.
//
// Solidity: function getHistory(address user, uint256 key) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCallerSession) GetHistory(user common.Address, key *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistory(&_StateManager.CallOpts, user, key)
}

// GetHistoryAfterBlockNumber is a free data retrieval call binding the contract method 0x937cad89.
//
// Solidity: function getHistoryAfterBlockNumber(address user, uint256 key, uint256 blockNumber) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCaller) GetHistoryAfterBlockNumber(opts *bind.CallOpts, user common.Address, key *big.Int, blockNumber *big.Int) ([]IStateManagerHistory, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getHistoryAfterBlockNumber", user, key, blockNumber)

	if err != nil {
		return *new([]IStateManagerHistory), err
	}

	out0 := *abi.ConvertType(out[0], new([]IStateManagerHistory)).(*[]IStateManagerHistory)

	return out0, err

}

// GetHistoryAfterBlockNumber is a free data retrieval call binding the contract method 0x937cad89.
//
// Solidity: function getHistoryAfterBlockNumber(address user, uint256 key, uint256 blockNumber) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerSession) GetHistoryAfterBlockNumber(user common.Address, key *big.Int, blockNumber *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryAfterBlockNumber(&_StateManager.CallOpts, user, key, blockNumber)
}

// GetHistoryAfterBlockNumber is a free data retrieval call binding the contract method 0x937cad89.
//
// Solidity: function getHistoryAfterBlockNumber(address user, uint256 key, uint256 blockNumber) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCallerSession) GetHistoryAfterBlockNumber(user common.Address, key *big.Int, blockNumber *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryAfterBlockNumber(&_StateManager.CallOpts, user, key, blockNumber)
}

// GetHistoryAfterTimestamp is a free data retrieval call binding the contract method 0xbadd55e6.
//
// Solidity: function getHistoryAfterTimestamp(address user, uint256 key, uint256 timestamp) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCaller) GetHistoryAfterTimestamp(opts *bind.CallOpts, user common.Address, key *big.Int, timestamp *big.Int) ([]IStateManagerHistory, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getHistoryAfterTimestamp", user, key, timestamp)

	if err != nil {
		return *new([]IStateManagerHistory), err
	}

	out0 := *abi.ConvertType(out[0], new([]IStateManagerHistory)).(*[]IStateManagerHistory)

	return out0, err

}

// GetHistoryAfterTimestamp is a free data retrieval call binding the contract method 0xbadd55e6.
//
// Solidity: function getHistoryAfterTimestamp(address user, uint256 key, uint256 timestamp) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerSession) GetHistoryAfterTimestamp(user common.Address, key *big.Int, timestamp *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryAfterTimestamp(&_StateManager.CallOpts, user, key, timestamp)
}

// GetHistoryAfterTimestamp is a free data retrieval call binding the contract method 0xbadd55e6.
//
// Solidity: function getHistoryAfterTimestamp(address user, uint256 key, uint256 timestamp) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCallerSession) GetHistoryAfterTimestamp(user common.Address, key *big.Int, timestamp *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryAfterTimestamp(&_StateManager.CallOpts, user, key, timestamp)
}

// GetHistoryAt is a free data retrieval call binding the contract method 0xa0f72401.
//
// Solidity: function getHistoryAt(address user, uint256 key, uint256 index) view returns((uint256,uint64,uint48))
func (_StateManager *StateManagerCaller) GetHistoryAt(opts *bind.CallOpts, user common.Address, key *big.Int, index *big.Int) (IStateManagerHistory, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getHistoryAt", user, key, index)

	if err != nil {
		return *new(IStateManagerHistory), err
	}

	out0 := *abi.ConvertType(out[0], new(IStateManagerHistory)).(*IStateManagerHistory)

	return out0, err

}

// GetHistoryAt is a free data retrieval call binding the contract method 0xa0f72401.
//
// Solidity: function getHistoryAt(address user, uint256 key, uint256 index) view returns((uint256,uint64,uint48))
func (_StateManager *StateManagerSession) GetHistoryAt(user common.Address, key *big.Int, index *big.Int) (IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryAt(&_StateManager.CallOpts, user, key, index)
}

// GetHistoryAt is a free data retrieval call binding the contract method 0xa0f72401.
//
// Solidity: function getHistoryAt(address user, uint256 key, uint256 index) view returns((uint256,uint64,uint48))
func (_StateManager *StateManagerCallerSession) GetHistoryAt(user common.Address, key *big.Int, index *big.Int) (IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryAt(&_StateManager.CallOpts, user, key, index)
}

// GetHistoryAtBlock is a free data retrieval call binding the contract method 0x35c7fa54.
//
// Solidity: function getHistoryAtBlock(address user, uint256 key, uint256 blockNumber) view returns((uint256,uint64,uint48))
func (_StateManager *StateManagerCaller) GetHistoryAtBlock(opts *bind.CallOpts, user common.Address, key *big.Int, blockNumber *big.Int) (IStateManagerHistory, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getHistoryAtBlock", user, key, blockNumber)

	if err != nil {
		return *new(IStateManagerHistory), err
	}

	out0 := *abi.ConvertType(out[0], new(IStateManagerHistory)).(*IStateManagerHistory)

	return out0, err

}

// GetHistoryAtBlock is a free data retrieval call binding the contract method 0x35c7fa54.
//
// Solidity: function getHistoryAtBlock(address user, uint256 key, uint256 blockNumber) view returns((uint256,uint64,uint48))
func (_StateManager *StateManagerSession) GetHistoryAtBlock(user common.Address, key *big.Int, blockNumber *big.Int) (IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryAtBlock(&_StateManager.CallOpts, user, key, blockNumber)
}

// GetHistoryAtBlock is a free data retrieval call binding the contract method 0x35c7fa54.
//
// Solidity: function getHistoryAtBlock(address user, uint256 key, uint256 blockNumber) view returns((uint256,uint64,uint48))
func (_StateManager *StateManagerCallerSession) GetHistoryAtBlock(user common.Address, key *big.Int, blockNumber *big.Int) (IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryAtBlock(&_StateManager.CallOpts, user, key, blockNumber)
}

// GetHistoryAtTimestamp is a free data retrieval call binding the contract method 0x2edc01b1.
//
// Solidity: function getHistoryAtTimestamp(address user, uint256 key, uint256 timestamp) view returns((uint256,uint64,uint48))
func (_StateManager *StateManagerCaller) GetHistoryAtTimestamp(opts *bind.CallOpts, user common.Address, key *big.Int, timestamp *big.Int) (IStateManagerHistory, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getHistoryAtTimestamp", user, key, timestamp)

	if err != nil {
		return *new(IStateManagerHistory), err
	}

	out0 := *abi.ConvertType(out[0], new(IStateManagerHistory)).(*IStateManagerHistory)

	return out0, err

}

// GetHistoryAtTimestamp is a free data retrieval call binding the contract method 0x2edc01b1.
//
// Solidity: function getHistoryAtTimestamp(address user, uint256 key, uint256 timestamp) view returns((uint256,uint64,uint48))
func (_StateManager *StateManagerSession) GetHistoryAtTimestamp(user common.Address, key *big.Int, timestamp *big.Int) (IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryAtTimestamp(&_StateManager.CallOpts, user, key, timestamp)
}

// GetHistoryAtTimestamp is a free data retrieval call binding the contract method 0x2edc01b1.
//
// Solidity: function getHistoryAtTimestamp(address user, uint256 key, uint256 timestamp) view returns((uint256,uint64,uint48))
func (_StateManager *StateManagerCallerSession) GetHistoryAtTimestamp(user common.Address, key *big.Int, timestamp *big.Int) (IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryAtTimestamp(&_StateManager.CallOpts, user, key, timestamp)
}

// GetHistoryBeforeBlockNumber is a free data retrieval call binding the contract method 0xde90452a.
//
// Solidity: function getHistoryBeforeBlockNumber(address user, uint256 key, uint256 blockNumber) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCaller) GetHistoryBeforeBlockNumber(opts *bind.CallOpts, user common.Address, key *big.Int, blockNumber *big.Int) ([]IStateManagerHistory, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getHistoryBeforeBlockNumber", user, key, blockNumber)

	if err != nil {
		return *new([]IStateManagerHistory), err
	}

	out0 := *abi.ConvertType(out[0], new([]IStateManagerHistory)).(*[]IStateManagerHistory)

	return out0, err

}

// GetHistoryBeforeBlockNumber is a free data retrieval call binding the contract method 0xde90452a.
//
// Solidity: function getHistoryBeforeBlockNumber(address user, uint256 key, uint256 blockNumber) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerSession) GetHistoryBeforeBlockNumber(user common.Address, key *big.Int, blockNumber *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryBeforeBlockNumber(&_StateManager.CallOpts, user, key, blockNumber)
}

// GetHistoryBeforeBlockNumber is a free data retrieval call binding the contract method 0xde90452a.
//
// Solidity: function getHistoryBeforeBlockNumber(address user, uint256 key, uint256 blockNumber) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCallerSession) GetHistoryBeforeBlockNumber(user common.Address, key *big.Int, blockNumber *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryBeforeBlockNumber(&_StateManager.CallOpts, user, key, blockNumber)
}

// GetHistoryBeforeTimestamp is a free data retrieval call binding the contract method 0xa80832f2.
//
// Solidity: function getHistoryBeforeTimestamp(address user, uint256 key, uint256 timestamp) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCaller) GetHistoryBeforeTimestamp(opts *bind.CallOpts, user common.Address, key *big.Int, timestamp *big.Int) ([]IStateManagerHistory, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getHistoryBeforeTimestamp", user, key, timestamp)

	if err != nil {
		return *new([]IStateManagerHistory), err
	}

	out0 := *abi.ConvertType(out[0], new([]IStateManagerHistory)).(*[]IStateManagerHistory)

	return out0, err

}

// GetHistoryBeforeTimestamp is a free data retrieval call binding the contract method 0xa80832f2.
//
// Solidity: function getHistoryBeforeTimestamp(address user, uint256 key, uint256 timestamp) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerSession) GetHistoryBeforeTimestamp(user common.Address, key *big.Int, timestamp *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryBeforeTimestamp(&_StateManager.CallOpts, user, key, timestamp)
}

// GetHistoryBeforeTimestamp is a free data retrieval call binding the contract method 0xa80832f2.
//
// Solidity: function getHistoryBeforeTimestamp(address user, uint256 key, uint256 timestamp) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCallerSession) GetHistoryBeforeTimestamp(user common.Address, key *big.Int, timestamp *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryBeforeTimestamp(&_StateManager.CallOpts, user, key, timestamp)
}

// GetHistoryBetweenBlockNumbers is a free data retrieval call binding the contract method 0xc4c31cf9.
//
// Solidity: function getHistoryBetweenBlockNumbers(address user, uint256 key, uint256 fromBlock, uint256 toBlock) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCaller) GetHistoryBetweenBlockNumbers(opts *bind.CallOpts, user common.Address, key *big.Int, fromBlock *big.Int, toBlock *big.Int) ([]IStateManagerHistory, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getHistoryBetweenBlockNumbers", user, key, fromBlock, toBlock)

	if err != nil {
		return *new([]IStateManagerHistory), err
	}

	out0 := *abi.ConvertType(out[0], new([]IStateManagerHistory)).(*[]IStateManagerHistory)

	return out0, err

}

// GetHistoryBetweenBlockNumbers is a free data retrieval call binding the contract method 0xc4c31cf9.
//
// Solidity: function getHistoryBetweenBlockNumbers(address user, uint256 key, uint256 fromBlock, uint256 toBlock) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerSession) GetHistoryBetweenBlockNumbers(user common.Address, key *big.Int, fromBlock *big.Int, toBlock *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryBetweenBlockNumbers(&_StateManager.CallOpts, user, key, fromBlock, toBlock)
}

// GetHistoryBetweenBlockNumbers is a free data retrieval call binding the contract method 0xc4c31cf9.
//
// Solidity: function getHistoryBetweenBlockNumbers(address user, uint256 key, uint256 fromBlock, uint256 toBlock) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCallerSession) GetHistoryBetweenBlockNumbers(user common.Address, key *big.Int, fromBlock *big.Int, toBlock *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryBetweenBlockNumbers(&_StateManager.CallOpts, user, key, fromBlock, toBlock)
}

// GetHistoryBetweenTimestamps is a free data retrieval call binding the contract method 0x08244ee5.
//
// Solidity: function getHistoryBetweenTimestamps(address user, uint256 key, uint256 fromTimestamp, uint256 toTimestamp) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCaller) GetHistoryBetweenTimestamps(opts *bind.CallOpts, user common.Address, key *big.Int, fromTimestamp *big.Int, toTimestamp *big.Int) ([]IStateManagerHistory, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getHistoryBetweenTimestamps", user, key, fromTimestamp, toTimestamp)

	if err != nil {
		return *new([]IStateManagerHistory), err
	}

	out0 := *abi.ConvertType(out[0], new([]IStateManagerHistory)).(*[]IStateManagerHistory)

	return out0, err

}

// GetHistoryBetweenTimestamps is a free data retrieval call binding the contract method 0x08244ee5.
//
// Solidity: function getHistoryBetweenTimestamps(address user, uint256 key, uint256 fromTimestamp, uint256 toTimestamp) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerSession) GetHistoryBetweenTimestamps(user common.Address, key *big.Int, fromTimestamp *big.Int, toTimestamp *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryBetweenTimestamps(&_StateManager.CallOpts, user, key, fromTimestamp, toTimestamp)
}

// GetHistoryBetweenTimestamps is a free data retrieval call binding the contract method 0x08244ee5.
//
// Solidity: function getHistoryBetweenTimestamps(address user, uint256 key, uint256 fromTimestamp, uint256 toTimestamp) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCallerSession) GetHistoryBetweenTimestamps(user common.Address, key *big.Int, fromTimestamp *big.Int, toTimestamp *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetHistoryBetweenTimestamps(&_StateManager.CallOpts, user, key, fromTimestamp, toTimestamp)
}

// GetHistoryCount is a free data retrieval call binding the contract method 0xfa749363.
//
// Solidity: function getHistoryCount(address user, uint256 key) view returns(uint256)
func (_StateManager *StateManagerCaller) GetHistoryCount(opts *bind.CallOpts, user common.Address, key *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getHistoryCount", user, key)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetHistoryCount is a free data retrieval call binding the contract method 0xfa749363.
//
// Solidity: function getHistoryCount(address user, uint256 key) view returns(uint256)
func (_StateManager *StateManagerSession) GetHistoryCount(user common.Address, key *big.Int) (*big.Int, error) {
	return _StateManager.Contract.GetHistoryCount(&_StateManager.CallOpts, user, key)
}

// GetHistoryCount is a free data retrieval call binding the contract method 0xfa749363.
//
// Solidity: function getHistoryCount(address user, uint256 key) view returns(uint256)
func (_StateManager *StateManagerCallerSession) GetHistoryCount(user common.Address, key *big.Int) (*big.Int, error) {
	return _StateManager.Contract.GetHistoryCount(&_StateManager.CallOpts, user, key)
}

// GetLatestHistory is a free data retrieval call binding the contract method 0x173b2fdd.
//
// Solidity: function getLatestHistory(address user, uint256 n) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCaller) GetLatestHistory(opts *bind.CallOpts, user common.Address, n *big.Int) ([]IStateManagerHistory, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getLatestHistory", user, n)

	if err != nil {
		return *new([]IStateManagerHistory), err
	}

	out0 := *abi.ConvertType(out[0], new([]IStateManagerHistory)).(*[]IStateManagerHistory)

	return out0, err

}

// GetLatestHistory is a free data retrieval call binding the contract method 0x173b2fdd.
//
// Solidity: function getLatestHistory(address user, uint256 n) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerSession) GetLatestHistory(user common.Address, n *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetLatestHistory(&_StateManager.CallOpts, user, n)
}

// GetLatestHistory is a free data retrieval call binding the contract method 0x173b2fdd.
//
// Solidity: function getLatestHistory(address user, uint256 n) view returns((uint256,uint64,uint48)[])
func (_StateManager *StateManagerCallerSession) GetLatestHistory(user common.Address, n *big.Int) ([]IStateManagerHistory, error) {
	return _StateManager.Contract.GetLatestHistory(&_StateManager.CallOpts, user, n)
}

// GetUsedKeys is a free data retrieval call binding the contract method 0x3503418f.
//
// Solidity: function getUsedKeys(address user) view returns(uint256[])
func (_StateManager *StateManagerCaller) GetUsedKeys(opts *bind.CallOpts, user common.Address) ([]*big.Int, error) {
	var out []interface{}
	err := _StateManager.contract.Call(opts, &out, "getUsedKeys", user)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// GetUsedKeys is a free data retrieval call binding the contract method 0x3503418f.
//
// Solidity: function getUsedKeys(address user) view returns(uint256[])
func (_StateManager *StateManagerSession) GetUsedKeys(user common.Address) ([]*big.Int, error) {
	return _StateManager.Contract.GetUsedKeys(&_StateManager.CallOpts, user)
}

// GetUsedKeys is a free data retrieval call binding the contract method 0x3503418f.
//
// Solidity: function getUsedKeys(address user) view returns(uint256[])
func (_StateManager *StateManagerCallerSession) GetUsedKeys(user common.Address) ([]*big.Int, error) {
	return _StateManager.Contract.GetUsedKeys(&_StateManager.CallOpts, user)
}

// BatchSetValues is a paid mutator transaction binding the contract method 0x3d4868f8.
//
// Solidity: function batchSetValues((uint256,uint256)[] params) returns()
func (_StateManager *StateManagerTransactor) BatchSetValues(opts *bind.TransactOpts, params []IStateManagerSetValueParams) (*types.Transaction, error) {
	return _StateManager.contract.Transact(opts, "batchSetValues", params)
}

// BatchSetValues is a paid mutator transaction binding the contract method 0x3d4868f8.
//
// Solidity: function batchSetValues((uint256,uint256)[] params) returns()
func (_StateManager *StateManagerSession) BatchSetValues(params []IStateManagerSetValueParams) (*types.Transaction, error) {
	return _StateManager.Contract.BatchSetValues(&_StateManager.TransactOpts, params)
}

// BatchSetValues is a paid mutator transaction binding the contract method 0x3d4868f8.
//
// Solidity: function batchSetValues((uint256,uint256)[] params) returns()
func (_StateManager *StateManagerTransactorSession) BatchSetValues(params []IStateManagerSetValueParams) (*types.Transaction, error) {
	return _StateManager.Contract.BatchSetValues(&_StateManager.TransactOpts, params)
}

// SetValue is a paid mutator transaction binding the contract method 0x7b8d56e3.
//
// Solidity: function setValue(uint256 key, uint256 value) returns()
func (_StateManager *StateManagerTransactor) SetValue(opts *bind.TransactOpts, key *big.Int, value *big.Int) (*types.Transaction, error) {
	return _StateManager.contract.Transact(opts, "setValue", key, value)
}

// SetValue is a paid mutator transaction binding the contract method 0x7b8d56e3.
//
// Solidity: function setValue(uint256 key, uint256 value) returns()
func (_StateManager *StateManagerSession) SetValue(key *big.Int, value *big.Int) (*types.Transaction, error) {
	return _StateManager.Contract.SetValue(&_StateManager.TransactOpts, key, value)
}

// SetValue is a paid mutator transaction binding the contract method 0x7b8d56e3.
//
// Solidity: function setValue(uint256 key, uint256 value) returns()
func (_StateManager *StateManagerTransactorSession) SetValue(key *big.Int, value *big.Int) (*types.Transaction, error) {
	return _StateManager.Contract.SetValue(&_StateManager.TransactOpts, key, value)
}

// StateManagerHistoryCommittedIterator is returned from FilterHistoryCommitted and is used to iterate over the raw logs and unpacked data for HistoryCommitted events raised by the StateManager contract.
type StateManagerHistoryCommittedIterator struct {
	Event *StateManagerHistoryCommitted // Event containing the contract specifics and raw log

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
func (it *StateManagerHistoryCommittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StateManagerHistoryCommitted)
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
		it.Event = new(StateManagerHistoryCommitted)
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
func (it *StateManagerHistoryCommittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StateManagerHistoryCommittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StateManagerHistoryCommitted represents a HistoryCommitted event raised by the StateManager contract.
type StateManagerHistoryCommitted struct {
	User        common.Address
	Key         *big.Int
	Value       *big.Int
	Timestamp   *big.Int
	BlockNumber *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterHistoryCommitted is a free log retrieval operation binding the contract event 0xa1a325dd3ab50a50da29b7e0c70017d7834b6d6565bd3518be12f123084109fe.
//
// Solidity: event HistoryCommitted(address indexed user, uint256 indexed key, uint256 value, uint256 timestamp, uint256 blockNumber)
func (_StateManager *StateManagerFilterer) FilterHistoryCommitted(opts *bind.FilterOpts, user []common.Address, key []*big.Int) (*StateManagerHistoryCommittedIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}

	logs, sub, err := _StateManager.contract.FilterLogs(opts, "HistoryCommitted", userRule, keyRule)
	if err != nil {
		return nil, err
	}
	return &StateManagerHistoryCommittedIterator{contract: _StateManager.contract, event: "HistoryCommitted", logs: logs, sub: sub}, nil
}

// WatchHistoryCommitted is a free log subscription operation binding the contract event 0xa1a325dd3ab50a50da29b7e0c70017d7834b6d6565bd3518be12f123084109fe.
//
// Solidity: event HistoryCommitted(address indexed user, uint256 indexed key, uint256 value, uint256 timestamp, uint256 blockNumber)
func (_StateManager *StateManagerFilterer) WatchHistoryCommitted(opts *bind.WatchOpts, sink chan<- *StateManagerHistoryCommitted, user []common.Address, key []*big.Int) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var keyRule []interface{}
	for _, keyItem := range key {
		keyRule = append(keyRule, keyItem)
	}

	logs, sub, err := _StateManager.contract.WatchLogs(opts, "HistoryCommitted", userRule, keyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StateManagerHistoryCommitted)
				if err := _StateManager.contract.UnpackLog(event, "HistoryCommitted", log); err != nil {
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

// ParseHistoryCommitted is a log parse operation binding the contract event 0xa1a325dd3ab50a50da29b7e0c70017d7834b6d6565bd3518be12f123084109fe.
//
// Solidity: event HistoryCommitted(address indexed user, uint256 indexed key, uint256 value, uint256 timestamp, uint256 blockNumber)
func (_StateManager *StateManagerFilterer) ParseHistoryCommitted(log types.Log) (*StateManagerHistoryCommitted, error) {
	event := new(StateManagerHistoryCommitted)
	if err := _StateManager.contract.UnpackLog(event, "HistoryCommitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
