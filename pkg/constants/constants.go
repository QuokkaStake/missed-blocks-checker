package constants

type EventName string
type ReporterName string
type QueryType string

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

	EventValidatorActive       EventName = "ValidatorActive"
	EventValidatorGroupChanged EventName = "ValidatorGroupChanged"
	EventValidatorInactive     EventName = "ValidatorInactive"
	EventValidatorJailed       EventName = "ValidatorJailed"
	EventValidatorUnjailed     EventName = "ValidatorUnjailed"
	EventValidatorTombstoned   EventName = "ValidatorTombstoned"

	TelegramReporterName ReporterName = "telegram"
	TestReporterName     ReporterName = "test"

	QueryTypeValidators           QueryType = "validators"
	QueryTypeSigningInfos         QueryType = "signing_infos"
	QueryTypeSigningInfo          QueryType = "signing_info"
	QueryTypeSlashingParams       QueryType = "slashing_params"
	QueryTypeHistoricalValidators QueryType = "historical_validators"
	QueryTypeBlock                QueryType = "block"
)

func GetEventNames() []EventName {
	return []EventName{
		EventValidatorActive,
		EventValidatorGroupChanged,
		EventValidatorInactive,
		EventValidatorJailed,
		EventValidatorUnjailed,
		EventValidatorTombstoned,
	}
}
