package app

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	gethcmn "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/smartbch/smartbch/internal/ethutils"
	"github.com/smartbch/smartbch/internal/testutils"
)

var _sep101ABI = testutils.MustParseABI(`
[
{
  "inputs": [
	{
	  "internalType": "bytes",
	  "name": "key",
	  "type": "bytes"
	},
	{
	  "internalType": "bytes",
	  "name": "value",
	  "type": "bytes"
	}
  ],
  "name": "set",
  "outputs": [],
  "stateMutability": "nonpayable",
  "type": "function"
},
{
  "inputs": [
	{
	  "internalType": "bytes",
	  "name": "key",
	  "type": "bytes"
	}
  ],
  "name": "get",
  "outputs": [
	{
	  "internalType": "bytes",
	  "name": "",
	  "type": "bytes"
	}
  ],
  "stateMutability": "nonpayable",
  "type": "function"
}
]
`)

// see testdata/seps/contracts/SEP101Proxy.sol
var _sep101ProxyCreationBytecode = testutils.HexToBytes(`
608060405234801561001057600080fd5b50610789806100206000396000f3fe
608060405234801561001057600080fd5b50600436106100415760003560e01c
8063a18c751e14610046578063d6d7d52514610062578063f5ff5c7614610092
575b600080fd5b610060600480360381019061005b9190610408565b6100b056
5b005b61007c600480360381019061007791906103c3565b6101d0565b604051
61008991906105f5565b60405180910390f35b61009a61030b565b6040516100
a7919061057b565b60405180910390f35b61271273ffffffffffffffffffffff
ffffffffffffffffff166040518060400160405280601081526020017f736574
2862797465732c62797465732900000000000000000000000000000000815250
805190602001208585858560405160240161011d94939291906105ba565b6040
51602081830303815290604052907bffffffffffffffffffffffffffffffffff
ffffffffffffffffffffff19166020820180517bffffffffffffffffffffffff
ffffffffffffffffffffffffffffffff83818316178352505050506040516101
879190610564565b600060405180830381855af49150503d80600081146101c2
576040519150601f19603f3d011682016040523d82523d6000602084013e6101
c7565b606091505b50505050505050565b606060008061271273ffffffffffff
ffffffffffffffffffffffffffff166040518060400160405280600a81526020
017f676574286279746573290000000000000000000000000000000000000000
000081525080519060200120868660405160240161023e929190610596565b60
4051602081830303815290604052907bffffffffffffffffffffffffffffffff
ffffffffffffffffffffffff19166020820180517bffffffffffffffffffffff
ffffffffffffffffffffffffffffffffff838183161783525050505060405161
02a89190610564565b600060405180830381855af49150503d80600081146102
e3576040519150601f19603f3d011682016040523d82523d6000602084013e61
02e8565b606091505b509150915080806020019051810190610301919061047d
565b9250505092915050565b61271281565b600061032461031f84610648565b
610617565b90508281526020810184848401111561033c57600080fd5b610347
8482856106e0565b509392505050565b60008083601f84011261036157600080
fd5b8235905067ffffffffffffffff81111561037a57600080fd5b6020830191
5083600182028301111561039257600080fd5b9250929050565b600082601f83
01126103aa57600080fd5b81516103ba848260208601610311565b9150509291
5050565b600080602083850312156103d657600080fd5b600083013567ffffff
ffffffffff8111156103f057600080fd5b6103fc8582860161034f565b925092
50509250929050565b6000806000806040858703121561041e57600080fd5b60
0085013567ffffffffffffffff81111561043857600080fd5b61044487828801
61034f565b9450945050602085013567ffffffffffffffff8111156104635760
0080fd5b61046f8782880161034f565b925092505092959194509250565b6000
6020828403121561048f57600080fd5b600082015167ffffffffffffffff8111
156104a957600080fd5b6104b584828501610399565b91505092915050565b61
04c78161069f565b82525050565b60006104d98385610683565b93506104e683
85846106d1565b6104ef83610742565b840190509392505050565b6000610505
82610678565b61050f8185610683565b935061051f8185602086016106e0565b
61052881610742565b840191505092915050565b600061053e82610678565b61
05488185610694565b93506105588185602086016106e0565b80840191505092
915050565b60006105708284610533565b915081905092915050565b60006020
8201905061059060008301846104be565b92915050565b600060208201905081
810360008301526105b18184866104cd565b90509392505050565b6000604082
01905081810360008301526105d58186886104cd565b90508181036020830152
6105ea8184866104cd565b905095945050505050565b60006020820190508181
03600083015261060f81846104fa565b905092915050565b6000604051905081
810181811067ffffffffffffffff8211171561063e5761063d610713565b5b80
60405250919050565b600067ffffffffffffffff821115610663576106626107
13565b5b601f19601f8301169050602081019050919050565b60008151905091
9050565b600082825260208201905092915050565b600081905092915050565b
60006106aa826106b1565b9050919050565b600073ffffffffffffffffffffff
ffffffffffffffffff82169050919050565b8281833760008383015250505056
5b60005b838110156106fe5780820151818401526020810190506106e3565b83
81111561070d576000848401525b50505050565b7f4e487b7100000000000000
0000000000000000000000000000000000000000006000526041600452602460
00fd5b6000601f19601f830116905091905056fea2646970667358221220fb43
e5de781849eed8012fcec370be5db5d6a1c9ad988f9fc8d3b49cde95ffe36473
6f6c63430008000033
`)

func deploySEP101Proxy(t *testing.T, _app *App, privKey string, senderAddr gethcmn.Address) gethcmn.Address {
	tx1 := gethtypes.NewContractCreation(0, big.NewInt(0), 1000000, big.NewInt(1),
		_sep101ProxyCreationBytecode)
	tx1 = ethutils.MustSignTx(tx1, _app.chainId.ToBig(), ethutils.MustHexToPrivKey(privKey))

	testutils.ExecTxInBlock(_app, 1, tx1)
	contractAddr := gethcrypto.CreateAddress(senderAddr, tx1.Nonce())
	code := getCode(_app, contractAddr)
	require.True(t, len(code) > 0)
	return contractAddr
}

func TestSEP101(t *testing.T) {
	privKey, addr := testutils.GenKeyAndAddr()
	_app := CreateTestApp(privKey)
	defer DestroyTestApp(_app)

	// deploy proxy
	contractAddr := deploySEP101Proxy(t, _app, privKey, addr)

	key := []byte{0xAB, 0xCD}
	val := bytes.Repeat([]byte{0x12, 0x34}, 500)

	// call set()
	data := _sep101ABI.MustPack("set", key, val)
	tx2 := gethtypes.NewTransaction(1, contractAddr, big.NewInt(0), 1000000, big.NewInt(1), data)
	tx2 = ethutils.MustSignTx(tx2, _app.chainId.ToBig(), ethutils.MustHexToPrivKey(privKey))
	testutils.ExecTxInBlock(_app, 3, tx2)

	blk3 := getBlock(_app, 3)
	require.Equal(t, int64(3), blk3.Number)
	require.Len(t, blk3.Transactions, 1)
	txInBlk3 := getTx(_app, blk3.Transactions[0])
	require.Equal(t, gethtypes.ReceiptStatusSuccessful, txInBlk3.Status)
	require.Equal(t, "success", txInBlk3.StatusStr)
	require.Equal(t, tx2.Hash(), gethcmn.Hash(txInBlk3.Hash))

	// call get()
	data = _sep101ABI.MustPack("get", key)
	tx4 := gethtypes.NewTransaction(0, contractAddr, big.NewInt(0), 10000000, big.NewInt(1), data)
	statusCode, statusStr, output := call(_app, addr, tx4)
	require.Equal(t, 0, statusCode)
	require.Equal(t, "success", statusStr)
	require.Equal(t, []interface{}{val}, _sep101ABI.MustUnpack("get", output))

	// get non-existing key
	data = _sep101ABI.MustPack("get", []byte{9, 9, 9})
	tx5 := gethtypes.NewTransaction(0, contractAddr, big.NewInt(0), 10000000, big.NewInt(1), data)
	statusCode, statusStr, output = call(_app, addr, tx5)
	require.Equal(t, "success", statusStr)
	require.Equal(t, 0, statusCode)
	require.Equal(t, []interface{}{[]byte{}}, _sep101ABI.MustUnpack("get", output))
}

func TestSEP101_setZeroLenKey(t *testing.T) {
	privKey, addr := testutils.GenKeyAndAddr()
	_app := CreateTestApp(privKey)
	defer DestroyTestApp(_app)

	// deploy proxy
	contractAddr := deploySEP101Proxy(t, _app, privKey, addr)

	// set() with zero-len key
	data := _sep101ABI.MustPack("set", []byte{}, []byte{1, 2, 3})
	tx2 := gethtypes.NewTransaction(1, contractAddr, big.NewInt(0), 1000000, big.NewInt(1), data)
	tx2 = ethutils.MustSignTx(tx2, _app.chainId.ToBig(), ethutils.MustHexToPrivKey(privKey))
	testutils.ExecTxInBlock(_app, 3, tx2)

	blk3 := getBlock(_app, 3)
	require.Equal(t, int64(3), blk3.Number)
	require.Len(t, blk3.Transactions, 1)
	txInBlk3 := getTx(_app, blk3.Transactions[0])
	//require.Equal(t, 2, txInBlk3.Status)
	require.Equal(t, "precompile-failure", txInBlk3.StatusStr)
	require.Equal(t, tx2.Hash(), gethcmn.Hash(txInBlk3.Hash))
}

func TestSEP101_setKeyTooLong(t *testing.T) {
	privKey, addr := testutils.GenKeyAndAddr()
	_app := CreateTestApp(privKey)
	defer DestroyTestApp(_app)

	// deploy proxy
	contractAddr := deploySEP101Proxy(t, _app, privKey, addr)

	// set() with looooong key
	data := _sep101ABI.MustPack("set", bytes.Repeat([]byte{39}, 257), []byte{1, 2, 3})
	tx2 := gethtypes.NewTransaction(1, contractAddr, big.NewInt(0), 1000000, big.NewInt(1), data)
	tx2 = ethutils.MustSignTx(tx2, _app.chainId.ToBig(), ethutils.MustHexToPrivKey(privKey))
	testutils.ExecTxInBlock(_app, 3, tx2)

	blk3 := getBlock(_app, 3)
	require.Equal(t, int64(3), blk3.Number)
	require.Len(t, blk3.Transactions, 1)
	txInBlk3 := getTx(_app, blk3.Transactions[0])
	//require.Equal(t, 2, txInBlk3.Status)
	require.Equal(t, "precompile-failure", txInBlk3.StatusStr)
	require.Equal(t, tx2.Hash(), gethcmn.Hash(txInBlk3.Hash))
}

func TestSEP101_setValTooLong(t *testing.T) {
	privKey, addr := testutils.GenKeyAndAddr()
	_app := CreateTestApp(privKey)
	defer DestroyTestApp(_app)

	// deploy proxy
	contractAddr := deploySEP101Proxy(t, _app, privKey, addr)

	// set() with looooong val
	data := _sep101ABI.MustPack("set", []byte{1, 2, 3}, bytes.Repeat([]byte{39}, 24*1024+1))
	tx2 := gethtypes.NewTransaction(1, contractAddr, big.NewInt(0), 1000000, big.NewInt(1), data)
	tx2 = ethutils.MustSignTx(tx2, _app.chainId.ToBig(), ethutils.MustHexToPrivKey(privKey))
	testutils.ExecTxInBlock(_app, 3, tx2)

	blk3 := getBlock(_app, 3)
	require.Equal(t, int64(3), blk3.Number)
	require.Len(t, blk3.Transactions, 1)
	txInBlk3 := getTx(_app, blk3.Transactions[0])
	//require.Equal(t, 2, txInBlk3.Status)
	require.Equal(t, "precompile-failure", txInBlk3.StatusStr)
	require.Equal(t, tx2.Hash(), gethcmn.Hash(txInBlk3.Hash))
}
