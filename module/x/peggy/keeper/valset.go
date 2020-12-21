package keeper

import (
	"github.com/althea-net/peggy/module/x/peggy/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// StoreValset save valset
func (k Keeper) StoreValset(ctx sdk.Context, valset *types.Valset) {
	k.storeValset(ctx, valset)
}

// GetLatestValset save valset
func (k Keeper) GetLatestValset(ctx sdk.Context) (valset *types.Valset) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.ValsetRequestKey)
	iter := prefixStore.ReverseIterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), valset)
		break
	}
	return
}
