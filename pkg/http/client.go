package http

import (
	"encoding/json"
	"io"
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

func (c *Client) GetInternal(
	url string,
	queryType constants.QueryType,
	headers map[string]string,
) (io.ReadCloser, error) {
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
		return nil, err
	}

	req.Header.Set("User-Agent", "missed-blocks-checker")

	for headerName, headerValue := range headers {
		req.Header.Set(headerName, headerValue)
	}

	c.logger.Trace().
		Str("url", fullURL).
		Msg("Doing a query...")

	res, err := client.Do(req)
	if err != nil {
		c.logger.Warn().Str("url", fullURL).Err(err).Msg("Query failed")
		c.metricsManager.LogQuery(c.chainName, queryInfo)
		return nil, err
	}

	c.logger.Debug().
		Str("url", fullURL).
		Dur("duration", time.Since(start)).
		Msg("Query is finished")

	queryInfo.Success = true
	c.metricsManager.LogQuery(c.chainName, queryInfo)

	return res.Body, nil
}

func (c *Client) GetPlain(
	url string,
	queryType constants.QueryType,
	headers map[string]string,
) ([]byte, error) {
	body, err := c.GetInternal(url, queryType, headers)
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (c *Client) Get(
	url string,
	queryType constants.QueryType,
	target interface{},
) error {
	body, err := c.GetInternal(url, queryType, map[string]string{})
	if err != nil {
		return err
	}

	fullURL := c.Host + url

	if jsonErr := json.NewDecoder(body).Decode(target); jsonErr != nil {
		c.logger.Warn().Str("url", fullURL).Err(jsonErr).Msg("Error decoding JSON from response")
		return jsonErr
	}

	if err := body.Close(); err != nil {
		return err
	}

	return nil
}
