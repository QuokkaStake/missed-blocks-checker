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
	config      configPkg.ChainConfig
	httpManager *tendermint.RPCManager
	converter   *converterPkg.Converter
}

func NewManager(
	logger zerolog.Logger,
	config configPkg.ChainConfig,
	httpManager *tendermint.RPCManager,
) *Manager {
	return &Manager{
		logger:      logger.With().Str("component", "data_manager").Logger(),
		config:      config,
		httpManager: httpManager,
		converter:   converterPkg.NewConverter(),
	}
}

func (m *Manager) GetValidators() (types.Validators, error) {
	if m.config.QueryEachSigningInfo.Bool {
		return m.GetValidatorsAndEachSigningInfo()
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
		validatorsResponse, validatorsError = m.httpManager.GetValidators()
		wg.Done()
	}()

	go func() {
		signingInfoResponse, signingInfoErr = m.httpManager.GetSigningInfos()
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
			return i.Address == consensusAddr
		})

		if !ok {
			m.logger.Warn().
				Str("operator_address", validatorRaw.OperatorAddress).
				Msg("Could not find signing info for validator")
		}

		validator := m.converter.ValidatorFromCosmosValidator(validatorRaw, signingInfo)

		validators[index] = validator
	}

	return validators, nil
}

func (m *Manager) GetValidatorsAndEachSigningInfo() (types.Validators, error) {
	validatorsResponse, validatorsError := m.httpManager.GetValidators()
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
			signingInfoResponse, signingInfoErr := m.httpManager.GetSigningInfo(consensusAddr)
			if signingInfoErr != nil {
				m.logger.Warn().
					Str("operator_address", validatorRaw.OperatorAddress).
					Err(signingInfoErr).
					Msg("Error fetching validator signing info")
			}

			var signingInfo slashingTypes.ValidatorSigningInfo
			if signingInfoResponse != nil {
				signingInfo = signingInfoResponse.ValSigningInfo
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
