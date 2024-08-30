package metrics

import (
	"io"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/events"
	loggerPkg "main/pkg/logger"
	"main/pkg/types"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func TestMetricsManagerStartDisabled(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(false)}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)
	manager.Start()

	require.True(t, true)
}

func TestMetricsManagerStartFail(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)
	manager.Start()
}

//nolint:paralleltest // disabled
func TestAppLoadConfigOk(t *testing.T) {
	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: ":9570"}
	logger := loggerPkg.GetNopLogger()
	metricsManager := NewManager(*logger, config)
	metricsManager.LogAppVersion("1.2.3")
	go metricsManager.Start()

	for {
		request, err := http.Get("http://localhost:9570/healthcheck")
		_ = request.Body.Close()
		if err == nil {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	response, err := http.Get("http://localhost:9570/metrics")
	require.NoError(t, err)
	require.NotEmpty(t, response)

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)

	err = response.Body.Close()
	require.NoError(t, err)

	metricsManager.Stop()
}

func TestMetricsManagerLogLastHeight(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)

	currentTime := time.Now()

	manager.LogLastHeight("chain", 123, currentTime)

	assert.Equal(t, 1, testutil.CollectAndCount(manager.lastBlockHeightCollector))
	assert.InDelta(t, 123, testutil.ToFloat64(manager.lastBlockHeightCollector.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)

	assert.Equal(t, 1, testutil.CollectAndCount(manager.lastBlockTimeCollector))
	assert.InDelta(t, currentTime.Unix(), testutil.ToFloat64(manager.lastBlockTimeCollector.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)
}

func TestMetricsManagerLogNodeConnection(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)

	manager.LogNodeConnection("chain", "node1", true)
	manager.LogNodeConnection("chain", "node2", false)

	assert.Equal(t, 2, testutil.CollectAndCount(manager.nodeConnectedCollector))
	assert.InDelta(t, 1, testutil.ToFloat64(manager.nodeConnectedCollector.With(prometheus.Labels{
		"chain": "chain",
		"node":  "node1",
	})), 0.01)
	assert.Zero(t, testutil.ToFloat64(manager.nodeConnectedCollector.With(prometheus.Labels{
		"chain": "chain",
		"node":  "node2",
	})))
}

func TestMetricsManagerLogQuery(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)

	manager.LogQuery("chain", types.QueryInfo{
		Success:   true,
		QueryType: constants.QueryTypeBlock,
		Node:      "node1",
	})
	manager.LogQuery("chain", types.QueryInfo{
		Success:   false,
		QueryType: constants.QueryTypeBlock,
		Node:      "node2",
	})

	assert.Equal(t, 1, testutil.CollectAndCount(manager.successfulQueriesCollector))
	assert.InDelta(t, 1, testutil.ToFloat64(manager.successfulQueriesCollector.With(prometheus.Labels{
		"chain": "chain",
		"node":  "node1",
		"type":  string(constants.QueryTypeBlock),
	})), 0.01)

	assert.Equal(t, 1, testutil.CollectAndCount(manager.failedQueriesCollector))
	assert.InDelta(t, 1, testutil.ToFloat64(manager.failedQueriesCollector.With(prometheus.Labels{
		"chain": "chain",
		"node":  "node2",
		"type":  string(constants.QueryTypeBlock),
	})), 0.01)
}

func TestMetricsManagerLogReport(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)

	manager.LogReport("chain", &types.Report{
		Events: []types.ReportEvent{
			events.ValidatorActive{},
		},
	})

	assert.Equal(t, 1, testutil.CollectAndCount(manager.reportsCounter))
	assert.InDelta(t, 1, testutil.ToFloat64(manager.reportsCounter.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)

	assert.Equal(t, 1, testutil.CollectAndCount(manager.reportEntriesCounter))
	assert.InDelta(t, 1, testutil.ToFloat64(manager.reportEntriesCounter.With(prometheus.Labels{
		"chain": "chain",
		"type":  string(constants.EventValidatorActive),
	})), 0.01)
}

func TestMetricsManagerLogTotalBlocksAmount(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)

	manager.LogTotalBlocksAmount("chain", 10)

	assert.Equal(t, 1, testutil.CollectAndCount(manager.totalBlocksGauge))
	assert.InDelta(t, 10, testutil.ToFloat64(manager.totalBlocksGauge.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)
}

func TestMetricsManagerLogReporterEnabled(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)

	manager.LogReporterEnabled("chain", "reporter1", true)
	manager.LogReporterEnabled("chain", "reporter2", false)

	assert.Equal(t, 2, testutil.CollectAndCount(manager.reporterEnabledGauge))
	assert.InDelta(t, 1, testutil.ToFloat64(manager.reporterEnabledGauge.With(prometheus.Labels{
		"chain": "chain",
		"name":  "reporter1",
	})), 0.01)
	assert.Zero(t, testutil.ToFloat64(manager.reporterEnabledGauge.With(prometheus.Labels{
		"chain": "chain",
		"name":  "reporter2",
	})))
}

func TestMetricsManagerLogAppVersion(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)

	manager.LogAppVersion("1.2.3")

	assert.Equal(t, 1, testutil.CollectAndCount(manager.appVersionGauge.With(prometheus.Labels{
		"version": "1.2.3",
	})))
}

func TestMetricsManagerLogWsEvent(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)

	manager.LogWSEvent("chain", "node1")

	assert.Equal(t, 1, testutil.CollectAndCount(manager.eventsCounter))
	assert.InDelta(t, 1, testutil.ToFloat64(manager.eventsCounter.With(prometheus.Labels{
		"chain": "chain",
		"node":  "node1",
	})), 0.01)
}

func TestMetricsManagerLogNodeReconnect(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)

	manager.LogNodeReconnect("chain", "node1")

	assert.Equal(t, 1, testutil.CollectAndCount(manager.reconnectsCounter))
	assert.InDelta(t, 1, testutil.ToFloat64(manager.reconnectsCounter.With(prometheus.Labels{
		"chain": "chain",
		"node":  "node1",
	})), 0.01)
}

func TestMetricsManagerLogValidatorStats(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)

	manager.LogValidatorStats("chain", &types.Entry{
		IsActive: true,
		Validator: &types.Validator{
			Moniker:                 "moniker",
			Description:             "description",
			Identity:                "identity",
			SecurityContact:         "contact",
			Website:                 "website",
			ConsensusAddressValcons: "valcons",
			ConsensusAddressHex:     "hex",
			OperatorAddress:         "valoper",
			Commission:              0.05,
			Jailed:                  false,
			SigningInfo: &types.SigningInfo{
				Tombstoned:          false,
				MissedBlocksCounter: 123,
			},
			VotingPowerPercent:           123,
			CumulativeVotingPowerPercent: 0.12,
			Rank:                         12,
		},
		SignatureInfo: types.SignatureInto{
			BlocksCount: 5,
			Signed:      1,
			NotSigned:   4,
			NoSignature: 0,
			NotActive:   0,
			Active:      5,
			Proposed:    0,
		},
	})

	assert.Equal(t, 1, testutil.CollectAndCount(manager.missingBlocksGauge))
	assert.InDelta(t, 4, testutil.ToFloat64(manager.missingBlocksGauge.With(prometheus.Labels{
		"chain":   "chain",
		"moniker": "moniker",
		"address": "valoper",
	})), 0.01)

	assert.Equal(t, 1, testutil.CollectAndCount(manager.activeBlocksGauge))
	assert.InDelta(t, 5, testutil.ToFloat64(manager.activeBlocksGauge.With(prometheus.Labels{
		"chain":   "chain",
		"moniker": "moniker",
		"address": "valoper",
	})), 0.01)

	assert.Equal(t, 1, testutil.CollectAndCount(manager.isActiveGauge))
	assert.InDelta(t, 1, testutil.ToFloat64(manager.isActiveGauge.With(prometheus.Labels{
		"chain":   "chain",
		"moniker": "moniker",
		"address": "valoper",
	})), 0.01)

	assert.Equal(t, 1, testutil.CollectAndCount(manager.isJailedGauge))
	assert.Zero(t, testutil.ToFloat64(manager.isJailedGauge.With(prometheus.Labels{
		"chain":   "chain",
		"moniker": "moniker",
		"address": "valoper",
	})))

	assert.Equal(t, 1, testutil.CollectAndCount(manager.isTombstonedGauge))
	assert.Zero(t, testutil.ToFloat64(manager.isTombstonedGauge.With(prometheus.Labels{
		"chain":   "chain",
		"moniker": "moniker",
		"address": "valoper",
	})))

	assert.Equal(t, 1, testutil.CollectAndCount(manager.votingPowerGauge))
	assert.InDelta(t, 123, testutil.ToFloat64(manager.votingPowerGauge.With(prometheus.Labels{
		"chain":   "chain",
		"moniker": "moniker",
		"address": "valoper",
	})), 0.01)

	assert.Equal(t, 1, testutil.CollectAndCount(manager.cumulativeVotingPowerGauge))
	assert.InDelta(t, 0.12, testutil.ToFloat64(manager.cumulativeVotingPowerGauge.With(prometheus.Labels{
		"chain":   "chain",
		"moniker": "moniker",
		"address": "valoper",
	})), 0.01)

	assert.Equal(t, 1, testutil.CollectAndCount(manager.validatorRankGauge))
	assert.InDelta(t, 12, testutil.ToFloat64(manager.validatorRankGauge.With(prometheus.Labels{
		"chain":   "chain",
		"moniker": "moniker",
		"address": "valoper",
	})), 0.01)
}

func TestMetricsManagerLogSlashingParams(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)

	manager.LogSlashingParams("chain", 10000, 0.05, 20000)

	assert.Equal(t, 1, testutil.CollectAndCount(manager.signedBlocksWindowGauge))
	assert.InDelta(t, 10000, testutil.ToFloat64(manager.signedBlocksWindowGauge.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)

	assert.Equal(t, 1, testutil.CollectAndCount(manager.minSignedPerWindowGauge))
	assert.InDelta(t, 0.05, testutil.ToFloat64(manager.minSignedPerWindowGauge.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)

	assert.Equal(t, 1, testutil.CollectAndCount(manager.storeBlocksGauge))
	assert.InDelta(t, 20000, testutil.ToFloat64(manager.storeBlocksGauge.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)
}

func TestMetricsManagerLogChainInfo(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)

	manager.LogChainInfo("chain", "Pretty Name")

	assert.Equal(t, 1, testutil.CollectAndCount(manager.chainInfoGauge.With(prometheus.Labels{
		"chain":       "chain",
		"pretty_name": "Pretty Name",
	})))
}

func TestMetricsManagerLogReporterQuery(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)

	manager.LogReporterQuery("chain", constants.TelegramReporterName, "query")

	assert.Equal(t, 1, testutil.CollectAndCount(manager.reporterQueriesCounter))
	assert.InDelta(t, 1, testutil.ToFloat64(manager.reporterQueriesCounter.With(prometheus.Labels{
		"chain": "chain",
		"query": "query",
		"name":  string(constants.TelegramReporterName),
	})), 0.01)
}

func TestMetricsManagerSetDefaultMetrics(t *testing.T) {
	t.Parallel()

	config := configPkg.MetricsConfig{Enabled: null.BoolFrom(true), ListenAddr: "invalid"}
	logger := loggerPkg.GetNopLogger()
	manager := NewManager(*logger, config)

	chainConfig := &configPkg.ChainConfig{
		RPCEndpoints: []string{"node"},
	}

	manager.SetDefaultMetrics(chainConfig)

	assert.Equal(t, 1, testutil.CollectAndCount(manager.reportsCounter))
	assert.Zero(t, testutil.ToFloat64(manager.reportsCounter.With(prometheus.Labels{
		"chain": "chain",
	})))

	eventNames := constants.GetEventNames()
	assert.Equal(t, len(eventNames), testutil.CollectAndCount(manager.reportEntriesCounter))
	for _, name := range eventNames {
		assert.Zero(t, testutil.ToFloat64(manager.reportEntriesCounter.With(prometheus.Labels{
			"chain": "chain",
			"type":  string(name),
		})))
	}

	assert.Equal(t, 1, testutil.CollectAndCount(manager.eventsCounter))
	assert.Zero(t, testutil.ToFloat64(manager.eventsCounter.With(prometheus.Labels{
		"chain": "chain",
		"node":  "node",
	})))

	assert.Equal(t, 1, testutil.CollectAndCount(manager.reconnectsCounter))
	assert.Zero(t, testutil.ToFloat64(manager.reconnectsCounter.With(prometheus.Labels{
		"chain": "chain",
		"node":  "node",
	})))
}
