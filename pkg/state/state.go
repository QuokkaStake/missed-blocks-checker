package state

import (
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/types"
	"sync"
	"time"
)

type State struct {
	blocks              types.BlocksMap
	validators          types.ValidatorsMap
	activeSet           types.ActiveSet
	notifiers           *types.Notifiers
	lastBlockHeight     int64
	lastActiveSetHeight int64
	mutex               sync.Mutex
}

func NewState() *State {
	return &State{
		blocks:              make(types.BlocksMap),
		validators:          make(types.ValidatorsMap),
		activeSet:           make(types.ActiveSet),
		lastBlockHeight:     0,
		lastActiveSetHeight: 0,
	}
}

func (s *State) AddBlock(block *types.Block) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.blocks[block.Height] = block

	if block.Height > s.lastBlockHeight {
		s.lastBlockHeight = block.Height
	}
}

func (s *State) AddActiveSet(height int64, activeSet map[string]bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.activeSet[height] = activeSet

	if height > s.lastActiveSetHeight {
		s.lastActiveSetHeight = height
	}
}

func (s *State) GetBlocksCountSinceLatest(expected int64) int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var expectedCount int64 = 0

	for height := s.lastBlockHeight; height > s.lastBlockHeight-expected; height-- {
		if _, ok := s.blocks[height]; ok {
			expectedCount++
		}
	}

	return expectedCount
}

func (s *State) GetActiveSetsCountSinceLatest(expected int64) int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var expectedCount int64 = 0

	for height := s.lastActiveSetHeight; height > s.lastActiveSetHeight-expected; height-- {
		if _, ok := s.activeSet[height]; ok {
			expectedCount++
		}
	}

	return expectedCount
}

func (s *State) HasActiveSetAtHeight(height int64) bool {
	_, ok := s.activeSet[height]
	return ok
}

func (s *State) IsPopulated(appConfig *config.Config) bool {
	expected := appConfig.ChainConfig.BlocksWindow
	return s.GetActiveSetsCountSinceLatest(expected) >= expected &&
		s.GetBlocksCountSinceLatest(expected) >= expected
}

func (s *State) TrimBlocksBefore(trimHeight int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for height := range s.blocks {
		if height <= trimHeight {
			delete(s.blocks, height)
		}
	}
}

func (s *State) TrimActiveSetsBefore(trimHeight int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for height := range s.activeSet {
		if height <= trimHeight {
			delete(s.activeSet, height)
		}
	}
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
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.blocks = blocks
}

func (s *State) SetActiveSet(activeSet types.ActiveSet) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.activeSet = activeSet
}

func (s *State) AddNotifier(operatorAddress, reporter, notifier string) bool {
	notifiers, added := s.notifiers.AddNotifier(operatorAddress, reporter, notifier)
	if added {
		s.SetNotifiers(notifiers)
	}

	return added
}

func (s *State) RemoveNotifier(operatorAddress, reporter, notifier string) bool {
	notifiers, removed := s.notifiers.RemoveNotifier(operatorAddress, reporter, notifier)
	if removed {
		s.SetNotifiers(notifiers)
	}

	return removed
}

func (s *State) GetNotifiersForReporter(operatorAddress, reporter string) []string {
	return s.notifiers.GetNotifiersForReporter(operatorAddress, reporter)
}

func (s *State) GetValidatorsForNotifier(reporter, notifier string) []string {
	return s.notifiers.GetValidatorsForNotifier(reporter, notifier)
}

func (s *State) GetLastBlockHeight() int64 {
	return s.lastBlockHeight
}

func (s *State) GetLastActiveSetHeight() int64 {
	return s.lastActiveSetHeight
}

func (s *State) GetValidators() types.ValidatorsMap {
	return s.validators
}

func (s *State) IsValidatorActiveAtBlock(validator *types.Validator, height int64) bool {
	if _, ok := s.activeSet[height]; !ok {
		return false
	}

	_, ok := s.activeSet[height][validator.ConsensusAddress]
	return ok
}

func (s *State) GetValidator(operatorAddress string) (*types.Validator, bool) {
	validator, found := s.validators[operatorAddress]
	return validator, found
}

func (s *State) GetValidatorMissedBlocks(validator *types.Validator, blocksToCheck int64) types.SignatureInto {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	signatureInfo := types.SignatureInto{}

	for height := s.lastBlockHeight; height > s.lastBlockHeight-blocksToCheck; height-- {
		if _, ok := s.blocks[height]; !ok {
			continue
		}

		if !s.IsValidatorActiveAtBlock(validator, height) {
			signatureInfo.NotActive++
			continue
		}

		if s.blocks[height].Proposer == validator.ConsensusAddress {
			signatureInfo.Proposed++
		}

		value, ok := s.blocks[height].Signatures[validator.ConsensusAddress]

		if !ok {
			signatureInfo.NoSignature++
		} else if value != constants.ValidatorSigned && value != constants.ValidatorNilSignature {
			signatureInfo.NotSigned++
		} else {
			signatureInfo.Signed++
		}
	}

	return signatureInfo
}

func (s *State) GetEarliestBlock() *types.Block {
	earliestHeight := s.lastBlockHeight

	for height := range s.blocks {
		if height < earliestHeight {
			earliestHeight = height
		}
	}

	return s.blocks[earliestHeight]
}

func (s *State) GetBlockTime() time.Duration {
	latestHeight := s.lastBlockHeight
	latestBlock := s.blocks[latestHeight]

	earliestBlock := s.GetEarliestBlock()
	earliestHeight := earliestBlock.Height

	heightDiff := latestHeight - earliestHeight
	timeDiff := latestBlock.Time.Sub(earliestBlock.Time)

	timeDiffNano := timeDiff.Nanoseconds()
	blockTimeNano := timeDiffNano / heightDiff
	return time.Duration(blockTimeNano) * time.Nanosecond
}

func (s *State) GetTimeTillJail(
	validator *types.Validator,
	appConfig *config.Config,
) (time.Duration, bool) {
	validator, found := s.GetValidator(validator.OperatorAddress)
	if !found {
		return 0, false
	}

	missedBlocks := s.GetValidatorMissedBlocks(validator, appConfig.ChainConfig.StoreBlocks)
	needToSign := appConfig.ChainConfig.GetBlocksSignCount()
	blocksToJail := needToSign - missedBlocks.GetNotSigned()
	blockTime := s.GetBlockTime()
	nanoToJail := blockTime.Nanoseconds() * blocksToJail
	return time.Duration(nanoToJail) * time.Nanosecond, true
}
