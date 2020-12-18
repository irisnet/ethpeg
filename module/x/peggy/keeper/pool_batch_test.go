package keeper

import (
	"testing"

	"github.com/althea-net/peggy/module/x/peggy/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	mySender, _         = sdk.AccAddressFromBech32("cosmos1ahx7f8wyertuus9r20284ej0asrs085case3kn")
	myReceiver          = "0xd041c41EA1bf0F006ADBb6d2c9ef9D425dE5eaD7"
	myTokenContractAddr = "0x429881672B9AE42b8EbA0E26cD9C73711b891Ca5"
)

func TestPushToOutgoingPool(t *testing.T) {
	k, ctx, keepers := CreateTestEnv(t)

	pushToOutgoingPool(t, ctx, k, keepers)
	// then
	var got []*types.OutgoingTx
	k.IteratePoolTxByFee(ctx, func(_ uint64, tx *types.OutgoingTx) bool {
		got = append(got, tx)
		return false
	})
	exp := []*types.OutgoingTx{
		{
			BridgeFee: types.NewERC20Token(3, myTokenContractAddr).PeggyCoin(),
			Sender:    mySender.String(),
			DestAddr:  myReceiver,
			Amount:    types.NewERC20Token(101, myTokenContractAddr).PeggyCoin(),
		},
		{
			BridgeFee: types.NewERC20Token(2, myTokenContractAddr).PeggyCoin(),
			Sender:    mySender.String(),
			DestAddr:  myReceiver,
			Amount:    types.NewERC20Token(100, myTokenContractAddr).PeggyCoin(),
		},
		{
			BridgeFee: types.NewERC20Token(2, myTokenContractAddr).PeggyCoin(),
			Sender:    mySender.String(),
			DestAddr:  myReceiver,
			Amount:    types.NewERC20Token(102, myTokenContractAddr).PeggyCoin(),
		},
		{
			BridgeFee: types.NewERC20Token(1, myTokenContractAddr).PeggyCoin(),
			Sender:    mySender.String(),
			DestAddr:  myReceiver,
			Amount:    types.NewERC20Token(103, myTokenContractAddr).PeggyCoin(),
		},
	}

	assert.Equal(t, len(exp), int(k.GetUnbatchedTxCnt(ctx)))
	assert.Equal(t, exp, got)
}

func TestBuildTxBatch(t *testing.T) {
	k, ctx, keepers := CreateTestEnv(t)

	pushToOutgoingPool(t, ctx, k, keepers)
	batchID, err := k.BuildTxBatch(ctx, 5)
	assert.NoError(t, err)

	txBatch := k.GetTxBatch(ctx, batchID)

	exp := []*types.OutgoingTransferTx{
		{
			Id:          2,
			Sender:      mySender.String(),
			DestAddress: myReceiver,
			Erc20Token:  types.NewERC20Token(101, myTokenContractAddr),
		},
		{
			Id:          1,
			Sender:      mySender.String(),
			DestAddress: myReceiver,
			Erc20Token:  types.NewERC20Token(100, myTokenContractAddr),
		},
		{
			Id:          3,
			Sender:      mySender.String(),
			DestAddress: myReceiver,
			Erc20Token:  types.NewERC20Token(102, myTokenContractAddr),
		},
		{
			Id:          4,
			Sender:      mySender.String(),
			DestAddress: myReceiver,
			Erc20Token:  types.NewERC20Token(103, myTokenContractAddr),
		},
	}

	assert.EqualValues(t, exp, txBatch.Transactions)
	assert.Equal(t, 0, int(k.GetUnbatchedTxCnt(ctx)))
}

func pushToOutgoingPool(t *testing.T, ctx sdk.Context, k Keeper, tk TestKeepers) {
	// mint some voucher first
	allVouchers := sdk.Coins{types.NewERC20Token(99999, myTokenContractAddr).PeggyCoin()}
	err := tk.BankKeeper.MintCoins(ctx, types.ModuleName, allVouchers)
	require.NoError(t, err)

	// set senders balance
	tk.AccountKeeper.NewAccountWithAddress(ctx, mySender)
	err = tk.BankKeeper.SetBalances(ctx, mySender, allVouchers)
	require.NoError(t, err)

	// when
	for i, v := range []uint64{2, 3, 2, 1} {
		amount := types.NewERC20Token(uint64(i+100), myTokenContractAddr).PeggyCoin()
		fee := types.NewERC20Token(v, myTokenContractAddr).PeggyCoin()
		r, err := k.PushToOutgoingPool(ctx, mySender, myReceiver, amount, fee)
		require.NoError(t, err)
		t.Logf("___ response: %#v", r)
	}
}
