package fetchers

import (
	"encoding/json"
	"errors"
	"fmt"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/http"
	"main/pkg/metrics"
	"main/pkg/types/responses"
	"main/pkg/utils"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/std"
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"

	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	providerTypes "github.com/cosmos/interchain-security/v3/x/ccv/provider/types"
	"github.com/rs/zerolog"
)

type CosmosLCDFetcher struct {
	config         *configPkg.ChainConfig
	metricsManager *metrics.Manager
	logger         zerolog.Logger

	clients         []*http.Client
	providerClients []*http.Client

	parseCodec *codec.ProtoCodec
}

func NewCosmosLCDFetcher(
	config *configPkg.ChainConfig,
	logger zerolog.Logger,
	metricsManager *metrics.Manager,
) *CosmosLCDFetcher {
	clients := make([]*http.Client, len(config.LCDEndpoints))
	for index, host := range config.LCDEndpoints {
		clients[index] = http.NewClient(logger, metricsManager, host, config.Name)
	}

	providerClients := make([]*http.Client, len(config.ProviderLCDEndpoints))
	for index, host := range config.ProviderLCDEndpoints {
		providerClients[index] = http.NewClient(logger, metricsManager, host, config.Name)
	}

	interfaceRegistry := codecTypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	parseCodec := codec.NewProtoCodec(interfaceRegistry)

	return &CosmosLCDFetcher{
		config:          config,
		metricsManager:  metricsManager,
		logger:          logger.With().Str("component", "cosmos_lcd_fetcher").Logger(),
		clients:         clients,
		providerClients: providerClients,
		parseCodec:      parseCodec,
	}
}

func (f *CosmosLCDFetcher) GetConsumerOrProviderClients() []*http.Client {
	if f.config.IsConsumer.Bool {
		return f.providerClients
	}

	return f.clients
}

func (f *CosmosLCDFetcher) GetValidators(height int64) (*stakingTypes.QueryValidatorsResponse, error) {
	var validatorsResponse stakingTypes.QueryValidatorsResponse

	if err := f.Get(
		"/cosmos/staking/v1beta1/validators?pagination.limit=1000",
		constants.QueryTypeValidators,
		&validatorsResponse,
		f.GetConsumerOrProviderClients(),
		height,
		func(v interface{}) error {
			response, ok := v.(*stakingTypes.QueryValidatorsResponse)
			if !ok {
				return errors.New("error converting validators response")
			}

			if len(response.Validators) == 0 {
				return errors.New("malformed response: got 0 validators")
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &validatorsResponse, nil
}

func (f *CosmosLCDFetcher) GetSigningInfos(height int64) (*slashingTypes.QuerySigningInfosResponse, error) {
	var response slashingTypes.QuerySigningInfosResponse

	if err := f.Get(
		"/cosmos/slashing/v1beta1/signing_infos?pagination.limit=1000",
		constants.QueryTypeSigningInfos,
		&response,
		f.clients,
		height,
		func(v interface{}) error {
			response, ok := v.(*slashingTypes.QuerySigningInfosResponse)
			if !ok {
				return errors.New("error converting signing infos response")
			}

			if len(response.Info) == 0 {
				return errors.New("malformed response: got 0 signing infos")
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *CosmosLCDFetcher) GetSigningInfo(valcons string, height int64) (*slashingTypes.QuerySigningInfoResponse, error) {
	var response slashingTypes.QuerySigningInfoResponse

	if err := f.Get(
		"/cosmos/slashing/v1beta1/signing_infos/"+valcons,
		constants.QueryTypeSigningInfo,
		&response,
		f.clients,
		height,
		func(v interface{}) error {
			_, ok := v.(*slashingTypes.QuerySigningInfoResponse)
			if !ok {
				return errors.New("error converting signing info response")
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *CosmosLCDFetcher) GetValidatorAssignedConsumerKey(
	providerValcons string,
	height int64,
) (*providerTypes.QueryValidatorConsumerAddrResponse, error) {
	var response providerTypes.QueryValidatorConsumerAddrResponse

	if err := f.Get(
		fmt.Sprintf(
			"/interchain_security/ccv/provider/validator_consumer_addr?chain_id=%s&provider_address=%s",
			f.config.ConsumerChainID,
			providerValcons,
		),
		constants.QueryTypeConsumerAddr,
		&response,
		f.providerClients,
		height,
		func(v interface{}) error {
			_, ok := v.(*providerTypes.QueryValidatorConsumerAddrResponse)
			if !ok {
				return errors.New("error converting assigned consumer key response")
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *CosmosLCDFetcher) GetSlashingParams(height int64) (*slashingTypes.QueryParamsResponse, error) {
	var response slashingTypes.QueryParamsResponse

	if err := f.Get(
		"/cosmos/slashing/v1beta1/params",
		constants.QueryTypeSlashingParams,
		&response,
		f.clients,
		height,
		func(v interface{}) error {
			response, ok := v.(*slashingTypes.QueryParamsResponse)
			if !ok {
				return errors.New("error converting slashing params response")
			}

			if response.Params.SignedBlocksWindow == 0 {
				return errors.New("malformed response: got 0 as signed blocks window")
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *CosmosLCDFetcher) GetConsumerSoftOutOutThreshold(height int64) (*paramsTypes.QueryParamsResponse, error) {
	var response paramsTypes.QueryParamsResponse

	if err := f.Get(
		"/cosmos/params/v1beta1/params?subspace=ccvconsumer&key=SoftOptOutThreshold",
		constants.QueryTypeSubspaceParams,
		&response,
		f.clients,
		height,
		func(v interface{}) error {
			response, ok := v.(*paramsTypes.QueryParamsResponse)
			if !ok {
				return errors.New("error converting subspace param response")
			}

			if response.Param.Value == "" {
				return errors.New("malformed response: got empty subspace param")
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *CosmosLCDFetcher) Get(
	url string,
	queryType constants.QueryType,
	target proto.Message,
	clients []*http.Client,
	height int64,
	predicate func(interface{}) error,
) error {
	errorsArray := make([]error, len(clients))

	indexesShuffled := utils.MakeShuffledArray(len(clients))
	clientsShuffled := make([]*http.Client, len(clients))

	for index, indexShuffled := range indexesShuffled {
		clientsShuffled[index] = clients[indexShuffled]
	}

	headers := map[string]string{
		"x-cosmos-block-height": strconv.FormatInt(height, 10),
	}

	for index := range indexesShuffled {
		client := clientsShuffled[index]

		fullURL := client.Host + url
		f.logger.Trace().Str("url", fullURL).Msg("Trying making request to LCD")

		bytes, err := client.GetPlain(
			url,
			queryType,
			headers,
		)

		if err != nil {
			f.logger.Warn().Str("url", fullURL).Err(err).Msg("LCD request failed")
			errorsArray[index] = err
			continue
		}

		// check whether the response is error first
		var errorResponse responses.LCDError
		if err := json.Unmarshal(bytes, &errorResponse); err == nil {
			// if we successfully unmarshalled it into LCDError, so err == nil,
			// that means the response is indeed an error.
			if errorResponse.Code != 0 {
				// code = NotFound desc = SigningInfo not found for validator xxx: key not found
				if queryType == constants.QueryTypeSigningInfo && errorResponse.Code == 5 {
					return nil
				}

				f.logger.Warn().Str("url", fullURL).
					Err(err).
					Int("code", errorResponse.Code).
					Str("message", errorResponse.Message).
					Msg("LCD request returned an error")
				errorsArray[index] = errors.New(errorResponse.Message)
				continue
			}
		}

		if err := f.parseCodec.UnmarshalJSON(bytes, target); err != nil {
			f.logger.Warn().Str("url", fullURL).Err(err).Msg("JSON unmarshalling failed")
			errorsArray[index] = err
			continue
		}

		if predicateErr := predicate(target); predicateErr != nil {
			f.logger.Warn().Str("url", fullURL).Err(predicateErr).Msg("LCD precondition failed")
			errorsArray[index] = fmt.Errorf("precondition failed: %s", predicateErr)
			continue
		}

		return nil
	}

	f.logger.Warn().Str("url", url).Msg("All LCD requests failed")

	var sb strings.Builder

	sb.WriteString("All LCD requests failed:\n")
	for index, client := range clientsShuffled {
		sb.WriteString(fmt.Sprintf("#%d: %s -> %s\n", index+1, client.Host, errorsArray[index]))
	}

	return errors.New(sb.String())
}
