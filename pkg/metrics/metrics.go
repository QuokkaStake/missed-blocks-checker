package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	configPkg "main/pkg/config"
	"main/pkg/report"
	"main/pkg/types"
	"main/pkg/utils"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Manager struct {
	logger                   zerolog.Logger
	config                   *configPkg.Config
	lastBlockHeightCollector *prometheus.GaugeVec
	lastBlockTimeCollector   *prometheus.GaugeVec

	nodeConnectedCollector     *prometheus.GaugeVec
	successfulQueriesCollector *prometheus.CounterVec
	failedQueriesCollector     *prometheus.CounterVec

	reportsCounter       *prometheus.CounterVec
	reportEntriesCounter *prometheus.CounterVec

	totalBlocksGauge               *prometheus.GaugeVec
	totalHistoricalValidatorsGauge *prometheus.GaugeVec
}

func NewManager(logger zerolog.Logger, config *configPkg.Config) *Manager {
	return &Manager{
		logger: logger.With().Str("component", "metrics").Logger(),
		config: config,
		lastBlockHeightCollector: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "missed_blocks_checker_last_height",
			Help: "Height of the last block processed",
		}, []string{"chain"}),
		lastBlockTimeCollector: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "missed_blocks_checker_last_time",
			Help: "Time of the last block processed",
		}, []string{"chain"}),
		nodeConnectedCollector: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "missed_blocks_node_connected",
			Help: "Whether the node is successfully connected (1 if yes, 0 if no)",
		}, []string{"chain", "node"}),
		successfulQueriesCollector: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "missed_blocks_node_successful_queries",
			Help: "Counter of successful node queries",
		}, []string{"chain", "node", "type"}),
		failedQueriesCollector: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "missed_blocks_node_failed_queries",
			Help: "Counter of failed node queries",
		}, []string{"chain", "node", "type"}),
		reportsCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "missed_blocks_node_reports",
			Help: "Counter of reports to send",
		}, []string{"chain"}),
		reportEntriesCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "missed_blocks_node_report_entries",
			Help: "Counter of report entries send",
		}, []string{"chain", "type"}),
		totalBlocksGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "missed_blocks_node_blocks",
			Help: "Total amount of blocks stored",
		}, []string{"chain"}),
		totalHistoricalValidatorsGauge: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "missed_blocks_node_historical_validators",
			Help: "Total amount of historical validators stored",
		}, []string{"chain"}),
	}
}

func (m *Manager) Start() {
	if !m.config.MetricsConfig.Enabled.Bool {
		m.logger.Info().Msg("Metrics not enabled")
		return
	}

	m.logger.Info().
		Str("addr", m.config.MetricsConfig.ListenAddr).
		Msg("Metrics handler listening")

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(m.config.MetricsConfig.ListenAddr, nil); err != nil {
		m.logger.Fatal().
			Err(err).
			Str("addr", m.config.MetricsConfig.ListenAddr).
			Msg("Cannot start metrics handler")
	}
}

func (m *Manager) LogLastHeight(height int64, blockTime time.Time) {
	m.lastBlockHeightCollector.
		With(prometheus.Labels{"chain": m.config.ChainConfig.Name}).
		Set(float64(height))

	m.lastBlockTimeCollector.
		With(prometheus.Labels{"chain": m.config.ChainConfig.Name}).
		Set(float64(blockTime.Unix()))
}

func (m *Manager) LogNodeConnection(node string, connected bool) {
	m.nodeConnectedCollector.
		With(prometheus.Labels{"chain": m.config.ChainConfig.Name, "node": node}).
		Set(utils.BoolToFloat64(connected))
}

func (m *Manager) LogTendermintQuery(query types.QueryInfo) {
	if query.Success {
		m.successfulQueriesCollector.
			With(prometheus.Labels{
				"chain": m.config.ChainConfig.Name,
				"node":  query.Node,
				"type":  query.QueryType,
			}).Inc()
	} else {
		m.failedQueriesCollector.
			With(prometheus.Labels{
				"chain": m.config.ChainConfig.Name,
				"node":  query.Node,
				"type":  query.QueryType,
			}).Inc()
	}
}

func (m *Manager) LogReport(report *report.Report) {
	m.reportsCounter.
		With(prometheus.Labels{"chain": m.config.ChainConfig.Name}).
		Inc()

	for _, entry := range report.Entries {
		m.reportEntriesCounter.
			With(prometheus.Labels{
				"chain": m.config.ChainConfig.Name,
				"type":  entry.Type(),
			}).
			Inc()
	}
}

func (m *Manager) LogTotalBlocksAmount(amount int64) {
	m.totalBlocksGauge.
		With(prometheus.Labels{"chain": m.config.ChainConfig.Name}).
		Set(float64(amount))
}

func (m *Manager) LogTotalHistoricalValidatorsAmount(amount int64) {
	m.totalHistoricalValidatorsGauge.
		With(prometheus.Labels{"chain": m.config.ChainConfig.Name}).
		Set(float64(amount))
}
