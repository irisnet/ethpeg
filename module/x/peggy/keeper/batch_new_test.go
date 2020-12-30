package keeper

import (
	"encoding/hex"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestTxBatchExecuted(t *testing.T) {
	k, ctx, keepers := CreateTestEnv(t)

	maxBatchNum := 2
	// add tx to pool,current unbatched tx : 4
	pushToOutgoingPool(t, ctx, k, keepers)

	// build a batch for tx
	preBatchID, _, err := k.BuildTxBatch(ctx, maxBatchNum)
	assert.NoError(t, err)

	// build a batch for tx again
	nextBatchID, _, err := k.BuildTxBatch(ctx, maxBatchNum)
	assert.NoError(t, err)
	assert.Equal(t, 0, int(k.GetUnbatchedTxCnt(ctx)))

	// execute a batch and repush unexecute tx to pool
	err = k.TxBatchExecuted(ctx, nextBatchID)
	assert.NoError(t, err)
	assert.Equal(t, 2, int(k.GetUnbatchedTxCnt(ctx)))
	assert.Nil(t, k.GetTxBatch(ctx, preBatchID))
	assert.Nil(t, k.GetTxBatch(ctx, nextBatchID))

	// package unbatchted tx again
	nextBatchID, _, err = k.BuildTxBatch(ctx, maxBatchNum)
	assert.NoError(t, err)

	// execute a batch again
	err = k.TxBatchExecuted(ctx, nextBatchID)
	assert.NoError(t, err)
	assert.Equal(t, 0, int(k.GetUnbatchedTxCnt(ctx)))
	assert.Nil(t, k.GetTxBatch(ctx, nextBatchID))

	addr, err := sdk.AccAddressFromBech32("cosmos1pzs4v88qj6u7ar3rh0g8jwtf3ngz9jjvud9jre")
	assert.NoError(t, err)
	fmt.Println("1111111")
	fmt.Println(hex.EncodeToString(addr.Bytes()))
	fmt.Println(len(addr.Bytes()))
}
