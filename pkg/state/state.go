package state

import (
	"fmt"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/types"
	"sync"
	"time"
)

type LastBlockHeight struct {
	signingInfos int64
	validators   int64
	report       int64
}

type State struct {
	blocks          *Blocks
	validators      types.ValidatorsMap
	notifiers       *types.Notifiers
	lastBlockHeight *LastBlockHeight
	mutex           sync.RWMutex
}

func NewState() *State {
	return &State{
		blocks:     NewBlocks(),
		validators: make(types.ValidatorsMap),
		notifiers:  &types.Notifiers{},
		lastBlockHeight: &LastBlockHeight{
			signingInfos: 0,
			validators:   0,
			report:       0,
		},
	}
}

func (s *State) AddBlock(block *types.Block) {
	s.blocks.AddBlock(block)
}

func (s *State) GetBlocksCountSinceLatest(expected int64) int64 {
	return s.blocks.GetCountSinceLatest(expected)
}

func (s *State) GetMissingBlocksSinceLatest(expected int64) []int64 {
	return s.blocks.GetMissingSinceLatest(expected)
}

func (s *State) HasBlockAtHeight(height int64) bool {
	return s.blocks.HasBlockAtHeight(height)
}

func (s *State) TrimBlocksBefore(trimHeight int64) {
	s.blocks.TrimBefore(trimHeight)
}

func (s *State) SetValidators(validators types.ValidatorsMap) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.validators = validators
}

func (s *State) SetNotifiers(notifiers *types.Notifiers) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.notifiers = notifiers
}

func (s *State) SetBlocks(blocks map[int64]*types.Block) {
	s.blocks.SetBlocks(blocks)
}

func (s *State) AddNotifier(
	operatorAddress string,
	reporter constants.ReporterName,
	userID string,
	userName string,
) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	notifiers, added := s.notifiers.AddNotifier(operatorAddress, reporter, userID, userName)
	if added {
		s.notifiers = notifiers
	}

	return added
}

func (s *State) RemoveNotifier(
	operatorAddress string,
	reporter constants.ReporterName,
	notifier string,
) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	notifiers, removed := s.notifiers.RemoveNotifier(operatorAddress, reporter, notifier)
	if removed {
		s.notifiers = notifiers
	}

	return removed
}

func (s *State) GetNotifiersForReporter(
	operatorAddress string,
	reporter constants.ReporterName,
) []*types.Notifier {
	return s.notifiers.GetNotifiersForReporter(operatorAddress, reporter)
}

func (s *State) GetValidatorsForNotifier(
	reporter constants.ReporterName,
	notifier string,
) []string {
	return s.notifiers.GetValidatorsForNotifier(reporter, notifier)
}

func (s *State) GetLastBlockHeight() int64 {
	return s.blocks.lastHeight
}

func (s *State) GetValidators() types.ValidatorsMap {
	return s.validators
}

func (s *State) GetValidator(operatorAddress string) (*types.Validator, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	validator, found := s.validators[operatorAddress]
	return validator, found
}

func (s *State) GetValidatorMissedBlocks(
	validator *types.Validator,
	blocksToCheck int64,
) (types.SignatureInto, error) {
	signatureInfo := types.SignatureInto{}

	errors := 0

	for height := s.blocks.lastHeight; height > s.blocks.lastHeight-blocksToCheck; height-- {
		block, exists := s.blocks.GetBlock(height)
		if !exists {
			errors += 1
			continue
		}

		signatureInfo.BlocksCount++

		if _, ok := block.Validators[validator.ConsensusAddressHex]; !ok {
			signatureInfo.NotActive++
			continue
		} else {
			signatureInfo.Active++
		}

		if block.Proposer == validator.ConsensusAddressHex {
			signatureInfo.Proposed++
		}

		value, ok := block.Signatures[validator.ConsensusAddressHex]

		if !ok {
			signatureInfo.NoSignature++
		} else if value != constants.ValidatorSigned && value != constants.ValidatorNilSignature {
			signatureInfo.NotSigned++
		} else {
			signatureInfo.Signed++
		}
	}

	// if a validator was not active during the whole period,
	// we do not know for sure the missed blocks counter for this validator
	// and therefore are taking it from signing-info
	if signatureInfo.NotActive > 0 && validator.SigningInfo != nil {
		signatureInfo.NoSignature = 0
		signatureInfo.NotSigned = validator.SigningInfo.MissedBlocksCounter
		signatureInfo.Signed = blocksToCheck - validator.SigningInfo.MissedBlocksCounter
	}

	if errors > 0 {
		return signatureInfo, fmt.Errorf("could not get info on %d blocks", errors)
	}

	return signatureInfo, nil
}

func (s *State) GetEarliestBlock() *types.Block {
	return s.blocks.GetEarliestBlock()
}

func (s *State) GetBlockTime() time.Duration {
	latestHeight := s.blocks.lastHeight
	latestBlock := s.blocks.GetLatestBlock()

	earliestBlock := s.GetEarliestBlock()
	earliestHeight := earliestBlock.Height

	heightDiff := latestHeight - earliestHeight
	timeDiff := latestBlock.Time.Sub(earliestBlock.Time)

	timeDiffNano := timeDiff.Nanoseconds()
	blockTimeNano := timeDiffNano / heightDiff
	return time.Duration(blockTimeNano) * time.Nanosecond
}

func (s *State) GetTimeTillJail(
	chainConfig *config.ChainConfig,
	missedBlocks int64,
) time.Duration {
	needToSign := chainConfig.GetBlocksSignCount()
	blocksToJail := needToSign - missedBlocks
	blockTime := s.GetBlockTime()
	nanoToJail := blockTime.Nanoseconds() * blocksToJail
	return time.Duration(nanoToJail) * time.Nanosecond
}
