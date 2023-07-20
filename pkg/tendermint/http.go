package tendermint

import (
	"encoding/json"
	"fmt"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/metrics"
	"main/pkg/utils"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	queryTypes "github.com/cosmos/cosmos-sdk/types/query"

	"main/pkg/types"

	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	providerTypes "github.com/cosmos/interchain-security/x/ccv/provider/types"
	"github.com/rs/zerolog"
)

type RPC struct {
	config         *configPkg.ChainConfig
	metricsManager *metrics.Manager
	logger         zerolog.Logger
}

func NewRPC(config *configPkg.ChainConfig, logger zerolog.Logger, metricsManager *metrics.Manager) *RPC {
	return &RPC{
		config:         config,
		metricsManager: metricsManager,
		logger:         logger.With().Str("component", "rpc").Logger(),
	}
}

func (rpc *RPC) GetConsumerOrProviderHosts() []string {
	if rpc.config.IsConsumer.Bool {
		return rpc.config.ProviderRPCEndpoints
	}

	return rpc.config.RPCEndpoints
}

func (rpc *RPC) GetBlock(height int64) (*types.SingleBlockResponse, error) {
	queryURL := "/block"
	if height != 0 {
		queryURL = fmt.Sprintf("/block?height=%d", height)
	}

	var response types.SingleBlockResponse
	if err := rpc.Get(queryURL, constants.QueryTypeBlock, &response, rpc.config.RPCEndpoints, func(v interface{}) error {
		response, ok := v.(*types.SingleBlockResponse)
		if !ok {
			return fmt.Errorf("error converting block")
		}

		if response.Error != nil {
			return fmt.Errorf("error in Tendermint response: %s", response.Error.Data)
		}

		if response.Result.Block.Header.Height == "" {
			return fmt.Errorf("malformed result of block: empty block height")
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
	hosts []string,
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

	var response types.AbciQueryResponse
	if err := rpc.Get(queryURL, constants.QueryType("abci_"+string(queryType)), &response, hosts, func(v interface{}) error {
		response, ok := v.(*types.AbciQueryResponse)
		if !ok {
			return fmt.Errorf("error converting ABCI response")
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
			Limit: constants.ValidatorsQueryPagination,
		},
	}

	var validatorsResponse stakingTypes.QueryValidatorsResponse
	if err := rpc.AbciQuery(
		"/cosmos.staking.v1beta1.Query/Validators",
		&query,
		height,
		constants.QueryTypeValidators,
		&validatorsResponse,
		rpc.GetConsumerOrProviderHosts(),
	); err != nil {
		return nil, err
	}

	return &validatorsResponse, nil
}

func (rpc *RPC) GetSigningInfos(height int64) (*slashingTypes.QuerySigningInfosResponse, error) {
	query := slashingTypes.QuerySigningInfosRequest{
		Pagination: &queryTypes.PageRequest{
			Limit: constants.SigningInfosQueryPagination,
		},
	}

	var response slashingTypes.QuerySigningInfosResponse
	if err := rpc.AbciQuery(
		"/cosmos.slashing.v1beta1.Query/SigningInfos",
		&query,
		height,
		constants.QueryTypeSigningInfos,
		&response,
		rpc.config.RPCEndpoints,
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
		rpc.config.RPCEndpoints,
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
		ChainId:         "neutron-1",
		ProviderAddress: providerValcons,
	}

	var response providerTypes.QueryValidatorConsumerAddrResponse
	if err := rpc.AbciQuery(
		"/interchain_security.ccv.provider.v1.Query/QueryValidatorConsumerAddr",
		&query,
		height,
		constants.QueryTypeConsumerAddr,
		&response,
		rpc.config.ProviderRPCEndpoints,
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
		rpc.config.RPCEndpoints,
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
			"/validators?height=%d&per_page=%d&page=%d",
			height,
			constants.ActiveSetPagination,
			page,
		)

		var response types.ValidatorsResponse
		if err := rpc.Get(queryURL, constants.QueryTypeHistoricalValidators, &response, rpc.config.RPCEndpoints, func(v interface{}) error {
			response, ok := v.(*types.ValidatorsResponse)
			if !ok {
				return fmt.Errorf("error converting validators")
			}

			if response.Error != nil {
				return fmt.Errorf("error in Tendermint response: %s", response.Error.Data)
			}

			if len(response.Result.Validators) == 0 {
				return fmt.Errorf("malformed result of validators active set: got 0 validators")
			}

			return nil
		}); err != nil {
			return nil, err
		}

		for _, validator := range response.Result.Validators {
			activeSetMap[validator.Address] = true
		}

		if len(response.Result.Validators) <= constants.ActiveSetPagination {
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
	hosts []string,
	predicate func(interface{}) error,
) error {
	errors := make([]error, len(hosts))

	indexes := utils.MakeShuffledArray(len(hosts))

	for _, index := range indexes {
		lcd := hosts[index]

		fullURL := lcd + url
		rpc.logger.Trace().Str("url", fullURL).Msg("Trying making request to LCD")

		err := rpc.GetFull(
			lcd,
			url,
			queryType,
			target,
		)

		if err != nil {
			rpc.logger.Warn().Str("url", fullURL).Err(err).Msg("LCD request failed")
			errors[index] = err
			continue
		}

		if predicateErr := predicate(target); predicateErr != nil {
			rpc.logger.Warn().Str("url", fullURL).Err(predicateErr).Msg("LCD precondition failed")
			errors[index] = fmt.Errorf("precondition failed")
			continue
		}

		return nil
	}

	rpc.logger.Warn().Str("url", url).Msg("All LCD requests failed")

	var sb strings.Builder

	sb.WriteString("All LCD requests failed:\n")
	for index, nodeURL := range hosts {
		sb.WriteString(fmt.Sprintf("#%d: %s -> %s\n", index+1, nodeURL, errors[index]))
	}

	return fmt.Errorf(sb.String())
}

func (rpc *RPC) GetFull(
	host, url string,
	queryType constants.QueryType,
	target interface{},
) error {
	client := &http.Client{Timeout: 60 * 1000000000}
	start := time.Now()

	fullURL := host + url

	queryInfo := types.QueryInfo{
		Success:   false,
		Node:      host,
		QueryType: queryType,
	}

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "missed-blocks-checker")

	rpc.logger.Trace().
		Str("url", fullURL).
		Msg("Doing a query...")

	res, err := client.Do(req)
	if err != nil {
		rpc.logger.Warn().Str("url", fullURL).Err(err).Msg("Query failed")
		rpc.metricsManager.LogTendermintQuery(rpc.config.Name, queryInfo)
		return err
	}
	defer res.Body.Close()

	rpc.logger.Debug().
		Str("url", fullURL).
		Dur("duration", time.Since(start)).
		Msg("Query is finished")

	if jsonErr := json.NewDecoder(res.Body).Decode(target); jsonErr != nil {
		rpc.logger.Warn().Str("url", fullURL).Err(jsonErr).Msg("Error decoding JSON from response")
		rpc.metricsManager.LogTendermintQuery(rpc.config.Name, queryInfo)
		return jsonErr
	}

	queryInfo.Success = true
	rpc.metricsManager.LogTendermintQuery(rpc.config.Name, queryInfo)

	return nil
}
