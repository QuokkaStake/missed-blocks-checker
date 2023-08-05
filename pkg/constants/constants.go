package constants

type EventName string
type ReporterName string
type QueryType string
type FormatType string

const (
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
	EventValidatorCreated      EventName = "ValidatorCreated"

	TelegramReporterName ReporterName = "telegram"
	DiscordReporterName  ReporterName = "discord"
	TestReporterName     ReporterName = "test"

	QueryTypeValidators   QueryType = "validators"
	QueryTypeSigningInfos QueryType = "signing_infos"
	QueryTypeSigningInfo  QueryType = "signing_info"
	QueryTypeConsumerAddr QueryType = "consumer_addr"

	QueryTypeSlashingParams       QueryType = "slashing_params"
	QueryTypeHistoricalValidators QueryType = "historical_validators"
	QueryTypeBlock                QueryType = "block"

	FormatTypeHTML     FormatType = "html"
	FormatTypeMarkdown FormatType = "markdown"

	DatabaseTypeSqlite   string = "sqlite"
	DatabaseTypePostgres string = "postgres"
)

func GetEventNames() []EventName {
	return []EventName{
		EventValidatorActive,
		EventValidatorGroupChanged,
		EventValidatorInactive,
		EventValidatorJailed,
		EventValidatorUnjailed,
		EventValidatorTombstoned,
		EventValidatorCreated,
	}
}

func GetDatabaseTypes() []string {
	return []string{
		DatabaseTypeSqlite,
		DatabaseTypePostgres,
	}
}
