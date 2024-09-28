package fetchers

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"main/pkg/metrics"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

//nolint:paralleltest // disabled due to httpmock usage
func TestNewCosmosRPCFetcherRpcError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosRPCFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.staking.v1beta1.Query%2FValidators%22&data=0x1200",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	response, err := lcdFetcher.GetValidators(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestNewCosmosRPCFetcherGetValidatorsFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosRPCFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.staking.v1beta1.Query%2FValidators%22&data=0x1200",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-error.json")),
	)

	response, err := lcdFetcher.GetValidators(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "expected code 0, but got 18")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestNewCosmosRPCFetcherGetValidatorsOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosRPCFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.staking.v1beta1.Query%2FValidators%22&data=0x1200&height=123",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-validators.json")),
	)

	response, err := lcdFetcher.GetValidators(123)

	require.NoError(t, err)
	require.NotNil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestNewCosmosRPCFetcherGetValidatorsOkConsumer(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:                 "chain",
		ProviderRPCEndpoints: []string{"https://example.com"},
		IsConsumer:           null.BoolFrom(true),
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosRPCFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.staking.v1beta1.Query%2FValidators%22&data=0x1200&height=123",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-validators.json")),
	)

	response, err := lcdFetcher.GetValidators(123)

	require.NoError(t, err)
	require.NotNil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestNewCosmosRPCFetcherGetSigningInfosFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosRPCFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.slashing.v1beta1.Query%2FSigningInfos%22&data=0x0a00",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	response, err := lcdFetcher.GetSigningInfos(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestNewCosmosRPCFetcherGetSigningInfosOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosRPCFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.slashing.v1beta1.Query%2FSigningInfos%22&data=0x0a00",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-signing-infos.json")),
	)

	response, err := lcdFetcher.GetSigningInfos(0)

	require.NoError(t, err)
	require.NotNil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestNewCosmosRPCFetcherGetSlashingParamsFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosRPCFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.slashing.v1beta1.Query%2FParams%22&data=0x",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	response, err := lcdFetcher.GetSlashingParams(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestNewCosmosRPCFetcherGetSlashingParamsOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosRPCFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.slashing.v1beta1.Query%2FParams%22&data=0x",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-slashing-params.json")),
	)

	response, err := lcdFetcher.GetSlashingParams(0)

	require.NoError(t, err)
	require.NotNil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestNewCosmosRPCFetcherGetAssignedKeysFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:                 "chain",
		ProviderRPCEndpoints: []string{"https://example.com"},
		ConsumerID:           "neutron-1",
		IsConsumer:           null.BoolFrom(true),
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosRPCFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Finterchain_security.ccv.provider.v1.Query%2FQueryAllPairsValConsAddrByConsumer%22&data=0x0a096e657574726f6e2d31",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	response, err := lcdFetcher.GetValidatorsAssignedConsumerKeys(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestNewCosmosRPCFetcherGetAssignedKeysOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:                 "chain",
		ProviderRPCEndpoints: []string{"https://example.com"},
		ConsumerID:           "neutron-1",
		IsConsumer:           null.BoolFrom(true),
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosRPCFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Finterchain_security.ccv.provider.v1.Query%2FQueryAllPairsValConsAddrByConsumer%22&data=0x0a096e657574726f6e2d31",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-assigned-keys.json")),
	)

	response, err := lcdFetcher.GetValidatorsAssignedConsumerKeys(0)

	require.NoError(t, err)
	require.NotNil(t, response)
}
