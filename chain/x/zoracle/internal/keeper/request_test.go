package keeper

import (
	"testing"
	"time"

	"github.com/bandprotocol/d3n/chain/x/zoracle/internal/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestGetterSetterRequest(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)

	_, err := keeper.GetRequest(ctx, 1)
	require.NotNil(t, err)

	request := newDefaultRequest()

	keeper.SetRequest(ctx, 1, request)
	actualRequest, err := keeper.GetRequest(ctx, 1)
	require.Nil(t, err)
	require.Equal(t, request, actualRequest)
}

func TestRequest(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)

	ctx = ctx.WithBlockHeight(2)
	ctx = ctx.WithBlockTime(time.Unix(int64(1581589790), 0))
	calldata := []byte("calldata")
	_, err := keeper.AddRequest(ctx, 1, calldata, 2, 2, 100, 20000)
	require.NotNil(t, err)

	script := GetTestOracleScript("../../../../owasm/res/silly.wasm")
	keeper.SetOracleScript(ctx, 1, script)
	_, err = keeper.AddRequest(ctx, 1, calldata, 2, 2, 100, 20000)
	require.NotNil(t, err)

	pubStr := []string{
		"03d03708f161d1583f49e4260a42b2b08d3ba186d7803a23cc3acd12f074d9d76f",
		"03f57f3997a4e81d8f321e9710927e22c2e6d30fb6d8f749a9e4a07afb3b3b7909",
	}

	validatorAddress1 := SetupTestValidator(
		ctx,
		keeper,
		pubStr[0],
		10,
	)
	_, err = keeper.AddRequest(ctx, 1, calldata, 2, 2, 100, 20000)
	require.NotNil(t, err)

	validatorAddress2 := SetupTestValidator(
		ctx,
		keeper,
		pubStr[1],
		100,
	)
	requestID, err := keeper.AddRequest(ctx, 1, calldata, 2, 2, 100, 20000)
	require.Nil(t, err)
	require.Equal(t, types.RequestID(1), requestID)

	actualRequest, err := keeper.GetRequest(ctx, 1)
	require.Nil(t, err)
	expectRequest := types.NewRequest(1, calldata,
		[]sdk.ValAddress{validatorAddress2, validatorAddress1}, 2,
		2, 1581589790, 102, 20000,
	)
	require.Equal(t, expectRequest, actualRequest)
}

func TestRequestCallDataSizeTooBig(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)

	script := GetTestOracleScript("../../../../owasm/res/silly.wasm")
	keeper.SetOracleScript(ctx, 1, script)

	SetupTestValidator(
		ctx,
		keeper,
		"03d03708f161d1583f49e4260a42b2b08d3ba186d7803a23cc3acd12f074d9d76f",
		10,
	)
	SetupTestValidator(
		ctx,
		keeper,
		"03f57f3997a4e81d8f321e9710927e22c2e6d30fb6d8f749a9e4a07afb3b3b7909",
		100,
	)

	// Set MaxCalldataSize to 0
	keeper.SetMaxCalldataSize(ctx, 0)
	// Should fail because size of "calldata" is > 0
	_, err := keeper.AddRequest(ctx, 1, []byte("calldata"), 2, 2, 100, 20000)
	require.NotNil(t, err)

	// Set MaxCalldataSize to 20
	keeper.SetMaxCalldataSize(ctx, 20)
	// Should pass because size of "calldata" is < 20
	_, err = keeper.AddRequest(ctx, 1, []byte("calldata"), 2, 2, 100, 20000)
	require.Nil(t, err)
}

func TestRequestExceedEndBlockExecuteGasLimit(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)

	script := GetTestOracleScript("../../../../owasm/res/silly.wasm")
	keeper.SetOracleScript(ctx, 1, script)

	SetupTestValidator(
		ctx,
		keeper,
		"03d03708f161d1583f49e4260a42b2b08d3ba186d7803a23cc3acd12f074d9d76f",
		10,
	)
	SetupTestValidator(
		ctx,
		keeper,
		"03f57f3997a4e81d8f321e9710927e22c2e6d30fb6d8f749a9e4a07afb3b3b7909",
		100,
	)

	// Set EndBlockExecuteGasLimit to 10000
	keeper.SetEndBlockExecuteGasLimit(ctx, 10000)
	// Should fail because required execute gas is > 10000
	_, err := keeper.AddRequest(ctx, 1, []byte("calldata"), 2, 2, 100, 20000)
	require.NotNil(t, err)

	// Set EndBlockExecuteGasLimit to 30000
	keeper.SetEndBlockExecuteGasLimit(ctx, 30000)
	// Should fail because required execute gas is < 30000
	_, err = keeper.AddRequest(ctx, 1, []byte("calldata"), 2, 2, 100, 20000)
	require.Nil(t, err)
}

// TestAddNewReceiveValidator tests keeper can add valid validator to request
func TestAddNewReceiveValidator(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)
	request := newDefaultRequest()

	keeper.SetRequest(ctx, 1, request)

	err := keeper.AddNewReceiveValidator(ctx, 1, sdk.ValAddress([]byte("validator1")))
	require.Nil(t, err)

	actualRequest, err := keeper.GetRequest(ctx, 1)
	request.ReceivedValidators = []sdk.ValAddress{sdk.ValAddress([]byte("validator1"))}
	require.Nil(t, err)
	require.Equal(t, request, actualRequest)
}

// TestAddNewReceiveValidatorOnInvalidRequest tests keeper must return if add on invalid request
func TestAddNewReceiveValidatorOnInvalidRequest(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)
	request := newDefaultRequest()

	keeper.SetRequest(ctx, 1, request)
	err := keeper.AddNewReceiveValidator(ctx, 2, sdk.ValAddress([]byte("validator1")))
	require.Equal(t, types.CodeRequestNotFound, err.Code())
}

// TestAddInvalidValidator tests keeper return error if try to add new validator that doesn't contain in list.
func TestAddInvalidValidator(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)
	request := newDefaultRequest()

	keeper.SetRequest(ctx, 1, request)

	err := keeper.AddNewReceiveValidator(ctx, 1, sdk.ValAddress([]byte("validator3")))
	require.Equal(t, types.CodeInvalidValidator, err.Code())

	actualRequest, err := keeper.GetRequest(ctx, 1)
	require.Nil(t, err)
	require.Equal(t, request, actualRequest)
}

// TestAddDuplicateValidator tests keeper return error if try to add new validator that already in list.
func TestAddDuplicateValidator(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)
	request := newDefaultRequest()

	keeper.SetRequest(ctx, 1, request)
	// First add must return nil
	err := keeper.AddNewReceiveValidator(ctx, 1, sdk.ValAddress([]byte("validator1")))
	require.Nil(t, err)

	// Second add must return duplicate error
	err = keeper.AddNewReceiveValidator(ctx, 1, sdk.ValAddress([]byte("validator1")))
	require.Equal(t, types.CodeDuplicateValidator, err.Code())

	// Check final output
	actualRequest, err := keeper.GetRequest(ctx, 1)
	request.ReceivedValidators = []sdk.ValAddress{sdk.ValAddress([]byte("validator1"))}
	require.Nil(t, err)
	require.Equal(t, request, actualRequest)
}

// TestSetResolved tests keeper can set resolved status to request
func TestSetResolved(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)
	request := newDefaultRequest()

	keeper.SetRequest(ctx, 1, request)

	err := keeper.SetResolve(ctx, 1, true)
	require.Nil(t, err)

	actualRequest, err := keeper.GetRequest(ctx, 1)
	request.IsResolved = true
	require.Nil(t, err)
	require.Equal(t, request, actualRequest)
}

// TestSetResolvedOnInvalidRequest tests keeper must return if set on invalid request
func TestSetResolvedOnInvalidRequest(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)
	request := newDefaultRequest()

	keeper.SetRequest(ctx, 1, request)
	err := keeper.SetResolve(ctx, 2, true)
	require.Equal(t, types.CodeRequestNotFound, err.Code())
}

// TestConsumeGasForExecute tests keeper must consume gas from context correctly.
func TestConsumeGasForExecute(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)

	script := GetTestOracleScript("../../../../owasm/res/silly.wasm")
	keeper.SetOracleScript(ctx, 1, script)

	SetupTestValidator(
		ctx,
		keeper,
		"03d03708f161d1583f49e4260a42b2b08d3ba186d7803a23cc3acd12f074d9d76f",
		10,
	)
	SetupTestValidator(
		ctx,
		keeper,
		"03f57f3997a4e81d8f321e9710927e22c2e6d30fb6d8f749a9e4a07afb3b3b7909",
		100,
	)

	// Consume 20000 gas in request 1
	beforeGas := ctx.GasMeter().GasConsumed()
	_, err := keeper.AddRequest(ctx, 1, []byte("calldata"), 2, 2, 100, 20000)
	require.Nil(t, err)
	afterGas := ctx.GasMeter().GasConsumed()

	gasUsed1 := afterGas - beforeGas

	// Consume 40000 gas in request 2
	beforeGas = ctx.GasMeter().GasConsumed()
	_, err = keeper.AddRequest(ctx, 1, []byte("calldata"), 2, 2, 100, 40000)
	require.Nil(t, err)
	afterGas = ctx.GasMeter().GasConsumed()
	gasUsed2 := afterGas - beforeGas

	require.True(t, 19800 <= gasUsed2-gasUsed1 && gasUsed2-gasUsed1 <= 20200)

	actualRequest, err := keeper.GetRequest(ctx, 1)
	require.Nil(t, err)
	require.Equal(t, uint64(20000), actualRequest.ExecuteGas)

	actualRequest, err = keeper.GetRequest(ctx, 2)
	require.Nil(t, err)
	require.Equal(t, uint64(40000), actualRequest.ExecuteGas)
}

// Can get/set pending request correctly and set empty case
func TestGetSetPendingRequests(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)

	reqIDs := keeper.GetPendingResolveList(ctx)

	require.Equal(t, []types.RequestID{}, reqIDs)

	keeper.SetPendingResolveList(ctx, []types.RequestID{1, 2, 3})

	reqIDs = keeper.GetPendingResolveList(ctx)
	require.Equal(t, []types.RequestID{1, 2, 3}, reqIDs)

	keeper.SetPendingResolveList(ctx, []types.RequestID{})
	reqIDs = keeper.GetPendingResolveList(ctx)
	require.Equal(t, []types.RequestID{}, reqIDs)
}

// Can add new pending request if request doesn't exist in list,
// and return error if request has already existed in list.
func TestAddPendingRequest(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)

	reqIDs := keeper.GetPendingResolveList(ctx)

	require.Equal(t, []types.RequestID{}, reqIDs)

	keeper.SetPendingResolveList(ctx, []types.RequestID{1, 2})
	err := keeper.AddPendingRequest(ctx, 3)
	require.Nil(t, err)
	reqIDs = keeper.GetPendingResolveList(ctx)
	require.Equal(t, []types.RequestID{1, 2, 3}, reqIDs)

	err = keeper.AddPendingRequest(ctx, 3)
	require.Equal(t, types.CodeDuplicateRequest, err.Code())
	reqIDs = keeper.GetPendingResolveList(ctx)
	require.Equal(t, []types.RequestID{1, 2, 3}, reqIDs)
}

func TestHasToPutInPendingList(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)

	require.False(t, keeper.ShouldBecomePendingResolve(ctx, 1))
	request := newDefaultRequest()
	request.SufficientValidatorCount = 1
	keeper.SetRequest(ctx, 1, request)
	require.False(t, keeper.ShouldBecomePendingResolve(ctx, 1))

	err := keeper.AddNewReceiveValidator(ctx, 1, sdk.ValAddress([]byte("validator1")))
	require.Nil(t, err)
	require.True(t, keeper.ShouldBecomePendingResolve(ctx, 1))

	err = keeper.AddNewReceiveValidator(ctx, 1, sdk.ValAddress([]byte("validator2")))
	require.Nil(t, err)
	require.False(t, keeper.ShouldBecomePendingResolve(ctx, 1))
}

func TestValidateDataSourceCount(t *testing.T) {
	ctx, keeper := CreateTestInput(t, false)
	// Set MaxDataSourceCountPerRequest to 3
	keeper.SetMaxDataSourceCountPerRequest(ctx, 3)

	request := newDefaultRequest()
	keeper.SetRequest(ctx, 1, request)

	keeper.SetRawDataRequest(ctx, 1, 101, types.NewRawDataRequest(0, []byte("calldata1")))
	err := keeper.ValidateDataSourceCount(ctx, 1)
	require.Nil(t, err)

	keeper.SetRawDataRequest(ctx, 1, 102, types.NewRawDataRequest(0, []byte("calldata2")))
	err = keeper.ValidateDataSourceCount(ctx, 1)
	require.Nil(t, err)

	keeper.SetRawDataRequest(ctx, 1, 103, types.NewRawDataRequest(0, []byte("calldata3")))
	err = keeper.ValidateDataSourceCount(ctx, 1)
	require.Nil(t, err)

	// Validation of "104" will return an error because MaxDataSourceCountPerRequest was set to 3.
	keeper.SetRawDataRequest(ctx, 1, 104, types.NewRawDataRequest(0, []byte("calldata4")))
	err = keeper.ValidateDataSourceCount(ctx, 1)
	require.NotNil(t, err)
}
