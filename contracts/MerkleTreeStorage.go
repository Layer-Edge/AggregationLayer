package contracts

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

// MerkleTreeStorageMetaData contains all meta data concerning the MerkleTreeStorage contract.
var MerkleTreeStorageMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"getAllRoots\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOwner\",\"inputs\":[{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTreeCount\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getTreeInfo\",\"inputs\":[{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"leafCount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"createdAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"leaves\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"initialOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"storeTree\",\"inputs\":[{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"leaves\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"treeExists\",\"inputs\":[{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TreeCreated\",\"inputs\":[{\"name\":\"merkleRoot\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"leafCount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]}]",
	Bin: "0x6080604052348015600e575f5ffd5b506112d78061001c5f395ff3fe608060405234801561000f575f5ffd5b506004361061009c575f3560e01c8063bada8fc411610064578063bada8fc414610149578063c4d66de814610167578063deb931a214610183578063f26f68bf146101b3578063f2fde38b146101e35761009c565b80630e998e5a146100a05780635028f70b146100d0578063715018a6146100ee57806384ab4db1146100f85780638da5cb5b1461012b575b5f5ffd5b6100ba60048036038101906100b59190610d55565b6101ff565b6040516100c79190610dc1565b60405180910390f35b6100d8610497565b6040516100e59190610df2565b60405180910390f35b6100f66104a3565b005b610112600480360381019061010d9190610e0b565b6104b6565b6040516101229493929190610f2c565b60405180910390f35b610133610604565b6040516101409190610f76565b60405180910390f35b610151610639565b60405161015e9190610f8f565b60405180910390f35b610181600480360381019061017c9190610fd9565b61068f565b005b61019d60048036038101906101989190610e0b565b610810565b6040516101aa9190610f76565b60405180910390f35b6101cd60048036038101906101c89190610e0b565b6108ef565b6040516101da919061101e565b60405180910390f35b6101fd60048036038101906101f89190610fd9565b610917565b005b5f835f5f1b8103610245576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161023c90611091565b60405180910390fd5b5f848490501161028a576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016102819061111f565b60405180910390fd5b5f5f8681526020019081526020015f206002015f9054906101000a900460ff16156102ea576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016102e1906111ad565b60405180910390fd5b600185908060018154018082558091505060019003905f5260205f20015f90919091909150556040518060a001604052803373ffffffffffffffffffffffffffffffffffffffff1681526020018585905081526020016001151581526020014281526020018585808060200260200160405190810160405280939291908181526020018383602002808284375f81840152601f19601f820116905080830192505050505050508152505f5f8781526020019081526020015f205f820151815f015f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550602082015181600101556040820151816002015f6101000a81548160ff021916908315150217905550606082015181600301556080820151816004019080519060200190610436929190610c53565b509050503373ffffffffffffffffffffffffffffffffffffffff16857f624f7f220d905c42f60c100138f05dc7c8e910d04f88bf7bde83a3d98389add8868690506040516104849190610df2565b60405180910390a3849150509392505050565b5f600180549050905090565b6104ab61099b565b6104b45f610a22565b565b5f5f5f6060845f5f1b8103610500576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104f790611091565b60405180910390fd5b5f5f8781526020019081526020015f206002015f9054906101000a900460ff1661055f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161055690611215565b60405180910390fd5b5f5f5f8881526020019081526020015f209050805f015f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff168160010154826003015483600401808054806020026020016040519081016040528092919081815260200182805480156105ec57602002820191905f5260205f20905b8154815260200190600101908083116105d8575b50505050509050955095509550955050509193509193565b5f5f61060e610af3565b9050805f015f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1691505090565b6060600180548060200260200160405190810160405280929190818152602001828054801561068557602002820191905f5260205f20905b815481526020019060010190808311610671575b5050505050905090565b5f610698610b1a565b90505f815f0160089054906101000a900460ff161590505f825f015f9054906101000a900467ffffffffffffffff1690505f5f8267ffffffffffffffff161480156106e05750825b90505f60018367ffffffffffffffff1614801561071357505f3073ffffffffffffffffffffffffffffffffffffffff163b145b905081158015610721575080155b15610758576040517ff92ee8a900000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6001855f015f6101000a81548167ffffffffffffffff021916908367ffffffffffffffff16021790555083156107a5576001855f0160086101000a81548160ff0219169083151502179055505b6107ae86610b2d565b8315610808575f855f0160086101000a81548160ff0219169083151502179055507fc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d260016040516107ff9190611288565b60405180910390a15b505050505050565b5f815f5f1b8103610856576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161084d90611091565b60405180910390fd5b5f5f8481526020019081526020015f206002015f9054906101000a900460ff166108b5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016108ac90611215565b60405180910390fd5b5f5f8481526020019081526020015f205f015f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff16915050919050565b5f5f5f8381526020019081526020015f206002015f9054906101000a900460ff169050919050565b61091f61099b565b5f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff160361098f575f6040517f1e4fbdf70000000000000000000000000000000000000000000000000000000081526004016109869190610f76565b60405180910390fd5b61099881610a22565b50565b6109a3610b41565b73ffffffffffffffffffffffffffffffffffffffff166109c1610604565b73ffffffffffffffffffffffffffffffffffffffff1614610a20576109e4610b41565b6040517f118cdaa7000000000000000000000000000000000000000000000000000000008152600401610a179190610f76565b60405180910390fd5b565b5f610a2b610af3565b90505f815f015f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905082825f015f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508273ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a3505050565b5f7f9016d09d72d40fdae2fd8ceac6b6234c7706214fd39c1cd1e609a0528c199300905090565b5f5f610b24610b48565b90508091505090565b610b35610b71565b610b3e81610bb1565b50565b5f33905090565b5f7ff0c57e16840df040f15088dc2f81fe391c3923bec73e23a9662efc9c229c6a005f1b905090565b610b79610c35565b610baf576040517fd7e6bcf800000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b565b610bb9610b71565b5f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1603610c29575f6040517f1e4fbdf7000000000000000000000000000000000000000000000000000000008152600401610c209190610f76565b60405180910390fd5b610c3281610a22565b50565b5f610c3e610b1a565b5f0160089054906101000a900460ff16905090565b828054828255905f5260205f20908101928215610c8d579160200282015b82811115610c8c578251825591602001919060010190610c71565b5b509050610c9a9190610c9e565b5090565b5b80821115610cb5575f815f905550600101610c9f565b5090565b5f5ffd5b5f5ffd5b5f819050919050565b610cd381610cc1565b8114610cdd575f5ffd5b50565b5f81359050610cee81610cca565b92915050565b5f5ffd5b5f5ffd5b5f5ffd5b5f5f83601f840112610d1557610d14610cf4565b5b8235905067ffffffffffffffff811115610d3257610d31610cf8565b5b602083019150836020820283011115610d4e57610d4d610cfc565b5b9250929050565b5f5f5f60408486031215610d6c57610d6b610cb9565b5b5f610d7986828701610ce0565b935050602084013567ffffffffffffffff811115610d9a57610d99610cbd565b5b610da686828701610d00565b92509250509250925092565b610dbb81610cc1565b82525050565b5f602082019050610dd45f830184610db2565b92915050565b5f819050919050565b610dec81610dda565b82525050565b5f602082019050610e055f830184610de3565b92915050565b5f60208284031215610e2057610e1f610cb9565b5b5f610e2d84828501610ce0565b91505092915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610e5f82610e36565b9050919050565b610e6f81610e55565b82525050565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b610ea781610cc1565b82525050565b5f610eb88383610e9e565b60208301905092915050565b5f602082019050919050565b5f610eda82610e75565b610ee48185610e7f565b9350610eef83610e8f565b805f5b83811015610f1f578151610f068882610ead565b9750610f1183610ec4565b925050600181019050610ef2565b5085935050505092915050565b5f608082019050610f3f5f830187610e66565b610f4c6020830186610de3565b610f596040830185610de3565b8181036060830152610f6b8184610ed0565b905095945050505050565b5f602082019050610f895f830184610e66565b92915050565b5f6020820190508181035f830152610fa78184610ed0565b905092915050565b610fb881610e55565b8114610fc2575f5ffd5b50565b5f81359050610fd381610faf565b92915050565b5f60208284031215610fee57610fed610cb9565b5b5f610ffb84828501610fc5565b91505092915050565b5f8115159050919050565b61101881611004565b82525050565b5f6020820190506110315f83018461100f565b92915050565b5f82825260208201905092915050565b7f496e76616c6964206d65726b6c6520726f6f74000000000000000000000000005f82015250565b5f61107b601383611037565b915061108682611047565b602082019050919050565b5f6020820190508181035f8301526110a88161106f565b9050919050565b7f4c65616620636f756e74206d7573742062652067726561746572207468616e205f8201527f3000000000000000000000000000000000000000000000000000000000000000602082015250565b5f611109602183611037565b9150611114826110af565b604082019050919050565b5f6020820190508181035f830152611136816110fd565b9050919050565b7f547265652077697468207468697320726f6f7420616c726561647920657869735f8201527f7473000000000000000000000000000000000000000000000000000000000000602082015250565b5f611197602283611037565b91506111a28261113d565b604082019050919050565b5f6020820190508181035f8301526111c48161118b565b9050919050565b7f5472656520646f6573206e6f74206578697374000000000000000000000000005f82015250565b5f6111ff601383611037565b915061120a826111cb565b602082019050919050565b5f6020820190508181035f83015261122c816111f3565b9050919050565b5f819050919050565b5f67ffffffffffffffff82169050919050565b5f819050919050565b5f61127261126d61126884611233565b61124f565b61123c565b9050919050565b61128281611258565b82525050565b5f60208201905061129b5f830184611279565b9291505056fea2646970667358221220e8b06b5cb53bad6cf5d64e3200de0a6700603f3982f6cc9dc61a571388d04db164736f6c634300081e0033",
}

// MerkleTreeStorageABI is the input ABI used to generate the binding from.
// Deprecated: Use MerkleTreeStorageMetaData.ABI instead.
var MerkleTreeStorageABI = MerkleTreeStorageMetaData.ABI

// MerkleTreeStorageBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use MerkleTreeStorageMetaData.Bin instead.
var MerkleTreeStorageBin = MerkleTreeStorageMetaData.Bin

// DeployMerkleTreeStorage deploys a new Ethereum contract, binding an instance of MerkleTreeStorage to it.
func DeployMerkleTreeStorage(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *MerkleTreeStorage, error) {
	parsed, err := MerkleTreeStorageMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(MerkleTreeStorageBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &MerkleTreeStorage{MerkleTreeStorageCaller: MerkleTreeStorageCaller{contract: contract}, MerkleTreeStorageTransactor: MerkleTreeStorageTransactor{contract: contract}, MerkleTreeStorageFilterer: MerkleTreeStorageFilterer{contract: contract}}, nil
}

// MerkleTreeStorage is an auto generated Go binding around an Ethereum contract.
type MerkleTreeStorage struct {
	MerkleTreeStorageCaller     // Read-only binding to the contract
	MerkleTreeStorageTransactor // Write-only binding to the contract
	MerkleTreeStorageFilterer   // Log filterer for contract events
}

// MerkleTreeStorageCaller is an auto generated read-only Go binding around an Ethereum contract.
type MerkleTreeStorageCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MerkleTreeStorageTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MerkleTreeStorageTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MerkleTreeStorageFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MerkleTreeStorageFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MerkleTreeStorageSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MerkleTreeStorageSession struct {
	Contract     *MerkleTreeStorage // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// MerkleTreeStorageCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MerkleTreeStorageCallerSession struct {
	Contract *MerkleTreeStorageCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// MerkleTreeStorageTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MerkleTreeStorageTransactorSession struct {
	Contract     *MerkleTreeStorageTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// MerkleTreeStorageRaw is an auto generated low-level Go binding around an Ethereum contract.
type MerkleTreeStorageRaw struct {
	Contract *MerkleTreeStorage // Generic contract binding to access the raw methods on
}

// MerkleTreeStorageCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MerkleTreeStorageCallerRaw struct {
	Contract *MerkleTreeStorageCaller // Generic read-only contract binding to access the raw methods on
}

// MerkleTreeStorageTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MerkleTreeStorageTransactorRaw struct {
	Contract *MerkleTreeStorageTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMerkleTreeStorage creates a new instance of MerkleTreeStorage, bound to a specific deployed contract.
func NewMerkleTreeStorage(address common.Address, backend bind.ContractBackend) (*MerkleTreeStorage, error) {
	contract, err := bindMerkleTreeStorage(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MerkleTreeStorage{MerkleTreeStorageCaller: MerkleTreeStorageCaller{contract: contract}, MerkleTreeStorageTransactor: MerkleTreeStorageTransactor{contract: contract}, MerkleTreeStorageFilterer: MerkleTreeStorageFilterer{contract: contract}}, nil
}

// NewMerkleTreeStorageCaller creates a new read-only instance of MerkleTreeStorage, bound to a specific deployed contract.
func NewMerkleTreeStorageCaller(address common.Address, caller bind.ContractCaller) (*MerkleTreeStorageCaller, error) {
	contract, err := bindMerkleTreeStorage(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MerkleTreeStorageCaller{contract: contract}, nil
}

// NewMerkleTreeStorageTransactor creates a new write-only instance of MerkleTreeStorage, bound to a specific deployed contract.
func NewMerkleTreeStorageTransactor(address common.Address, transactor bind.ContractTransactor) (*MerkleTreeStorageTransactor, error) {
	contract, err := bindMerkleTreeStorage(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MerkleTreeStorageTransactor{contract: contract}, nil
}

// NewMerkleTreeStorageFilterer creates a new log filterer instance of MerkleTreeStorage, bound to a specific deployed contract.
func NewMerkleTreeStorageFilterer(address common.Address, filterer bind.ContractFilterer) (*MerkleTreeStorageFilterer, error) {
	contract, err := bindMerkleTreeStorage(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MerkleTreeStorageFilterer{contract: contract}, nil
}

// bindMerkleTreeStorage binds a generic wrapper to an already deployed contract.
func bindMerkleTreeStorage(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := MerkleTreeStorageMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MerkleTreeStorage *MerkleTreeStorageRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MerkleTreeStorage.Contract.MerkleTreeStorageCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MerkleTreeStorage *MerkleTreeStorageRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MerkleTreeStorage.Contract.MerkleTreeStorageTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MerkleTreeStorage *MerkleTreeStorageRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MerkleTreeStorage.Contract.MerkleTreeStorageTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MerkleTreeStorage *MerkleTreeStorageCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MerkleTreeStorage.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MerkleTreeStorage *MerkleTreeStorageTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MerkleTreeStorage.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MerkleTreeStorage *MerkleTreeStorageTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MerkleTreeStorage.Contract.contract.Transact(opts, method, params...)
}

// GetAllRoots is a free data retrieval call binding the contract method 0xbada8fc4.
//
// Solidity: function getAllRoots() view returns(bytes32[])
func (_MerkleTreeStorage *MerkleTreeStorageCaller) GetAllRoots(opts *bind.CallOpts) ([][32]byte, error) {
	var out []interface{}
	err := _MerkleTreeStorage.contract.Call(opts, &out, "getAllRoots")

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetAllRoots is a free data retrieval call binding the contract method 0xbada8fc4.
//
// Solidity: function getAllRoots() view returns(bytes32[])
func (_MerkleTreeStorage *MerkleTreeStorageSession) GetAllRoots() ([][32]byte, error) {
	return _MerkleTreeStorage.Contract.GetAllRoots(&_MerkleTreeStorage.CallOpts)
}

// GetAllRoots is a free data retrieval call binding the contract method 0xbada8fc4.
//
// Solidity: function getAllRoots() view returns(bytes32[])
func (_MerkleTreeStorage *MerkleTreeStorageCallerSession) GetAllRoots() ([][32]byte, error) {
	return _MerkleTreeStorage.Contract.GetAllRoots(&_MerkleTreeStorage.CallOpts)
}

// GetOwner is a free data retrieval call binding the contract method 0xdeb931a2.
//
// Solidity: function getOwner(bytes32 merkleRoot) view returns(address)
func (_MerkleTreeStorage *MerkleTreeStorageCaller) GetOwner(opts *bind.CallOpts, merkleRoot [32]byte) (common.Address, error) {
	var out []interface{}
	err := _MerkleTreeStorage.contract.Call(opts, &out, "getOwner", merkleRoot)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetOwner is a free data retrieval call binding the contract method 0xdeb931a2.
//
// Solidity: function getOwner(bytes32 merkleRoot) view returns(address)
func (_MerkleTreeStorage *MerkleTreeStorageSession) GetOwner(merkleRoot [32]byte) (common.Address, error) {
	return _MerkleTreeStorage.Contract.GetOwner(&_MerkleTreeStorage.CallOpts, merkleRoot)
}

// GetOwner is a free data retrieval call binding the contract method 0xdeb931a2.
//
// Solidity: function getOwner(bytes32 merkleRoot) view returns(address)
func (_MerkleTreeStorage *MerkleTreeStorageCallerSession) GetOwner(merkleRoot [32]byte) (common.Address, error) {
	return _MerkleTreeStorage.Contract.GetOwner(&_MerkleTreeStorage.CallOpts, merkleRoot)
}

// GetTreeCount is a free data retrieval call binding the contract method 0x5028f70b.
//
// Solidity: function getTreeCount() view returns(uint256)
func (_MerkleTreeStorage *MerkleTreeStorageCaller) GetTreeCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MerkleTreeStorage.contract.Call(opts, &out, "getTreeCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTreeCount is a free data retrieval call binding the contract method 0x5028f70b.
//
// Solidity: function getTreeCount() view returns(uint256)
func (_MerkleTreeStorage *MerkleTreeStorageSession) GetTreeCount() (*big.Int, error) {
	return _MerkleTreeStorage.Contract.GetTreeCount(&_MerkleTreeStorage.CallOpts)
}

// GetTreeCount is a free data retrieval call binding the contract method 0x5028f70b.
//
// Solidity: function getTreeCount() view returns(uint256)
func (_MerkleTreeStorage *MerkleTreeStorageCallerSession) GetTreeCount() (*big.Int, error) {
	return _MerkleTreeStorage.Contract.GetTreeCount(&_MerkleTreeStorage.CallOpts)
}

// GetTreeInfo is a free data retrieval call binding the contract method 0x84ab4db1.
//
// Solidity: function getTreeInfo(bytes32 merkleRoot) view returns(address owner, uint256 leafCount, uint256 createdAt, bytes32[] leaves)
func (_MerkleTreeStorage *MerkleTreeStorageCaller) GetTreeInfo(opts *bind.CallOpts, merkleRoot [32]byte) (struct {
	Owner     common.Address
	LeafCount *big.Int
	CreatedAt *big.Int
	Leaves    [][32]byte
}, error) {
	var out []interface{}
	err := _MerkleTreeStorage.contract.Call(opts, &out, "getTreeInfo", merkleRoot)

	outstruct := new(struct {
		Owner     common.Address
		LeafCount *big.Int
		CreatedAt *big.Int
		Leaves    [][32]byte
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Owner = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.LeafCount = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.CreatedAt = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Leaves = *abi.ConvertType(out[3], new([][32]byte)).(*[][32]byte)

	return *outstruct, err

}

// GetTreeInfo is a free data retrieval call binding the contract method 0x84ab4db1.
//
// Solidity: function getTreeInfo(bytes32 merkleRoot) view returns(address owner, uint256 leafCount, uint256 createdAt, bytes32[] leaves)
func (_MerkleTreeStorage *MerkleTreeStorageSession) GetTreeInfo(merkleRoot [32]byte) (struct {
	Owner     common.Address
	LeafCount *big.Int
	CreatedAt *big.Int
	Leaves    [][32]byte
}, error) {
	return _MerkleTreeStorage.Contract.GetTreeInfo(&_MerkleTreeStorage.CallOpts, merkleRoot)
}

// GetTreeInfo is a free data retrieval call binding the contract method 0x84ab4db1.
//
// Solidity: function getTreeInfo(bytes32 merkleRoot) view returns(address owner, uint256 leafCount, uint256 createdAt, bytes32[] leaves)
func (_MerkleTreeStorage *MerkleTreeStorageCallerSession) GetTreeInfo(merkleRoot [32]byte) (struct {
	Owner     common.Address
	LeafCount *big.Int
	CreatedAt *big.Int
	Leaves    [][32]byte
}, error) {
	return _MerkleTreeStorage.Contract.GetTreeInfo(&_MerkleTreeStorage.CallOpts, merkleRoot)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_MerkleTreeStorage *MerkleTreeStorageCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MerkleTreeStorage.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_MerkleTreeStorage *MerkleTreeStorageSession) Owner() (common.Address, error) {
	return _MerkleTreeStorage.Contract.Owner(&_MerkleTreeStorage.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_MerkleTreeStorage *MerkleTreeStorageCallerSession) Owner() (common.Address, error) {
	return _MerkleTreeStorage.Contract.Owner(&_MerkleTreeStorage.CallOpts)
}

// TreeExists is a free data retrieval call binding the contract method 0xf26f68bf.
//
// Solidity: function treeExists(bytes32 merkleRoot) view returns(bool)
func (_MerkleTreeStorage *MerkleTreeStorageCaller) TreeExists(opts *bind.CallOpts, merkleRoot [32]byte) (bool, error) {
	var out []interface{}
	err := _MerkleTreeStorage.contract.Call(opts, &out, "treeExists", merkleRoot)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// TreeExists is a free data retrieval call binding the contract method 0xf26f68bf.
//
// Solidity: function treeExists(bytes32 merkleRoot) view returns(bool)
func (_MerkleTreeStorage *MerkleTreeStorageSession) TreeExists(merkleRoot [32]byte) (bool, error) {
	return _MerkleTreeStorage.Contract.TreeExists(&_MerkleTreeStorage.CallOpts, merkleRoot)
}

// TreeExists is a free data retrieval call binding the contract method 0xf26f68bf.
//
// Solidity: function treeExists(bytes32 merkleRoot) view returns(bool)
func (_MerkleTreeStorage *MerkleTreeStorageCallerSession) TreeExists(merkleRoot [32]byte) (bool, error) {
	return _MerkleTreeStorage.Contract.TreeExists(&_MerkleTreeStorage.CallOpts, merkleRoot)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address initialOwner) returns()
func (_MerkleTreeStorage *MerkleTreeStorageTransactor) Initialize(opts *bind.TransactOpts, initialOwner common.Address) (*types.Transaction, error) {
	return _MerkleTreeStorage.contract.Transact(opts, "initialize", initialOwner)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address initialOwner) returns()
func (_MerkleTreeStorage *MerkleTreeStorageSession) Initialize(initialOwner common.Address) (*types.Transaction, error) {
	return _MerkleTreeStorage.Contract.Initialize(&_MerkleTreeStorage.TransactOpts, initialOwner)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address initialOwner) returns()
func (_MerkleTreeStorage *MerkleTreeStorageTransactorSession) Initialize(initialOwner common.Address) (*types.Transaction, error) {
	return _MerkleTreeStorage.Contract.Initialize(&_MerkleTreeStorage.TransactOpts, initialOwner)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_MerkleTreeStorage *MerkleTreeStorageTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MerkleTreeStorage.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_MerkleTreeStorage *MerkleTreeStorageSession) RenounceOwnership() (*types.Transaction, error) {
	return _MerkleTreeStorage.Contract.RenounceOwnership(&_MerkleTreeStorage.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_MerkleTreeStorage *MerkleTreeStorageTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _MerkleTreeStorage.Contract.RenounceOwnership(&_MerkleTreeStorage.TransactOpts)
}

// StoreTree is a paid mutator transaction binding the contract method 0x0e998e5a.
//
// Solidity: function storeTree(bytes32 merkleRoot, bytes32[] leaves) returns(bytes32)
func (_MerkleTreeStorage *MerkleTreeStorageTransactor) StoreTree(opts *bind.TransactOpts, merkleRoot [32]byte, leaves [][32]byte) (*types.Transaction, error) {
	return _MerkleTreeStorage.contract.Transact(opts, "storeTree", merkleRoot, leaves)
}

// StoreTree is a paid mutator transaction binding the contract method 0x0e998e5a.
//
// Solidity: function storeTree(bytes32 merkleRoot, bytes32[] leaves) returns(bytes32)
func (_MerkleTreeStorage *MerkleTreeStorageSession) StoreTree(merkleRoot [32]byte, leaves [][32]byte) (*types.Transaction, error) {
	return _MerkleTreeStorage.Contract.StoreTree(&_MerkleTreeStorage.TransactOpts, merkleRoot, leaves)
}

// StoreTree is a paid mutator transaction binding the contract method 0x0e998e5a.
//
// Solidity: function storeTree(bytes32 merkleRoot, bytes32[] leaves) returns(bytes32)
func (_MerkleTreeStorage *MerkleTreeStorageTransactorSession) StoreTree(merkleRoot [32]byte, leaves [][32]byte) (*types.Transaction, error) {
	return _MerkleTreeStorage.Contract.StoreTree(&_MerkleTreeStorage.TransactOpts, merkleRoot, leaves)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_MerkleTreeStorage *MerkleTreeStorageTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _MerkleTreeStorage.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_MerkleTreeStorage *MerkleTreeStorageSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _MerkleTreeStorage.Contract.TransferOwnership(&_MerkleTreeStorage.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_MerkleTreeStorage *MerkleTreeStorageTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _MerkleTreeStorage.Contract.TransferOwnership(&_MerkleTreeStorage.TransactOpts, newOwner)
}

// MerkleTreeStorageInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the MerkleTreeStorage contract.
type MerkleTreeStorageInitializedIterator struct {
	Event *MerkleTreeStorageInitialized // Event containing the contract specifics and raw log

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
func (it *MerkleTreeStorageInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MerkleTreeStorageInitialized)
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
		it.Event = new(MerkleTreeStorageInitialized)
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
func (it *MerkleTreeStorageInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MerkleTreeStorageInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MerkleTreeStorageInitialized represents a Initialized event raised by the MerkleTreeStorage contract.
type MerkleTreeStorageInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_MerkleTreeStorage *MerkleTreeStorageFilterer) FilterInitialized(opts *bind.FilterOpts) (*MerkleTreeStorageInitializedIterator, error) {

	logs, sub, err := _MerkleTreeStorage.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &MerkleTreeStorageInitializedIterator{contract: _MerkleTreeStorage.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_MerkleTreeStorage *MerkleTreeStorageFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *MerkleTreeStorageInitialized) (event.Subscription, error) {

	logs, sub, err := _MerkleTreeStorage.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MerkleTreeStorageInitialized)
				if err := _MerkleTreeStorage.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_MerkleTreeStorage *MerkleTreeStorageFilterer) ParseInitialized(log types.Log) (*MerkleTreeStorageInitialized, error) {
	event := new(MerkleTreeStorageInitialized)
	if err := _MerkleTreeStorage.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MerkleTreeStorageOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the MerkleTreeStorage contract.
type MerkleTreeStorageOwnershipTransferredIterator struct {
	Event *MerkleTreeStorageOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *MerkleTreeStorageOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MerkleTreeStorageOwnershipTransferred)
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
		it.Event = new(MerkleTreeStorageOwnershipTransferred)
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
func (it *MerkleTreeStorageOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MerkleTreeStorageOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MerkleTreeStorageOwnershipTransferred represents a OwnershipTransferred event raised by the MerkleTreeStorage contract.
type MerkleTreeStorageOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_MerkleTreeStorage *MerkleTreeStorageFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*MerkleTreeStorageOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _MerkleTreeStorage.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &MerkleTreeStorageOwnershipTransferredIterator{contract: _MerkleTreeStorage.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_MerkleTreeStorage *MerkleTreeStorageFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *MerkleTreeStorageOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _MerkleTreeStorage.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MerkleTreeStorageOwnershipTransferred)
				if err := _MerkleTreeStorage.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_MerkleTreeStorage *MerkleTreeStorageFilterer) ParseOwnershipTransferred(log types.Log) (*MerkleTreeStorageOwnershipTransferred, error) {
	event := new(MerkleTreeStorageOwnershipTransferred)
	if err := _MerkleTreeStorage.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MerkleTreeStorageTreeCreatedIterator is returned from FilterTreeCreated and is used to iterate over the raw logs and unpacked data for TreeCreated events raised by the MerkleTreeStorage contract.
type MerkleTreeStorageTreeCreatedIterator struct {
	Event *MerkleTreeStorageTreeCreated // Event containing the contract specifics and raw log

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
func (it *MerkleTreeStorageTreeCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MerkleTreeStorageTreeCreated)
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
		it.Event = new(MerkleTreeStorageTreeCreated)
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
func (it *MerkleTreeStorageTreeCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MerkleTreeStorageTreeCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MerkleTreeStorageTreeCreated represents a TreeCreated event raised by the MerkleTreeStorage contract.
type MerkleTreeStorageTreeCreated struct {
	MerkleRoot [32]byte
	Owner      common.Address
	LeafCount  *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterTreeCreated is a free log retrieval operation binding the contract event 0x624f7f220d905c42f60c100138f05dc7c8e910d04f88bf7bde83a3d98389add8.
//
// Solidity: event TreeCreated(bytes32 indexed merkleRoot, address indexed owner, uint256 leafCount)
func (_MerkleTreeStorage *MerkleTreeStorageFilterer) FilterTreeCreated(opts *bind.FilterOpts, merkleRoot [][32]byte, owner []common.Address) (*MerkleTreeStorageTreeCreatedIterator, error) {

	var merkleRootRule []interface{}
	for _, merkleRootItem := range merkleRoot {
		merkleRootRule = append(merkleRootRule, merkleRootItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _MerkleTreeStorage.contract.FilterLogs(opts, "TreeCreated", merkleRootRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &MerkleTreeStorageTreeCreatedIterator{contract: _MerkleTreeStorage.contract, event: "TreeCreated", logs: logs, sub: sub}, nil
}

// WatchTreeCreated is a free log subscription operation binding the contract event 0x624f7f220d905c42f60c100138f05dc7c8e910d04f88bf7bde83a3d98389add8.
//
// Solidity: event TreeCreated(bytes32 indexed merkleRoot, address indexed owner, uint256 leafCount)
func (_MerkleTreeStorage *MerkleTreeStorageFilterer) WatchTreeCreated(opts *bind.WatchOpts, sink chan<- *MerkleTreeStorageTreeCreated, merkleRoot [][32]byte, owner []common.Address) (event.Subscription, error) {

	var merkleRootRule []interface{}
	for _, merkleRootItem := range merkleRoot {
		merkleRootRule = append(merkleRootRule, merkleRootItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _MerkleTreeStorage.contract.WatchLogs(opts, "TreeCreated", merkleRootRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MerkleTreeStorageTreeCreated)
				if err := _MerkleTreeStorage.contract.UnpackLog(event, "TreeCreated", log); err != nil {
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

// ParseTreeCreated is a log parse operation binding the contract event 0x624f7f220d905c42f60c100138f05dc7c8e910d04f88bf7bde83a3d98389add8.
//
// Solidity: event TreeCreated(bytes32 indexed merkleRoot, address indexed owner, uint256 leafCount)
func (_MerkleTreeStorage *MerkleTreeStorageFilterer) ParseTreeCreated(log types.Log) (*MerkleTreeStorageTreeCreated, error) {
	event := new(MerkleTreeStorageTreeCreated)
	if err := _MerkleTreeStorage.contract.UnpackLog(event, "TreeCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
