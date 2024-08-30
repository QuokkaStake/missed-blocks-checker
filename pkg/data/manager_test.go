package data

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	loggerPkg "main/pkg/logger"
	"main/pkg/metrics"
	"main/pkg/types"
	"main/pkg/utils"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

//nolint:paralleltest // disabled due to httpmock usage
func TestGetBlock(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	dataManager := NewManager(*logger, config, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/block?height=123",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-block.json")),
	)

	response, err := dataManager.GetBlock(123)

	require.NoError(t, err)
	require.NotNil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetSlashingParams(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	dataManager := NewManager(*logger, config, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.slashing.v1beta1.Query%2FParams%22&data=0x",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-slashing-params.json")),
	)

	response, err := dataManager.GetSlashingParams(0)

	require.NoError(t, err)
	require.NotNil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetValidators(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	dataManager := NewManager(*logger, config, metricsManager)

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

	response, err := dataManager.GetActiveSetAtBlock(123)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, response, 180)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetSigningInfos(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	dataManager := NewManager(*logger, config, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/abci_query?path=%22%2Fcosmos.slashing.v1beta1.Query%2FSigningInfos%22&data=0x0a00",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-signing-infos.json")),
	)

	response, err := dataManager.GetSigningInfos(0)

	require.NoError(t, err)
	require.NotNil(t, response)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetBlocksAndValidatorsFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	dataManager := NewManager(*logger, config, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/validators?height=123&per_page=100&page=1",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/block?height=123",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	_, _, errs := dataManager.GetBlocksAndValidatorsAtHeights([]int64{123})

	require.Len(t, errs, 2)
	require.ErrorContains(t, errs[0], "custom error")
	require.ErrorContains(t, errs[1], "custom error")
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetBlocksAndValidatorsOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		RPCEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	dataManager := NewManager(*logger, config, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/block?height=123",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("rpc-block.json")),
	)

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

	blocks, validators, errs := dataManager.GetBlocksAndValidatorsAtHeights([]int64{123})

	require.Empty(t, errs)
	require.NotEmpty(t, blocks)
	require.NotEmpty(t, validators)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetValidatorsValidatorsFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		FetcherType:  constants.FetcherTypeCosmosLCD,
		LCDEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	dataManager := NewManager(*logger, config, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/staking/v1beta1/validators?pagination.limit=1000",
		httpmock.NewErrorResponder(errors.New("validators custom error")),
	)

	validators, err := dataManager.GetValidators(123)

	require.Error(t, err)
	require.ErrorContains(t, err, "validators custom error")
	require.Nil(t, validators)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetValidatorsSigningInfosFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		FetcherType:  constants.FetcherTypeCosmosLCD,
		LCDEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	dataManager := NewManager(*logger, config, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/staking/v1beta1/validators?pagination.limit=1000",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("lcd-validators-sentinel.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/slashing/v1beta1/signing_infos?pagination.limit=1000",
		httpmock.NewErrorResponder(errors.New("signing infos custom error")),
	)

	validators, err := dataManager.GetValidators(123)

	require.Error(t, err)
	require.ErrorContains(t, err, "signing infos custom error")
	require.Nil(t, validators)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestGetValidatorsSigningInfosOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{
		Name:         "chain",
		FetcherType:  constants.FetcherTypeCosmosLCD,
		LCDEndpoints: []string{"https://example.com"},
	}
	logger := loggerPkg.GetNopLogger()

	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{Enabled: null.BoolFrom(false)})
	dataManager := NewManager(*logger, config, metricsManager)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/staking/v1beta1/validators?pagination.limit=1000",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("lcd-validators-sentinel.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/slashing/v1beta1/signing_infos?pagination.limit=1000",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("lcd-signing-infos-sentinel.json")),
	)

	validators, err := dataManager.GetValidators(123)

	require.NoError(t, err)
	require.NotEmpty(t, validators)
	require.Len(t, validators, 178)

	withoutSigningInfo := utils.Filter(validators, func(v *types.Validator) bool {
		return v.SigningInfo == nil
	})
	require.Len(t, withoutSigningInfo, 4)
}
