package keeper

import (
	"fmt"

	"github.com/althea-net/peggy/module/x/peggy/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
			Fee:         tx.BridgeFee,
		}
		selectedTx = append(selectedTx, txOut)
		err = k.removeFromUnbatchedTXIndex(ctx, tx.BridgeFee, txID)
		return err != nil || len(selectedTx) == maxBatchNum
	})
	return selectedTx, err
}

// TxBatchExecuted is run when the Cosmos chain detects that a batch has been executed on Ethereum
// It frees all the transactions in the batch, then cancels all earlier batches
func (k Keeper) TxBatchExecuted(ctx sdk.Context, nonce uint64) error {
	b := k.GetTxBatch(ctx, nonce)
	if b == nil {
		return sdkerrors.Wrap(types.ErrUnknown, "nonce")
	}

	// cleanup outgoing TX pool
	for _, tx := range b.Transactions {
		k.removePoolEntry(ctx, tx.Id)
	}

	// Iterate through remaining batches
	k.IterateOutgoingTXBatches(ctx, func(key []byte, iter_batch *types.OutgoingTxBatch) bool {
		// If the iterated batches nonce is lower than the one that was just executed, cancel it
		// TODO: iterate only over batches we need to iterate over
		if iter_batch.BatchNonce < b.BatchNonce {
			k.CancelTxBatch(ctx, iter_batch.BatchNonce)
		}
		return false
	})

	// Delete batch since it is finished
	k.deleteBatch(ctx, *b)
	return nil
}

// CancelTxBatch releases all TX in the batch and deletes the batch
func (k Keeper) CancelTxBatch(ctx sdk.Context, nonce uint64) error {
	batch := k.GetTxBatch(ctx, nonce)
	if batch == nil {
		return types.ErrUnknown
	}
	for _, tx := range batch.Transactions {
		//TODO refund coin
		k.prependToUnbatchedTXIndex(ctx, tx.Fee, tx.Id)
	}

	// Add unbatchedTx count
	k.incrUnbatchedTxCnt(ctx, uint64(len(batch.Transactions)))
	// Delete batch since it is finished
	k.deleteBatch(ctx, *batch)

	batchEvent := sdk.NewEvent(
		types.EventTypeOutgoingBatchCanceled,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		//sdk.NewAttribute(types.AttributeKeyContract, k.GetBridgeContractAddress(ctx)),
		//sdk.NewAttribute(types.AttributeKeyBridgeChainID, strconv.Itoa(int(k.GetBridgeChainID(ctx)))),
		sdk.NewAttribute(types.AttributeKeyOutgoingBatchID, fmt.Sprint(nonce)),
		sdk.NewAttribute(types.AttributeKeyNonce, fmt.Sprint(nonce)),
	)
	ctx.EventManager().EmitEvent(batchEvent)
	return nil
}
