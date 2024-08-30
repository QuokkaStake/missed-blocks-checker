package tendermint

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
func TestLcdGetBlockFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewRPC(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/block",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	response, err := lcdFetcher.GetBlock(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetBlockError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewRPC(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/block",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-tendermint-error.json")),
	)

	response, err := lcdFetcher.GetBlock(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "error in Tendermint response")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetBlockInvalid(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewRPC(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/block",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-block-invalid.json")),
	)

	response, err := lcdFetcher.GetBlock(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "malformed result of block")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetBlockOkLatest(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewRPC(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/block",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-block.json")),
	)

	response, err := lcdFetcher.GetBlock(0)

	require.NoError(t, err)
	require.NotNil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetBlockOkNotLatest(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewRPC(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/block?height=123",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-block.json")),
	)

	response, err := lcdFetcher.GetBlock(123)

	require.NoError(t, err)
	require.NotNil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetValidatorsFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewRPC(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/validators?height=123&per_page=100&page=1",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	response, err := lcdFetcher.GetActiveSetAtBlock(123)

	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetValidatorsTendermintError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewRPC(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/validators?height=123&per_page=100&page=1",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-tendermint-error.json")),
	)

	response, err := lcdFetcher.GetActiveSetAtBlock(123)

	require.Error(t, err)
	require.ErrorContains(t, err, "error in Tendermint response")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetValidatorsNoValidators(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewRPC(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/validators?height=123&per_page=100&page=1",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-tendermint-validators-empty.json")),
	)

	response, err := lcdFetcher.GetActiveSetAtBlock(123)

	require.Error(t, err)
	require.ErrorContains(t, err, "malformed result of validators active set")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetValidatorsInvalidPagination(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewRPC(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/validators?height=123&per_page=100&page=1",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-tendermint-validators-invalid-pagination.json")),
	)

	response, err := lcdFetcher.GetActiveSetAtBlock(123)

	require.Error(t, err)
	require.ErrorContains(t, err, "invalid syntax")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetValidatorsOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewRPC(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/validators?height=123&per_page=100&page=1",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-tendermint-validators-1.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/validators?height=123&per_page=100&page=2",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-tendermint-validators-2.json")),
	)

	response, err := lcdFetcher.GetActiveSetAtBlock(123)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, response, 180)
}
