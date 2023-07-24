package data

import (
	configPkg "main/pkg/config"
	converterPkg "main/pkg/converter"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"

	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/rs/zerolog"
)

type Manager struct {
	logger      zerolog.Logger
	config      *configPkg.ChainConfig
	httpManager *tendermint.RPCManager
	converter   *converterPkg.Converter
}

func NewManager(
	logger zerolog.Logger,
	config *configPkg.ChainConfig,
	httpManager *tendermint.RPCManager,
) *Manager {
	return &Manager{
		logger:      logger.With().Str("component", "data_manager").Logger(),
		config:      config,
		httpManager: httpManager,
		converter:   converterPkg.NewConverter(),
	}
}

func (m *Manager) GetValidators(height int64) (types.Validators, error) {
	if m.config.IsConsumer.Bool {
		return m.GetValidatorsAndSigningInfoForConsumerChain(height)
	}

	if m.config.QueryEachSigningInfo.Bool {
		return m.GetValidatorsAndEachSigningInfo(height)
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
		validatorsResponse, validatorsError = m.httpManager.GetValidators(height)
		wg.Done()
	}()

	go func() {
		signingInfoResponse, signingInfoErr = m.httpManager.GetSigningInfos(height)
		wg.Done()
	}()

	wg.Wait()

	if validatorsError != nil {
		return nil, validatorsError
	}

	if signingInfoErr != nil {
		return nil, signingInfoErr
	}

	validators := make([]*types.Validator, len(validatorsResponse.Validators))
	for index, validatorRaw := range validatorsResponse.Validators {
		consensusAddr := m.converter.GetConsensusAddress(validatorRaw)

		signingInfo, ok := utils.Find(signingInfoResponse.Info, func(i slashingTypes.ValidatorSigningInfo) bool {
			equal, compareErr := utils.CompareTwoBech32(i.Address, consensusAddr)
			if compareErr != nil {
				m.logger.Error().
					Str("operator_address", validatorRaw.OperatorAddress).
					Str("first", i.Address).
					Str("second", consensusAddr).
					Msg("Error converting bech32 address")
				return false
			}

			return equal
		})

		if !ok {
			m.logger.Debug().
				Str("operator_address", validatorRaw.OperatorAddress).
				Msg("Could not find signing info for validator")
		}

		validator := m.converter.ValidatorFromCosmosValidator(validatorRaw, &signingInfo)

		validators[index] = validator
	}

	return validators, nil
}

func (m *Manager) GetValidatorsAndEachSigningInfo(height int64) (types.Validators, error) {
	validatorsResponse, validatorsError := m.httpManager.GetValidators(height)
	if validatorsError != nil {
		return nil, validatorsError
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	validators := make([]*types.Validator, len(validatorsResponse.Validators))
	for index, validatorRaw := range validatorsResponse.Validators {
		wg.Add(1)
		go func(validatorRaw stakingTypes.Validator, index int) {
			defer wg.Done()

			consensusAddr := m.converter.GetConsensusAddress(validatorRaw)
			signingInfoResponse, signingInfoErr := m.httpManager.GetSigningInfo(consensusAddr, height)
			if signingInfoErr != nil {
				m.logger.Warn().
					Str("operator_address", validatorRaw.OperatorAddress).
					Err(signingInfoErr).
					Msg("Error fetching validator signing info")
			}

			var signingInfo *slashingTypes.ValidatorSigningInfo
			if signingInfoResponse != nil {
				signingInfo = &signingInfoResponse.ValSigningInfo
			}

			validator := m.converter.ValidatorFromCosmosValidator(validatorRaw, signingInfo)

			mutex.Lock()
			validators[index] = validator
			mutex.Unlock()
		}(validatorRaw, index)
	}

	wg.Wait()

	return validators, nil
}

func (m *Manager) GetValidatorsAndSigningInfoForConsumerChain(height int64) (types.Validators, error) {
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
		validatorsResponse, validatorsError = m.httpManager.GetValidators(0)
		wg.Done()
	}()

	go func() {
		signingInfoResponse, signingInfoErr = m.httpManager.GetSigningInfos(height)
		wg.Done()
	}()

	wg.Wait()

	if validatorsError != nil {
		return nil, validatorsError
	}

	if signingInfoErr != nil {
		return nil, signingInfoErr
	}

	validators := make([]*types.Validator, len(validatorsResponse.Validators))

	for index, validatorRaw := range validatorsResponse.Validators {
		if m.config.ConsumerValidatorPrefix != "" {
			if newOperatorAddress, convertErr := utils.ConvertBech32Prefix(
				validatorRaw.OperatorAddress,
				m.config.ConsumerValidatorPrefix,
			); convertErr != nil {
				m.logger.Error().
					Str("operator_address", validatorRaw.OperatorAddress).
					Msg("Error converting operator address to a new prefix")
			} else {
				validatorRaw.OperatorAddress = newOperatorAddress
			}
		}

		wg.Add(1)
		go func(validatorRaw stakingTypes.Validator, index int) {
			defer wg.Done()

			consensusAddrProvider := m.converter.GetConsensusAddress(validatorRaw)
			consensusAddr := consensusAddrProvider

			consensusAddrConsumer, err := m.httpManager.GetValidatorAssignedConsumerKey(consensusAddrProvider, 0)
			if err != nil {
				m.logger.Warn().
					Str("operator_address", validatorRaw.OperatorAddress).
					Err(err).
					Msg("Error fetching validator assigned consumer key")
			} else if consensusAddrConsumer.ConsumerAddress != "" {
				consensusAddr = consensusAddrConsumer.ConsumerAddress
			}

			signingInfo, ok := utils.Find(signingInfoResponse.Info, func(i slashingTypes.ValidatorSigningInfo) bool {
				equal, compareErr := utils.CompareTwoBech32(i.Address, consensusAddr)
				if compareErr != nil {
					m.logger.Error().
						Str("operator_address", validatorRaw.OperatorAddress).
						Str("first", i.Address).
						Str("second", consensusAddr).
						Msg("Error converting bech32 address")
					return false
				}

				return equal
			})

			if !ok {
				m.logger.Debug().
					Str("operator_address", validatorRaw.OperatorAddress).
					Msg("Could not find signing info for validator")
			}

			validator := m.converter.ValidatorFromCosmosValidator(validatorRaw, &signingInfo)

			mutex.Lock()
			validators[index] = validator
			mutex.Unlock()
		}(validatorRaw, index)
	}

	wg.Wait()

	return validators, nil
}
