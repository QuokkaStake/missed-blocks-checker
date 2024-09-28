package fetchers

import (
	"errors"
	"fmt"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/http"
	"main/pkg/metrics"
	"main/pkg/types/responses"
	"main/pkg/utils"
	"net/url"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"

	queryTypes "github.com/cosmos/cosmos-sdk/types/query"

	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	providerTypes "github.com/cosmos/interchain-security/v6/x/ccv/provider/types"
	"github.com/rs/zerolog"
)

type CosmosRPCFetcher struct {
	config         *configPkg.ChainConfig
	metricsManager *metrics.Manager
	logger         zerolog.Logger

	clients         []*http.Client
	providerClients []*http.Client
}

func NewCosmosRPCFetcher(
	config *configPkg.ChainConfig,
	logger zerolog.Logger,
	metricsManager *metrics.Manager,
) *CosmosRPCFetcher {
	clients := make([]*http.Client, len(config.RPCEndpoints))
	for index, host := range config.RPCEndpoints {
		clients[index] = http.NewClient(logger, metricsManager, host, config.Name)
	}

	providerClients := make([]*http.Client, len(config.ProviderRPCEndpoints))
	for index, host := range config.ProviderRPCEndpoints {
		providerClients[index] = http.NewClient(logger, metricsManager, host, config.Name)
	}

	return &CosmosRPCFetcher{
		config:          config,
		metricsManager:  metricsManager,
		logger:          logger.With().Str("component", "cosmos_rpc_fetcher").Logger(),
		clients:         clients,
		providerClients: providerClients,
	}
}

func (f *CosmosRPCFetcher) GetConsumerOrProviderClients() []*http.Client {
	if f.config.IsConsumer.Bool {
		return f.providerClients
	}

	return f.clients
}

func (f *CosmosRPCFetcher) AbciQuery(
	method string,
	message codec.ProtoMarshaler, //nolint:staticcheck
	height int64,
	queryType constants.QueryType,
	output codec.ProtoMarshaler, //nolint:staticcheck
	clients []*http.Client,
) error {
	dataBytes, _ := message.Marshal()
	methodName := fmt.Sprintf("\"%s\"", method)
	queryURL := fmt.Sprintf(
		"/abci_query?path=%s&data=0x%x",
		url.QueryEscape(methodName),
		dataBytes,
	)

	if height != 0 {
		queryURL += fmt.Sprintf("&height=%d", height)
	}

	var response responses.AbciQueryResponse
	if err := f.Get(queryURL, constants.QueryType("abci_"+string(queryType)), &response, clients, func(v *responses.AbciQueryResponse) error {
		if v.Result.Response.Code != 0 {
			return fmt.Errorf(
				"error in Tendermint response: expected code 0, but got %d, error: %s",
				v.Result.Response.Code,
				v.Result.Response.Log,
			)
		}

		return nil
	}); err != nil {
		return err
	}

	return output.Unmarshal(response.Result.Response.Value)
}

func (f *CosmosRPCFetcher) GetValidators(height int64) (*stakingTypes.QueryValidatorsResponse, error) {
	query := stakingTypes.QueryValidatorsRequest{
		Pagination: &queryTypes.PageRequest{
			Limit: f.config.Pagination.ValidatorsList,
		},
	}

	var validatorsResponse stakingTypes.QueryValidatorsResponse
	if err := f.AbciQuery(
		"/cosmos.staking.v1beta1.Query/Validators",
		&query,
		height,
		constants.QueryTypeValidators,
		&validatorsResponse,
		f.GetConsumerOrProviderClients(),
	); err != nil {
		return nil, err
	}

	return &validatorsResponse, nil
}

func (f *CosmosRPCFetcher) GetSigningInfos(height int64) (*slashingTypes.QuerySigningInfosResponse, error) {
	query := slashingTypes.QuerySigningInfosRequest{
		Pagination: &queryTypes.PageRequest{
			Limit: f.config.Pagination.SigningInfos,
		},
	}

	var response slashingTypes.QuerySigningInfosResponse
	if err := f.AbciQuery(
		"/cosmos.slashing.v1beta1.Query/SigningInfos",
		&query,
		height,
		constants.QueryTypeSigningInfos,
		&response,
		f.clients,
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *CosmosRPCFetcher) GetValidatorsAssignedConsumerKeys(
	height int64,
) (*providerTypes.QueryAllPairsValConsAddrByConsumerResponse, error) {
	query := providerTypes.QueryAllPairsValConsAddrByConsumerRequest{
		ConsumerId: f.config.ConsumerID,
	}

	var response providerTypes.QueryAllPairsValConsAddrByConsumerResponse
	if err := f.AbciQuery(
		"/interchain_security.ccv.provider.v1.Query/QueryAllPairsValConsAddrByConsumer",
		&query,
		height,
		constants.QueryTypeConsumerAddrs,
		&response,
		f.providerClients,
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *CosmosRPCFetcher) GetSlashingParams(height int64) (*slashingTypes.QueryParamsResponse, error) {
	var response slashingTypes.QueryParamsResponse
	if err := f.AbciQuery(
		"/cosmos.slashing.v1beta1.Query/Params",
		&slashingTypes.QueryParamsRequest{},
		height,
		constants.QueryTypeSlashingParams,
		&response,
		f.clients,
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *CosmosRPCFetcher) Get(
	url string,
	queryType constants.QueryType,
	target *responses.AbciQueryResponse,
	clients []*http.Client,
	predicate func(response *responses.AbciQueryResponse) error,
) error {
	errorsArray := make([]error, len(clients))

	indexesShuffled := utils.MakeShuffledArray(len(clients))
	clientsShuffled := make([]*http.Client, len(clients))

	for index, indexShuffled := range indexesShuffled {
		clientsShuffled[index] = clients[indexShuffled]
	}

	for index := range indexesShuffled {
		client := clientsShuffled[index]

		fullURL := client.Host + url
		f.logger.Trace().Str("url", fullURL).Msg("Trying making request to LCD")

		err := client.Get(
			url,
			queryType,
			target,
		)

		if err != nil {
			f.logger.Warn().Str("url", fullURL).Err(err).Msg("LCD request failed")
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
