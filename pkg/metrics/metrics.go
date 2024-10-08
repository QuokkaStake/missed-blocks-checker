package metrics

import (
	"context"
	"errors"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/types"
	"main/pkg/utils"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Manager struct {
	logger zerolog.Logger
	config configPkg.MetricsConfig

	registry *prometheus.Registry
	server   *http.Server

	lastBlockHeightCollector   *prometheus.GaugeVec
	lastBlockTimeCollector     *prometheus.GaugeVec
	nodeConnectedCollector     *prometheus.GaugeVec
	successfulQueriesCollector *prometheus.CounterVec
	failedQueriesCollector     *prometheus.CounterVec
	eventsCounter              *prometheus.CounterVec
	reconnectsCounter          *prometheus.CounterVec

	reportsCounter       *prometheus.CounterVec
	reportEntriesCounter *prometheus.CounterVec

	totalBlocksGauge *prometheus.GaugeVec

	reporterEnabledGauge   *prometheus.GaugeVec
	reporterQueriesCounter *prometheus.CounterVec

	missingBlocksGauge         *prometheus.GaugeVec
	activeBlocksGauge          *prometheus.GaugeVec
	votingPowerGauge           *prometheus.GaugeVec
	cumulativeVotingPowerGauge *prometheus.GaugeVec
	validatorRankGauge         *prometheus.GaugeVec

	isActiveGauge     *prometheus.GaugeVec
	isJailedGauge     *prometheus.GaugeVec
	isTombstonedGauge *prometheus.GaugeVec

	appVersionGauge *prometheus.GaugeVec
	startTimeGauge  *prometheus.GaugeVec

	chainInfoGauge          *prometheus.GaugeVec
	signedBlocksWindowGauge *prometheus.GaugeVec
	minSignedPerWindowGauge *prometheus.GaugeVec

	storeBlocksGauge *prometheus.GaugeVec
}

func NewManager(logger zerolog.Logger, config configPkg.MetricsConfig) *Manager {
	registry := prometheus.NewRegistry()

	lastBlockHeightCollector := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "last_height",
		Help: "Height of the last block processed",
	}, []string{"chain"})
	lastBlockTimeCollector := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "last_time",
		Help: "Time of the last block processed",
	}, []string{"chain"})
	nodeConnectedCollector := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "node_connected",
		Help: "Whether the node is successfully connected (1 if yes, 0 if no)",
	}, []string{"chain", "node"})
	successfulQueriesCollector := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: constants.PrometheusMetricsPrefix + "node_successful_queries_total",
		Help: "Counter of successful node queries",
	}, []string{"chain", "node", "type"})
	failedQueriesCollector := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: constants.PrometheusMetricsPrefix + "node_failed_queries_total",
		Help: "Counter of failed node queries",
	}, []string{"chain", "node", "type"})
	reportsCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: constants.PrometheusMetricsPrefix + "node_reports",
		Help: "Counter of reports to send",
	}, []string{"chain"})
	reportEntriesCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: constants.PrometheusMetricsPrefix + "node_report_entries_total",
		Help: "Counter of report entries send",
	}, []string{"chain", "type"})
	totalBlocksGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "node_blocks",
		Help: "Total amount of blocks stored",
	}, []string{"chain"})
	reporterEnabledGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "reporter_enabled",
		Help: "Whether the reporter is enabled (1 if yes, 0 if no)",
	}, []string{"chain", "name"})
	reporterQueriesCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: constants.PrometheusMetricsPrefix + "reporter_queries",
		Help: "Reporters' queries count ",
	}, []string{"chain", "name", "query"})
	appVersionGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "version",
		Help: "App version",
	}, []string{"version"})
	startTimeGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "start_time",
		Help: "Unix timestamp on when the app was started. Useful for annotations.",
	}, []string{})
	eventsCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: constants.PrometheusMetricsPrefix + "events_total",
		Help: "WebSocket events received by node",
	}, []string{"chain", "node"})
	reconnectsCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: constants.PrometheusMetricsPrefix + "reconnects_total",
		Help: "Node reconnects count",
	}, []string{"chain", "node"})
	missingBlocksGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "missed_blocks",
		Help: "Validators' missed blocks count",
	}, []string{"chain", "moniker", "address"})
	activeBlocksGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "active_blocks",
		Help: "Count of each validator's blocks during which they were active",
	}, []string{"chain", "moniker", "address"})
	votingPowerGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "voting_power",
		Help: "Voting power % of the validator",
	}, []string{"chain", "moniker", "address"})
	cumulativeVotingPowerGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "cumulative_voting_power",
		Help: "Cumulative voting power % of the validator",
	}, []string{"chain", "moniker", "address"})
	validatorRankGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "validator_rank",
		Help: "Validator rank on chain",
	}, []string{"chain", "moniker", "address"})
	isActiveGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "active",
		Help: "Whether the validator is active",
	}, []string{"chain", "moniker", "address"})
	isJailedGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "jailed",
		Help: "Whether the validator is jailed",
	}, []string{"chain", "moniker", "address"})
	isTombstonedGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "tombstoned",
		Help: "Whether the validator is tombstoned",
	}, []string{"chain", "moniker", "address"})
	signedBlocksWindowGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "window",
		Help: "A window in which validator needs to sign blocks",
	}, []string{"chain"})
	storeBlocksGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "store_blocks",
		Help: "How much blocks at max should be stored in a database",
	}, []string{"chain"})
	minSignedPerWindowGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "min_signed",
		Help: "A % of blocks validator needs to sign within window",
	}, []string{"chain"})
	chainInfoGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: constants.PrometheusMetricsPrefix + "chain_info",
		Help: "Chain info, with constant 1 as value and pretty_name and chain as labels",
	}, []string{"chain", "pretty_name"})

	registry.MustRegister(lastBlockHeightCollector)
	registry.MustRegister(lastBlockTimeCollector)
	registry.MustRegister(nodeConnectedCollector)
	registry.MustRegister(successfulQueriesCollector)
	registry.MustRegister(failedQueriesCollector)
	registry.MustRegister(reportsCounter)
	registry.MustRegister(reportEntriesCounter)
	registry.MustRegister(totalBlocksGauge)
	registry.MustRegister(reporterEnabledGauge)
	registry.MustRegister(reporterQueriesCounter)
	registry.MustRegister(appVersionGauge)
	registry.MustRegister(startTimeGauge)
	registry.MustRegister(eventsCounter)
	registry.MustRegister(reconnectsCounter)
	registry.MustRegister(missingBlocksGauge)
	registry.MustRegister(activeBlocksGauge)
	registry.MustRegister(votingPowerGauge)
	registry.MustRegister(cumulativeVotingPowerGauge)
	registry.MustRegister(validatorRankGauge)
	registry.MustRegister(isActiveGauge)
	registry.MustRegister(isJailedGauge)
	registry.MustRegister(isTombstonedGauge)
	registry.MustRegister(signedBlocksWindowGauge)
	registry.MustRegister(storeBlocksGauge)
	registry.MustRegister(minSignedPerWindowGauge)
	registry.MustRegister(chainInfoGauge)

	startTimeGauge.
		With(prometheus.Labels{}).
		Set(float64(time.Now().Unix()))

	server := &http.Server{Addr: config.ListenAddr, Handler: nil}

	return &Manager{
		logger:                     logger.With().Str("component", "metrics").Logger(),
		config:                     config,
		registry:                   registry,
		lastBlockHeightCollector:   lastBlockHeightCollector,
		lastBlockTimeCollector:     lastBlockTimeCollector,
		nodeConnectedCollector:     nodeConnectedCollector,
		successfulQueriesCollector: successfulQueriesCollector,
		failedQueriesCollector:     failedQueriesCollector,
		reportsCounter:             reportsCounter,
		reportEntriesCounter:       reportEntriesCounter,
		totalBlocksGauge:           totalBlocksGauge,
		reporterEnabledGauge:       reporterEnabledGauge,
		reporterQueriesCounter:     reporterQueriesCounter,
		appVersionGauge:            appVersionGauge,
		startTimeGauge:             startTimeGauge,
		eventsCounter:              eventsCounter,
		reconnectsCounter:          reconnectsCounter,
		missingBlocksGauge:         missingBlocksGauge,
		activeBlocksGauge:          activeBlocksGauge,
		cumulativeVotingPowerGauge: cumulativeVotingPowerGauge,
		votingPowerGauge:           votingPowerGauge,
		validatorRankGauge:         validatorRankGauge,
		isActiveGauge:              isActiveGauge,
		isJailedGauge:              isJailedGauge,
		isTombstonedGauge:          isTombstonedGauge,
		signedBlocksWindowGauge:    signedBlocksWindowGauge,
		storeBlocksGauge:           storeBlocksGauge,
		minSignedPerWindowGauge:    minSignedPerWindowGauge,
		chainInfoGauge:             chainInfoGauge,
		server:                     server,
	}
}

func (m *Manager) SetDefaultMetrics(chain *configPkg.ChainConfig) {
	m.reportsCounter.
		With(prometheus.Labels{"chain": chain.Name}).
		Add(0)

	for _, eventName := range constants.GetEventNames() {
		m.reportEntriesCounter.
			With(prometheus.Labels{
				"chain": chain.Name,
				"type":  string(eventName),
			}).
			Add(0)
	}

	for _, node := range chain.RPCEndpoints {
		m.eventsCounter.
			With(prometheus.Labels{"chain": chain.Name, "node": node}).
			Add(0)

		m.reconnectsCounter.
			With(prometheus.Labels{"chain": chain.Name, "node": node}).
			Add(0)
	}
}

func (m *Manager) Start() {
	if !m.config.Enabled.Bool {
		m.logger.Info().Msg("Metrics not enabled")
		return
	}

	m.logger.Info().
		Str("addr", m.config.ListenAddr).
		Msg("Metrics handler listening")

	handler := http.NewServeMux()
	handler.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))
	handler.HandleFunc("/healthcheck", m.Healthcheck)
	m.server.Handler = handler

	if err := m.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		m.logger.Panic().
			Err(err).
			Str("addr", m.config.ListenAddr).
			Msg("Cannot start metrics handler")
	}
}

func (m *Manager) Healthcheck(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("ok"))
}

func (m *Manager) Stop() {
	m.logger.Info().Str("addr", m.config.ListenAddr).Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = m.server.Shutdown(ctx)
}

func (m *Manager) LogLastHeight(chain string, height int64, blockTime time.Time) {
	m.lastBlockHeightCollector.
		With(prometheus.Labels{"chain": chain}).
		Set(float64(height))

	m.lastBlockTimeCollector.
		With(prometheus.Labels{"chain": chain}).
		Set(float64(blockTime.Unix()))
}

func (m *Manager) LogNodeConnection(chain, node string, connected bool) {
	m.nodeConnectedCollector.
		With(prometheus.Labels{"chain": chain, "node": node}).
		Set(utils.BoolToFloat64(connected))
}

func (m *Manager) LogQuery(chain string, query types.QueryInfo) {
	if query.Success {
		m.successfulQueriesCollector.
			With(prometheus.Labels{
				"chain": chain,
				"node":  query.Node,
				"type":  string(query.QueryType),
			}).Inc()
	} else {
		m.failedQueriesCollector.
			With(prometheus.Labels{
				"chain": chain,
				"node":  query.Node,
				"type":  string(query.QueryType),
			}).Inc()
	}
}

func (m *Manager) LogReport(chain string, report *types.Report) {
	m.reportsCounter.
		With(prometheus.Labels{"chain": chain}).
		Inc()

	for _, event := range report.Events {
		m.reportEntriesCounter.
			With(prometheus.Labels{
				"chain": chain,
				"type":  string(event.Type()),
			}).
			Inc()
	}
}

func (m *Manager) LogTotalBlocksAmount(chain string, amount int64) {
	m.totalBlocksGauge.
		With(prometheus.Labels{"chain": chain}).
		Set(float64(amount))
}

func (m *Manager) LogReporterEnabled(chain string, name constants.ReporterName, enabled bool) {
	m.reporterEnabledGauge.
		With(prometheus.Labels{"chain": chain, "name": string(name)}).
		Set(utils.BoolToFloat64(enabled))
}

func (m *Manager) LogAppVersion(version string) {
	m.appVersionGauge.
		With(prometheus.Labels{"version": version}).
		Set(1)
}

func (m *Manager) LogWSEvent(chain string, node string) {
	m.eventsCounter.
		With(prometheus.Labels{"chain": chain, "node": node}).
		Inc()
}

func (m *Manager) LogNodeReconnect(chain string, node string) {
	m.reconnectsCounter.
		With(prometheus.Labels{"chain": chain, "node": node}).
		Inc()
}

func (m *Manager) LogValidatorStats(
	chain string,
	entry *types.Entry,
) {
	m.missingBlocksGauge.
		With(prometheus.Labels{
			"chain":   chain,
			"moniker": entry.Validator.Moniker,
			"address": entry.Validator.OperatorAddress,
		}).
		Set(float64(entry.SignatureInfo.GetNotSigned()))

	m.activeBlocksGauge.
		With(prometheus.Labels{
			"chain":   chain,
			"moniker": entry.Validator.Moniker,
			"address": entry.Validator.OperatorAddress,
		}).
		Set(float64(entry.SignatureInfo.Active))

	m.isActiveGauge.
		With(prometheus.Labels{
			"chain":   chain,
			"moniker": entry.Validator.Moniker,
			"address": entry.Validator.OperatorAddress,
		}).
		Set(utils.BoolToFloat64(entry.IsActive))

	m.isJailedGauge.
		With(prometheus.Labels{
			"chain":   chain,
			"moniker": entry.Validator.Moniker,
			"address": entry.Validator.OperatorAddress,
		}).
		Set(utils.BoolToFloat64(entry.Validator.Jailed))

	m.missingBlocksGauge.
		With(prometheus.Labels{
			"chain":   chain,
			"moniker": entry.Validator.Moniker,
			"address": entry.Validator.OperatorAddress,
		}).
		Set(float64(entry.SignatureInfo.GetNotSigned()))

	if entry.Validator.SigningInfo != nil {
		m.isTombstonedGauge.
			With(prometheus.Labels{
				"chain":   chain,
				"moniker": entry.Validator.Moniker,
				"address": entry.Validator.OperatorAddress,
			}).
			Set(utils.BoolToFloat64(entry.Validator.SigningInfo.Tombstoned))
	}

	if entry.IsActive {
		m.votingPowerGauge.
			With(prometheus.Labels{
				"chain":   chain,
				"moniker": entry.Validator.Moniker,
				"address": entry.Validator.OperatorAddress,
			}).
			Set(entry.Validator.VotingPowerPercent)

		m.cumulativeVotingPowerGauge.
			With(prometheus.Labels{
				"chain":   chain,
				"moniker": entry.Validator.Moniker,
				"address": entry.Validator.OperatorAddress,
			}).
			Set(entry.Validator.CumulativeVotingPowerPercent)

		m.validatorRankGauge.
			With(prometheus.Labels{
				"chain":   chain,
				"moniker": entry.Validator.Moniker,
				"address": entry.Validator.OperatorAddress,
			}).
			Set(float64(entry.Validator.Rank))
	}
}

func (m *Manager) LogSlashingParams(
	chain string,
	window int64,
	minSigned float64,
	storeBlocks int64,
) {
	m.signedBlocksWindowGauge.
		With(prometheus.Labels{"chain": chain}).
		Set(float64(window))

	m.minSignedPerWindowGauge.
		With(prometheus.Labels{"chain": chain}).
		Set(minSigned)

	m.storeBlocksGauge.
		With(prometheus.Labels{"chain": chain}).
		Set(float64(storeBlocks))
}

func (m *Manager) LogChainInfo(chain string, prettyName string) {
	m.chainInfoGauge.
		With(prometheus.Labels{"chain": chain, "pretty_name": prettyName}).
		Set(1)
}

func (m *Manager) LogReporterQuery(chain string, reporter constants.ReporterName, query string) {
	m.reporterQueriesCounter.
		With(prometheus.Labels{
			"chain": chain,
			"name":  string(reporter),
			"query": query,
		}).
		Inc()
}
