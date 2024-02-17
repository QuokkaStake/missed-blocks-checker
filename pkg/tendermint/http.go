package tendermint

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
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"

	queryTypes "github.com/cosmos/cosmos-sdk/types/query"

	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	providerTypes "github.com/cosmos/interchain-security/x/ccv/provider/types"
	"github.com/rs/zerolog"
)

type RPC struct {
	config         *configPkg.ChainConfig
	metricsManager *metrics.Manager
	logger         zerolog.Logger

	clients         []*http.Client
	providerClients []*http.Client
}

func NewRPC(config *configPkg.ChainConfig, logger zerolog.Logger, metricsManager *metrics.Manager) *RPC {
	clients := make([]*http.Client, len(config.RPCEndpoints))
	for index, host := range config.RPCEndpoints {
		clients[index] = http.NewClient(logger, metricsManager, host, config.Name)
	}

	providerClients := make([]*http.Client, len(config.ProviderRPCEndpoints))
	for index, host := range config.ProviderRPCEndpoints {
		providerClients[index] = http.NewClient(logger, metricsManager, host, config.Name)
	}

	return &RPC{
		config:          config,
		metricsManager:  metricsManager,
		logger:          logger.With().Str("component", "rpc").Logger(),
		clients:         clients,
		providerClients: providerClients,
	}
}

func (rpc *RPC) GetConsumerOrProviderClients() []*http.Client {
	if rpc.config.IsConsumer.Bool {
		return rpc.providerClients
	}

	return rpc.clients
}

func (rpc *RPC) GetBlock(height int64) (*responses.SingleBlockResponse, error) {
	queryURL := "/block"
	if height != 0 {
		queryURL = fmt.Sprintf("/block?height=%d", height)
	}

	var response responses.SingleBlockResponse
	if err := rpc.Get(queryURL, constants.QueryTypeBlock, &response, rpc.clients, func(v interface{}) error {
		response, ok := v.(*responses.SingleBlockResponse)
		if !ok {
			return errors.New("error converting block")
		}

		if response.Error != nil {
			return fmt.Errorf("error in Tendermint response: %s", response.Error.Data)
		}

		if response.Result.Block.Header.Height == "" {
			return errors.New("malformed result of block: empty block height")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &response, nil
}

func (rpc *RPC) AbciQuery(
	method string,
	message codec.ProtoMarshaler,
	height int64,
	queryType constants.QueryType,
	output codec.ProtoMarshaler,
	clients []*http.Client,
) error {
	dataBytes, err := message.Marshal()
	if err != nil {
		return err
	}

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
	if err := rpc.Get(queryURL, constants.QueryType("abci_"+string(queryType)), &response, clients, func(v interface{}) error {
		response, ok := v.(*responses.AbciQueryResponse)
		if !ok {
			return errors.New("error converting ABCI response")
		}

		// code = NotFound desc = SigningInfo not found for validator xxx: key not found
		if queryType == constants.QueryTypeSigningInfo && response.Result.Response.Code == 22 {
			return nil
		}

		if response.Result.Response.Code != 0 {
			return fmt.Errorf(
				"error in Tendermint response: expected code 0, but got %d, error: %s",
				response.Result.Response.Code,
				response.Result.Response.Log,
			)
		}

		return nil
	}); err != nil {
		return err
	}

	return output.Unmarshal(response.Result.Response.Value)
}

func (rpc *RPC) GetValidators(height int64) (*stakingTypes.QueryValidatorsResponse, error) {
	query := stakingTypes.QueryValidatorsRequest{
		Pagination: &queryTypes.PageRequest{
			Limit: rpc.config.Pagination.ValidatorsList,
		},
	}

	var validatorsResponse stakingTypes.QueryValidatorsResponse
	if err := rpc.AbciQuery(
		"/cosmos.staking.v1beta1.Query/Validators",
		&query,
		height,
		constants.QueryTypeValidators,
		&validatorsResponse,
		rpc.GetConsumerOrProviderClients(),
	); err != nil {
		return nil, err
	}

	return &validatorsResponse, nil
}

func (rpc *RPC) GetSigningInfos(height int64) (*slashingTypes.QuerySigningInfosResponse, error) {
	query := slashingTypes.QuerySigningInfosRequest{
		Pagination: &queryTypes.PageRequest{
			Limit: rpc.config.Pagination.SigningInfos,
		},
	}

	var response slashingTypes.QuerySigningInfosResponse
	if err := rpc.AbciQuery(
		"/cosmos.slashing.v1beta1.Query/SigningInfos",
		&query,
		height,
		constants.QueryTypeSigningInfos,
		&response,
		rpc.clients,
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (rpc *RPC) GetSigningInfo(valcons string, height int64) (*slashingTypes.QuerySigningInfoResponse, error) {
	query := slashingTypes.QuerySigningInfoRequest{
		ConsAddress: valcons,
	}

	var response slashingTypes.QuerySigningInfoResponse
	if err := rpc.AbciQuery(
		"/cosmos.slashing.v1beta1.Query/SigningInfo",
		&query,
		height,
		constants.QueryTypeSigningInfo,
		&response,
		rpc.clients,
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (rpc *RPC) GetValidatorAssignedConsumerKey(
	providerValcons string,
	height int64,
) (*providerTypes.QueryValidatorConsumerAddrResponse, error) {
	query := providerTypes.QueryValidatorConsumerAddrRequest{
		ChainId:         rpc.config.ConsumerChainID,
		ProviderAddress: providerValcons,
	}

	var response providerTypes.QueryValidatorConsumerAddrResponse
	if err := rpc.AbciQuery(
		"/interchain_security.ccv.provider.v1.Query/QueryValidatorConsumerAddr",
		&query,
		height,
		constants.QueryTypeConsumerAddr,
		&response,
		rpc.providerClients,
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (rpc *RPC) GetSlashingParams(height int64) (*slashingTypes.QueryParamsResponse, error) {
	var response slashingTypes.QueryParamsResponse
	if err := rpc.AbciQuery(
		"/cosmos.slashing.v1beta1.Query/Params",
		&slashingTypes.QueryParamsRequest{},
		height,
		constants.QueryTypeSlashingParams,
		&response,
		rpc.clients,
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (rpc *RPC) GetConsumerSoftOutOutThreshold(height int64) (*paramsTypes.QueryParamsResponse, error) {
	query := paramsTypes.QueryParamsRequest{
		Subspace: "ccvconsumer",
		Key:      "SoftOptOutThreshold",
	}

	var response paramsTypes.QueryParamsResponse
	if err := rpc.AbciQuery(
		"/cosmos.params.v1beta1.Query/Params",
		&query,
		height,
		constants.QueryTypeSubspaceParams,
		&response,
		rpc.clients,
	); err != nil {
		return nil, err
	}

	return &response, nil
}

func (rpc *RPC) GetActiveSetAtBlock(height int64) (map[string]bool, error) {
	page := 1

	activeSetMap := make(map[string]bool)

	for {
		queryURL := fmt.Sprintf(
			"/validators?height=%d&per_page=100&page=%d",
			height,
			page,
		)

		var response responses.ValidatorsResponse
		if err := rpc.Get(queryURL, constants.QueryTypeHistoricalValidators, &response, rpc.clients, func(v interface{}) error {
			response, ok := v.(*responses.ValidatorsResponse)
			if !ok {
				return errors.New("error converting validators")
			}

			if response.Error != nil {
				return fmt.Errorf("error in Tendermint response: %s", response.Error.Data)
			}

			if len(response.Result.Validators) == 0 {
				return errors.New("malformed result of validators active set: got 0 validators")
			}

			return nil
		}); err != nil {
			return nil, err
		}

		validatorsCount, err := strconv.Atoi(response.Result.Total)
		if err != nil {
			rpc.logger.Warn().
				Err(err).
				Msg("Error parsing validators count from response")
			return nil, err
		}

		for _, validator := range response.Result.Validators {
			activeSetMap[validator.Address] = true
		}

		if len(activeSetMap) >= validatorsCount {
			break
		}

		page += 1
	}

	return activeSetMap, nil
}

func (rpc *RPC) Get(
	url string,
	queryType constants.QueryType,
	target interface{},
	clients []*http.Client,
	predicate func(interface{}) error,
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
		rpc.logger.Trace().Str("url", fullURL).Msg("Trying making request to LCD")

		err := client.Get(
			url,
			queryType,
			target,
		)

		if err != nil {
			rpc.logger.Warn().Str("url", fullURL).Err(err).Msg("LCD request failed")
			errorsArray[index] = err
			continue
		}

		if predicateErr := predicate(target); predicateErr != nil {
			rpc.logger.Warn().Str("url", fullURL).Err(predicateErr).Msg("LCD precondition failed")
			errorsArray[index] = fmt.Errorf("precondition failed: %s", predicateErr)
			continue
		}

		return nil
	}

	rpc.logger.Warn().Str("url", url).Msg("All LCD requests failed")

	var sb strings.Builder

	sb.WriteString("All LCD requests failed:\n")
	for index, client := range clientsShuffled {
		sb.WriteString(fmt.Sprintf("#%d: %s -> %s\n", index+1, client.Host, errorsArray[index]))
	}

	return errors.New(sb.String())
}
