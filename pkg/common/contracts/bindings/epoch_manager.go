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

// EpochManagerMetaData contains all meta data concerning the EpochManager contract.
var EpochManagerMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_registryStateSender\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"EPOCH_LENGTH\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"GENESIS_BLOCK\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"GRACE_PERIOD\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"REGISTRY_STATE_SENDER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"advanceEpoch\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"blocksUntilGracePeriod\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"blocksUntilNextEpoch\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"canAdvanceEpoch\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"currentEpoch\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"epochBlocks\",\"inputs\":[{\"name\":\"startBlock\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"outputs\":[{\"name\":\"epochNumber\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"epochUpdateCounts\",\"inputs\":[{\"name\":\"epochNumber\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"updateCount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getCurrentEpoch\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEffectiveEpoch\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEffectiveEpochForBlock\",\"inputs\":[{\"name\":\"blockNumber\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEpochInterval\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"startBlock\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"graceBlock\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"endBlock\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isInGracePeriod\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isUpdatable\",\"inputs\":[{\"name\":\"targetEpoch\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"lastEpochBlock\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"lastUpdatedEpoch\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"nextEpochBlock\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"queueStateUpdate\",\"inputs\":[{\"name\":\"updateType\",\"type\":\"uint8\",\"internalType\":\"enumIEpochManager.MessageType\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"revertEpoch\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"sendStateUpdates\",\"inputs\":[{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"EpochAdvanced\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"lastEpochBlock\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"},{\"name\":\"nextEpochBlock\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EpochReverted\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"uint8\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"StateUpdateQueued\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"updateType\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumIEpochManager.MessageType\"},{\"name\":\"data\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"StateUpdatesSent\",\"inputs\":[{\"name\":\"epoch\",\"type\":\"uint32\",\"indexed\":true,\"internalType\":\"uint32\"},{\"name\":\"updatesCount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"EpochManager__EpochNotReady\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EpochManager__InvalidEpochForUpdate\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EpochManager__InvalidEpochLength\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EpochManager__InvalidEpochRevert\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EpochManager__InvalidGracePeriod\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EpochManager__InvalidPeriodLength\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EpochManager__UpdateAlreadyProcessed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EpochManager__UpdateNotAllowed\",\"inputs\":[]}]",
}

// EpochManagerABI is the input ABI used to generate the binding from.
// Deprecated: Use EpochManagerMetaData.ABI instead.
var EpochManagerABI = EpochManagerMetaData.ABI

// EpochManager is an auto generated Go binding around an Ethereum contract.
type EpochManager struct {
	EpochManagerCaller     // Read-only binding to the contract
	EpochManagerTransactor // Write-only binding to the contract
	EpochManagerFilterer   // Log filterer for contract events
}

// EpochManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type EpochManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EpochManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EpochManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EpochManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EpochManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EpochManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EpochManagerSession struct {
	Contract     *EpochManager     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EpochManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EpochManagerCallerSession struct {
	Contract *EpochManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// EpochManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EpochManagerTransactorSession struct {
	Contract     *EpochManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// EpochManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type EpochManagerRaw struct {
	Contract *EpochManager // Generic contract binding to access the raw methods on
}

// EpochManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EpochManagerCallerRaw struct {
	Contract *EpochManagerCaller // Generic read-only contract binding to access the raw methods on
}

// EpochManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EpochManagerTransactorRaw struct {
	Contract *EpochManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEpochManager creates a new instance of EpochManager, bound to a specific deployed contract.
func NewEpochManager(address common.Address, backend bind.ContractBackend) (*EpochManager, error) {
	contract, err := bindEpochManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &EpochManager{EpochManagerCaller: EpochManagerCaller{contract: contract}, EpochManagerTransactor: EpochManagerTransactor{contract: contract}, EpochManagerFilterer: EpochManagerFilterer{contract: contract}}, nil
}

// NewEpochManagerCaller creates a new read-only instance of EpochManager, bound to a specific deployed contract.
func NewEpochManagerCaller(address common.Address, caller bind.ContractCaller) (*EpochManagerCaller, error) {
	contract, err := bindEpochManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EpochManagerCaller{contract: contract}, nil
}

// NewEpochManagerTransactor creates a new write-only instance of EpochManager, bound to a specific deployed contract.
func NewEpochManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*EpochManagerTransactor, error) {
	contract, err := bindEpochManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EpochManagerTransactor{contract: contract}, nil
}

// NewEpochManagerFilterer creates a new log filterer instance of EpochManager, bound to a specific deployed contract.
func NewEpochManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*EpochManagerFilterer, error) {
	contract, err := bindEpochManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EpochManagerFilterer{contract: contract}, nil
}

// bindEpochManager binds a generic wrapper to an already deployed contract.
func bindEpochManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := EpochManagerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EpochManager *EpochManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _EpochManager.Contract.EpochManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EpochManager *EpochManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EpochManager.Contract.EpochManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EpochManager *EpochManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EpochManager.Contract.EpochManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_EpochManager *EpochManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _EpochManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_EpochManager *EpochManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EpochManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_EpochManager *EpochManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _EpochManager.Contract.contract.Transact(opts, method, params...)
}

// EPOCHLENGTH is a free data retrieval call binding the contract method 0xac4746ab.
//
// Solidity: function EPOCH_LENGTH() view returns(uint64)
func (_EpochManager *EpochManagerCaller) EPOCHLENGTH(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "EPOCH_LENGTH")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// EPOCHLENGTH is a free data retrieval call binding the contract method 0xac4746ab.
//
// Solidity: function EPOCH_LENGTH() view returns(uint64)
func (_EpochManager *EpochManagerSession) EPOCHLENGTH() (uint64, error) {
	return _EpochManager.Contract.EPOCHLENGTH(&_EpochManager.CallOpts)
}

// EPOCHLENGTH is a free data retrieval call binding the contract method 0xac4746ab.
//
// Solidity: function EPOCH_LENGTH() view returns(uint64)
func (_EpochManager *EpochManagerCallerSession) EPOCHLENGTH() (uint64, error) {
	return _EpochManager.Contract.EPOCHLENGTH(&_EpochManager.CallOpts)
}

// GENESISBLOCK is a free data retrieval call binding the contract method 0xddbfc5c6.
//
// Solidity: function GENESIS_BLOCK() view returns(uint64)
func (_EpochManager *EpochManagerCaller) GENESISBLOCK(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "GENESIS_BLOCK")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GENESISBLOCK is a free data retrieval call binding the contract method 0xddbfc5c6.
//
// Solidity: function GENESIS_BLOCK() view returns(uint64)
func (_EpochManager *EpochManagerSession) GENESISBLOCK() (uint64, error) {
	return _EpochManager.Contract.GENESISBLOCK(&_EpochManager.CallOpts)
}

// GENESISBLOCK is a free data retrieval call binding the contract method 0xddbfc5c6.
//
// Solidity: function GENESIS_BLOCK() view returns(uint64)
func (_EpochManager *EpochManagerCallerSession) GENESISBLOCK() (uint64, error) {
	return _EpochManager.Contract.GENESISBLOCK(&_EpochManager.CallOpts)
}

// GRACEPERIOD is a free data retrieval call binding the contract method 0xc1a287e2.
//
// Solidity: function GRACE_PERIOD() view returns(uint64)
func (_EpochManager *EpochManagerCaller) GRACEPERIOD(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "GRACE_PERIOD")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GRACEPERIOD is a free data retrieval call binding the contract method 0xc1a287e2.
//
// Solidity: function GRACE_PERIOD() view returns(uint64)
func (_EpochManager *EpochManagerSession) GRACEPERIOD() (uint64, error) {
	return _EpochManager.Contract.GRACEPERIOD(&_EpochManager.CallOpts)
}

// GRACEPERIOD is a free data retrieval call binding the contract method 0xc1a287e2.
//
// Solidity: function GRACE_PERIOD() view returns(uint64)
func (_EpochManager *EpochManagerCallerSession) GRACEPERIOD() (uint64, error) {
	return _EpochManager.Contract.GRACEPERIOD(&_EpochManager.CallOpts)
}

// REGISTRYSTATESENDER is a free data retrieval call binding the contract method 0x9c732b69.
//
// Solidity: function REGISTRY_STATE_SENDER() view returns(address)
func (_EpochManager *EpochManagerCaller) REGISTRYSTATESENDER(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "REGISTRY_STATE_SENDER")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// REGISTRYSTATESENDER is a free data retrieval call binding the contract method 0x9c732b69.
//
// Solidity: function REGISTRY_STATE_SENDER() view returns(address)
func (_EpochManager *EpochManagerSession) REGISTRYSTATESENDER() (common.Address, error) {
	return _EpochManager.Contract.REGISTRYSTATESENDER(&_EpochManager.CallOpts)
}

// REGISTRYSTATESENDER is a free data retrieval call binding the contract method 0x9c732b69.
//
// Solidity: function REGISTRY_STATE_SENDER() view returns(address)
func (_EpochManager *EpochManagerCallerSession) REGISTRYSTATESENDER() (common.Address, error) {
	return _EpochManager.Contract.REGISTRYSTATESENDER(&_EpochManager.CallOpts)
}

// BlocksUntilGracePeriod is a free data retrieval call binding the contract method 0xedf29448.
//
// Solidity: function blocksUntilGracePeriod() view returns(uint64)
func (_EpochManager *EpochManagerCaller) BlocksUntilGracePeriod(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "blocksUntilGracePeriod")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// BlocksUntilGracePeriod is a free data retrieval call binding the contract method 0xedf29448.
//
// Solidity: function blocksUntilGracePeriod() view returns(uint64)
func (_EpochManager *EpochManagerSession) BlocksUntilGracePeriod() (uint64, error) {
	return _EpochManager.Contract.BlocksUntilGracePeriod(&_EpochManager.CallOpts)
}

// BlocksUntilGracePeriod is a free data retrieval call binding the contract method 0xedf29448.
//
// Solidity: function blocksUntilGracePeriod() view returns(uint64)
func (_EpochManager *EpochManagerCallerSession) BlocksUntilGracePeriod() (uint64, error) {
	return _EpochManager.Contract.BlocksUntilGracePeriod(&_EpochManager.CallOpts)
}

// BlocksUntilNextEpoch is a free data retrieval call binding the contract method 0x4f3d90b1.
//
// Solidity: function blocksUntilNextEpoch() view returns(uint64)
func (_EpochManager *EpochManagerCaller) BlocksUntilNextEpoch(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "blocksUntilNextEpoch")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// BlocksUntilNextEpoch is a free data retrieval call binding the contract method 0x4f3d90b1.
//
// Solidity: function blocksUntilNextEpoch() view returns(uint64)
func (_EpochManager *EpochManagerSession) BlocksUntilNextEpoch() (uint64, error) {
	return _EpochManager.Contract.BlocksUntilNextEpoch(&_EpochManager.CallOpts)
}

// BlocksUntilNextEpoch is a free data retrieval call binding the contract method 0x4f3d90b1.
//
// Solidity: function blocksUntilNextEpoch() view returns(uint64)
func (_EpochManager *EpochManagerCallerSession) BlocksUntilNextEpoch() (uint64, error) {
	return _EpochManager.Contract.BlocksUntilNextEpoch(&_EpochManager.CallOpts)
}

// CanAdvanceEpoch is a free data retrieval call binding the contract method 0x24ca9275.
//
// Solidity: function canAdvanceEpoch() view returns(bool)
func (_EpochManager *EpochManagerCaller) CanAdvanceEpoch(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "canAdvanceEpoch")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CanAdvanceEpoch is a free data retrieval call binding the contract method 0x24ca9275.
//
// Solidity: function canAdvanceEpoch() view returns(bool)
func (_EpochManager *EpochManagerSession) CanAdvanceEpoch() (bool, error) {
	return _EpochManager.Contract.CanAdvanceEpoch(&_EpochManager.CallOpts)
}

// CanAdvanceEpoch is a free data retrieval call binding the contract method 0x24ca9275.
//
// Solidity: function canAdvanceEpoch() view returns(bool)
func (_EpochManager *EpochManagerCallerSession) CanAdvanceEpoch() (bool, error) {
	return _EpochManager.Contract.CanAdvanceEpoch(&_EpochManager.CallOpts)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint32)
func (_EpochManager *EpochManagerCaller) CurrentEpoch(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "currentEpoch")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint32)
func (_EpochManager *EpochManagerSession) CurrentEpoch() (uint32, error) {
	return _EpochManager.Contract.CurrentEpoch(&_EpochManager.CallOpts)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint32)
func (_EpochManager *EpochManagerCallerSession) CurrentEpoch() (uint32, error) {
	return _EpochManager.Contract.CurrentEpoch(&_EpochManager.CallOpts)
}

// EpochBlocks is a free data retrieval call binding the contract method 0x10c7da6c.
//
// Solidity: function epochBlocks(uint64 startBlock) view returns(uint32 epochNumber)
func (_EpochManager *EpochManagerCaller) EpochBlocks(opts *bind.CallOpts, startBlock uint64) (uint32, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "epochBlocks", startBlock)

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// EpochBlocks is a free data retrieval call binding the contract method 0x10c7da6c.
//
// Solidity: function epochBlocks(uint64 startBlock) view returns(uint32 epochNumber)
func (_EpochManager *EpochManagerSession) EpochBlocks(startBlock uint64) (uint32, error) {
	return _EpochManager.Contract.EpochBlocks(&_EpochManager.CallOpts, startBlock)
}

// EpochBlocks is a free data retrieval call binding the contract method 0x10c7da6c.
//
// Solidity: function epochBlocks(uint64 startBlock) view returns(uint32 epochNumber)
func (_EpochManager *EpochManagerCallerSession) EpochBlocks(startBlock uint64) (uint32, error) {
	return _EpochManager.Contract.EpochBlocks(&_EpochManager.CallOpts, startBlock)
}

// EpochUpdateCounts is a free data retrieval call binding the contract method 0x6d96256d.
//
// Solidity: function epochUpdateCounts(uint32 epochNumber) view returns(uint256 updateCount)
func (_EpochManager *EpochManagerCaller) EpochUpdateCounts(opts *bind.CallOpts, epochNumber uint32) (*big.Int, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "epochUpdateCounts", epochNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// EpochUpdateCounts is a free data retrieval call binding the contract method 0x6d96256d.
//
// Solidity: function epochUpdateCounts(uint32 epochNumber) view returns(uint256 updateCount)
func (_EpochManager *EpochManagerSession) EpochUpdateCounts(epochNumber uint32) (*big.Int, error) {
	return _EpochManager.Contract.EpochUpdateCounts(&_EpochManager.CallOpts, epochNumber)
}

// EpochUpdateCounts is a free data retrieval call binding the contract method 0x6d96256d.
//
// Solidity: function epochUpdateCounts(uint32 epochNumber) view returns(uint256 updateCount)
func (_EpochManager *EpochManagerCallerSession) EpochUpdateCounts(epochNumber uint32) (*big.Int, error) {
	return _EpochManager.Contract.EpochUpdateCounts(&_EpochManager.CallOpts, epochNumber)
}

// GetCurrentEpoch is a free data retrieval call binding the contract method 0xb97dd9e2.
//
// Solidity: function getCurrentEpoch() view returns(uint32)
func (_EpochManager *EpochManagerCaller) GetCurrentEpoch(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "getCurrentEpoch")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// GetCurrentEpoch is a free data retrieval call binding the contract method 0xb97dd9e2.
//
// Solidity: function getCurrentEpoch() view returns(uint32)
func (_EpochManager *EpochManagerSession) GetCurrentEpoch() (uint32, error) {
	return _EpochManager.Contract.GetCurrentEpoch(&_EpochManager.CallOpts)
}

// GetCurrentEpoch is a free data retrieval call binding the contract method 0xb97dd9e2.
//
// Solidity: function getCurrentEpoch() view returns(uint32)
func (_EpochManager *EpochManagerCallerSession) GetCurrentEpoch() (uint32, error) {
	return _EpochManager.Contract.GetCurrentEpoch(&_EpochManager.CallOpts)
}

// GetEffectiveEpoch is a free data retrieval call binding the contract method 0x319c63b2.
//
// Solidity: function getEffectiveEpoch() view returns(uint32)
func (_EpochManager *EpochManagerCaller) GetEffectiveEpoch(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "getEffectiveEpoch")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// GetEffectiveEpoch is a free data retrieval call binding the contract method 0x319c63b2.
//
// Solidity: function getEffectiveEpoch() view returns(uint32)
func (_EpochManager *EpochManagerSession) GetEffectiveEpoch() (uint32, error) {
	return _EpochManager.Contract.GetEffectiveEpoch(&_EpochManager.CallOpts)
}

// GetEffectiveEpoch is a free data retrieval call binding the contract method 0x319c63b2.
//
// Solidity: function getEffectiveEpoch() view returns(uint32)
func (_EpochManager *EpochManagerCallerSession) GetEffectiveEpoch() (uint32, error) {
	return _EpochManager.Contract.GetEffectiveEpoch(&_EpochManager.CallOpts)
}

// GetEffectiveEpochForBlock is a free data retrieval call binding the contract method 0x82678ee4.
//
// Solidity: function getEffectiveEpochForBlock(uint64 blockNumber) view returns(uint32)
func (_EpochManager *EpochManagerCaller) GetEffectiveEpochForBlock(opts *bind.CallOpts, blockNumber uint64) (uint32, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "getEffectiveEpochForBlock", blockNumber)

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// GetEffectiveEpochForBlock is a free data retrieval call binding the contract method 0x82678ee4.
//
// Solidity: function getEffectiveEpochForBlock(uint64 blockNumber) view returns(uint32)
func (_EpochManager *EpochManagerSession) GetEffectiveEpochForBlock(blockNumber uint64) (uint32, error) {
	return _EpochManager.Contract.GetEffectiveEpochForBlock(&_EpochManager.CallOpts, blockNumber)
}

// GetEffectiveEpochForBlock is a free data retrieval call binding the contract method 0x82678ee4.
//
// Solidity: function getEffectiveEpochForBlock(uint64 blockNumber) view returns(uint32)
func (_EpochManager *EpochManagerCallerSession) GetEffectiveEpochForBlock(blockNumber uint64) (uint32, error) {
	return _EpochManager.Contract.GetEffectiveEpochForBlock(&_EpochManager.CallOpts, blockNumber)
}

// GetEpochInterval is a free data retrieval call binding the contract method 0xa1116c9a.
//
// Solidity: function getEpochInterval(uint32 epoch) view returns(uint64 startBlock, uint64 graceBlock, uint64 endBlock)
func (_EpochManager *EpochManagerCaller) GetEpochInterval(opts *bind.CallOpts, epoch uint32) (struct {
	StartBlock uint64
	GraceBlock uint64
	EndBlock   uint64
}, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "getEpochInterval", epoch)

	outstruct := new(struct {
		StartBlock uint64
		GraceBlock uint64
		EndBlock   uint64
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.StartBlock = *abi.ConvertType(out[0], new(uint64)).(*uint64)
	outstruct.GraceBlock = *abi.ConvertType(out[1], new(uint64)).(*uint64)
	outstruct.EndBlock = *abi.ConvertType(out[2], new(uint64)).(*uint64)

	return *outstruct, err

}

// GetEpochInterval is a free data retrieval call binding the contract method 0xa1116c9a.
//
// Solidity: function getEpochInterval(uint32 epoch) view returns(uint64 startBlock, uint64 graceBlock, uint64 endBlock)
func (_EpochManager *EpochManagerSession) GetEpochInterval(epoch uint32) (struct {
	StartBlock uint64
	GraceBlock uint64
	EndBlock   uint64
}, error) {
	return _EpochManager.Contract.GetEpochInterval(&_EpochManager.CallOpts, epoch)
}

// GetEpochInterval is a free data retrieval call binding the contract method 0xa1116c9a.
//
// Solidity: function getEpochInterval(uint32 epoch) view returns(uint64 startBlock, uint64 graceBlock, uint64 endBlock)
func (_EpochManager *EpochManagerCallerSession) GetEpochInterval(epoch uint32) (struct {
	StartBlock uint64
	GraceBlock uint64
	EndBlock   uint64
}, error) {
	return _EpochManager.Contract.GetEpochInterval(&_EpochManager.CallOpts, epoch)
}

// IsInGracePeriod is a free data retrieval call binding the contract method 0xcf27d016.
//
// Solidity: function isInGracePeriod() view returns(bool)
func (_EpochManager *EpochManagerCaller) IsInGracePeriod(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "isInGracePeriod")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsInGracePeriod is a free data retrieval call binding the contract method 0xcf27d016.
//
// Solidity: function isInGracePeriod() view returns(bool)
func (_EpochManager *EpochManagerSession) IsInGracePeriod() (bool, error) {
	return _EpochManager.Contract.IsInGracePeriod(&_EpochManager.CallOpts)
}

// IsInGracePeriod is a free data retrieval call binding the contract method 0xcf27d016.
//
// Solidity: function isInGracePeriod() view returns(bool)
func (_EpochManager *EpochManagerCallerSession) IsInGracePeriod() (bool, error) {
	return _EpochManager.Contract.IsInGracePeriod(&_EpochManager.CallOpts)
}

// IsUpdatable is a free data retrieval call binding the contract method 0xac901751.
//
// Solidity: function isUpdatable(uint32 targetEpoch) view returns(bool)
func (_EpochManager *EpochManagerCaller) IsUpdatable(opts *bind.CallOpts, targetEpoch uint32) (bool, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "isUpdatable", targetEpoch)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsUpdatable is a free data retrieval call binding the contract method 0xac901751.
//
// Solidity: function isUpdatable(uint32 targetEpoch) view returns(bool)
func (_EpochManager *EpochManagerSession) IsUpdatable(targetEpoch uint32) (bool, error) {
	return _EpochManager.Contract.IsUpdatable(&_EpochManager.CallOpts, targetEpoch)
}

// IsUpdatable is a free data retrieval call binding the contract method 0xac901751.
//
// Solidity: function isUpdatable(uint32 targetEpoch) view returns(bool)
func (_EpochManager *EpochManagerCallerSession) IsUpdatable(targetEpoch uint32) (bool, error) {
	return _EpochManager.Contract.IsUpdatable(&_EpochManager.CallOpts, targetEpoch)
}

// LastEpochBlock is a free data retrieval call binding the contract method 0xc2362dd5.
//
// Solidity: function lastEpochBlock() view returns(uint64)
func (_EpochManager *EpochManagerCaller) LastEpochBlock(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "lastEpochBlock")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// LastEpochBlock is a free data retrieval call binding the contract method 0xc2362dd5.
//
// Solidity: function lastEpochBlock() view returns(uint64)
func (_EpochManager *EpochManagerSession) LastEpochBlock() (uint64, error) {
	return _EpochManager.Contract.LastEpochBlock(&_EpochManager.CallOpts)
}

// LastEpochBlock is a free data retrieval call binding the contract method 0xc2362dd5.
//
// Solidity: function lastEpochBlock() view returns(uint64)
func (_EpochManager *EpochManagerCallerSession) LastEpochBlock() (uint64, error) {
	return _EpochManager.Contract.LastEpochBlock(&_EpochManager.CallOpts)
}

// LastUpdatedEpoch is a free data retrieval call binding the contract method 0xb7642a43.
//
// Solidity: function lastUpdatedEpoch() view returns(uint32)
func (_EpochManager *EpochManagerCaller) LastUpdatedEpoch(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "lastUpdatedEpoch")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// LastUpdatedEpoch is a free data retrieval call binding the contract method 0xb7642a43.
//
// Solidity: function lastUpdatedEpoch() view returns(uint32)
func (_EpochManager *EpochManagerSession) LastUpdatedEpoch() (uint32, error) {
	return _EpochManager.Contract.LastUpdatedEpoch(&_EpochManager.CallOpts)
}

// LastUpdatedEpoch is a free data retrieval call binding the contract method 0xb7642a43.
//
// Solidity: function lastUpdatedEpoch() view returns(uint32)
func (_EpochManager *EpochManagerCallerSession) LastUpdatedEpoch() (uint32, error) {
	return _EpochManager.Contract.LastUpdatedEpoch(&_EpochManager.CallOpts)
}

// NextEpochBlock is a free data retrieval call binding the contract method 0x00640c2e.
//
// Solidity: function nextEpochBlock() view returns(uint64)
func (_EpochManager *EpochManagerCaller) NextEpochBlock(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "nextEpochBlock")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// NextEpochBlock is a free data retrieval call binding the contract method 0x00640c2e.
//
// Solidity: function nextEpochBlock() view returns(uint64)
func (_EpochManager *EpochManagerSession) NextEpochBlock() (uint64, error) {
	return _EpochManager.Contract.NextEpochBlock(&_EpochManager.CallOpts)
}

// NextEpochBlock is a free data retrieval call binding the contract method 0x00640c2e.
//
// Solidity: function nextEpochBlock() view returns(uint64)
func (_EpochManager *EpochManagerCallerSession) NextEpochBlock() (uint64, error) {
	return _EpochManager.Contract.NextEpochBlock(&_EpochManager.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_EpochManager *EpochManagerCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _EpochManager.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_EpochManager *EpochManagerSession) Owner() (common.Address, error) {
	return _EpochManager.Contract.Owner(&_EpochManager.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_EpochManager *EpochManagerCallerSession) Owner() (common.Address, error) {
	return _EpochManager.Contract.Owner(&_EpochManager.CallOpts)
}

// AdvanceEpoch is a paid mutator transaction binding the contract method 0x3cf80e6c.
//
// Solidity: function advanceEpoch() returns()
func (_EpochManager *EpochManagerTransactor) AdvanceEpoch(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EpochManager.contract.Transact(opts, "advanceEpoch")
}

// AdvanceEpoch is a paid mutator transaction binding the contract method 0x3cf80e6c.
//
// Solidity: function advanceEpoch() returns()
func (_EpochManager *EpochManagerSession) AdvanceEpoch() (*types.Transaction, error) {
	return _EpochManager.Contract.AdvanceEpoch(&_EpochManager.TransactOpts)
}

// AdvanceEpoch is a paid mutator transaction binding the contract method 0x3cf80e6c.
//
// Solidity: function advanceEpoch() returns()
func (_EpochManager *EpochManagerTransactorSession) AdvanceEpoch() (*types.Transaction, error) {
	return _EpochManager.Contract.AdvanceEpoch(&_EpochManager.TransactOpts)
}

// QueueStateUpdate is a paid mutator transaction binding the contract method 0x1e926e0a.
//
// Solidity: function queueStateUpdate(uint8 updateType, bytes data) returns()
func (_EpochManager *EpochManagerTransactor) QueueStateUpdate(opts *bind.TransactOpts, updateType uint8, data []byte) (*types.Transaction, error) {
	return _EpochManager.contract.Transact(opts, "queueStateUpdate", updateType, data)
}

// QueueStateUpdate is a paid mutator transaction binding the contract method 0x1e926e0a.
//
// Solidity: function queueStateUpdate(uint8 updateType, bytes data) returns()
func (_EpochManager *EpochManagerSession) QueueStateUpdate(updateType uint8, data []byte) (*types.Transaction, error) {
	return _EpochManager.Contract.QueueStateUpdate(&_EpochManager.TransactOpts, updateType, data)
}

// QueueStateUpdate is a paid mutator transaction binding the contract method 0x1e926e0a.
//
// Solidity: function queueStateUpdate(uint8 updateType, bytes data) returns()
func (_EpochManager *EpochManagerTransactorSession) QueueStateUpdate(updateType uint8, data []byte) (*types.Transaction, error) {
	return _EpochManager.Contract.QueueStateUpdate(&_EpochManager.TransactOpts, updateType, data)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_EpochManager *EpochManagerTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _EpochManager.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_EpochManager *EpochManagerSession) RenounceOwnership() (*types.Transaction, error) {
	return _EpochManager.Contract.RenounceOwnership(&_EpochManager.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_EpochManager *EpochManagerTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _EpochManager.Contract.RenounceOwnership(&_EpochManager.TransactOpts)
}

// RevertEpoch is a paid mutator transaction binding the contract method 0xd2e636f8.
//
// Solidity: function revertEpoch(uint32 epoch) returns()
func (_EpochManager *EpochManagerTransactor) RevertEpoch(opts *bind.TransactOpts, epoch uint32) (*types.Transaction, error) {
	return _EpochManager.contract.Transact(opts, "revertEpoch", epoch)
}

// RevertEpoch is a paid mutator transaction binding the contract method 0xd2e636f8.
//
// Solidity: function revertEpoch(uint32 epoch) returns()
func (_EpochManager *EpochManagerSession) RevertEpoch(epoch uint32) (*types.Transaction, error) {
	return _EpochManager.Contract.RevertEpoch(&_EpochManager.TransactOpts, epoch)
}

// RevertEpoch is a paid mutator transaction binding the contract method 0xd2e636f8.
//
// Solidity: function revertEpoch(uint32 epoch) returns()
func (_EpochManager *EpochManagerTransactorSession) RevertEpoch(epoch uint32) (*types.Transaction, error) {
	return _EpochManager.Contract.RevertEpoch(&_EpochManager.TransactOpts, epoch)
}

// SendStateUpdates is a paid mutator transaction binding the contract method 0xca21380c.
//
// Solidity: function sendStateUpdates(uint256 chainId) payable returns()
func (_EpochManager *EpochManagerTransactor) SendStateUpdates(opts *bind.TransactOpts, chainId *big.Int) (*types.Transaction, error) {
	return _EpochManager.contract.Transact(opts, "sendStateUpdates", chainId)
}

// SendStateUpdates is a paid mutator transaction binding the contract method 0xca21380c.
//
// Solidity: function sendStateUpdates(uint256 chainId) payable returns()
func (_EpochManager *EpochManagerSession) SendStateUpdates(chainId *big.Int) (*types.Transaction, error) {
	return _EpochManager.Contract.SendStateUpdates(&_EpochManager.TransactOpts, chainId)
}

// SendStateUpdates is a paid mutator transaction binding the contract method 0xca21380c.
//
// Solidity: function sendStateUpdates(uint256 chainId) payable returns()
func (_EpochManager *EpochManagerTransactorSession) SendStateUpdates(chainId *big.Int) (*types.Transaction, error) {
	return _EpochManager.Contract.SendStateUpdates(&_EpochManager.TransactOpts, chainId)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_EpochManager *EpochManagerTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _EpochManager.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_EpochManager *EpochManagerSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _EpochManager.Contract.TransferOwnership(&_EpochManager.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_EpochManager *EpochManagerTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _EpochManager.Contract.TransferOwnership(&_EpochManager.TransactOpts, newOwner)
}

// EpochManagerEpochAdvancedIterator is returned from FilterEpochAdvanced and is used to iterate over the raw logs and unpacked data for EpochAdvanced events raised by the EpochManager contract.
type EpochManagerEpochAdvancedIterator struct {
	Event *EpochManagerEpochAdvanced // Event containing the contract specifics and raw log

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
func (it *EpochManagerEpochAdvancedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EpochManagerEpochAdvanced)
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
		it.Event = new(EpochManagerEpochAdvanced)
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
func (it *EpochManagerEpochAdvancedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EpochManagerEpochAdvancedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EpochManagerEpochAdvanced represents a EpochAdvanced event raised by the EpochManager contract.
type EpochManagerEpochAdvanced struct {
	Epoch          uint32
	LastEpochBlock uint64
	NextEpochBlock uint64
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterEpochAdvanced is a free log retrieval operation binding the contract event 0xa0069ea59ac3b1a224e81ad8293fa7007c1c633d8441a9a75d52b519e39db74e.
//
// Solidity: event EpochAdvanced(uint32 indexed epoch, uint64 lastEpochBlock, uint64 nextEpochBlock)
func (_EpochManager *EpochManagerFilterer) FilterEpochAdvanced(opts *bind.FilterOpts, epoch []uint32) (*EpochManagerEpochAdvancedIterator, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}

	logs, sub, err := _EpochManager.contract.FilterLogs(opts, "EpochAdvanced", epochRule)
	if err != nil {
		return nil, err
	}
	return &EpochManagerEpochAdvancedIterator{contract: _EpochManager.contract, event: "EpochAdvanced", logs: logs, sub: sub}, nil
}

// WatchEpochAdvanced is a free log subscription operation binding the contract event 0xa0069ea59ac3b1a224e81ad8293fa7007c1c633d8441a9a75d52b519e39db74e.
//
// Solidity: event EpochAdvanced(uint32 indexed epoch, uint64 lastEpochBlock, uint64 nextEpochBlock)
func (_EpochManager *EpochManagerFilterer) WatchEpochAdvanced(opts *bind.WatchOpts, sink chan<- *EpochManagerEpochAdvanced, epoch []uint32) (event.Subscription, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}

	logs, sub, err := _EpochManager.contract.WatchLogs(opts, "EpochAdvanced", epochRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EpochManagerEpochAdvanced)
				if err := _EpochManager.contract.UnpackLog(event, "EpochAdvanced", log); err != nil {
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

// ParseEpochAdvanced is a log parse operation binding the contract event 0xa0069ea59ac3b1a224e81ad8293fa7007c1c633d8441a9a75d52b519e39db74e.
//
// Solidity: event EpochAdvanced(uint32 indexed epoch, uint64 lastEpochBlock, uint64 nextEpochBlock)
func (_EpochManager *EpochManagerFilterer) ParseEpochAdvanced(log types.Log) (*EpochManagerEpochAdvanced, error) {
	event := new(EpochManagerEpochAdvanced)
	if err := _EpochManager.contract.UnpackLog(event, "EpochAdvanced", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EpochManagerEpochRevertedIterator is returned from FilterEpochReverted and is used to iterate over the raw logs and unpacked data for EpochReverted events raised by the EpochManager contract.
type EpochManagerEpochRevertedIterator struct {
	Event *EpochManagerEpochReverted // Event containing the contract specifics and raw log

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
func (it *EpochManagerEpochRevertedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EpochManagerEpochReverted)
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
		it.Event = new(EpochManagerEpochReverted)
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
func (it *EpochManagerEpochRevertedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EpochManagerEpochRevertedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EpochManagerEpochReverted represents a EpochReverted event raised by the EpochManager contract.
type EpochManagerEpochReverted struct {
	Epoch uint32
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterEpochReverted is a free log retrieval operation binding the contract event 0x4ac4edeb96958a51f320fe9e4d4347306818810aae96fe85d22a9dea7cd4ec70.
//
// Solidity: event EpochReverted(uint32 indexed epoch)
func (_EpochManager *EpochManagerFilterer) FilterEpochReverted(opts *bind.FilterOpts, epoch []uint32) (*EpochManagerEpochRevertedIterator, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}

	logs, sub, err := _EpochManager.contract.FilterLogs(opts, "EpochReverted", epochRule)
	if err != nil {
		return nil, err
	}
	return &EpochManagerEpochRevertedIterator{contract: _EpochManager.contract, event: "EpochReverted", logs: logs, sub: sub}, nil
}

// WatchEpochReverted is a free log subscription operation binding the contract event 0x4ac4edeb96958a51f320fe9e4d4347306818810aae96fe85d22a9dea7cd4ec70.
//
// Solidity: event EpochReverted(uint32 indexed epoch)
func (_EpochManager *EpochManagerFilterer) WatchEpochReverted(opts *bind.WatchOpts, sink chan<- *EpochManagerEpochReverted, epoch []uint32) (event.Subscription, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}

	logs, sub, err := _EpochManager.contract.WatchLogs(opts, "EpochReverted", epochRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EpochManagerEpochReverted)
				if err := _EpochManager.contract.UnpackLog(event, "EpochReverted", log); err != nil {
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

// ParseEpochReverted is a log parse operation binding the contract event 0x4ac4edeb96958a51f320fe9e4d4347306818810aae96fe85d22a9dea7cd4ec70.
//
// Solidity: event EpochReverted(uint32 indexed epoch)
func (_EpochManager *EpochManagerFilterer) ParseEpochReverted(log types.Log) (*EpochManagerEpochReverted, error) {
	event := new(EpochManagerEpochReverted)
	if err := _EpochManager.contract.UnpackLog(event, "EpochReverted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EpochManagerInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the EpochManager contract.
type EpochManagerInitializedIterator struct {
	Event *EpochManagerInitialized // Event containing the contract specifics and raw log

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
func (it *EpochManagerInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EpochManagerInitialized)
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
		it.Event = new(EpochManagerInitialized)
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
func (it *EpochManagerInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EpochManagerInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EpochManagerInitialized represents a Initialized event raised by the EpochManager contract.
type EpochManagerInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_EpochManager *EpochManagerFilterer) FilterInitialized(opts *bind.FilterOpts) (*EpochManagerInitializedIterator, error) {

	logs, sub, err := _EpochManager.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &EpochManagerInitializedIterator{contract: _EpochManager.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_EpochManager *EpochManagerFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *EpochManagerInitialized) (event.Subscription, error) {

	logs, sub, err := _EpochManager.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EpochManagerInitialized)
				if err := _EpochManager.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_EpochManager *EpochManagerFilterer) ParseInitialized(log types.Log) (*EpochManagerInitialized, error) {
	event := new(EpochManagerInitialized)
	if err := _EpochManager.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EpochManagerOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the EpochManager contract.
type EpochManagerOwnershipTransferredIterator struct {
	Event *EpochManagerOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *EpochManagerOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EpochManagerOwnershipTransferred)
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
		it.Event = new(EpochManagerOwnershipTransferred)
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
func (it *EpochManagerOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EpochManagerOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EpochManagerOwnershipTransferred represents a OwnershipTransferred event raised by the EpochManager contract.
type EpochManagerOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_EpochManager *EpochManagerFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*EpochManagerOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _EpochManager.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &EpochManagerOwnershipTransferredIterator{contract: _EpochManager.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_EpochManager *EpochManagerFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *EpochManagerOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _EpochManager.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EpochManagerOwnershipTransferred)
				if err := _EpochManager.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_EpochManager *EpochManagerFilterer) ParseOwnershipTransferred(log types.Log) (*EpochManagerOwnershipTransferred, error) {
	event := new(EpochManagerOwnershipTransferred)
	if err := _EpochManager.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EpochManagerStateUpdateQueuedIterator is returned from FilterStateUpdateQueued and is used to iterate over the raw logs and unpacked data for StateUpdateQueued events raised by the EpochManager contract.
type EpochManagerStateUpdateQueuedIterator struct {
	Event *EpochManagerStateUpdateQueued // Event containing the contract specifics and raw log

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
func (it *EpochManagerStateUpdateQueuedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EpochManagerStateUpdateQueued)
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
		it.Event = new(EpochManagerStateUpdateQueued)
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
func (it *EpochManagerStateUpdateQueuedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EpochManagerStateUpdateQueuedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EpochManagerStateUpdateQueued represents a StateUpdateQueued event raised by the EpochManager contract.
type EpochManagerStateUpdateQueued struct {
	Epoch      uint32
	UpdateType uint8
	Data       []byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterStateUpdateQueued is a free log retrieval operation binding the contract event 0x159625a0a529e2e14aeae1e0f0b54d82481734a584ff2017e1083c8fb52b2855.
//
// Solidity: event StateUpdateQueued(uint32 indexed epoch, uint8 updateType, bytes data)
func (_EpochManager *EpochManagerFilterer) FilterStateUpdateQueued(opts *bind.FilterOpts, epoch []uint32) (*EpochManagerStateUpdateQueuedIterator, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}

	logs, sub, err := _EpochManager.contract.FilterLogs(opts, "StateUpdateQueued", epochRule)
	if err != nil {
		return nil, err
	}
	return &EpochManagerStateUpdateQueuedIterator{contract: _EpochManager.contract, event: "StateUpdateQueued", logs: logs, sub: sub}, nil
}

// WatchStateUpdateQueued is a free log subscription operation binding the contract event 0x159625a0a529e2e14aeae1e0f0b54d82481734a584ff2017e1083c8fb52b2855.
//
// Solidity: event StateUpdateQueued(uint32 indexed epoch, uint8 updateType, bytes data)
func (_EpochManager *EpochManagerFilterer) WatchStateUpdateQueued(opts *bind.WatchOpts, sink chan<- *EpochManagerStateUpdateQueued, epoch []uint32) (event.Subscription, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}

	logs, sub, err := _EpochManager.contract.WatchLogs(opts, "StateUpdateQueued", epochRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EpochManagerStateUpdateQueued)
				if err := _EpochManager.contract.UnpackLog(event, "StateUpdateQueued", log); err != nil {
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

// ParseStateUpdateQueued is a log parse operation binding the contract event 0x159625a0a529e2e14aeae1e0f0b54d82481734a584ff2017e1083c8fb52b2855.
//
// Solidity: event StateUpdateQueued(uint32 indexed epoch, uint8 updateType, bytes data)
func (_EpochManager *EpochManagerFilterer) ParseStateUpdateQueued(log types.Log) (*EpochManagerStateUpdateQueued, error) {
	event := new(EpochManagerStateUpdateQueued)
	if err := _EpochManager.contract.UnpackLog(event, "StateUpdateQueued", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EpochManagerStateUpdatesSentIterator is returned from FilterStateUpdatesSent and is used to iterate over the raw logs and unpacked data for StateUpdatesSent events raised by the EpochManager contract.
type EpochManagerStateUpdatesSentIterator struct {
	Event *EpochManagerStateUpdatesSent // Event containing the contract specifics and raw log

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
func (it *EpochManagerStateUpdatesSentIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EpochManagerStateUpdatesSent)
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
		it.Event = new(EpochManagerStateUpdatesSent)
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
func (it *EpochManagerStateUpdatesSentIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EpochManagerStateUpdatesSentIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EpochManagerStateUpdatesSent represents a StateUpdatesSent event raised by the EpochManager contract.
type EpochManagerStateUpdatesSent struct {
	Epoch        uint32
	UpdatesCount *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterStateUpdatesSent is a free log retrieval operation binding the contract event 0xb1f32c40ed2f79eca2a73667dd6bfe460fc3cdf32e4d10c539c97e60a4c1487d.
//
// Solidity: event StateUpdatesSent(uint32 indexed epoch, uint256 updatesCount)
func (_EpochManager *EpochManagerFilterer) FilterStateUpdatesSent(opts *bind.FilterOpts, epoch []uint32) (*EpochManagerStateUpdatesSentIterator, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}

	logs, sub, err := _EpochManager.contract.FilterLogs(opts, "StateUpdatesSent", epochRule)
	if err != nil {
		return nil, err
	}
	return &EpochManagerStateUpdatesSentIterator{contract: _EpochManager.contract, event: "StateUpdatesSent", logs: logs, sub: sub}, nil
}

// WatchStateUpdatesSent is a free log subscription operation binding the contract event 0xb1f32c40ed2f79eca2a73667dd6bfe460fc3cdf32e4d10c539c97e60a4c1487d.
//
// Solidity: event StateUpdatesSent(uint32 indexed epoch, uint256 updatesCount)
func (_EpochManager *EpochManagerFilterer) WatchStateUpdatesSent(opts *bind.WatchOpts, sink chan<- *EpochManagerStateUpdatesSent, epoch []uint32) (event.Subscription, error) {

	var epochRule []interface{}
	for _, epochItem := range epoch {
		epochRule = append(epochRule, epochItem)
	}

	logs, sub, err := _EpochManager.contract.WatchLogs(opts, "StateUpdatesSent", epochRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EpochManagerStateUpdatesSent)
				if err := _EpochManager.contract.UnpackLog(event, "StateUpdatesSent", log); err != nil {
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

// ParseStateUpdatesSent is a log parse operation binding the contract event 0xb1f32c40ed2f79eca2a73667dd6bfe460fc3cdf32e4d10c539c97e60a4c1487d.
//
// Solidity: event StateUpdatesSent(uint32 indexed epoch, uint256 updatesCount)
func (_EpochManager *EpochManagerFilterer) ParseStateUpdatesSent(log types.Log) (*EpochManagerStateUpdatesSent, error) {
	event := new(EpochManagerStateUpdatesSent)
	if err := _EpochManager.contract.UnpackLog(event, "StateUpdatesSent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
