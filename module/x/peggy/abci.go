package peggy

import (
	"github.com/althea-net/peggy/module/x/peggy/keeper"
	"github.com/althea-net/peggy/module/x/peggy/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker handles block ending logic for peggy
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	param := k.GetParams(ctx)
	batchInterval := param.BatchInterval
	batchNum := param.BatchNum

	logger := ctx.Logger().With("module", types.ModuleName)
	if (ctx.BlockHeight()%int64(batchInterval) == 0) || (k.GetUnbatchedTxCnt(ctx) >= batchNum) {
		if _, err := k.BuildTxBatch(ctx, int(batchNum)); err != nil {
			logger.Error("build tx batch from pool failed", "err", err.Error())
		}
	}
}
