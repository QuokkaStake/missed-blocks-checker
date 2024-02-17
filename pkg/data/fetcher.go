package data

import (
	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	providerTypes "github.com/cosmos/interchain-security/x/ccv/provider/types"
)

type Fetcher interface {
	GetValidators(height int64) (*stakingTypes.QueryValidatorsResponse, error)
	GetSigningInfos(height int64) (*slashingTypes.QuerySigningInfosResponse, error)
	GetSigningInfo(valcons string, height int64) (*slashingTypes.QuerySigningInfoResponse, error)
	GetValidatorAssignedConsumerKey(
		providerValcons string,
		height int64,
	) (*providerTypes.QueryValidatorConsumerAddrResponse, error)
	GetSlashingParams(height int64) (*slashingTypes.QueryParamsResponse, error)
	GetConsumerSoftOutOutThreshold(height int64) (*paramsTypes.QueryParamsResponse, error)
}
