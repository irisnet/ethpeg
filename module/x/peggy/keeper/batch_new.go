package keeper

import (
	"github.com/althea-net/peggy/module/x/peggy/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// pickUnbatchedTx find TX in pool and remove from "available" second index
func (k Keeper) pickUnbatchedTx(ctx sdk.Context, maxBatchNum int) ([]*types.OutgoingTransferTx, error) {
	var selectedTx []*types.OutgoingTransferTx
	var err error
	var erc20Amount *types.ERC20Token
	k.IteratePoolTxByFee(ctx, func(txID uint64, tx *types.OutgoingTx) bool {
		erc20Amount, err = types.ERC20FromPeggyCoin(tx.Amount)
		if err != nil {
			return true
		}

		txOut := &types.OutgoingTransferTx{
			Id:          txID,
			Sender:      tx.Sender,
			DestAddress: tx.DestAddr,
			Erc20Token:  erc20Amount,
		}
		selectedTx = append(selectedTx, txOut)
		err = k.removeFromUnbatchedTXIndex(ctx, tx.BridgeFee, txID)
		return err != nil || len(selectedTx) == maxBatchNum
	})
	return selectedTx, err
}
