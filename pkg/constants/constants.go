package constants

const (
	BlockSearchPagination       int64  = 100
	ValidatorsQueryPagination   uint64 = 1000
	SigningInfosQueryPagination uint64 = 1000
	ActiveSetPagination         int    = 100
	ActiveSetsBulkQueryCount    int64  = 50

	NewBlocksQuery = "tm.event='NewBlock'"

	ValidatorBonded int32 = 3

	ValidatorSigned       = 2
	ValidatorNilSignature = 3

	PrometheusMetricsPrefix = "missed_blocks_checker_"
)
