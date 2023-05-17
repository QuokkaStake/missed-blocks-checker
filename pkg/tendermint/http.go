package tendermint

import (
	"encoding/json"
	"fmt"
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
	"github.com/rs/zerolog"
)

type RPC struct {
	urls           []string
	metricsManager *metrics.Manager
	logger         zerolog.Logger
}

func NewRPC(urls []string, logger zerolog.Logger, metricsManager *metrics.Manager) *RPC {
	return &RPC{
		urls:           urls,
		metricsManager: metricsManager,
		logger:         logger.With().Str("component", "rpc").Logger(),
	}
}

func (rpc *RPC) GetBlock(height int64) (*types.SingleBlockResponse, error) {
	queryURL := "/block"
	if height != 0 {
		queryURL = fmt.Sprintf("/block?height=%d", height)
	}

	var response types.SingleBlockResponse
	if err := rpc.Get(queryURL, "block", &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (rpc *RPC) AbciQuery(
	method string,
	message codec.ProtoMarshaler,
	queryType string,
	output codec.ProtoMarshaler,
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

	var response types.AbciQueryResponse
	if err := rpc.Get(queryURL, "abci_"+queryType, &response); err != nil {
		return err
	}

	return output.Unmarshal(response.Result.Response.Value)
}

func (rpc *RPC) GetValidators() (types.Validators, error) {
	query := stakingTypes.QueryValidatorsRequest{
		Pagination: &queryTypes.PageRequest{
			Limit: constants.ValidatorsQueryPagination,
		},
	}

	var validatorsResponse stakingTypes.QueryValidatorsResponse
	if err := rpc.AbciQuery("/cosmos.staking.v1beta1.Query/Validators", &query, "validators", &validatorsResponse); err != nil {
		return nil, err
	}

	return utils.Map(validatorsResponse.Validators, types.ValidatorFromCosmosValidator), nil
}

func (rpc *RPC) GetSigningInfo() error {
	query := slashingTypes.QuerySigningInfosRequest{
		Pagination: &queryTypes.PageRequest{
			Limit: constants.ValidatorsQueryPagination,
		},
	}

	var response slashingTypes.QuerySigningInfosResponse
	if err := rpc.AbciQuery("/cosmos.slashing.v1beta1.Query/SigningInfos", &query, "signing_infos", &response); err != nil {
		return err
	}

	return nil
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
		if err := rpc.Get(queryURL, "historical_validators", &response); err != nil {
			return nil, err
		}

		if len(response.Result.Validators) == 0 {
			return nil, fmt.Errorf("malformed result of validators active set: got 0 validators")
		}

		for _, validator := range response.Result.Validators {
			activeSetMap[validator.Address] = true
		}

		if len(response.Result.Validators) < constants.ActiveSetPagination {
			break
		}

		page += 1
	}

	return activeSetMap, nil
}

func (rpc *RPC) Get(url string, queryType string, target interface{}) error {
	errors := make([]error, len(rpc.urls))

	for index, lcd := range rpc.urls {
		fullURL := lcd + url
		rpc.logger.Trace().Str("url", fullURL).Msg("Trying making request to LCD")

		err := rpc.GetFull(
			lcd,
			url,
			queryType,
			target,
		)

		if err == nil {
			return nil
		}

		rpc.logger.Warn().Str("url", fullURL).Err(err).Msg("LCD request failed")
		errors[index] = err
	}

	rpc.logger.Warn().Str("url", url).Msg("All LCD requests failed")

	var sb strings.Builder

	sb.WriteString("All LCD requests failed:\n")
	for index, url := range rpc.urls {
		sb.WriteString(fmt.Sprintf("#%d: %s -> %s\n", index+1, url, errors[index]))
	}

	return fmt.Errorf(sb.String())
}

func (rpc *RPC) GetFull(host, url string, queryType string, target interface{}) error {
	client := &http.Client{Timeout: 10 * 1000000000}
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
		rpc.metricsManager.LogTendermintQuery(queryInfo)
		return err
	}
	defer res.Body.Close()

	rpc.logger.Debug().Str("url", url).Dur("duration", time.Since(start)).Msg("Query is finished")

	queryInfo.Success = true
	rpc.metricsManager.LogTendermintQuery(queryInfo)

	return json.NewDecoder(res.Body).Decode(target)
}
