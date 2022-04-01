package chaincode_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-samples/asset-transfer-basic/chaincode-go/chaincode"
	"github.com/hyperledger/fabric-samples/asset-transfer-basic/chaincode-go/chaincode/mocks"
	"github.com/stretchr/testify/require"
)

//go:generate counterfeiter -o mocks/transaction.go -fake-name TransactionContext . transactionContext
type transactionContext interface {
	contractapi.TransactionContextInterface
}

//go:generate counterfeiter -o mocks/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

//go:generate counterfeiter -o mocks/statequeryiterator.go -fake-name StateQueryIterator . stateQueryIterator
type stateQueryIterator interface {
	shim.StateQueryIteratorInterface
}

func TestInitLedger(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	assetTransfer := chaincode.SmartContract{}
	err := assetTransfer.InitLedger(transactionContext)
	require.NoError(t, err)

	errMsg := "failed inserting key"
	chaincodeStub.PutStateReturns(fmt.Errorf(errMsg))
	err = assetTransfer.InitLedger(transactionContext)
	require.NotNil(t, err, "should throw an err")
	require.Contains(t, err.Error(), errMsg, "should include the original error message")
}

func TestCreateAsset(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	assetTransfer := chaincode.SmartContract{}
	err := assetTransfer.CreateAsset(transactionContext, "", "", "", "")
	require.NoError(t, err)

	chaincodeStub.GetStateReturns([]byte{}, nil)
	err = assetTransfer.CreateAsset(transactionContext, "asset1", "", "", "")
	require.NotNil(t, err, "should be the error asset1 exists")
	require.Contains(t, err.Error(), "asset1", "should include the original error message")

	errMsg := "failed inserting key"
	chaincodeStub.GetStateReturns(nil, fmt.Errorf(errMsg))
	err = assetTransfer.CreateAsset(transactionContext, "asset1", "", "", "")
	require.NotNil(t, err, "should be the error failed to read")
	require.Contains(t, err.Error(), errMsg, "should include the original error message")
}

func TestReadAsset(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	expectedAsset := &chaincode.Asset{ID: "asset1"}
	bytes, err := json.Marshal(expectedAsset)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	assetTransfer := chaincode.SmartContract{}
	asset, err := assetTransfer.ReadAsset(transactionContext, "")
	require.NoError(t, err)
	require.Equal(t, expectedAsset, asset)

	errMsg := "unable to retrieve asset"
	chaincodeStub.GetStateReturns(nil, fmt.Errorf(errMsg))
	_, err = assetTransfer.ReadAsset(transactionContext, "")
	require.NotNil(t, err, "should be the error failed to read")
	require.Contains(t, err.Error(), errMsg, "should include the original error message")

	chaincodeStub.GetStateReturns(nil, nil)
	asset, err = assetTransfer.ReadAsset(transactionContext, "asset1")
	require.NotNil(t, err, "should be the error asset1 does not exists")
	require.Nil(t, asset)
}

func TestUpdateAsset(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	expectedAsset := &chaincode.Asset{ID: "asset1"}
	bytes, err := json.Marshal(expectedAsset)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	assetTransfer := chaincode.SmartContract{}
	err = assetTransfer.UpdateAsset(transactionContext, "", "", "", "")
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(nil, nil)
	err = assetTransfer.UpdateAsset(transactionContext, "asset1", "", "", "")
	require.NotNil(t, err, "should be the error asset1 does not exists")

	errMsg := "unable to retrieve asset"
	chaincodeStub.GetStateReturns(nil, fmt.Errorf(errMsg))
	err = assetTransfer.UpdateAsset(transactionContext, "asset1", "", "", "")
	require.NotNil(t, err, "should be the error failed to read")
	require.Contains(t, err.Error(), errMsg, "should include the original error message")
}

func TestDeleteAsset(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	asset := &chaincode.Asset{ID: "asset1"}
	bytes, err := json.Marshal(asset)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	chaincodeStub.DelStateReturns(nil)
	assetTransfer := chaincode.SmartContract{}
	err = assetTransfer.DeleteAsset(transactionContext, "")
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(nil, nil)
	err = assetTransfer.DeleteAsset(transactionContext, "asset1")
	require.NotNil(t, err, "should be the error asset1 does not exists")

	errMsg := "unable to retrieve asset"
	chaincodeStub.GetStateReturns(nil, fmt.Errorf(errMsg))
	err = assetTransfer.DeleteAsset(transactionContext, "")
	require.NotNil(t, err, "should be the error failed to read")
	require.Contains(t, err.Error(), errMsg, "should include the original error message")
}

func TestGetAllAssets(t *testing.T) {
	asset := &chaincode.Asset{ID: "asset1"}
	bytes, err := json.Marshal(asset)
	require.NoError(t, err)

	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturns(&queryresult.KV{Value: bytes}, nil)

	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	chaincodeStub.GetStateByRangeReturns(iterator, nil)
	assetTransfer := &chaincode.SmartContract{}
	assets, err := assetTransfer.GetAllAssets(transactionContext)
	require.NoError(t, err)
	require.Equal(t, []*chaincode.Asset{asset}, assets)

	iterator.HasNextReturns(true)
	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	assets, err = assetTransfer.GetAllAssets(transactionContext)
	require.Contains(t, err.Error(), "failed retrieving next item", "should failed retrieving next item")
	require.Nil(t, assets)

	chaincodeStub.GetStateByRangeReturns(nil, fmt.Errorf("failed retrieving all assets"))
	assets, err = assetTransfer.GetAllAssets(transactionContext)
	require.Contains(t, err.Error(), "failed retrieving all assets", "should failed retrieving all assets")
	require.Nil(t, assets)
}
