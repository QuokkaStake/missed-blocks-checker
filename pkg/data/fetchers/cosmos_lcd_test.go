package fetchers

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"main/pkg/metrics"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/jarcoal/httpmock"
)

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdFailToUnmarshal(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		LCDEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosLCDFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/staking/v1beta1/validators?pagination.limit=1000",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("invalid.json")),
	)

	response, err := lcdFetcher.GetValidators(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "invalid character")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		LCDEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosLCDFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/staking/v1beta1/validators?pagination.limit=1000",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("lcd-error.json")),
	)

	response, err := lcdFetcher.GetValidators(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "Not Implemented")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetValidatorsFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		LCDEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosLCDFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/staking/v1beta1/validators?pagination.limit=1000",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	response, err := lcdFetcher.GetValidators(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetValidatorsEmptyValidators(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		LCDEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosLCDFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/staking/v1beta1/validators?pagination.limit=1000",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("lcd-validators-empty.json")),
	)

	response, err := lcdFetcher.GetValidators(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "malformed response: got 0 validators")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetValidatorsOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		LCDEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosLCDFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/staking/v1beta1/validators?pagination.limit=1000",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("lcd-validators.json")),
	)

	response, err := lcdFetcher.GetValidators(0)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotEmpty(t, response.Validators)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetValidatorsOkConsumer(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:                 "chain",
		IsConsumer:           null.BoolFrom(true),
		ProviderLCDEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosLCDFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/staking/v1beta1/validators?pagination.limit=1000",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("lcd-validators.json")),
	)

	response, err := lcdFetcher.GetValidators(0)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotEmpty(t, response.Validators)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetSigningInfosFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		LCDEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosLCDFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/slashing/v1beta1/signing_infos?pagination.limit=1000",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	response, err := lcdFetcher.GetSigningInfos(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetSigningInfosEmpty(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		LCDEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosLCDFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/slashing/v1beta1/signing_infos?pagination.limit=1000",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("lcd-signing-infos-empty.json")),
	)

	response, err := lcdFetcher.GetSigningInfos(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "malformed response: got 0 signing infos")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetSigningInfosOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		LCDEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosLCDFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/slashing/v1beta1/signing_infos?pagination.limit=1000",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("lcd-signing-infos.json")),
	)

	response, err := lcdFetcher.GetSigningInfos(0)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotEmpty(t, response.Info)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetSlashingParamsFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		LCDEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosLCDFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/slashing/v1beta1/params",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	response, err := lcdFetcher.GetSlashingParams(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetSlashingParamsOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		LCDEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosLCDFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/slashing/v1beta1/params",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("lcd-slashing-params.json")),
	)

	response, err := lcdFetcher.GetSlashingParams(0)

	require.NoError(t, err)
	require.NotNil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetAssignedKeysFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:                 "chain",
		ConsumerID:           "consumer",
		LCDEndpoints:         []string{"https://consumer-example.com"},
		ProviderLCDEndpoints: []string{"https://provider-example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosLCDFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://provider-example.com/interchain_security/ccv/provider/address_pairs/consumer",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	response, err := lcdFetcher.GetValidatorsAssignedConsumerKeys(0)

	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
	require.Nil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestLcdGetAssignedKeysOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:                 "chain",
		ConsumerID:           "consumer",
		LCDEndpoints:         []string{"https://consumer-example.com"},
		ProviderLCDEndpoints: []string{"https://provider-example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	lcdFetcher := NewCosmosLCDFetcher(config, *logger, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://provider-example.com/interchain_security/ccv/provider/address_pairs/consumer",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("lcd-assigned-keys.json")),
	)

	response, err := lcdFetcher.GetValidatorsAssignedConsumerKeys(0)

	require.NoError(t, err)
	require.NotNil(t, response)
}
