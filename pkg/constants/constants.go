package constants

const (
	BlockSearchPagination     int64  = 100
	ValidatorsQueryPagination uint64 = 1000

	NewBlocksQuery = "tm.event='NewBlock'"

	ValidatorSigned       = 2
	ValidatorNilSignature = 3
)
