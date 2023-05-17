package tendermint

import (
	"context"
	"encoding/json"
	"main/pkg/config"
	"main/pkg/constants"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"main/pkg/types"

	"github.com/rs/zerolog"
	tmClient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	rpcTypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
)

type WebsocketClient struct {
	logger zerolog.Logger
	config *config.Config
	url    string
	client *tmClient.WSClient
	active bool
	error  error

	Channel chan types.WebsocketEmittable
}

func NewWebsocketClient(
	logger zerolog.Logger,
	url string,
	config *config.Config,
) *WebsocketClient {
	return &WebsocketClient{
		logger: logger.With().
			Str("component", "tendermint_ws_client").
			Str("url", url).
			Str("chain", config.ChainConfig.Name).
			Logger(),
		url:     url,
		config:  config,
		active:  false,
		Channel: make(chan types.WebsocketEmittable),
	}
}

func SetUnexportedField(field reflect.Value, value interface{}) {
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}

func (t *WebsocketClient) Listen() {
	client, err := tmClient.NewWSWithOptions(
		t.url,
		"/websocket",
		tmClient.WSOptions{PingPeriod: 1 * time.Second},
	)

	client.OnReconnect(func() {
		t.logger.Info().Msg("Reconnecting...")
		t.SubscribeToUpdates()
	})

	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to create a client")
		t.error = err
		t.Channel <- &types.WSError{Error: err}
		return
	}

	// Patching WSS connections
	if strings.HasPrefix(t.url, "https") {
		field := reflect.ValueOf(client).Elem().FieldByName("protocol")
		SetUnexportedField(field, "wss")
	}

	t.client = client

	t.logger.Trace().Msg("Connecting to a node...")

	if err = client.Start(); err != nil {
		t.error = err
		t.Channel <- &types.WSError{Error: err}
		t.logger.Warn().Err(err).Msg("Error connecting to node")
	} else {
		t.logger.Debug().Msg("Connected to a node")
		t.active = true
	}

	t.SubscribeToUpdates()

	for {
		select {
		case result := <-client.ResponsesCh:
			t.ProcessEvent(result)
		}
	}
}

func (t *WebsocketClient) Stop() {
	t.logger.Info().Msg("Stopping the node...")

	if t.client != nil {
		if err := t.client.Stop(); err != nil {
			t.logger.Warn().Err(err).Msg("Error stopping the node")
		}
	}
}

func (t *WebsocketClient) ProcessEvent(event rpcTypes.RPCResponse) {
	if event.Error != nil && event.Error.Message != "" {
		t.logger.Error().Str("msg", event.Error.Error()).Msg("Got error in RPC response")
		t.Channel <- &types.WSError{Error: event.Error}
		return
	}

	if len(event.Result) == 0 {
		return
	}

	var resultEvent types.EventResult
	if err := json.Unmarshal(event.Result, &resultEvent); err != nil {
		t.logger.Error().Err(err).Msg("Failed to parse event")
		t.Channel <- &types.WSError{Error: err}
		return
	}

	if resultEvent.Query == "" {
		t.logger.Debug().Msg("Event is empty, skipping.")
		return
	}

	if resultEvent.Query != constants.NewBlocksQuery {
		t.logger.Warn().Str("query", resultEvent.Query).Msg("Unsupported event, skipping")
		return
	}

	blockDataStr, err := json.Marshal(resultEvent.Data.Value)
	if err != nil {
		t.logger.Err(err).Err(err).Msg("Could not marshal block data to string")
		return
	}

	var blockData types.SingleBlockResult
	if err := json.Unmarshal(blockDataStr, &blockData); err != nil {
		t.logger.Error().Err(err).Msg("Failed to unmarshall event")
	}

	block := blockData.Block.ToBlock()
	t.Channel <- block
}

func (t *WebsocketClient) SubscribeToUpdates() {
	t.logger.Trace().Msg("Subscribing to updates...")

	queries := []string{
		constants.NewBlocksQuery,
	}

	for _, query := range queries {
		if err := t.client.Subscribe(context.Background(), query); err != nil {
			t.logger.Error().Err(err).Str("query", query).Msg("Failed to subscribe to query")
		} else {
			t.logger.Info().Str("query", query).Msg("Listening for incoming transactions")
		}
	}
}
