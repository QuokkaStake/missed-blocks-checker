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
	Logger  zerolog.Logger
	Config  *config.Config
	URL     string
	Client  *tmClient.WSClient
	Active  bool
	Error   error
	Channel chan types.WebsocketEmittable
}

func NewWebsocketClient(
	logger *zerolog.Logger,
	url string,
	config *config.Config,
) *WebsocketClient {
	return &WebsocketClient{
		Logger: logger.With().
			Str("component", "tendermint_ws_client").
			Str("url", url).
			Str("chain", config.ChainConfig.Name).
			Logger(),
		URL:     url,
		Config:  config,
		Active:  false,
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
		t.URL,
		"/websocket",
		tmClient.WSOptions{PingPeriod: 1 * time.Second},
	)

	client.OnReconnect(func() {
		t.Logger.Info().Msg("Reconnecting...")
		t.SubscribeToUpdates()
	})

	if err != nil {
		t.Logger.Error().Err(err).Msg("Failed to create a client")
		t.Error = err
		t.Channel <- &types.WSError{Error: err}
		return
	}

	// Patching WSS connections
	if strings.HasPrefix(t.URL, "https") {
		field := reflect.ValueOf(client).Elem().FieldByName("protocol")
		SetUnexportedField(field, "wss")
	}

	t.Client = client

	t.Logger.Trace().Msg("Connecting to a node...")

	if err = client.Start(); err != nil {
		t.Error = err
		t.Channel <- &types.WSError{Error: err}
		t.Logger.Warn().Err(err).Msg("Error connecting to node")
	} else {
		t.Logger.Debug().Msg("Connected to a node")
		t.Active = true
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
	t.Logger.Info().Msg("Stopping the node...")

	if t.Client != nil {
		if err := t.Client.Stop(); err != nil {
			t.Logger.Warn().Err(err).Msg("Error stopping the node")
		}
	}
}

func (t *WebsocketClient) ProcessEvent(event rpcTypes.RPCResponse) {
	if event.Error != nil && event.Error.Message != "" {
		t.Logger.Error().Str("msg", event.Error.Error()).Msg("Got error in RPC response")
		t.Channel <- &types.WSError{Error: event.Error}
		return
	}

	var resultEvent types.EventResult
	if err := json.Unmarshal(event.Result, &resultEvent); err != nil {
		t.Logger.Error().Err(err).Msg("Failed to parse event")
		t.Channel <- &types.WSError{Error: err}
		return
	}

	if resultEvent.Query == "" {
		t.Logger.Debug().Msg("Event is empty, skipping.")
		return
	}

	if resultEvent.Query != constants.NewBlocksQuery {
		t.Logger.Warn().Str("query", resultEvent.Query).Msg("Unsupported event, skipping")
		return
	}

	blockDataStr, err := json.Marshal(resultEvent.Data.Value)
	if err != nil {
		t.Logger.Err(err).Err(err).Msg("Could not marshal block data to string")
		return
	}

	var blockData types.SingleBlockResult
	if err := json.Unmarshal(blockDataStr, &blockData); err != nil {
		t.Logger.Error().Err(err).Msg("Failed to unmarshall event")
	}

	block := blockData.Block.ToBlock()
	t.Channel <- block
}

func (t *WebsocketClient) SubscribeToUpdates() {
	t.Logger.Trace().Msg("Subscribing to updates...")

	queries := []string{
		constants.NewBlocksQuery,
	}

	for _, query := range queries {
		if err := t.Client.Subscribe(context.Background(), query); err != nil {
			t.Logger.Error().Err(err).Str("query", query).Msg("Failed to subscribe to query")
		} else {
			t.Logger.Info().Str("query", query).Msg("Listening for incoming transactions")
		}
	}
}
