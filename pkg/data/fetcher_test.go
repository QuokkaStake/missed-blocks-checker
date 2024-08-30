package data

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/data/fetchers"
	loggerPkg "main/pkg/logger"
	"main/pkg/metrics"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestFetcherManager(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})

	fetcher1 := GetFetcher(&configPkg.ChainConfig{FetcherType: constants.FetcherTypeCosmosLCD}, *logger, metricsManager)
	require.IsType(t, &fetchers.CosmosLCDFetcher{}, fetcher1)

	fetcher2 := GetFetcher(&configPkg.ChainConfig{}, *logger, metricsManager)
	require.IsType(t, &fetchers.CosmosRPCFetcher{}, fetcher2)
}
