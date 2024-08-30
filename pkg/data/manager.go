package data

import (
	configPkg "main/pkg/config"
	converterPkg "main/pkg/converter"
	"main/pkg/metrics"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/types/responses"
	"main/pkg/utils"
	"sync"

	providerTypes "github.com/cosmos/interchain-security/v4/x/ccv/provider/types"

	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/rs/zerolog"
)

type Manager struct {
	logger    zerolog.Logger
	config    *configPkg.ChainConfig
	fetcher   Fetcher
	rpc       *tendermint.RPC
	converter *converterPkg.Converter
}

func NewManager(
	logger zerolog.Logger,
	config *configPkg.ChainConfig,
	metricsManager *metrics.Manager,
) *Manager {
	rpc := tendermint.NewRPC(config, logger, metricsManager)
	fetcher := GetFetcher(config, logger, metricsManager)

	return &Manager{
		logger:    logger.With().Str("component", "data_manager").Logger(),
		config:    config,
		fetcher:   fetcher,
		rpc:       rpc,
		converter: converterPkg.NewConverter(),
	}
}

func (manager *Manager) GetValidators(height int64) (types.Validators, error) {
	if manager.config.IsConsumer.Bool {
		return manager.GetValidatorsAndSigningInfoForConsumerChain(height)
	}

	var (
		wg                  sync.WaitGroup
		validatorsResponse  *stakingTypes.QueryValidatorsResponse
		validatorsError     error
		signingInfoResponse *slashingTypes.QuerySigningInfosResponse
		signingInfoErr      error
	)

	wg.Add(2)
	go func() {
		validatorsResponse, validatorsError = manager.fetcher.GetValidators(height)
		wg.Done()
	}()

	go func() {
		signingInfoResponse, signingInfoErr = manager.fetcher.GetSigningInfos(height)
		wg.Done()
	}()

	wg.Wait()

	if validatorsError != nil {
		return nil, validatorsError
	}

	if signingInfoErr != nil {
		return nil, signingInfoErr
	}

	signingInfosByAddr := make(map[string]*slashingTypes.ValidatorSigningInfo)
	for _, info := range signingInfoResponse.Info {
		signingInfosByAddr[utils.MustDecodeBech32(info.Address)] = &info //nolint:exportloopref
	}

	validators := make(types.Validators, len(validatorsResponse.Validators))
	for index, validatorRaw := range validatorsResponse.Validators {
		consensusAddr := manager.converter.GetConsensusAddress(validatorRaw)

		signingInfo, ok := signingInfosByAddr[utils.MustDecodeBech32(consensusAddr)]
		if !ok {
			manager.logger.Debug().
				Str("operator_address", validatorRaw.OperatorAddress).
				Msg("Could not find signing info for validator")
		}

		validator := manager.converter.ValidatorFromCosmosValidator(validatorRaw, signingInfo)
		validators[index] = validator
	}

	return validators, nil
}

func (manager *Manager) GetValidatorsAndSigningInfoForConsumerChain(height int64) (types.Validators, error) {
	var (
		wg                   sync.WaitGroup
		validatorsResponse   *stakingTypes.QueryValidatorsResponse
		validatorsError      error
		signingInfoResponse  *slashingTypes.QuerySigningInfosResponse
		signingInfoErr       error
		assignedKeysResponse *providerTypes.QueryAllPairsValConAddrByConsumerChainIDResponse
		assignedKeysError    error
		mutex                sync.Mutex
	)

	wg.Add(3)
	go func() {
		validatorsResponse, validatorsError = manager.fetcher.GetValidators(0)
		wg.Done()
	}()

	go func() {
		signingInfoResponse, signingInfoErr = manager.fetcher.GetSigningInfos(height)
		wg.Done()
	}()

	go func() {
		assignedKeysResponse, assignedKeysError = manager.fetcher.GetValidatorsAssignedConsumerKeys(0)
		wg.Done()
	}()

	wg.Wait()

	if validatorsError != nil {
		return nil, validatorsError
	}

	if signingInfoErr != nil {
		return nil, signingInfoErr
	}

	if assignedKeysError != nil {
		return nil, assignedKeysError
	}

	signingInfosByAddr := make(map[string]*slashingTypes.ValidatorSigningInfo)
	for _, info := range signingInfoResponse.Info {
		signingInfosByAddr[utils.MustDecodeBech32(info.Address)] = &info //nolint:exportloopref
	}

	assignedKeysByAddr := make(map[string]*providerTypes.PairValConAddrProviderAndConsumer)
	for _, assignedKey := range assignedKeysResponse.PairValConAddr {
		assignedKeysByAddr[utils.MustDecodeBech32(assignedKey.ProviderAddress)] = assignedKey 
	}

	validators := make(types.Validators, len(validatorsResponse.Validators))

	for index, validatorRaw := range validatorsResponse.Validators {
		if manager.config.ConsumerValidatorPrefix != "" {
			newOperatorAddress := utils.MustConvertBech32Prefix(
				validatorRaw.OperatorAddress,
				manager.config.ConsumerValidatorPrefix,
			)
			validatorRaw.OperatorAddress = newOperatorAddress
		}

		consensusAddrProvider := manager.converter.GetConsensusAddress(validatorRaw)
		consensusAddr := consensusAddrProvider

		assignedConsensusAddr, ok := assignedKeysByAddr[utils.MustDecodeBech32(consensusAddr)]
		if ok {
			consensusAddr = assignedConsensusAddr.ConsumerAddress
		}

		signingInfo, ok := signingInfosByAddr[utils.MustDecodeBech32(consensusAddr)]
		if !ok {
			manager.logger.Debug().
				Str("operator_address", validatorRaw.OperatorAddress).
				Msg("Could not find signing info for validator")
		}

		validator := manager.converter.ValidatorFromCosmosValidator(validatorRaw, signingInfo)
		manager.converter.MustSetValidatorConsumerConsensusAddr(validator, consensusAddr)

		mutex.Lock()
		validators[index] = validator
		mutex.Unlock()
	}

	return validators, nil
}

func (manager *Manager) GetBlock(height int64) (*responses.SingleBlockResponse, error) {
	return manager.rpc.GetBlock(height)
}

func (manager *Manager) GetSigningInfos(height int64) (*slashingTypes.QuerySigningInfosResponse, error) {
	return manager.fetcher.GetSigningInfos(height)
}

func (manager *Manager) GetSlashingParams(height int64) (*slashingTypes.QueryParamsResponse, error) {
	return manager.fetcher.GetSlashingParams(height)
}

func (manager *Manager) GetActiveSetAtBlock(height int64) (map[string]bool, error) {
	return manager.rpc.GetActiveSetAtBlock(height)
}

func (manager *Manager) GetBlocksAndValidatorsAtHeights(heights []int64) (
	map[int64]*responses.SingleBlockResponse,
	map[int64]map[string]bool,
	[]error,
) {
	blocksMap := make(map[int64]*responses.SingleBlockResponse)
	activeSetsMap := make(map[int64]map[string]bool)
	errors := make([]error, 0)

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, height := range heights {
		wg.Add(1)
		go func(height int64) {
			block, err := manager.rpc.GetBlock(height)
			mutex.Lock()
			defer mutex.Unlock()

			if err != nil {
				errors = append(errors, err)
			} else {
				blocksMap[height] = block
			}

			wg.Done()
		}(height)

		wg.Add(1)
		go func(height int64) {
			activeSet, err := manager.rpc.GetActiveSetAtBlock(height)
			mutex.Lock()
			defer mutex.Unlock()

			if err != nil {
				errors = append(errors, err)
			} else {
				activeSetsMap[height] = activeSet
			}

			wg.Done()
		}(height)
	}

	wg.Wait()
	return blocksMap, activeSetsMap, errors
}
