package metrics

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/report"
	"main/pkg/types"
	"main/pkg/utils"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Manager struct {
	logger zerolog.Logger
	config configPkg.MetricsConfig

	lastBlockHeightCollector   *prometheus.GaugeVec
	lastBlockTimeCollector     *prometheus.GaugeVec
	nodeConnectedCollector     *prometheus.GaugeVec
	successfulQueriesCollector *prometheus.CounterVec
	failedQueriesCollector     *prometheus.CounterVec
	eventsCounter              *prometheus.CounterVec
	reconnectsCounter          *prometheus.CounterVec

	reportsCounter       *prometheus.CounterVec
	reportEntriesCounter *prometheus.CounterVec

	totalBlocksGauge               *prometheus.GaugeVec
	totalHistoricalValidatorsGauge *prometheus.GaugeVec

	reporterEnabledGauge   *prometheus.GaugeVec
	reporterQueriesCounter *prometheus.CounterVec

	missingBlocksGauge *prometheus.GaugeVec
	activeBlocksGauge  *prometheus.GaugeVec
	isActiveGauge      *prometheus.GaugeVec
	isJailedGauge      *prometheus.GaugeVec
	isTombstonedGauge  *prometheus.GaugeVec

	appVersionGauge         *prometheus.GaugeVec
	chainInfoGauge          *prometheus.GaugeVec
	signedBlocksWindowGauge *prometheus.GaugeVec
	minSignedPerWindowGauge *prometheus.GaugeVec
}

func NewManager(logger zerolog.Logger, config configPkg.MetricsConfig) *Manager {
	return &Manager{
		logger: logger.With().Str("component", "metrics").Logger(),
		config: config,
		lastBlockHeightCollector: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "last_height",
			Help: "Height of the last block processed",
		}, []string{"chain"}),
		lastBlockTimeCollector: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "last_time",
			Help: "Time of the last block processed",
		}, []string{"chain"}),
		nodeConnectedCollector: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "node_connected",
			Help: "Whether the node is successfully connected (1 if yes, 0 if no)",
		}, []string{"chain", "node"}),
		successfulQueriesCollector: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: constants.PrometheusMetricsPrefix + "node_successful_queries_total",
			Help: "Counter of successful node queries",
		}, []string{"chain", "node", "type"}),
		failedQueriesCollector: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: constants.PrometheusMetricsPrefix + "node_failed_queries_total",
			Help: "Counter of failed node queries",
		}, []string{"chain", "node", "type"}),
		reportsCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: constants.PrometheusMetricsPrefix + "node_reports",
			Help: "Counter of reports to send",
		}, []string{"chain"}),
		reportEntriesCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: constants.PrometheusMetricsPrefix + "node_report_entries_total",
			Help: "Counter of report entries send",
		}, []string{"chain", "type"}),
		totalBlocksGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "node_blocks",
			Help: "Total amount of blocks stored",
		}, []string{"chain"}),
		totalHistoricalValidatorsGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "node_historical_validators",
			Help: "Total amount of historical validators stored",
		}, []string{"chain"}),
		reporterEnabledGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "reporter_enabled",
			Help: "Whether the reporter is enabled (1 if yes, 0 if no)",
		}, []string{"chain", "name"}),
		reporterQueriesCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: constants.PrometheusMetricsPrefix + "reporter_queries",
			Help: "Reporters' queries count ",
		}, []string{"chain", "name", "query"}),
		appVersionGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "version",
			Help: "App version",
		}, []string{"version"}),
		eventsCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: constants.PrometheusMetricsPrefix + "events_total",
			Help: "WebSocket events received by node",
		}, []string{"chain", "node"}),
		reconnectsCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: constants.PrometheusMetricsPrefix + "reconnects_total",
			Help: "Node reconnects count",
		}, []string{"chain", "node"}),
		missingBlocksGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "missed_blocks",
			Help: "Validators' missed blocks count",
		}, []string{"chain", "moniker", "address"}),
		activeBlocksGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "active_blocks",
			Help: "Count of each validator's blocks during which they were active",
		}, []string{"chain", "moniker", "address"}),
		isActiveGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "active",
			Help: "Whether the validator is active",
		}, []string{"chain", "moniker", "address"}),
		isJailedGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "jailed",
			Help: "Whether the validator is jailed",
		}, []string{"chain", "moniker", "address"}),
		isTombstonedGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "tombstoned",
			Help: "Whether the validator is tombstoned",
		}, []string{"chain", "moniker", "address"}),
		signedBlocksWindowGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "window",
			Help: "A window in which validator needs to sign blocks",
		}, []string{"chain"}),
		minSignedPerWindowGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "min_signed",
			Help: "A % of blocks validator needs to sign within window",
		}, []string{"chain"}),
		chainInfoGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: constants.PrometheusMetricsPrefix + "chain_info",
			Help: "Chain info, with constant 1 as value and pretty_name and chain as labels",
		}, []string{"chain", "pretty_name"}),
	}
}

func (m *Manager) SetDefaultMetrics(chain *configPkg.ChainConfig) {
	m.reportsCounter.
		With(prometheus.Labels{"chain": chain.Name}).
		Add(0)

	m.reportEntriesCounter.
		With(prometheus.Labels{
			"chain": chain.Name,
			"type":  string(constants.EventValidatorActive),
		}).
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

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(m.config.ListenAddr, nil); err != nil {
		m.logger.Fatal().
			Err(err).
			Str("addr", m.config.ListenAddr).
			Msg("Cannot start metrics handler")
	}
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

func (m *Manager) LogTendermintQuery(chain string, query types.QueryInfo) {
	if query.Success {
		m.successfulQueriesCollector.
			With(prometheus.Labels{
				"chain": chain,
				"node":  query.Node,
				"type":  query.QueryType,
			}).Inc()
	} else {
		m.failedQueriesCollector.
			With(prometheus.Labels{
				"chain": chain,
				"node":  query.Node,
				"type":  query.QueryType,
			}).Inc()
	}
}

func (m *Manager) LogReport(chain string, report *report.Report) {
	m.reportsCounter.
		With(prometheus.Labels{"chain": chain}).
		Inc()

	for _, entry := range report.Entries {
		m.reportEntriesCounter.
			With(prometheus.Labels{
				"chain": chain,
				"type":  string(entry.Type()),
			}).
			Inc()
	}
}

func (m *Manager) LogTotalBlocksAmount(chain string, amount int64) {
	m.totalBlocksGauge.
		With(prometheus.Labels{"chain": chain}).
		Set(float64(amount))
}

func (m *Manager) LogTotalHistoricalValidatorsAmount(chain string, amount int64) {
	m.totalHistoricalValidatorsGauge.
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
	validator *types.Validator,
	signatureInfo types.SignatureInto,
) {
	m.missingBlocksGauge.
		With(prometheus.Labels{
			"chain":   chain,
			"moniker": validator.Moniker,
			"address": validator.OperatorAddress,
		}).
		Set(float64(signatureInfo.GetNotSigned()))

	m.activeBlocksGauge.
		With(prometheus.Labels{
			"chain":   chain,
			"moniker": validator.Moniker,
			"address": validator.OperatorAddress,
		}).
		Set(float64(signatureInfo.Active))

	m.isActiveGauge.
		With(prometheus.Labels{
			"chain":   chain,
			"moniker": validator.Moniker,
			"address": validator.OperatorAddress,
		}).
		Set(utils.BoolToFloat64(validator.Active()))

	m.isJailedGauge.
		With(prometheus.Labels{
			"chain":   chain,
			"moniker": validator.Moniker,
			"address": validator.OperatorAddress,
		}).
		Set(utils.BoolToFloat64(validator.Jailed))

	if validator.SigningInfo != nil {
		m.isTombstonedGauge.
			With(prometheus.Labels{
				"chain":   chain,
				"moniker": validator.Moniker,
				"address": validator.OperatorAddress,
			}).
			Set(utils.BoolToFloat64(validator.SigningInfo.Tombstoned))
	}
}

func (m *Manager) LogSlashingParams(chain string, window int64, minSigned float64) {
	m.signedBlocksWindowGauge.
		With(prometheus.Labels{"chain": chain}).
		Set(float64(window))

	m.minSignedPerWindowGauge.
		With(prometheus.Labels{"chain": chain}).
		Set(minSigned)
}

func (m *Manager) LogChainInfo(chain string, prettyName string) {
	m.chainInfoGauge.
		With(prometheus.Labels{"chain": chain, "pretty_name": prettyName}).
		Set(1)
}

func (m *Manager) LogReporterQuery(chain string, reporter constants.ReporterName, query string) {
	m.reporterQueriesCounter.
		With(prometheus.Labels{
			"chain":    chain,
			"reporter": string(reporter),
			"query":    query,
		}).
		Inc()
}
