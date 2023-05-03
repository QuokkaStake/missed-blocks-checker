package state

import (
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/types"
	"sync"
	"time"
)

type LastBlockHeight struct {
	signingInfos int64
	block        int64
	activeSet    int64
	validators   int64
}

type State struct {
	blocks          types.BlocksMap
	validators      types.ValidatorsMap
	activeSet       types.ActiveSet
	notifiers       *types.Notifiers
	lastBlockHeight *LastBlockHeight
	mutex           sync.RWMutex
}

func NewState() *State {
	return &State{
		blocks:     make(types.BlocksMap),
		validators: make(types.ValidatorsMap),
		activeSet:  make(types.ActiveSet),
		lastBlockHeight: &LastBlockHeight{
			signingInfos: 0,
			block:        0,
			activeSet:    0,
			validators:   0,
		},
	}
}

func (s *State) GetLatestBlock() int64 {
	return s.lastBlockHeight.block
}

func (s *State) AddBlock(block *types.Block) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.blocks[block.Height] = block

	if block.Height > s.lastBlockHeight.block {
		s.lastBlockHeight.block = block.Height
	}
}

func (s *State) AddActiveSet(height int64, activeSet map[string]bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.activeSet[height] = activeSet

	if height > s.lastBlockHeight.activeSet {
		s.lastBlockHeight.activeSet = height
	}
}

func (s *State) GetBlocksCountSinceLatest(expected int64) int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var expectedCount int64 = 0

	for height := s.lastBlockHeight.block; height > s.lastBlockHeight.block-expected; height-- {
		if _, ok := s.blocks[height]; ok {
			expectedCount++
		}
	}

	return expectedCount
}

func (s *State) GetActiveSetsCountSinceLatest(expected int64) int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var expectedCount int64 = 0

	for height := s.lastBlockHeight.activeSet; height > s.lastBlockHeight.activeSet-expected; height-- {
		if _, ok := s.activeSet[height]; ok {
			expectedCount++
		}
	}

	return expectedCount
}

func (s *State) HasBlockAtHeight(height int64) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, ok := s.blocks[height]
	return ok
}

func (s *State) HasActiveSetAtHeight(height int64) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

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
	s.mutex.Lock()
	defer s.mutex.Unlock()

	notifiers, added := s.notifiers.AddNotifier(operatorAddress, reporter, notifier)
	if added {
		s.SetNotifiers(notifiers)
	}

	return added
}

func (s *State) RemoveNotifier(operatorAddress, reporter, notifier string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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
	return s.lastBlockHeight.block
}

func (s *State) GetLastActiveSetHeight() int64 {
	return s.lastBlockHeight.activeSet
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
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	validator, found := s.validators[operatorAddress]
	return validator, found
}

func (s *State) GetValidatorMissedBlocks(validator *types.Validator, blocksToCheck int64) types.SignatureInto {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	signatureInfo := types.SignatureInto{}

	for height := s.lastBlockHeight.block; height > s.lastBlockHeight.block-blocksToCheck; height-- {
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
	earliestHeight := s.lastBlockHeight.block

	for height := range s.blocks {
		if height < earliestHeight {
			earliestHeight = height
		}
	}

	return s.blocks[earliestHeight]
}

func (s *State) GetBlockTime() time.Duration {
	latestHeight := s.lastBlockHeight.block
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
