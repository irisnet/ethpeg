package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisStateValidate(t *testing.T) {
	specs := map[string]struct {
		src    *GenesisState
		expErr bool
	}{
		"default params": {src: DefaultGenesisState(), expErr: false},
		"empty params":   {src: &GenesisState{Params: &Params{}}, expErr: false},
		"invalid params": {src: &GenesisState{
			// Params: &Params{
			// 	PeggyId:            "foo",
			// 	ContractSourceHash: "laksdjflasdkfja",
			// 	EthereumAddress:    "invalid-eth-address",
			// 	BridgeChainId:      3279089,
			// },
			Params: &Params{
				// TODO
				PeggyId:              "",
				ProxyContractHash:    "",
				ProxyContractAddress: "",
				LogicContractHash:    "",
				LogicContractAddress: "",
				Version:              "",
				StartThreshold:       uint64(0),
				BridgeChainId:        uint64(0),
				BootstrapValsetNonce: uint64(0),
				BatchInterval:        uint64(0),
				BatchNum:             uint64(0),
				ValsetInterval:       uint64(0),
				ValsetChange:         uint64(0),
			},
		}, expErr: true},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			err := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestStringToByteArray(t *testing.T) {
	specs := map[string]struct {
		testString string
		expErr     bool
	}{
		"16 bytes": {"lakjsdflaksdjfds", false},
		"32 bytes": {"lakjsdflaksdjfdslakjsdflaksdjfds", false},
		"33 bytes": {"€€€€€€€€€€€", true},
	}

	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			_, err := strToFixByteArray(spec.testString)
			if spec.expErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
