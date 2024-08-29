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

	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	providerTypes "github.com/cosmos/interchain-security/v4/x/ccv/provider/types"
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
		func(v proto.Message) error {
			response, _ := v.(*stakingTypes.QueryValidatorsResponse)
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
		func(v proto.Message) error {
			responseInternal, _ := v.(*slashingTypes.QuerySigningInfosResponse)
			if len(responseInternal.Info) == 0 {
				return errors.New("malformed response: got 0 signing infos")
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *CosmosLCDFetcher) GetValidatorsAssignedConsumerKeys(
	height int64,
) (*providerTypes.QueryAllPairsValConAddrByConsumerChainIDResponse, error) {
	var response providerTypes.QueryAllPairsValConAddrByConsumerChainIDResponse

	if err := f.Get(
		"/interchain_security/ccv/provider/consumer_chain_id?chain_id="+f.config.ConsumerChainID,
		constants.QueryTypeConsumerAddrs,
		&response,
		f.providerClients,
		height,
		func(v proto.Message) error {
			return nil
		},
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *CosmosLCDFetcher) GetSlashingParams(height int64) (*slashingTypes.QueryParamsResponse, error) {
	var slashingParamsResponse slashingTypes.QueryParamsResponse

	if err := f.Get(
		"/cosmos/slashing/v1beta1/params",
		constants.QueryTypeSlashingParams,
		&slashingParamsResponse,
		f.clients,
		height,
		func(v proto.Message) error {
			return nil
		},
	); err != nil {
		return nil, err
	}

	return &slashingParamsResponse, nil
}

func (f *CosmosLCDFetcher) Get(
	url string,
	queryType constants.QueryType,
	target proto.Message,
	clients []*http.Client,
	height int64,
	predicate func(proto.Message) error,
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
