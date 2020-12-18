package peggy

import (
	"github.com/althea-net/peggy/module/x/peggy/keeper"
	"github.com/althea-net/peggy/module/x/peggy/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker handles block ending logic for peggy
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	param := k.GetParams(ctx)
	batchNum := param.BatchNum
	logger := ctx.Logger().With("module", types.ModuleName)

	// make tx batch from unbatched tx pool
	unBatchedTxCnt := k.GetUnbatchedTxCnt(ctx)
	if (ctx.BlockHeight()%int64(param.BatchInterval) == 0) || (unBatchedTxCnt >= batchNum) {
		batchID, err := k.BuildTxBatch(ctx, int(batchNum))
		if err != nil {
			logger.Error("build tx batch from pool failed", "err", err.Error())
			return
		}
		logger.Info("Build tx batch",
			"batchInterval", param.BatchInterval, "batchID", batchID,
			"unBatchedTxCnt", unBatchedTxCnt, "batchNum", batchNum,
		)
	}

	// update valset

	// get valset at current heigth
	currentValset := k.GetCurrentValset(ctx)
	// get the last saved valset
	latestValset := k.GetLatestValset(ctx)
	if latestValset == nil {
		latestValset = currentValset
	}

	valsetMap := currentValset.MapValset()

	//the power of the last validator set in this validator set
	latestTotalPower := uint64(0)
	//the total power in current valset
	currentTotalPower := uint64(0)
	for _, val := range latestValset.Members {
		latestVal, ok := valsetMap[val.EthereumAddress]
		if ok {
			latestTotalPower += latestVal.Power
		}
		currentTotalPower += val.Power
	}

	var diffPower = currentTotalPower - latestTotalPower
	if diffPower < 0 {
		diffPower = -diffPower
	}

	valsetChangePower := sdk.NewIntFromUint64(diffPower)
	valsetChangeThreshold := sdk.NewIntFromUint64(currentTotalPower).Mul(sdk.NewIntFromUint64(param.ValsetChange)).Quo(sdk.NewInt(100))
	if valsetChangePower.GTE(valsetChangeThreshold) || ctx.BlockHeight()%int64(param.ValsetInterval) == 0 {
		logger.Info("Update valset",
			"diffPower", diffPower, "valsetChangeThreshold", valsetChangeThreshold,
			"valsetInterval", param.ValsetInterval,
		)
		k.StoreValset(ctx, latestValset)
	}
}
