package tendermint

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"main/pkg/types"

	"github.com/rs/zerolog"
)

const PaginationLimit = 100

type RPC struct {
	URLs   []string
	Logger zerolog.Logger
}

func NewRPC(urls []string, logger *zerolog.Logger) *RPC {
	return &RPC{
		URLs:   urls,
		Logger: logger.With().Str("component", "rpc").Logger(),
	}
}

func (rpc *RPC) GetLatestBlock() (*types.SingleBlockResponse, error) {
	var response types.SingleBlockResponse
	if err := rpc.Get("/block", &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (rpc *RPC) GetBlocksFromTo(from, to, limit int64) (*types.BlockSearchResponse, error) {
	query := fmt.Sprintf(
		"\"block.height >= %d AND block.height < %d\"",
		from,
		to,
	)
	queryURL := fmt.Sprintf(
		"/block_search?query=%s&per_page=%d",
		url.QueryEscape(query),
		limit,
	)

	var response types.BlockSearchResponse
	if err := rpc.Get(queryURL, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (rpc *RPC) Get(url string, target interface{}) error {
	errors := make([]error, len(rpc.URLs))

	for index, lcd := range rpc.URLs {
		fullURL := lcd + url
		rpc.Logger.Trace().Str("url", fullURL).Msg("Trying making request to LCD")

		err := rpc.GetFull(
			fullURL,
			target,
		)

		if err == nil {
			return nil
		}

		rpc.Logger.Warn().Str("url", fullURL).Err(err).Msg("LCD request failed")
		errors[index] = err
	}

	rpc.Logger.Warn().Str("url", url).Msg("All LCD requests failed")

	var sb strings.Builder

	sb.WriteString("All LCD requests failed:\n")
	for index, url := range rpc.URLs {
		sb.WriteString(fmt.Sprintf("#%d: %s -> %s\n", index+1, url, errors[index]))
	}

	return fmt.Errorf(sb.String())
}

func (rpc *RPC) GetFull(url string, target interface{}) error {
	client := &http.Client{Timeout: 10 * 1000000000}
	start := time.Now()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "missed-blocks-checker")

	rpc.Logger.Debug().Str("url", url).Msg("Doing a query...")

	res, err := client.Do(req)
	if err != nil {
		rpc.Logger.Warn().Str("url", url).Err(err).Msg("Query failed")
		return err
	}
	defer res.Body.Close()

	rpc.Logger.Debug().Str("url", url).Dur("duration", time.Since(start)).Msg("Query is finished")

	return json.NewDecoder(res.Body).Decode(target)
}