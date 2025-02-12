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

// ISignatureUtilsSignatureWithSaltAndExpiry is an auto generated low-level Go binding around an user-defined struct.
type ISignatureUtilsSignatureWithSaltAndExpiry struct {
	Signature []byte
	Salt      [32]byte
	Expiry    *big.Int
}

// Quorum is an auto generated low-level Go binding around an user-defined struct.
type Quorum struct {
	Strategies []StrategyParams
}

// StrategyParams is an auto generated low-level Go binding around an user-defined struct.
type StrategyParams struct {
	Strategy   common.Address
	Multiplier *big.Int
}

// ECDSAStakeRegistryMetaData contains all meta data concerning the ECDSAStakeRegistry contract.
var ECDSAStakeRegistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_delegationManager\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_epochManager\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_serviceManager\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"deregisterOperator\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getLastCheckpointOperatorWeight\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getLastCheckpointThresholdWeight\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getLastCheckpointTotalWeight\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getLastestOperatorP2pKey\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getLastestOperatorSigningKey\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorP2pKeyAtEpoch\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_epochNumber\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorSigningKeyAtEpoch\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_epochNumber\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorWeight\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorWeightAtEpoch\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_epochNumber\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getThresholdStake\",\"inputs\":[{\"name\":\"_referenceEpoch\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getThresholdWeightAtEpoch\",\"inputs\":[{\"name\":\"_epochNumber\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTotalWeightAtEpoch\",\"inputs\":[{\"name\":\"_epochNumber\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"_thresholdWeight\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"quorumParams\",\"type\":\"tuple\",\"internalType\":\"structQuorum\",\"components\":[{\"name\":\"strategies\",\"type\":\"tuple[]\",\"internalType\":\"structStrategyParams[]\",\"components\":[{\"name\":\"strategy\",\"type\":\"address\",\"internalType\":\"contractIStrategy\"},{\"name\":\"multiplier\",\"type\":\"uint96\",\"internalType\":\"uint96\"}]}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"isValidSignature\",\"inputs\":[{\"name\":\"_dataHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"_signatureData\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes4\",\"internalType\":\"bytes4\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"minimumWeight\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorRegistered\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"quorum\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structQuorum\",\"components\":[{\"name\":\"strategies\",\"type\":\"tuple[]\",\"internalType\":\"structStrategyParams[]\",\"components\":[{\"name\":\"strategy\",\"type\":\"address\",\"internalType\":\"contractIStrategy\"},{\"name\":\"multiplier\",\"type\":\"uint96\",\"internalType\":\"uint96\"}]}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerOperatorWithSignature\",\"inputs\":[{\"name\":\"_operatorSignature\",\"type\":\"tuple\",\"internalType\":\"structISignatureUtils.SignatureWithSaltAndExpiry\",\"components\":[{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"expiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"_signingKey\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_p2pKey\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateMinimumWeight\",\"inputs\":[{\"name\":\"_newMinimumWeight\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_operators\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateOperatorP2pKey\",\"inputs\":[{\"name\":\"_newP2pKey\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateOperatorSigningKey\",\"inputs\":[{\"name\":\"_newSigningKey\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateOperators\",\"inputs\":[{\"name\":\"_operators\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateOperatorsForQuorum\",\"inputs\":[{\"name\":\"operatorsPerQuorum\",\"type\":\"address[][]\",\"internalType\":\"address[][]\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateQuorumConfig\",\"inputs\":[{\"name\":\"newQuorumConfig\",\"type\":\"tuple\",\"internalType\":\"structQuorum\",\"components\":[{\"name\":\"strategies\",\"type\":\"tuple[]\",\"internalType\":\"structStrategyParams[]\",\"components\":[{\"name\":\"strategy\",\"type\":\"address\",\"internalType\":\"contractIStrategy\"},{\"name\":\"multiplier\",\"type\":\"uint96\",\"internalType\":\"uint96\"}]}]},{\"name\":\"_operators\",\"type\":\"address[]\",\"internalType\":\"address[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateStakeThreshold\",\"inputs\":[{\"name\":\"_thresholdWeight\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"uint8\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MinimumWeightUpdated\",\"inputs\":[{\"name\":\"_old\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"_new\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorDeregistered\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"blockNumber\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"_avs\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorRegistered\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"blockNumber\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"_p2pKey\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"_signingKey\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"_avs\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorWeightUpdated\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"oldWeight\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newWeight\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"P2pKeyUpdate\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newP2pKey\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"oldP2pKey\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"QuorumUpdated\",\"inputs\":[{\"name\":\"_old\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structQuorum\",\"components\":[{\"name\":\"strategies\",\"type\":\"tuple[]\",\"internalType\":\"structStrategyParams[]\",\"components\":[{\"name\":\"strategy\",\"type\":\"address\",\"internalType\":\"contractIStrategy\"},{\"name\":\"multiplier\",\"type\":\"uint96\",\"internalType\":\"uint96\"}]}]},{\"name\":\"_new\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structQuorum\",\"components\":[{\"name\":\"strategies\",\"type\":\"tuple[]\",\"internalType\":\"structStrategyParams[]\",\"components\":[{\"name\":\"strategy\",\"type\":\"address\",\"internalType\":\"contractIStrategy\"},{\"name\":\"multiplier\",\"type\":\"uint96\",\"internalType\":\"uint96\"}]}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SigningKeyUpdate\",\"inputs\":[{\"name\":\"operator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newSigningKey\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"oldSigningKey\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ThresholdWeightUpdated\",\"inputs\":[{\"name\":\"_thresholdWeight\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TotalWeightUpdated\",\"inputs\":[{\"name\":\"oldTotalWeight\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newTotalWeight\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UpdateMinimumWeight\",\"inputs\":[{\"name\":\"oldMinimumWeight\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newMinimumWeight\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"InsufficientSignedStake\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientWeight\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidEpoch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidEpoch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidLength\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidQuorum\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidReferenceBlock\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidReferenceEpoch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidSignedWeight\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidThreshold\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"LengthMismatch\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"MustUpdateAllOperators\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotSorted\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorAlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OperatorNotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SigningKeyAlreadyExists\",\"inputs\":[]}]",
}

// ECDSAStakeRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use ECDSAStakeRegistryMetaData.ABI instead.
var ECDSAStakeRegistryABI = ECDSAStakeRegistryMetaData.ABI

// ECDSAStakeRegistry is an auto generated Go binding around an Ethereum contract.
type ECDSAStakeRegistry struct {
	ECDSAStakeRegistryCaller     // Read-only binding to the contract
	ECDSAStakeRegistryTransactor // Write-only binding to the contract
	ECDSAStakeRegistryFilterer   // Log filterer for contract events
}

// ECDSAStakeRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type ECDSAStakeRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECDSAStakeRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ECDSAStakeRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECDSAStakeRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ECDSAStakeRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECDSAStakeRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ECDSAStakeRegistrySession struct {
	Contract     *ECDSAStakeRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ECDSAStakeRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ECDSAStakeRegistryCallerSession struct {
	Contract *ECDSAStakeRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// ECDSAStakeRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ECDSAStakeRegistryTransactorSession struct {
	Contract     *ECDSAStakeRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// ECDSAStakeRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type ECDSAStakeRegistryRaw struct {
	Contract *ECDSAStakeRegistry // Generic contract binding to access the raw methods on
}

// ECDSAStakeRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ECDSAStakeRegistryCallerRaw struct {
	Contract *ECDSAStakeRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// ECDSAStakeRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ECDSAStakeRegistryTransactorRaw struct {
	Contract *ECDSAStakeRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewECDSAStakeRegistry creates a new instance of ECDSAStakeRegistry, bound to a specific deployed contract.
func NewECDSAStakeRegistry(address common.Address, backend bind.ContractBackend) (*ECDSAStakeRegistry, error) {
	contract, err := bindECDSAStakeRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistry{ECDSAStakeRegistryCaller: ECDSAStakeRegistryCaller{contract: contract}, ECDSAStakeRegistryTransactor: ECDSAStakeRegistryTransactor{contract: contract}, ECDSAStakeRegistryFilterer: ECDSAStakeRegistryFilterer{contract: contract}}, nil
}

// NewECDSAStakeRegistryCaller creates a new read-only instance of ECDSAStakeRegistry, bound to a specific deployed contract.
func NewECDSAStakeRegistryCaller(address common.Address, caller bind.ContractCaller) (*ECDSAStakeRegistryCaller, error) {
	contract, err := bindECDSAStakeRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistryCaller{contract: contract}, nil
}

// NewECDSAStakeRegistryTransactor creates a new write-only instance of ECDSAStakeRegistry, bound to a specific deployed contract.
func NewECDSAStakeRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*ECDSAStakeRegistryTransactor, error) {
	contract, err := bindECDSAStakeRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistryTransactor{contract: contract}, nil
}

// NewECDSAStakeRegistryFilterer creates a new log filterer instance of ECDSAStakeRegistry, bound to a specific deployed contract.
func NewECDSAStakeRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*ECDSAStakeRegistryFilterer, error) {
	contract, err := bindECDSAStakeRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistryFilterer{contract: contract}, nil
}

// bindECDSAStakeRegistry binds a generic wrapper to an already deployed contract.
func bindECDSAStakeRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ECDSAStakeRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ECDSAStakeRegistry *ECDSAStakeRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ECDSAStakeRegistry.Contract.ECDSAStakeRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ECDSAStakeRegistry *ECDSAStakeRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.ECDSAStakeRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ECDSAStakeRegistry *ECDSAStakeRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.ECDSAStakeRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ECDSAStakeRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.contract.Transact(opts, method, params...)
}

// GetLastCheckpointOperatorWeight is a free data retrieval call binding the contract method 0x3b242e4a.
//
// Solidity: function getLastCheckpointOperatorWeight(address _operator) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) GetLastCheckpointOperatorWeight(opts *bind.CallOpts, _operator common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "getLastCheckpointOperatorWeight", _operator)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLastCheckpointOperatorWeight is a free data retrieval call binding the contract method 0x3b242e4a.
//
// Solidity: function getLastCheckpointOperatorWeight(address _operator) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) GetLastCheckpointOperatorWeight(_operator common.Address) (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetLastCheckpointOperatorWeight(&_ECDSAStakeRegistry.CallOpts, _operator)
}

// GetLastCheckpointOperatorWeight is a free data retrieval call binding the contract method 0x3b242e4a.
//
// Solidity: function getLastCheckpointOperatorWeight(address _operator) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) GetLastCheckpointOperatorWeight(_operator common.Address) (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetLastCheckpointOperatorWeight(&_ECDSAStakeRegistry.CallOpts, _operator)
}

// GetLastCheckpointThresholdWeight is a free data retrieval call binding the contract method 0xb933fa74.
//
// Solidity: function getLastCheckpointThresholdWeight() view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) GetLastCheckpointThresholdWeight(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "getLastCheckpointThresholdWeight")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLastCheckpointThresholdWeight is a free data retrieval call binding the contract method 0xb933fa74.
//
// Solidity: function getLastCheckpointThresholdWeight() view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) GetLastCheckpointThresholdWeight() (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetLastCheckpointThresholdWeight(&_ECDSAStakeRegistry.CallOpts)
}

// GetLastCheckpointThresholdWeight is a free data retrieval call binding the contract method 0xb933fa74.
//
// Solidity: function getLastCheckpointThresholdWeight() view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) GetLastCheckpointThresholdWeight() (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetLastCheckpointThresholdWeight(&_ECDSAStakeRegistry.CallOpts)
}

// GetLastCheckpointTotalWeight is a free data retrieval call binding the contract method 0x314f3a49.
//
// Solidity: function getLastCheckpointTotalWeight() view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) GetLastCheckpointTotalWeight(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "getLastCheckpointTotalWeight")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLastCheckpointTotalWeight is a free data retrieval call binding the contract method 0x314f3a49.
//
// Solidity: function getLastCheckpointTotalWeight() view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) GetLastCheckpointTotalWeight() (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetLastCheckpointTotalWeight(&_ECDSAStakeRegistry.CallOpts)
}

// GetLastCheckpointTotalWeight is a free data retrieval call binding the contract method 0x314f3a49.
//
// Solidity: function getLastCheckpointTotalWeight() view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) GetLastCheckpointTotalWeight() (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetLastCheckpointTotalWeight(&_ECDSAStakeRegistry.CallOpts)
}

// GetLastestOperatorP2pKey is a free data retrieval call binding the contract method 0xfc53b930.
//
// Solidity: function getLastestOperatorP2pKey(address _operator) view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) GetLastestOperatorP2pKey(opts *bind.CallOpts, _operator common.Address) (common.Address, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "getLastestOperatorP2pKey", _operator)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetLastestOperatorP2pKey is a free data retrieval call binding the contract method 0xfc53b930.
//
// Solidity: function getLastestOperatorP2pKey(address _operator) view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) GetLastestOperatorP2pKey(_operator common.Address) (common.Address, error) {
	return _ECDSAStakeRegistry.Contract.GetLastestOperatorP2pKey(&_ECDSAStakeRegistry.CallOpts, _operator)
}

// GetLastestOperatorP2pKey is a free data retrieval call binding the contract method 0xfc53b930.
//
// Solidity: function getLastestOperatorP2pKey(address _operator) view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) GetLastestOperatorP2pKey(_operator common.Address) (common.Address, error) {
	return _ECDSAStakeRegistry.Contract.GetLastestOperatorP2pKey(&_ECDSAStakeRegistry.CallOpts, _operator)
}

// GetLastestOperatorSigningKey is a free data retrieval call binding the contract method 0xcdcd3581.
//
// Solidity: function getLastestOperatorSigningKey(address _operator) view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) GetLastestOperatorSigningKey(opts *bind.CallOpts, _operator common.Address) (common.Address, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "getLastestOperatorSigningKey", _operator)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetLastestOperatorSigningKey is a free data retrieval call binding the contract method 0xcdcd3581.
//
// Solidity: function getLastestOperatorSigningKey(address _operator) view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) GetLastestOperatorSigningKey(_operator common.Address) (common.Address, error) {
	return _ECDSAStakeRegistry.Contract.GetLastestOperatorSigningKey(&_ECDSAStakeRegistry.CallOpts, _operator)
}

// GetLastestOperatorSigningKey is a free data retrieval call binding the contract method 0xcdcd3581.
//
// Solidity: function getLastestOperatorSigningKey(address _operator) view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) GetLastestOperatorSigningKey(_operator common.Address) (common.Address, error) {
	return _ECDSAStakeRegistry.Contract.GetLastestOperatorSigningKey(&_ECDSAStakeRegistry.CallOpts, _operator)
}

// GetOperatorP2pKeyAtEpoch is a free data retrieval call binding the contract method 0x6158fbb1.
//
// Solidity: function getOperatorP2pKeyAtEpoch(address _operator, uint32 _epochNumber) view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) GetOperatorP2pKeyAtEpoch(opts *bind.CallOpts, _operator common.Address, _epochNumber uint32) (common.Address, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "getOperatorP2pKeyAtEpoch", _operator, _epochNumber)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetOperatorP2pKeyAtEpoch is a free data retrieval call binding the contract method 0x6158fbb1.
//
// Solidity: function getOperatorP2pKeyAtEpoch(address _operator, uint32 _epochNumber) view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) GetOperatorP2pKeyAtEpoch(_operator common.Address, _epochNumber uint32) (common.Address, error) {
	return _ECDSAStakeRegistry.Contract.GetOperatorP2pKeyAtEpoch(&_ECDSAStakeRegistry.CallOpts, _operator, _epochNumber)
}

// GetOperatorP2pKeyAtEpoch is a free data retrieval call binding the contract method 0x6158fbb1.
//
// Solidity: function getOperatorP2pKeyAtEpoch(address _operator, uint32 _epochNumber) view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) GetOperatorP2pKeyAtEpoch(_operator common.Address, _epochNumber uint32) (common.Address, error) {
	return _ECDSAStakeRegistry.Contract.GetOperatorP2pKeyAtEpoch(&_ECDSAStakeRegistry.CallOpts, _operator, _epochNumber)
}

// GetOperatorSigningKeyAtEpoch is a free data retrieval call binding the contract method 0x63125d32.
//
// Solidity: function getOperatorSigningKeyAtEpoch(address _operator, uint32 _epochNumber) view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) GetOperatorSigningKeyAtEpoch(opts *bind.CallOpts, _operator common.Address, _epochNumber uint32) (common.Address, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "getOperatorSigningKeyAtEpoch", _operator, _epochNumber)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetOperatorSigningKeyAtEpoch is a free data retrieval call binding the contract method 0x63125d32.
//
// Solidity: function getOperatorSigningKeyAtEpoch(address _operator, uint32 _epochNumber) view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) GetOperatorSigningKeyAtEpoch(_operator common.Address, _epochNumber uint32) (common.Address, error) {
	return _ECDSAStakeRegistry.Contract.GetOperatorSigningKeyAtEpoch(&_ECDSAStakeRegistry.CallOpts, _operator, _epochNumber)
}

// GetOperatorSigningKeyAtEpoch is a free data retrieval call binding the contract method 0x63125d32.
//
// Solidity: function getOperatorSigningKeyAtEpoch(address _operator, uint32 _epochNumber) view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) GetOperatorSigningKeyAtEpoch(_operator common.Address, _epochNumber uint32) (common.Address, error) {
	return _ECDSAStakeRegistry.Contract.GetOperatorSigningKeyAtEpoch(&_ECDSAStakeRegistry.CallOpts, _operator, _epochNumber)
}

// GetOperatorWeight is a free data retrieval call binding the contract method 0x98ec1ac9.
//
// Solidity: function getOperatorWeight(address _operator) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) GetOperatorWeight(opts *bind.CallOpts, _operator common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "getOperatorWeight", _operator)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOperatorWeight is a free data retrieval call binding the contract method 0x98ec1ac9.
//
// Solidity: function getOperatorWeight(address _operator) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) GetOperatorWeight(_operator common.Address) (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetOperatorWeight(&_ECDSAStakeRegistry.CallOpts, _operator)
}

// GetOperatorWeight is a free data retrieval call binding the contract method 0x98ec1ac9.
//
// Solidity: function getOperatorWeight(address _operator) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) GetOperatorWeight(_operator common.Address) (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetOperatorWeight(&_ECDSAStakeRegistry.CallOpts, _operator)
}

// GetOperatorWeightAtEpoch is a free data retrieval call binding the contract method 0x338a0b2a.
//
// Solidity: function getOperatorWeightAtEpoch(address _operator, uint32 _epochNumber) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) GetOperatorWeightAtEpoch(opts *bind.CallOpts, _operator common.Address, _epochNumber uint32) (*big.Int, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "getOperatorWeightAtEpoch", _operator, _epochNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOperatorWeightAtEpoch is a free data retrieval call binding the contract method 0x338a0b2a.
//
// Solidity: function getOperatorWeightAtEpoch(address _operator, uint32 _epochNumber) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) GetOperatorWeightAtEpoch(_operator common.Address, _epochNumber uint32) (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetOperatorWeightAtEpoch(&_ECDSAStakeRegistry.CallOpts, _operator, _epochNumber)
}

// GetOperatorWeightAtEpoch is a free data retrieval call binding the contract method 0x338a0b2a.
//
// Solidity: function getOperatorWeightAtEpoch(address _operator, uint32 _epochNumber) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) GetOperatorWeightAtEpoch(_operator common.Address, _epochNumber uint32) (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetOperatorWeightAtEpoch(&_ECDSAStakeRegistry.CallOpts, _operator, _epochNumber)
}

// GetThresholdStake is a free data retrieval call binding the contract method 0xb208cd38.
//
// Solidity: function getThresholdStake(uint32 _referenceEpoch) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) GetThresholdStake(opts *bind.CallOpts, _referenceEpoch uint32) (*big.Int, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "getThresholdStake", _referenceEpoch)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetThresholdStake is a free data retrieval call binding the contract method 0xb208cd38.
//
// Solidity: function getThresholdStake(uint32 _referenceEpoch) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) GetThresholdStake(_referenceEpoch uint32) (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetThresholdStake(&_ECDSAStakeRegistry.CallOpts, _referenceEpoch)
}

// GetThresholdStake is a free data retrieval call binding the contract method 0xb208cd38.
//
// Solidity: function getThresholdStake(uint32 _referenceEpoch) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) GetThresholdStake(_referenceEpoch uint32) (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetThresholdStake(&_ECDSAStakeRegistry.CallOpts, _referenceEpoch)
}

// GetThresholdWeightAtEpoch is a free data retrieval call binding the contract method 0xbb12059f.
//
// Solidity: function getThresholdWeightAtEpoch(uint32 _epochNumber) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) GetThresholdWeightAtEpoch(opts *bind.CallOpts, _epochNumber uint32) (*big.Int, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "getThresholdWeightAtEpoch", _epochNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetThresholdWeightAtEpoch is a free data retrieval call binding the contract method 0xbb12059f.
//
// Solidity: function getThresholdWeightAtEpoch(uint32 _epochNumber) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) GetThresholdWeightAtEpoch(_epochNumber uint32) (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetThresholdWeightAtEpoch(&_ECDSAStakeRegistry.CallOpts, _epochNumber)
}

// GetThresholdWeightAtEpoch is a free data retrieval call binding the contract method 0xbb12059f.
//
// Solidity: function getThresholdWeightAtEpoch(uint32 _epochNumber) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) GetThresholdWeightAtEpoch(_epochNumber uint32) (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetThresholdWeightAtEpoch(&_ECDSAStakeRegistry.CallOpts, _epochNumber)
}

// GetTotalWeightAtEpoch is a free data retrieval call binding the contract method 0xa440d6f8.
//
// Solidity: function getTotalWeightAtEpoch(uint32 _epochNumber) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) GetTotalWeightAtEpoch(opts *bind.CallOpts, _epochNumber uint32) (*big.Int, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "getTotalWeightAtEpoch", _epochNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalWeightAtEpoch is a free data retrieval call binding the contract method 0xa440d6f8.
//
// Solidity: function getTotalWeightAtEpoch(uint32 _epochNumber) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) GetTotalWeightAtEpoch(_epochNumber uint32) (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetTotalWeightAtEpoch(&_ECDSAStakeRegistry.CallOpts, _epochNumber)
}

// GetTotalWeightAtEpoch is a free data retrieval call binding the contract method 0xa440d6f8.
//
// Solidity: function getTotalWeightAtEpoch(uint32 _epochNumber) view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) GetTotalWeightAtEpoch(_epochNumber uint32) (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.GetTotalWeightAtEpoch(&_ECDSAStakeRegistry.CallOpts, _epochNumber)
}

// IsValidSignature is a free data retrieval call binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 _dataHash, bytes _signatureData) view returns(bytes4)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) IsValidSignature(opts *bind.CallOpts, _dataHash [32]byte, _signatureData []byte) ([4]byte, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "isValidSignature", _dataHash, _signatureData)

	if err != nil {
		return *new([4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([4]byte)).(*[4]byte)

	return out0, err

}

// IsValidSignature is a free data retrieval call binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 _dataHash, bytes _signatureData) view returns(bytes4)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) IsValidSignature(_dataHash [32]byte, _signatureData []byte) ([4]byte, error) {
	return _ECDSAStakeRegistry.Contract.IsValidSignature(&_ECDSAStakeRegistry.CallOpts, _dataHash, _signatureData)
}

// IsValidSignature is a free data retrieval call binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 _dataHash, bytes _signatureData) view returns(bytes4)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) IsValidSignature(_dataHash [32]byte, _signatureData []byte) ([4]byte, error) {
	return _ECDSAStakeRegistry.Contract.IsValidSignature(&_ECDSAStakeRegistry.CallOpts, _dataHash, _signatureData)
}

// MinimumWeight is a free data retrieval call binding the contract method 0x40bf2fb7.
//
// Solidity: function minimumWeight() view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) MinimumWeight(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "minimumWeight")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinimumWeight is a free data retrieval call binding the contract method 0x40bf2fb7.
//
// Solidity: function minimumWeight() view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) MinimumWeight() (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.MinimumWeight(&_ECDSAStakeRegistry.CallOpts)
}

// MinimumWeight is a free data retrieval call binding the contract method 0x40bf2fb7.
//
// Solidity: function minimumWeight() view returns(uint256)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) MinimumWeight() (*big.Int, error) {
	return _ECDSAStakeRegistry.Contract.MinimumWeight(&_ECDSAStakeRegistry.CallOpts)
}

// OperatorRegistered is a free data retrieval call binding the contract method 0xec7fbb31.
//
// Solidity: function operatorRegistered(address _operator) view returns(bool)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) OperatorRegistered(opts *bind.CallOpts, _operator common.Address) (bool, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "operatorRegistered", _operator)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// OperatorRegistered is a free data retrieval call binding the contract method 0xec7fbb31.
//
// Solidity: function operatorRegistered(address _operator) view returns(bool)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) OperatorRegistered(_operator common.Address) (bool, error) {
	return _ECDSAStakeRegistry.Contract.OperatorRegistered(&_ECDSAStakeRegistry.CallOpts, _operator)
}

// OperatorRegistered is a free data retrieval call binding the contract method 0xec7fbb31.
//
// Solidity: function operatorRegistered(address _operator) view returns(bool)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) OperatorRegistered(_operator common.Address) (bool, error) {
	return _ECDSAStakeRegistry.Contract.OperatorRegistered(&_ECDSAStakeRegistry.CallOpts, _operator)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) Owner() (common.Address, error) {
	return _ECDSAStakeRegistry.Contract.Owner(&_ECDSAStakeRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) Owner() (common.Address, error) {
	return _ECDSAStakeRegistry.Contract.Owner(&_ECDSAStakeRegistry.CallOpts)
}

// Quorum is a free data retrieval call binding the contract method 0x1703a018.
//
// Solidity: function quorum() view returns(((address,uint96)[]))
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCaller) Quorum(opts *bind.CallOpts) (Quorum, error) {
	var out []interface{}
	err := _ECDSAStakeRegistry.contract.Call(opts, &out, "quorum")

	if err != nil {
		return *new(Quorum), err
	}

	out0 := *abi.ConvertType(out[0], new(Quorum)).(*Quorum)

	return out0, err

}

// Quorum is a free data retrieval call binding the contract method 0x1703a018.
//
// Solidity: function quorum() view returns(((address,uint96)[]))
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) Quorum() (Quorum, error) {
	return _ECDSAStakeRegistry.Contract.Quorum(&_ECDSAStakeRegistry.CallOpts)
}

// Quorum is a free data retrieval call binding the contract method 0x1703a018.
//
// Solidity: function quorum() view returns(((address,uint96)[]))
func (_ECDSAStakeRegistry *ECDSAStakeRegistryCallerSession) Quorum() (Quorum, error) {
	return _ECDSAStakeRegistry.Contract.Quorum(&_ECDSAStakeRegistry.CallOpts)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x857dc190.
//
// Solidity: function deregisterOperator() returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactor) DeregisterOperator(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.contract.Transact(opts, "deregisterOperator")
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x857dc190.
//
// Solidity: function deregisterOperator() returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) DeregisterOperator() (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.DeregisterOperator(&_ECDSAStakeRegistry.TransactOpts)
}

// DeregisterOperator is a paid mutator transaction binding the contract method 0x857dc190.
//
// Solidity: function deregisterOperator() returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactorSession) DeregisterOperator() (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.DeregisterOperator(&_ECDSAStakeRegistry.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x4251a489.
//
// Solidity: function initialize(uint256 _thresholdWeight, ((address,uint96)[]) quorumParams) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactor) Initialize(opts *bind.TransactOpts, _thresholdWeight *big.Int, quorumParams Quorum) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.contract.Transact(opts, "initialize", _thresholdWeight, quorumParams)
}

// Initialize is a paid mutator transaction binding the contract method 0x4251a489.
//
// Solidity: function initialize(uint256 _thresholdWeight, ((address,uint96)[]) quorumParams) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) Initialize(_thresholdWeight *big.Int, quorumParams Quorum) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.Initialize(&_ECDSAStakeRegistry.TransactOpts, _thresholdWeight, quorumParams)
}

// Initialize is a paid mutator transaction binding the contract method 0x4251a489.
//
// Solidity: function initialize(uint256 _thresholdWeight, ((address,uint96)[]) quorumParams) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactorSession) Initialize(_thresholdWeight *big.Int, quorumParams Quorum) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.Initialize(&_ECDSAStakeRegistry.TransactOpts, _thresholdWeight, quorumParams)
}

// RegisterOperatorWithSignature is a paid mutator transaction binding the contract method 0x2a855b17.
//
// Solidity: function registerOperatorWithSignature((bytes,bytes32,uint256) _operatorSignature, address _signingKey, address _p2pKey) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactor) RegisterOperatorWithSignature(opts *bind.TransactOpts, _operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry, _signingKey common.Address, _p2pKey common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.contract.Transact(opts, "registerOperatorWithSignature", _operatorSignature, _signingKey, _p2pKey)
}

// RegisterOperatorWithSignature is a paid mutator transaction binding the contract method 0x2a855b17.
//
// Solidity: function registerOperatorWithSignature((bytes,bytes32,uint256) _operatorSignature, address _signingKey, address _p2pKey) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) RegisterOperatorWithSignature(_operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry, _signingKey common.Address, _p2pKey common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.RegisterOperatorWithSignature(&_ECDSAStakeRegistry.TransactOpts, _operatorSignature, _signingKey, _p2pKey)
}

// RegisterOperatorWithSignature is a paid mutator transaction binding the contract method 0x2a855b17.
//
// Solidity: function registerOperatorWithSignature((bytes,bytes32,uint256) _operatorSignature, address _signingKey, address _p2pKey) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactorSession) RegisterOperatorWithSignature(_operatorSignature ISignatureUtilsSignatureWithSaltAndExpiry, _signingKey common.Address, _p2pKey common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.RegisterOperatorWithSignature(&_ECDSAStakeRegistry.TransactOpts, _operatorSignature, _signingKey, _p2pKey)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.RenounceOwnership(&_ECDSAStakeRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.RenounceOwnership(&_ECDSAStakeRegistry.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.TransferOwnership(&_ECDSAStakeRegistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.TransferOwnership(&_ECDSAStakeRegistry.TransactOpts, newOwner)
}

// UpdateMinimumWeight is a paid mutator transaction binding the contract method 0x696255be.
//
// Solidity: function updateMinimumWeight(uint256 _newMinimumWeight, address[] _operators) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactor) UpdateMinimumWeight(opts *bind.TransactOpts, _newMinimumWeight *big.Int, _operators []common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.contract.Transact(opts, "updateMinimumWeight", _newMinimumWeight, _operators)
}

// UpdateMinimumWeight is a paid mutator transaction binding the contract method 0x696255be.
//
// Solidity: function updateMinimumWeight(uint256 _newMinimumWeight, address[] _operators) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) UpdateMinimumWeight(_newMinimumWeight *big.Int, _operators []common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.UpdateMinimumWeight(&_ECDSAStakeRegistry.TransactOpts, _newMinimumWeight, _operators)
}

// UpdateMinimumWeight is a paid mutator transaction binding the contract method 0x696255be.
//
// Solidity: function updateMinimumWeight(uint256 _newMinimumWeight, address[] _operators) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactorSession) UpdateMinimumWeight(_newMinimumWeight *big.Int, _operators []common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.UpdateMinimumWeight(&_ECDSAStakeRegistry.TransactOpts, _newMinimumWeight, _operators)
}

// UpdateOperatorP2pKey is a paid mutator transaction binding the contract method 0xca02e133.
//
// Solidity: function updateOperatorP2pKey(address _newP2pKey) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactor) UpdateOperatorP2pKey(opts *bind.TransactOpts, _newP2pKey common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.contract.Transact(opts, "updateOperatorP2pKey", _newP2pKey)
}

// UpdateOperatorP2pKey is a paid mutator transaction binding the contract method 0xca02e133.
//
// Solidity: function updateOperatorP2pKey(address _newP2pKey) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) UpdateOperatorP2pKey(_newP2pKey common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.UpdateOperatorP2pKey(&_ECDSAStakeRegistry.TransactOpts, _newP2pKey)
}

// UpdateOperatorP2pKey is a paid mutator transaction binding the contract method 0xca02e133.
//
// Solidity: function updateOperatorP2pKey(address _newP2pKey) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactorSession) UpdateOperatorP2pKey(_newP2pKey common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.UpdateOperatorP2pKey(&_ECDSAStakeRegistry.TransactOpts, _newP2pKey)
}

// UpdateOperatorSigningKey is a paid mutator transaction binding the contract method 0x743c31f4.
//
// Solidity: function updateOperatorSigningKey(address _newSigningKey) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactor) UpdateOperatorSigningKey(opts *bind.TransactOpts, _newSigningKey common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.contract.Transact(opts, "updateOperatorSigningKey", _newSigningKey)
}

// UpdateOperatorSigningKey is a paid mutator transaction binding the contract method 0x743c31f4.
//
// Solidity: function updateOperatorSigningKey(address _newSigningKey) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) UpdateOperatorSigningKey(_newSigningKey common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.UpdateOperatorSigningKey(&_ECDSAStakeRegistry.TransactOpts, _newSigningKey)
}

// UpdateOperatorSigningKey is a paid mutator transaction binding the contract method 0x743c31f4.
//
// Solidity: function updateOperatorSigningKey(address _newSigningKey) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactorSession) UpdateOperatorSigningKey(_newSigningKey common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.UpdateOperatorSigningKey(&_ECDSAStakeRegistry.TransactOpts, _newSigningKey)
}

// UpdateOperators is a paid mutator transaction binding the contract method 0x00cf2ab5.
//
// Solidity: function updateOperators(address[] _operators) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactor) UpdateOperators(opts *bind.TransactOpts, _operators []common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.contract.Transact(opts, "updateOperators", _operators)
}

// UpdateOperators is a paid mutator transaction binding the contract method 0x00cf2ab5.
//
// Solidity: function updateOperators(address[] _operators) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) UpdateOperators(_operators []common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.UpdateOperators(&_ECDSAStakeRegistry.TransactOpts, _operators)
}

// UpdateOperators is a paid mutator transaction binding the contract method 0x00cf2ab5.
//
// Solidity: function updateOperators(address[] _operators) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactorSession) UpdateOperators(_operators []common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.UpdateOperators(&_ECDSAStakeRegistry.TransactOpts, _operators)
}

// UpdateOperatorsForQuorum is a paid mutator transaction binding the contract method 0x5140a548.
//
// Solidity: function updateOperatorsForQuorum(address[][] operatorsPerQuorum, bytes ) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactor) UpdateOperatorsForQuorum(opts *bind.TransactOpts, operatorsPerQuorum [][]common.Address, arg1 []byte) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.contract.Transact(opts, "updateOperatorsForQuorum", operatorsPerQuorum, arg1)
}

// UpdateOperatorsForQuorum is a paid mutator transaction binding the contract method 0x5140a548.
//
// Solidity: function updateOperatorsForQuorum(address[][] operatorsPerQuorum, bytes ) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) UpdateOperatorsForQuorum(operatorsPerQuorum [][]common.Address, arg1 []byte) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.UpdateOperatorsForQuorum(&_ECDSAStakeRegistry.TransactOpts, operatorsPerQuorum, arg1)
}

// UpdateOperatorsForQuorum is a paid mutator transaction binding the contract method 0x5140a548.
//
// Solidity: function updateOperatorsForQuorum(address[][] operatorsPerQuorum, bytes ) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactorSession) UpdateOperatorsForQuorum(operatorsPerQuorum [][]common.Address, arg1 []byte) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.UpdateOperatorsForQuorum(&_ECDSAStakeRegistry.TransactOpts, operatorsPerQuorum, arg1)
}

// UpdateQuorumConfig is a paid mutator transaction binding the contract method 0xdec5d1f6.
//
// Solidity: function updateQuorumConfig(((address,uint96)[]) newQuorumConfig, address[] _operators) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactor) UpdateQuorumConfig(opts *bind.TransactOpts, newQuorumConfig Quorum, _operators []common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.contract.Transact(opts, "updateQuorumConfig", newQuorumConfig, _operators)
}

// UpdateQuorumConfig is a paid mutator transaction binding the contract method 0xdec5d1f6.
//
// Solidity: function updateQuorumConfig(((address,uint96)[]) newQuorumConfig, address[] _operators) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) UpdateQuorumConfig(newQuorumConfig Quorum, _operators []common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.UpdateQuorumConfig(&_ECDSAStakeRegistry.TransactOpts, newQuorumConfig, _operators)
}

// UpdateQuorumConfig is a paid mutator transaction binding the contract method 0xdec5d1f6.
//
// Solidity: function updateQuorumConfig(((address,uint96)[]) newQuorumConfig, address[] _operators) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactorSession) UpdateQuorumConfig(newQuorumConfig Quorum, _operators []common.Address) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.UpdateQuorumConfig(&_ECDSAStakeRegistry.TransactOpts, newQuorumConfig, _operators)
}

// UpdateStakeThreshold is a paid mutator transaction binding the contract method 0x5ef53329.
//
// Solidity: function updateStakeThreshold(uint256 _thresholdWeight) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactor) UpdateStakeThreshold(opts *bind.TransactOpts, _thresholdWeight *big.Int) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.contract.Transact(opts, "updateStakeThreshold", _thresholdWeight)
}

// UpdateStakeThreshold is a paid mutator transaction binding the contract method 0x5ef53329.
//
// Solidity: function updateStakeThreshold(uint256 _thresholdWeight) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistrySession) UpdateStakeThreshold(_thresholdWeight *big.Int) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.UpdateStakeThreshold(&_ECDSAStakeRegistry.TransactOpts, _thresholdWeight)
}

// UpdateStakeThreshold is a paid mutator transaction binding the contract method 0x5ef53329.
//
// Solidity: function updateStakeThreshold(uint256 _thresholdWeight) returns()
func (_ECDSAStakeRegistry *ECDSAStakeRegistryTransactorSession) UpdateStakeThreshold(_thresholdWeight *big.Int) (*types.Transaction, error) {
	return _ECDSAStakeRegistry.Contract.UpdateStakeThreshold(&_ECDSAStakeRegistry.TransactOpts, _thresholdWeight)
}

// ECDSAStakeRegistryInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryInitializedIterator struct {
	Event *ECDSAStakeRegistryInitialized // Event containing the contract specifics and raw log

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
func (it *ECDSAStakeRegistryInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ECDSAStakeRegistryInitialized)
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
		it.Event = new(ECDSAStakeRegistryInitialized)
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
func (it *ECDSAStakeRegistryInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ECDSAStakeRegistryInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ECDSAStakeRegistryInitialized represents a Initialized event raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) FilterInitialized(opts *bind.FilterOpts) (*ECDSAStakeRegistryInitializedIterator, error) {

	logs, sub, err := _ECDSAStakeRegistry.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistryInitializedIterator{contract: _ECDSAStakeRegistry.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ECDSAStakeRegistryInitialized) (event.Subscription, error) {

	logs, sub, err := _ECDSAStakeRegistry.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ECDSAStakeRegistryInitialized)
				if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) ParseInitialized(log types.Log) (*ECDSAStakeRegistryInitialized, error) {
	event := new(ECDSAStakeRegistryInitialized)
	if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ECDSAStakeRegistryMinimumWeightUpdatedIterator is returned from FilterMinimumWeightUpdated and is used to iterate over the raw logs and unpacked data for MinimumWeightUpdated events raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryMinimumWeightUpdatedIterator struct {
	Event *ECDSAStakeRegistryMinimumWeightUpdated // Event containing the contract specifics and raw log

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
func (it *ECDSAStakeRegistryMinimumWeightUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ECDSAStakeRegistryMinimumWeightUpdated)
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
		it.Event = new(ECDSAStakeRegistryMinimumWeightUpdated)
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
func (it *ECDSAStakeRegistryMinimumWeightUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ECDSAStakeRegistryMinimumWeightUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ECDSAStakeRegistryMinimumWeightUpdated represents a MinimumWeightUpdated event raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryMinimumWeightUpdated struct {
	Old *big.Int
	New *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterMinimumWeightUpdated is a free log retrieval operation binding the contract event 0x713ca53b88d6eb63f5b1854cb8cbdd736ec51eda225e46791aa9298b0160648f.
//
// Solidity: event MinimumWeightUpdated(uint256 _old, uint256 _new)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) FilterMinimumWeightUpdated(opts *bind.FilterOpts) (*ECDSAStakeRegistryMinimumWeightUpdatedIterator, error) {

	logs, sub, err := _ECDSAStakeRegistry.contract.FilterLogs(opts, "MinimumWeightUpdated")
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistryMinimumWeightUpdatedIterator{contract: _ECDSAStakeRegistry.contract, event: "MinimumWeightUpdated", logs: logs, sub: sub}, nil
}

// WatchMinimumWeightUpdated is a free log subscription operation binding the contract event 0x713ca53b88d6eb63f5b1854cb8cbdd736ec51eda225e46791aa9298b0160648f.
//
// Solidity: event MinimumWeightUpdated(uint256 _old, uint256 _new)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) WatchMinimumWeightUpdated(opts *bind.WatchOpts, sink chan<- *ECDSAStakeRegistryMinimumWeightUpdated) (event.Subscription, error) {

	logs, sub, err := _ECDSAStakeRegistry.contract.WatchLogs(opts, "MinimumWeightUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ECDSAStakeRegistryMinimumWeightUpdated)
				if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "MinimumWeightUpdated", log); err != nil {
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

// ParseMinimumWeightUpdated is a log parse operation binding the contract event 0x713ca53b88d6eb63f5b1854cb8cbdd736ec51eda225e46791aa9298b0160648f.
//
// Solidity: event MinimumWeightUpdated(uint256 _old, uint256 _new)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) ParseMinimumWeightUpdated(log types.Log) (*ECDSAStakeRegistryMinimumWeightUpdated, error) {
	event := new(ECDSAStakeRegistryMinimumWeightUpdated)
	if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "MinimumWeightUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ECDSAStakeRegistryOperatorDeregisteredIterator is returned from FilterOperatorDeregistered and is used to iterate over the raw logs and unpacked data for OperatorDeregistered events raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryOperatorDeregisteredIterator struct {
	Event *ECDSAStakeRegistryOperatorDeregistered // Event containing the contract specifics and raw log

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
func (it *ECDSAStakeRegistryOperatorDeregisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ECDSAStakeRegistryOperatorDeregistered)
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
		it.Event = new(ECDSAStakeRegistryOperatorDeregistered)
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
func (it *ECDSAStakeRegistryOperatorDeregisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ECDSAStakeRegistryOperatorDeregisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ECDSAStakeRegistryOperatorDeregistered represents a OperatorDeregistered event raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryOperatorDeregistered struct {
	Operator    common.Address
	BlockNumber *big.Int
	Avs         common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterOperatorDeregistered is a free log retrieval operation binding the contract event 0xed218afd65e8eaf426500b12bc170cf79b3b2496d94468127f79d95fb45ccae9.
//
// Solidity: event OperatorDeregistered(address indexed _operator, uint256 indexed blockNumber, address indexed _avs)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) FilterOperatorDeregistered(opts *bind.FilterOpts, _operator []common.Address, blockNumber []*big.Int, _avs []common.Address) (*ECDSAStakeRegistryOperatorDeregisteredIterator, error) {

	var _operatorRule []interface{}
	for _, _operatorItem := range _operator {
		_operatorRule = append(_operatorRule, _operatorItem)
	}
	var blockNumberRule []interface{}
	for _, blockNumberItem := range blockNumber {
		blockNumberRule = append(blockNumberRule, blockNumberItem)
	}
	var _avsRule []interface{}
	for _, _avsItem := range _avs {
		_avsRule = append(_avsRule, _avsItem)
	}

	logs, sub, err := _ECDSAStakeRegistry.contract.FilterLogs(opts, "OperatorDeregistered", _operatorRule, blockNumberRule, _avsRule)
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistryOperatorDeregisteredIterator{contract: _ECDSAStakeRegistry.contract, event: "OperatorDeregistered", logs: logs, sub: sub}, nil
}

// WatchOperatorDeregistered is a free log subscription operation binding the contract event 0xed218afd65e8eaf426500b12bc170cf79b3b2496d94468127f79d95fb45ccae9.
//
// Solidity: event OperatorDeregistered(address indexed _operator, uint256 indexed blockNumber, address indexed _avs)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) WatchOperatorDeregistered(opts *bind.WatchOpts, sink chan<- *ECDSAStakeRegistryOperatorDeregistered, _operator []common.Address, blockNumber []*big.Int, _avs []common.Address) (event.Subscription, error) {

	var _operatorRule []interface{}
	for _, _operatorItem := range _operator {
		_operatorRule = append(_operatorRule, _operatorItem)
	}
	var blockNumberRule []interface{}
	for _, blockNumberItem := range blockNumber {
		blockNumberRule = append(blockNumberRule, blockNumberItem)
	}
	var _avsRule []interface{}
	for _, _avsItem := range _avs {
		_avsRule = append(_avsRule, _avsItem)
	}

	logs, sub, err := _ECDSAStakeRegistry.contract.WatchLogs(opts, "OperatorDeregistered", _operatorRule, blockNumberRule, _avsRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ECDSAStakeRegistryOperatorDeregistered)
				if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "OperatorDeregistered", log); err != nil {
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

// ParseOperatorDeregistered is a log parse operation binding the contract event 0xed218afd65e8eaf426500b12bc170cf79b3b2496d94468127f79d95fb45ccae9.
//
// Solidity: event OperatorDeregistered(address indexed _operator, uint256 indexed blockNumber, address indexed _avs)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) ParseOperatorDeregistered(log types.Log) (*ECDSAStakeRegistryOperatorDeregistered, error) {
	event := new(ECDSAStakeRegistryOperatorDeregistered)
	if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "OperatorDeregistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ECDSAStakeRegistryOperatorRegisteredIterator is returned from FilterOperatorRegistered and is used to iterate over the raw logs and unpacked data for OperatorRegistered events raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryOperatorRegisteredIterator struct {
	Event *ECDSAStakeRegistryOperatorRegistered // Event containing the contract specifics and raw log

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
func (it *ECDSAStakeRegistryOperatorRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ECDSAStakeRegistryOperatorRegistered)
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
		it.Event = new(ECDSAStakeRegistryOperatorRegistered)
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
func (it *ECDSAStakeRegistryOperatorRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ECDSAStakeRegistryOperatorRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ECDSAStakeRegistryOperatorRegistered represents a OperatorRegistered event raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryOperatorRegistered struct {
	Operator    common.Address
	BlockNumber *big.Int
	P2pKey      common.Address
	SigningKey  common.Address
	Avs         common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterOperatorRegistered is a free log retrieval operation binding the contract event 0xe0a0b7f17ef76bcee73033890bc3864e96d49ea12881e0b4b0b4b865c890ff4e.
//
// Solidity: event OperatorRegistered(address indexed _operator, uint256 indexed blockNumber, address indexed _p2pKey, address _signingKey, address _avs)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) FilterOperatorRegistered(opts *bind.FilterOpts, _operator []common.Address, blockNumber []*big.Int, _p2pKey []common.Address) (*ECDSAStakeRegistryOperatorRegisteredIterator, error) {

	var _operatorRule []interface{}
	for _, _operatorItem := range _operator {
		_operatorRule = append(_operatorRule, _operatorItem)
	}
	var blockNumberRule []interface{}
	for _, blockNumberItem := range blockNumber {
		blockNumberRule = append(blockNumberRule, blockNumberItem)
	}
	var _p2pKeyRule []interface{}
	for _, _p2pKeyItem := range _p2pKey {
		_p2pKeyRule = append(_p2pKeyRule, _p2pKeyItem)
	}

	logs, sub, err := _ECDSAStakeRegistry.contract.FilterLogs(opts, "OperatorRegistered", _operatorRule, blockNumberRule, _p2pKeyRule)
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistryOperatorRegisteredIterator{contract: _ECDSAStakeRegistry.contract, event: "OperatorRegistered", logs: logs, sub: sub}, nil
}

// WatchOperatorRegistered is a free log subscription operation binding the contract event 0xe0a0b7f17ef76bcee73033890bc3864e96d49ea12881e0b4b0b4b865c890ff4e.
//
// Solidity: event OperatorRegistered(address indexed _operator, uint256 indexed blockNumber, address indexed _p2pKey, address _signingKey, address _avs)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) WatchOperatorRegistered(opts *bind.WatchOpts, sink chan<- *ECDSAStakeRegistryOperatorRegistered, _operator []common.Address, blockNumber []*big.Int, _p2pKey []common.Address) (event.Subscription, error) {

	var _operatorRule []interface{}
	for _, _operatorItem := range _operator {
		_operatorRule = append(_operatorRule, _operatorItem)
	}
	var blockNumberRule []interface{}
	for _, blockNumberItem := range blockNumber {
		blockNumberRule = append(blockNumberRule, blockNumberItem)
	}
	var _p2pKeyRule []interface{}
	for _, _p2pKeyItem := range _p2pKey {
		_p2pKeyRule = append(_p2pKeyRule, _p2pKeyItem)
	}

	logs, sub, err := _ECDSAStakeRegistry.contract.WatchLogs(opts, "OperatorRegistered", _operatorRule, blockNumberRule, _p2pKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ECDSAStakeRegistryOperatorRegistered)
				if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "OperatorRegistered", log); err != nil {
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

// ParseOperatorRegistered is a log parse operation binding the contract event 0xe0a0b7f17ef76bcee73033890bc3864e96d49ea12881e0b4b0b4b865c890ff4e.
//
// Solidity: event OperatorRegistered(address indexed _operator, uint256 indexed blockNumber, address indexed _p2pKey, address _signingKey, address _avs)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) ParseOperatorRegistered(log types.Log) (*ECDSAStakeRegistryOperatorRegistered, error) {
	event := new(ECDSAStakeRegistryOperatorRegistered)
	if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "OperatorRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ECDSAStakeRegistryOperatorWeightUpdatedIterator is returned from FilterOperatorWeightUpdated and is used to iterate over the raw logs and unpacked data for OperatorWeightUpdated events raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryOperatorWeightUpdatedIterator struct {
	Event *ECDSAStakeRegistryOperatorWeightUpdated // Event containing the contract specifics and raw log

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
func (it *ECDSAStakeRegistryOperatorWeightUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ECDSAStakeRegistryOperatorWeightUpdated)
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
		it.Event = new(ECDSAStakeRegistryOperatorWeightUpdated)
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
func (it *ECDSAStakeRegistryOperatorWeightUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ECDSAStakeRegistryOperatorWeightUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ECDSAStakeRegistryOperatorWeightUpdated represents a OperatorWeightUpdated event raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryOperatorWeightUpdated struct {
	Operator  common.Address
	OldWeight *big.Int
	NewWeight *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterOperatorWeightUpdated is a free log retrieval operation binding the contract event 0x88770dc862e47a7ed586907857eb1b75e4c5ffc8b707c7ee10eb74d6885fe594.
//
// Solidity: event OperatorWeightUpdated(address indexed _operator, uint256 oldWeight, uint256 newWeight)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) FilterOperatorWeightUpdated(opts *bind.FilterOpts, _operator []common.Address) (*ECDSAStakeRegistryOperatorWeightUpdatedIterator, error) {

	var _operatorRule []interface{}
	for _, _operatorItem := range _operator {
		_operatorRule = append(_operatorRule, _operatorItem)
	}

	logs, sub, err := _ECDSAStakeRegistry.contract.FilterLogs(opts, "OperatorWeightUpdated", _operatorRule)
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistryOperatorWeightUpdatedIterator{contract: _ECDSAStakeRegistry.contract, event: "OperatorWeightUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorWeightUpdated is a free log subscription operation binding the contract event 0x88770dc862e47a7ed586907857eb1b75e4c5ffc8b707c7ee10eb74d6885fe594.
//
// Solidity: event OperatorWeightUpdated(address indexed _operator, uint256 oldWeight, uint256 newWeight)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) WatchOperatorWeightUpdated(opts *bind.WatchOpts, sink chan<- *ECDSAStakeRegistryOperatorWeightUpdated, _operator []common.Address) (event.Subscription, error) {

	var _operatorRule []interface{}
	for _, _operatorItem := range _operator {
		_operatorRule = append(_operatorRule, _operatorItem)
	}

	logs, sub, err := _ECDSAStakeRegistry.contract.WatchLogs(opts, "OperatorWeightUpdated", _operatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ECDSAStakeRegistryOperatorWeightUpdated)
				if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "OperatorWeightUpdated", log); err != nil {
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

// ParseOperatorWeightUpdated is a log parse operation binding the contract event 0x88770dc862e47a7ed586907857eb1b75e4c5ffc8b707c7ee10eb74d6885fe594.
//
// Solidity: event OperatorWeightUpdated(address indexed _operator, uint256 oldWeight, uint256 newWeight)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) ParseOperatorWeightUpdated(log types.Log) (*ECDSAStakeRegistryOperatorWeightUpdated, error) {
	event := new(ECDSAStakeRegistryOperatorWeightUpdated)
	if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "OperatorWeightUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ECDSAStakeRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryOwnershipTransferredIterator struct {
	Event *ECDSAStakeRegistryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ECDSAStakeRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ECDSAStakeRegistryOwnershipTransferred)
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
		it.Event = new(ECDSAStakeRegistryOwnershipTransferred)
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
func (it *ECDSAStakeRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ECDSAStakeRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ECDSAStakeRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ECDSAStakeRegistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ECDSAStakeRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistryOwnershipTransferredIterator{contract: _ECDSAStakeRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ECDSAStakeRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ECDSAStakeRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ECDSAStakeRegistryOwnershipTransferred)
				if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) ParseOwnershipTransferred(log types.Log) (*ECDSAStakeRegistryOwnershipTransferred, error) {
	event := new(ECDSAStakeRegistryOwnershipTransferred)
	if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ECDSAStakeRegistryP2pKeyUpdateIterator is returned from FilterP2pKeyUpdate and is used to iterate over the raw logs and unpacked data for P2pKeyUpdate events raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryP2pKeyUpdateIterator struct {
	Event *ECDSAStakeRegistryP2pKeyUpdate // Event containing the contract specifics and raw log

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
func (it *ECDSAStakeRegistryP2pKeyUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ECDSAStakeRegistryP2pKeyUpdate)
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
		it.Event = new(ECDSAStakeRegistryP2pKeyUpdate)
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
func (it *ECDSAStakeRegistryP2pKeyUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ECDSAStakeRegistryP2pKeyUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ECDSAStakeRegistryP2pKeyUpdate represents a P2pKeyUpdate event raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryP2pKeyUpdate struct {
	Operator  common.Address
	NewP2pKey common.Address
	OldP2pKey common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterP2pKeyUpdate is a free log retrieval operation binding the contract event 0x5921178687094d1c42916434af2b2f18071190946686fd8b599c0f957b9d1372.
//
// Solidity: event P2pKeyUpdate(address indexed operator, address indexed newP2pKey, address oldP2pKey)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) FilterP2pKeyUpdate(opts *bind.FilterOpts, operator []common.Address, newP2pKey []common.Address) (*ECDSAStakeRegistryP2pKeyUpdateIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var newP2pKeyRule []interface{}
	for _, newP2pKeyItem := range newP2pKey {
		newP2pKeyRule = append(newP2pKeyRule, newP2pKeyItem)
	}

	logs, sub, err := _ECDSAStakeRegistry.contract.FilterLogs(opts, "P2pKeyUpdate", operatorRule, newP2pKeyRule)
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistryP2pKeyUpdateIterator{contract: _ECDSAStakeRegistry.contract, event: "P2pKeyUpdate", logs: logs, sub: sub}, nil
}

// WatchP2pKeyUpdate is a free log subscription operation binding the contract event 0x5921178687094d1c42916434af2b2f18071190946686fd8b599c0f957b9d1372.
//
// Solidity: event P2pKeyUpdate(address indexed operator, address indexed newP2pKey, address oldP2pKey)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) WatchP2pKeyUpdate(opts *bind.WatchOpts, sink chan<- *ECDSAStakeRegistryP2pKeyUpdate, operator []common.Address, newP2pKey []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var newP2pKeyRule []interface{}
	for _, newP2pKeyItem := range newP2pKey {
		newP2pKeyRule = append(newP2pKeyRule, newP2pKeyItem)
	}

	logs, sub, err := _ECDSAStakeRegistry.contract.WatchLogs(opts, "P2pKeyUpdate", operatorRule, newP2pKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ECDSAStakeRegistryP2pKeyUpdate)
				if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "P2pKeyUpdate", log); err != nil {
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

// ParseP2pKeyUpdate is a log parse operation binding the contract event 0x5921178687094d1c42916434af2b2f18071190946686fd8b599c0f957b9d1372.
//
// Solidity: event P2pKeyUpdate(address indexed operator, address indexed newP2pKey, address oldP2pKey)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) ParseP2pKeyUpdate(log types.Log) (*ECDSAStakeRegistryP2pKeyUpdate, error) {
	event := new(ECDSAStakeRegistryP2pKeyUpdate)
	if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "P2pKeyUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ECDSAStakeRegistryQuorumUpdatedIterator is returned from FilterQuorumUpdated and is used to iterate over the raw logs and unpacked data for QuorumUpdated events raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryQuorumUpdatedIterator struct {
	Event *ECDSAStakeRegistryQuorumUpdated // Event containing the contract specifics and raw log

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
func (it *ECDSAStakeRegistryQuorumUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ECDSAStakeRegistryQuorumUpdated)
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
		it.Event = new(ECDSAStakeRegistryQuorumUpdated)
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
func (it *ECDSAStakeRegistryQuorumUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ECDSAStakeRegistryQuorumUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ECDSAStakeRegistryQuorumUpdated represents a QuorumUpdated event raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryQuorumUpdated struct {
	Old Quorum
	New Quorum
	Raw types.Log // Blockchain specific contextual infos
}

// FilterQuorumUpdated is a free log retrieval operation binding the contract event 0x23aad4e61744ece164130aa415c1616e80136b0f0770e56589438b90b269265e.
//
// Solidity: event QuorumUpdated(((address,uint96)[]) _old, ((address,uint96)[]) _new)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) FilterQuorumUpdated(opts *bind.FilterOpts) (*ECDSAStakeRegistryQuorumUpdatedIterator, error) {

	logs, sub, err := _ECDSAStakeRegistry.contract.FilterLogs(opts, "QuorumUpdated")
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistryQuorumUpdatedIterator{contract: _ECDSAStakeRegistry.contract, event: "QuorumUpdated", logs: logs, sub: sub}, nil
}

// WatchQuorumUpdated is a free log subscription operation binding the contract event 0x23aad4e61744ece164130aa415c1616e80136b0f0770e56589438b90b269265e.
//
// Solidity: event QuorumUpdated(((address,uint96)[]) _old, ((address,uint96)[]) _new)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) WatchQuorumUpdated(opts *bind.WatchOpts, sink chan<- *ECDSAStakeRegistryQuorumUpdated) (event.Subscription, error) {

	logs, sub, err := _ECDSAStakeRegistry.contract.WatchLogs(opts, "QuorumUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ECDSAStakeRegistryQuorumUpdated)
				if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "QuorumUpdated", log); err != nil {
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

// ParseQuorumUpdated is a log parse operation binding the contract event 0x23aad4e61744ece164130aa415c1616e80136b0f0770e56589438b90b269265e.
//
// Solidity: event QuorumUpdated(((address,uint96)[]) _old, ((address,uint96)[]) _new)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) ParseQuorumUpdated(log types.Log) (*ECDSAStakeRegistryQuorumUpdated, error) {
	event := new(ECDSAStakeRegistryQuorumUpdated)
	if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "QuorumUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ECDSAStakeRegistrySigningKeyUpdateIterator is returned from FilterSigningKeyUpdate and is used to iterate over the raw logs and unpacked data for SigningKeyUpdate events raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistrySigningKeyUpdateIterator struct {
	Event *ECDSAStakeRegistrySigningKeyUpdate // Event containing the contract specifics and raw log

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
func (it *ECDSAStakeRegistrySigningKeyUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ECDSAStakeRegistrySigningKeyUpdate)
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
		it.Event = new(ECDSAStakeRegistrySigningKeyUpdate)
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
func (it *ECDSAStakeRegistrySigningKeyUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ECDSAStakeRegistrySigningKeyUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ECDSAStakeRegistrySigningKeyUpdate represents a SigningKeyUpdate event raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistrySigningKeyUpdate struct {
	Operator      common.Address
	NewSigningKey common.Address
	OldSigningKey common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterSigningKeyUpdate is a free log retrieval operation binding the contract event 0x6f32abd4797f9f6fd30d9393963a8dc7ab716a4355ab392a6c8ac5be1911d4e2.
//
// Solidity: event SigningKeyUpdate(address indexed operator, address indexed newSigningKey, address oldSigningKey)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) FilterSigningKeyUpdate(opts *bind.FilterOpts, operator []common.Address, newSigningKey []common.Address) (*ECDSAStakeRegistrySigningKeyUpdateIterator, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var newSigningKeyRule []interface{}
	for _, newSigningKeyItem := range newSigningKey {
		newSigningKeyRule = append(newSigningKeyRule, newSigningKeyItem)
	}

	logs, sub, err := _ECDSAStakeRegistry.contract.FilterLogs(opts, "SigningKeyUpdate", operatorRule, newSigningKeyRule)
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistrySigningKeyUpdateIterator{contract: _ECDSAStakeRegistry.contract, event: "SigningKeyUpdate", logs: logs, sub: sub}, nil
}

// WatchSigningKeyUpdate is a free log subscription operation binding the contract event 0x6f32abd4797f9f6fd30d9393963a8dc7ab716a4355ab392a6c8ac5be1911d4e2.
//
// Solidity: event SigningKeyUpdate(address indexed operator, address indexed newSigningKey, address oldSigningKey)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) WatchSigningKeyUpdate(opts *bind.WatchOpts, sink chan<- *ECDSAStakeRegistrySigningKeyUpdate, operator []common.Address, newSigningKey []common.Address) (event.Subscription, error) {

	var operatorRule []interface{}
	for _, operatorItem := range operator {
		operatorRule = append(operatorRule, operatorItem)
	}
	var newSigningKeyRule []interface{}
	for _, newSigningKeyItem := range newSigningKey {
		newSigningKeyRule = append(newSigningKeyRule, newSigningKeyItem)
	}

	logs, sub, err := _ECDSAStakeRegistry.contract.WatchLogs(opts, "SigningKeyUpdate", operatorRule, newSigningKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ECDSAStakeRegistrySigningKeyUpdate)
				if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "SigningKeyUpdate", log); err != nil {
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

// ParseSigningKeyUpdate is a log parse operation binding the contract event 0x6f32abd4797f9f6fd30d9393963a8dc7ab716a4355ab392a6c8ac5be1911d4e2.
//
// Solidity: event SigningKeyUpdate(address indexed operator, address indexed newSigningKey, address oldSigningKey)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) ParseSigningKeyUpdate(log types.Log) (*ECDSAStakeRegistrySigningKeyUpdate, error) {
	event := new(ECDSAStakeRegistrySigningKeyUpdate)
	if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "SigningKeyUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ECDSAStakeRegistryThresholdWeightUpdatedIterator is returned from FilterThresholdWeightUpdated and is used to iterate over the raw logs and unpacked data for ThresholdWeightUpdated events raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryThresholdWeightUpdatedIterator struct {
	Event *ECDSAStakeRegistryThresholdWeightUpdated // Event containing the contract specifics and raw log

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
func (it *ECDSAStakeRegistryThresholdWeightUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ECDSAStakeRegistryThresholdWeightUpdated)
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
		it.Event = new(ECDSAStakeRegistryThresholdWeightUpdated)
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
func (it *ECDSAStakeRegistryThresholdWeightUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ECDSAStakeRegistryThresholdWeightUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ECDSAStakeRegistryThresholdWeightUpdated represents a ThresholdWeightUpdated event raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryThresholdWeightUpdated struct {
	ThresholdWeight *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterThresholdWeightUpdated is a free log retrieval operation binding the contract event 0x9324f7e5a7c0288808a634ccde44b8e979676474b22e29ee9dd569b55e791a4b.
//
// Solidity: event ThresholdWeightUpdated(uint256 _thresholdWeight)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) FilterThresholdWeightUpdated(opts *bind.FilterOpts) (*ECDSAStakeRegistryThresholdWeightUpdatedIterator, error) {

	logs, sub, err := _ECDSAStakeRegistry.contract.FilterLogs(opts, "ThresholdWeightUpdated")
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistryThresholdWeightUpdatedIterator{contract: _ECDSAStakeRegistry.contract, event: "ThresholdWeightUpdated", logs: logs, sub: sub}, nil
}

// WatchThresholdWeightUpdated is a free log subscription operation binding the contract event 0x9324f7e5a7c0288808a634ccde44b8e979676474b22e29ee9dd569b55e791a4b.
//
// Solidity: event ThresholdWeightUpdated(uint256 _thresholdWeight)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) WatchThresholdWeightUpdated(opts *bind.WatchOpts, sink chan<- *ECDSAStakeRegistryThresholdWeightUpdated) (event.Subscription, error) {

	logs, sub, err := _ECDSAStakeRegistry.contract.WatchLogs(opts, "ThresholdWeightUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ECDSAStakeRegistryThresholdWeightUpdated)
				if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "ThresholdWeightUpdated", log); err != nil {
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

// ParseThresholdWeightUpdated is a log parse operation binding the contract event 0x9324f7e5a7c0288808a634ccde44b8e979676474b22e29ee9dd569b55e791a4b.
//
// Solidity: event ThresholdWeightUpdated(uint256 _thresholdWeight)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) ParseThresholdWeightUpdated(log types.Log) (*ECDSAStakeRegistryThresholdWeightUpdated, error) {
	event := new(ECDSAStakeRegistryThresholdWeightUpdated)
	if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "ThresholdWeightUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ECDSAStakeRegistryTotalWeightUpdatedIterator is returned from FilterTotalWeightUpdated and is used to iterate over the raw logs and unpacked data for TotalWeightUpdated events raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryTotalWeightUpdatedIterator struct {
	Event *ECDSAStakeRegistryTotalWeightUpdated // Event containing the contract specifics and raw log

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
func (it *ECDSAStakeRegistryTotalWeightUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ECDSAStakeRegistryTotalWeightUpdated)
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
		it.Event = new(ECDSAStakeRegistryTotalWeightUpdated)
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
func (it *ECDSAStakeRegistryTotalWeightUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ECDSAStakeRegistryTotalWeightUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ECDSAStakeRegistryTotalWeightUpdated represents a TotalWeightUpdated event raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryTotalWeightUpdated struct {
	OldTotalWeight *big.Int
	NewTotalWeight *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterTotalWeightUpdated is a free log retrieval operation binding the contract event 0x86dcf86b12dfeedea74ae9300dbdaa193bcce5809369c8177ea2f4eaaa65729b.
//
// Solidity: event TotalWeightUpdated(uint256 oldTotalWeight, uint256 newTotalWeight)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) FilterTotalWeightUpdated(opts *bind.FilterOpts) (*ECDSAStakeRegistryTotalWeightUpdatedIterator, error) {

	logs, sub, err := _ECDSAStakeRegistry.contract.FilterLogs(opts, "TotalWeightUpdated")
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistryTotalWeightUpdatedIterator{contract: _ECDSAStakeRegistry.contract, event: "TotalWeightUpdated", logs: logs, sub: sub}, nil
}

// WatchTotalWeightUpdated is a free log subscription operation binding the contract event 0x86dcf86b12dfeedea74ae9300dbdaa193bcce5809369c8177ea2f4eaaa65729b.
//
// Solidity: event TotalWeightUpdated(uint256 oldTotalWeight, uint256 newTotalWeight)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) WatchTotalWeightUpdated(opts *bind.WatchOpts, sink chan<- *ECDSAStakeRegistryTotalWeightUpdated) (event.Subscription, error) {

	logs, sub, err := _ECDSAStakeRegistry.contract.WatchLogs(opts, "TotalWeightUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ECDSAStakeRegistryTotalWeightUpdated)
				if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "TotalWeightUpdated", log); err != nil {
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

// ParseTotalWeightUpdated is a log parse operation binding the contract event 0x86dcf86b12dfeedea74ae9300dbdaa193bcce5809369c8177ea2f4eaaa65729b.
//
// Solidity: event TotalWeightUpdated(uint256 oldTotalWeight, uint256 newTotalWeight)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) ParseTotalWeightUpdated(log types.Log) (*ECDSAStakeRegistryTotalWeightUpdated, error) {
	event := new(ECDSAStakeRegistryTotalWeightUpdated)
	if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "TotalWeightUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ECDSAStakeRegistryUpdateMinimumWeightIterator is returned from FilterUpdateMinimumWeight and is used to iterate over the raw logs and unpacked data for UpdateMinimumWeight events raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryUpdateMinimumWeightIterator struct {
	Event *ECDSAStakeRegistryUpdateMinimumWeight // Event containing the contract specifics and raw log

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
func (it *ECDSAStakeRegistryUpdateMinimumWeightIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ECDSAStakeRegistryUpdateMinimumWeight)
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
		it.Event = new(ECDSAStakeRegistryUpdateMinimumWeight)
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
func (it *ECDSAStakeRegistryUpdateMinimumWeightIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ECDSAStakeRegistryUpdateMinimumWeightIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ECDSAStakeRegistryUpdateMinimumWeight represents a UpdateMinimumWeight event raised by the ECDSAStakeRegistry contract.
type ECDSAStakeRegistryUpdateMinimumWeight struct {
	OldMinimumWeight *big.Int
	NewMinimumWeight *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterUpdateMinimumWeight is a free log retrieval operation binding the contract event 0x1ea42186b305fa37310450d9fb87ea1e8f0c7f447e771479e3b27634bfe84dc1.
//
// Solidity: event UpdateMinimumWeight(uint256 oldMinimumWeight, uint256 newMinimumWeight)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) FilterUpdateMinimumWeight(opts *bind.FilterOpts) (*ECDSAStakeRegistryUpdateMinimumWeightIterator, error) {

	logs, sub, err := _ECDSAStakeRegistry.contract.FilterLogs(opts, "UpdateMinimumWeight")
	if err != nil {
		return nil, err
	}
	return &ECDSAStakeRegistryUpdateMinimumWeightIterator{contract: _ECDSAStakeRegistry.contract, event: "UpdateMinimumWeight", logs: logs, sub: sub}, nil
}

// WatchUpdateMinimumWeight is a free log subscription operation binding the contract event 0x1ea42186b305fa37310450d9fb87ea1e8f0c7f447e771479e3b27634bfe84dc1.
//
// Solidity: event UpdateMinimumWeight(uint256 oldMinimumWeight, uint256 newMinimumWeight)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) WatchUpdateMinimumWeight(opts *bind.WatchOpts, sink chan<- *ECDSAStakeRegistryUpdateMinimumWeight) (event.Subscription, error) {

	logs, sub, err := _ECDSAStakeRegistry.contract.WatchLogs(opts, "UpdateMinimumWeight")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ECDSAStakeRegistryUpdateMinimumWeight)
				if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "UpdateMinimumWeight", log); err != nil {
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

// ParseUpdateMinimumWeight is a log parse operation binding the contract event 0x1ea42186b305fa37310450d9fb87ea1e8f0c7f447e771479e3b27634bfe84dc1.
//
// Solidity: event UpdateMinimumWeight(uint256 oldMinimumWeight, uint256 newMinimumWeight)
func (_ECDSAStakeRegistry *ECDSAStakeRegistryFilterer) ParseUpdateMinimumWeight(log types.Log) (*ECDSAStakeRegistryUpdateMinimumWeight, error) {
	event := new(ECDSAStakeRegistryUpdateMinimumWeight)
	if err := _ECDSAStakeRegistry.contract.UnpackLog(event, "UpdateMinimumWeight", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
