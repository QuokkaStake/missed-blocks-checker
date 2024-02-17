package http

import (
	"encoding/json"
	"main/pkg/constants"
	"main/pkg/metrics"
	"main/pkg/types"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type Client struct {
	logger         zerolog.Logger
	metricsManager *metrics.Manager
	chainName      string

	Host string
}

func NewClient(
	logger zerolog.Logger,
	metricsManager *metrics.Manager,
	host string,
	chainName string,
) *Client {
	return &Client{
		logger: logger.With().
			Str("component", "http_client").
			Str("host", host).
			Logger(),
		metricsManager: metricsManager,
		chainName:      chainName,
		Host:           host,
	}
}

func (c *Client) Get(
	url string,
	queryType constants.QueryType,
	target interface{},
) error {
	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now()

	fullURL := c.Host + url

	queryInfo := types.QueryInfo{
		Success:   false,
		Node:      c.Host,
		QueryType: queryType,
	}

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "missed-blocks-checker")

	c.logger.Trace().
		Str("url", fullURL).
		Msg("Doing a query...")

	res, err := client.Do(req)
	if err != nil {
		c.logger.Warn().Str("url", fullURL).Err(err).Msg("Query failed")
		c.metricsManager.LogQuery(c.chainName, queryInfo)
		return err
	}
	defer res.Body.Close()

	c.logger.Debug().
		Str("url", fullURL).
		Dur("duration", time.Since(start)).
		Msg("Query is finished")

	if jsonErr := json.NewDecoder(res.Body).Decode(target); jsonErr != nil {
		c.logger.Warn().Str("url", fullURL).Err(jsonErr).Msg("Error decoding JSON from response")
		c.metricsManager.LogQuery(c.chainName, queryInfo)
		return jsonErr
	}

	queryInfo.Success = true
	c.metricsManager.LogQuery(c.chainName, queryInfo)

	return nil
}
