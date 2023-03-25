package constants

const (
	StoreBlocks               int64  = 20000
	BlocksWindow              int64  = 10000
	BlockSearchPagination     int64  = 100
	ValidatorsQueryPagination uint64 = 1000

	NewBlocksQuery = "tm.event='NewBlock'"

	ValidatorStatusBonded int32 = 3

	ValidatorSigned       = 2
	ValidatorNilSignature = 3
)
