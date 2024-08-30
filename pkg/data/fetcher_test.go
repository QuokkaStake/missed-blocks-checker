package data

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/data/fetchers"
	loggerPkg "main/pkg/logger"
	"main/pkg/metrics"
	"testing"
)

func TestFetcherManager(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})

	fetcher1 := GetFetcher(&configPkg.ChainConfig{FetcherType: constants.FetcherTypeCosmosLCD}, *logger, metricsManager)
	require.IsType(t, fetcher1, &fetchers.CosmosLCDFetcher{})

	fetcher2 := GetFetcher(&configPkg.ChainConfig{}, *logger, metricsManager)
	require.IsType(t, fetcher2, &fetchers.CosmosRPCFetcher{})
}
