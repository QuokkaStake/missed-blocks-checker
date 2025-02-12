package constants

type EventName string
type ReporterName string
type QueryType string
type FormatType string
type PopulatorType string

const (
	NewBlocksQuery = "tm.event='NewBlock'"

	ValidatorSigned       = 2
	ValidatorNilSignature = 3

	PrometheusMetricsPrefix = "missed_blocks_checker_"

	EventValidatorActive            EventName = "ValidatorActive"
	EventValidatorGroupChanged      EventName = "ValidatorGroupChanged"
	EventValidatorInactive          EventName = "ValidatorInactive"
	EventValidatorJailed            EventName = "ValidatorJailed"
	EventValidatorUnjailed          EventName = "ValidatorUnjailed"
	EventValidatorTombstoned        EventName = "ValidatorTombstoned"
	EventValidatorCreated           EventName = "ValidatorCreated"
	EventValidatorJoinedSignatory   EventName = "ValidatorJoinedSignatory"
	EventValidatorLeftSignatory     EventName = "ValidatorLeftSignatory"
	EventValidatorChangedKey        EventName = "ValidatorChangedKey"
	EventValidatorChangedMoniker    EventName = "ValidatorChangedMoniker"
	EventValidatorChangedCommission EventName = "ValidatorChangedCommission"

	TelegramReporterName ReporterName = "telegram"
	DiscordReporterName  ReporterName = "discord"
	TestReporterName     ReporterName = "test"

	QueryTypeValidators    QueryType = "validators"
	QueryTypeSigningInfos  QueryType = "signing_infos"
	QueryTypeConsumerAddrs QueryType = "consumer_addrs"

	QueryTypeSlashingParams QueryType = "slashing_params"

	QueryTypeHistoricalValidators QueryType = "historical_validators"
	QueryTypeBlock                QueryType = "block"

	FormatTypeHTML     FormatType = "html"
	FormatTypeMarkdown FormatType = "markdown"
	FormatTypeTest     FormatType = "test"

	DatabaseTypeSqlite   string = "sqlite"
	DatabaseTypePostgres string = "postgres"

	FetcherTypeCosmosRPC string = "cosmos-rpc"
	FetcherTypeCosmosLCD string = "cosmos-lcd"

	PopulatorSlashingParams = "slashing-params-populator"
	PopulatorTrimDatabase   = "trim-database-populator"

	LastEventsCount = 30
)

func GetEventNames() []EventName {
	return []EventName{
		EventValidatorTombstoned,
		EventValidatorJailed,
		EventValidatorInactive,
		EventValidatorUnjailed,
		EventValidatorActive,
		EventValidatorLeftSignatory,
		EventValidatorJoinedSignatory,
		EventValidatorChangedKey,
		EventValidatorChangedMoniker,
		EventValidatorChangedCommission,
		EventValidatorCreated,
		EventValidatorGroupChanged,
	}
}

func GetDatabaseTypes() []string {
	return []string{
		DatabaseTypeSqlite,
		DatabaseTypePostgres,
	}
}

func GetFetcherTypes() []string {
	return []string{
		FetcherTypeCosmosRPC,
		FetcherTypeCosmosLCD,
	}
}
