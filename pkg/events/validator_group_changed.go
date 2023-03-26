package events

type ValidatorGroupChanged struct {
	Moniker            string
	OperatorAddress    string
	MissedBlocksBefore int64
	MissedBlocksAfter  int64
}
