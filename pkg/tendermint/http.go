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
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

type RPC struct {
	config         *configPkg.ChainConfig
	metricsManager *metrics.Manager
	logger         zerolog.Logger

	clients []*http.Client
}

func NewRPC(config *configPkg.ChainConfig, logger zerolog.Logger, metricsManager *metrics.Manager) *RPC {
	clients := make([]*http.Client, len(config.RPCEndpoints))
	for index, host := range config.RPCEndpoints {
		clients[index] = http.NewClient(logger, metricsManager, host, config.Name)
	}

	return &RPC{
		config:         config,
		metricsManager: metricsManager,
		logger:         logger.With().Str("component", "rpc").Logger(),
		clients:        clients,
	}
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
