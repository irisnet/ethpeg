package types

import (
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// permissions
const (
	FeeCollectorName = "peggy_fee_collector"
)

// FeeCollectorAddress return the AccAddress from the hash of the FeeCollectorName
func FeeCollectorAddress() sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(FeeCollectorName)))
}
