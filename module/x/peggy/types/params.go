package types

import (
	"bytes"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyPeggyID              = []byte("PeggyID")
	KeyProxyContractHash    = []byte("ProxyContractHash")
	KeyProxyContractAddress = []byte("ProxyContractAddress")
	KeyLogicContractHash    = []byte("LogicContractHash")
	KeyLogicContractAddress = []byte("LogicContractAddress")
	KeyStartThreshold       = []byte("StartThreshold")
	KeyBridgeChainID        = []byte("BridgeChainID")
	BootstrapValsetNonce    = []byte("BootstrapValsetNonce")
	KeyBatchTime            = []byte("BatchTime")
	KeyBatchNum             = []byte("BatchNum")
	KeyUpdateValsetTime     = []byte("UpdateValsetTime")
	KeyUpdateValsetChange   = []byte("UpdateValsetChange")
)

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPeggyID, &p.PeggyId, validatePeggyID),
		paramtypes.NewParamSetPair(KeyProxyContractHash, &p.ProxyContractHash, validateProxyContractHash),
		paramtypes.NewParamSetPair(KeyProxyContractAddress, &p.ProxyContractAddress, validateProxyContractAddress),
		paramtypes.NewParamSetPair(KeyLogicContractHash, &p.LogicContractHash, validateLogicContractHash),
		paramtypes.NewParamSetPair(KeyLogicContractAddress, &p.LogicContractAddress, validateLogicContractAddress),
		paramtypes.NewParamSetPair(KeyStartThreshold, &p.StartThreshold, validateStartThreshold),
		paramtypes.NewParamSetPair(KeyBridgeChainID, &p.BridgeChainId, validateBridgeChainID),
		paramtypes.NewParamSetPair(BootstrapValsetNonce, &p.BootstrapValsetNonce, validateBootstrapValsetNonce),
		paramtypes.NewParamSetPair(KeyBatchTime, &p.BatchTime, validateBatchTime),
		paramtypes.NewParamSetPair(KeyBatchNum, &p.BatchNum, validateBatchNum),
		paramtypes.NewParamSetPair(KeyUpdateValsetTime, &p.UpdateValsetTime, validateUpdateValsetTime),
		paramtypes.NewParamSetPair(KeyUpdateValsetChange, &p.UpdateValsetChange, validateUpdateValsetChange),
	}
}

func NewParams(
	peggyID string,
	proxyContractHash string,
	proxyContractAddress string,
	logicContractHash string,
	logicContractAddress string,
	startThreshold string,
	bridgeChainId string,
	bootstrapValsetNonce string,
	batchTime string,
	batchNum string,
	updateValsetTime string,
	updateValsetChange string,
) Params {
	return Params{
		PeggyId:              peggyID,
		ProxyContractHash:    proxyContractHash,
		ProxyContractAddress: proxyContractAddress,
		LogicContractHash:    logicContractHash,
		LogicContractAddress: logicContractAddress,
		StartThreshold:       startThreshold,
		BridgeChainId:        bridgeChainId,
		BootstrapValsetNonce: bootstrapValsetNonce,
		BatchTime:            batchTime,
		BatchNum:             batchNum,
		UpdateValsetTime:     updateValsetTime,
		UpdateValsetChange:   updateValsetChange,
	}
}

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func DefaultParams() *Params {
	return &Params{
		PeggyId:              "",
		ProxyContractHash:    "",
		ProxyContractAddress: "",
		LogicContractHash:    "",
		LogicContractAddress: "",
		StartThreshold:       "",
		BridgeChainId:        "",
		BootstrapValsetNonce: "",
		BatchTime:            "",
		BatchNum:             "",
		UpdateValsetTime:     "",
		UpdateValsetChange:   "",
	}
}

func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

func (p Params) ValidateBasic() error {
	// TODO
	return nil
}

func validatePeggyID(i interface{}) error {
	// TODO
	return nil
}

func validateProxyContractHash(i interface{}) error {
	// TODO
	return nil
}

func validateProxyContractAddress(i interface{}) error {
	// TODO
	return nil
}

func validateLogicContractHash(i interface{}) error {
	// TODO
	return nil
}

func validateLogicContractAddress(i interface{}) error {
	// TODO
	return nil
}

func validateStartThreshold(i interface{}) error {
	// TODO
	return nil
}

func validateBridgeChainID(i interface{}) error {
	// TODO
	return nil
}

func validateBootstrapValsetNonce(i interface{}) error {
	// TODO
	return nil
}

func validateBatchTime(i interface{}) error {
	// TODO
	return nil
}

func validateBatchNum(i interface{}) error {
	// TODO
	return nil
}

func validateUpdateValsetTime(i interface{}) error {
	// TODO
	return nil
}

func validateUpdateValsetChange(i interface{}) error {
	// TODO
	return nil
}
