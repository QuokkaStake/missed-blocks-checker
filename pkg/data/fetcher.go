package data

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/data/fetchers"
	"main/pkg/metrics"

	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	providerTypes "github.com/cosmos/interchain-security/v3/x/ccv/provider/types"
	"github.com/rs/zerolog"
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

func GetFetcher(
	config *configPkg.ChainConfig,
	logger zerolog.Logger,
	metricsManager *metrics.Manager,
) Fetcher {
	if config.FetcherType == constants.FetcherTypeCosmosLCD {
		return fetchers.NewCosmosLCDFetcher(config, logger, metricsManager)
	}

	return fetchers.NewCosmosRPCFetcher(config, logger, metricsManager)
}
