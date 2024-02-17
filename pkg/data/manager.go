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

	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	providerTypes "github.com/cosmos/interchain-security/x/ccv/provider/types"

	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/rs/zerolog"
)

type Manager struct {
	logger    zerolog.Logger
	config    *configPkg.ChainConfig
	rpc       *tendermint.RPC
	converter *converterPkg.Converter
}

func NewManager(
	logger zerolog.Logger,
	config *configPkg.ChainConfig,
	metricsManager *metrics.Manager,
) *Manager {
	rpc := tendermint.NewRPC(config, logger, metricsManager)

	return &Manager{
		logger:    logger.With().Str("component", "data_manager").Logger(),
		config:    config,
		rpc:       rpc,
		converter: converterPkg.NewConverter(),
	}
}

func (manager *Manager) GetValidators(height int64) (types.Validators, []error) {
	if manager.config.IsConsumer.Bool {
		return manager.GetValidatorsAndSigningInfoForConsumerChain(height)
	}

	if manager.config.QueryEachSigningInfo.Bool {
		return manager.GetValidatorsAndEachSigningInfo(height)
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
		validatorsResponse, validatorsError = manager.rpc.GetValidators(height)
		wg.Done()
	}()

	go func() {
		signingInfoResponse, signingInfoErr = manager.rpc.GetSigningInfos(height)
		wg.Done()
	}()

	wg.Wait()

	if validatorsError != nil {
		return nil, []error{validatorsError}
	}

	if signingInfoErr != nil {
		return nil, []error{signingInfoErr}
	}

	validators := make([]*types.Validator, len(validatorsResponse.Validators))
	for index, validatorRaw := range validatorsResponse.Validators {
		consensusAddr := manager.converter.GetConsensusAddress(validatorRaw)

		signingInfo, ok := utils.Find(signingInfoResponse.Info, func(i slashingTypes.ValidatorSigningInfo) bool {
			equal, compareErr := utils.CompareTwoBech32(i.Address, consensusAddr)
			if compareErr != nil {
				manager.logger.Error().
					Str("operator_address", validatorRaw.OperatorAddress).
					Str("first", i.Address).
					Str("second", consensusAddr).
					Msg("Error converting bech32 address")
				return false
			}

			return equal
		})

		if !ok {
			manager.logger.Debug().
				Str("operator_address", validatorRaw.OperatorAddress).
				Msg("Could not find signing info for validator")
		}

		validator := manager.converter.ValidatorFromCosmosValidator(validatorRaw, &signingInfo)

		validators[index] = validator
	}

	return validators, nil
}

func (manager *Manager) GetValidatorsAndEachSigningInfo(height int64) (types.Validators, []error) {
	validatorsResponse, validatorsError := manager.rpc.GetValidators(height)
	if validatorsError != nil {
		return nil, []error{validatorsError}
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex
	errs := make([]error, 0)

	validators := make([]*types.Validator, len(validatorsResponse.Validators))
	for index, validatorRaw := range validatorsResponse.Validators {
		wg.Add(1)
		go func(validatorRaw stakingTypes.Validator, index int) {
			defer wg.Done()

			consensusAddr := manager.converter.GetConsensusAddress(validatorRaw)
			signingInfoResponse, signingInfoErr := manager.rpc.GetSigningInfo(consensusAddr, height)
			if signingInfoErr != nil {
				manager.logger.Warn().
					Str("operator_address", validatorRaw.OperatorAddress).
					Err(signingInfoErr).
					Msg("Error fetching validator signing info")
				mutex.Lock()
				errs = append(errs, signingInfoErr)
				mutex.Unlock()
				return
			}

			var signingInfo *slashingTypes.ValidatorSigningInfo
			if signingInfoResponse != nil {
				signingInfo = &signingInfoResponse.ValSigningInfo
			}

			validator := manager.converter.ValidatorFromCosmosValidator(validatorRaw, signingInfo)

			mutex.Lock()
			validators[index] = validator
			mutex.Unlock()
		}(validatorRaw, index)
	}

	wg.Wait()

	return validators, errs
}

func (manager *Manager) GetValidatorsAndSigningInfoForConsumerChain(height int64) (types.Validators, []error) {
	var (
		wg                  sync.WaitGroup
		validatorsResponse  *stakingTypes.QueryValidatorsResponse
		validatorsError     error
		signingInfoResponse *slashingTypes.QuerySigningInfosResponse
		signingInfoErr      error
		mutex               sync.Mutex
	)

	wg.Add(2)
	go func() {
		validatorsResponse, validatorsError = manager.rpc.GetValidators(0)
		wg.Done()
	}()

	go func() {
		signingInfoResponse, signingInfoErr = manager.rpc.GetSigningInfos(height)
		wg.Done()
	}()

	wg.Wait()

	if validatorsError != nil {
		return nil, []error{validatorsError}
	}

	if signingInfoErr != nil {
		return nil, []error{signingInfoErr}
	}

	validators := make([]*types.Validator, len(validatorsResponse.Validators))
	errs := make([]error, 0)

	for index, validatorRaw := range validatorsResponse.Validators {
		if manager.config.ConsumerValidatorPrefix != "" {
			if newOperatorAddress, convertErr := utils.ConvertBech32Prefix(
				validatorRaw.OperatorAddress,
				manager.config.ConsumerValidatorPrefix,
			); convertErr != nil {
				manager.logger.Error().
					Str("operator_address", validatorRaw.OperatorAddress).
					Msg("Error converting operator address to a new prefix")
			} else {
				validatorRaw.OperatorAddress = newOperatorAddress
			}
		}

		wg.Add(1)
		go func(validatorRaw stakingTypes.Validator, index int) {
			defer wg.Done()

			consensusAddrProvider := manager.converter.GetConsensusAddress(validatorRaw)
			consensusAddr := consensusAddrProvider

			consensusAddrConsumer, err := manager.rpc.GetValidatorAssignedConsumerKey(consensusAddrProvider, 0)
			if err != nil {
				manager.logger.Warn().
					Str("operator_address", validatorRaw.OperatorAddress).
					Err(err).
					Msg("Error fetching validator assigned consumer key")

				mutex.Lock()
				errs = append(errs, err)
				mutex.Unlock()
			} else if consensusAddrConsumer.ConsumerAddress != "" {
				consensusAddr = consensusAddrConsumer.ConsumerAddress
			}

			signingInfo, ok := utils.Find(signingInfoResponse.Info, func(i slashingTypes.ValidatorSigningInfo) bool {
				equal, compareErr := utils.CompareTwoBech32(i.Address, consensusAddr)
				if compareErr != nil {
					manager.logger.Error().
						Str("operator_address", validatorRaw.OperatorAddress).
						Str("first", i.Address).
						Str("second", consensusAddr).
						Msg("Error converting bech32 address")
					return false
				}

				return equal
			})

			if !ok {
				manager.logger.Debug().
					Str("operator_address", validatorRaw.OperatorAddress).
					Msg("Could not find signing info for validator")
			}

			validator := manager.converter.ValidatorFromCosmosValidator(validatorRaw, &signingInfo)
			if err := manager.converter.SetValidatorConsumerConsensusAddr(validator, consensusAddr); err != nil {
				manager.logger.Warn().Err(err).Msg("Could not set validator consumer consensus address")
			}

			mutex.Lock()
			validators[index] = validator
			mutex.Unlock()
		}(validatorRaw, index)
	}

	wg.Wait()

	return validators, errs
}

func (manager *Manager) GetBlock(height int64) (*responses.SingleBlockResponse, error) {
	return manager.rpc.GetBlock(height)
}

func (manager *Manager) GetValidatorAssignedConsumerKey(
	providerValcons string,
	height int64,
) (*providerTypes.QueryValidatorConsumerAddrResponse, error) {
	return manager.rpc.GetValidatorAssignedConsumerKey(providerValcons, height)
}

func (manager *Manager) GetSigningInfos(height int64) (*slashingTypes.QuerySigningInfosResponse, error) {
	return manager.rpc.GetSigningInfos(height)
}

func (manager *Manager) GetSigningInfo(valcons string, height int64) (*slashingTypes.QuerySigningInfoResponse, error) {
	return manager.rpc.GetSigningInfo(valcons, height)
}

func (manager *Manager) GetSlashingParams(height int64) (*slashingTypes.QueryParamsResponse, error) {
	return manager.rpc.GetSlashingParams(height)
}

func (manager *Manager) GetConsumerSoftOutOutThreshold(height int64) (*paramsTypes.QueryParamsResponse, error) {
	return manager.rpc.GetConsumerSoftOutOutThreshold(height)
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
