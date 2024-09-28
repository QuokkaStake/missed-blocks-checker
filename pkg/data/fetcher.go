package data

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/data/fetchers"
	"main/pkg/metrics"

	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	providerTypes "github.com/cosmos/interchain-security/v6/x/ccv/provider/types"
	"github.com/rs/zerolog"
)

type Fetcher interface {
	GetValidators(height int64) (*stakingTypes.QueryValidatorsResponse, error)
	GetSigningInfos(height int64) (*slashingTypes.QuerySigningInfosResponse, error)
	GetValidatorsAssignedConsumerKeys(
		height int64,
	) (*providerTypes.QueryAllPairsValConsAddrByConsumerResponse, error)
	GetSlashingParams(height int64) (*slashingTypes.QueryParamsResponse, error)
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
