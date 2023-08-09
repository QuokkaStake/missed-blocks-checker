package tendermint

import (
	"context"
	"encoding/json"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/metrics"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"main/pkg/types"

	"github.com/rs/zerolog"
	tmClient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	rpcTypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
)

const (
	StaleTime      = 10 * time.Minute
	StaleCheckTime = 1 * time.Minute
)

type WebsocketClient struct {
	logger         zerolog.Logger
	config         *configPkg.ChainConfig
	metricsManager *metrics.Manager
	url            string
	client         *tmClient.WSClient
	active         bool
	error          error
	reconnectTimer *time.Ticker
	lastEventTime  time.Time

	Channel chan types.WebsocketEmittable
}

func NewWebsocketClient(
	logger zerolog.Logger,
	url string,
	config *configPkg.ChainConfig,
	metricsManager *metrics.Manager,
) *WebsocketClient {
	return &WebsocketClient{
		logger: logger.With().
			Str("component", "tendermint_ws_client").
			Str("url", url).
			Str("chain", config.Name).
			Logger(),
		url:            url,
		config:         config,
		metricsManager: metricsManager,
		active:         false,
		Channel:        make(chan types.WebsocketEmittable),
		reconnectTimer: time.NewTicker(StaleCheckTime),
		lastEventTime:  time.Now(),
	}
}

func SetUnexportedField(field reflect.Value, value interface{}) {
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}

func (t *WebsocketClient) Listen() {
	client, err := tmClient.NewWS(
		t.url,
		"/websocket",
		tmClient.PingPeriod(1*time.Second),
		tmClient.ReadWait(10*time.Second),
		tmClient.WriteWait(10*time.Second),
		tmClient.OnReconnect(func() {
			t.metricsManager.LogNodeReconnect(t.config.Name, t.url)
			t.logger.Info().Msg("Reconnecting...")
			t.SubscribeToUpdates()
		}),
	)

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
	t.ConnectAndListen()
}

func (t *WebsocketClient) ConnectAndListen() {
	t.active = false
	t.metricsManager.LogNodeConnection(t.config.Name, t.url, false)

	t.logger.Info().Msg("Connecting to node...")

	for {
		if err := t.client.Start(); err != nil &&
			!strings.Contains(err.Error(), "client already running") &&
			!strings.Contains(err.Error(), "already started") {
			t.error = err
			t.Channel <- &types.WSError{Error: err}
			t.logger.Warn().Err(err).Msg("Error connecting to node")

			time.Sleep(time.Minute)
		} else {
			t.logger.Info().Msg("Connected to a node")
			t.active = true
			t.metricsManager.LogNodeConnection(t.config.Name, t.url, true)
			break
		}
	}

	t.logger.Info().Msg("Subscribing to updates...")

	t.SubscribeToUpdates()

	t.logger.Info().Msg("Listening for events...")

	loop := true
	for loop {
		select {
		case <-t.reconnectTimer.C:
			t.logger.Debug().Float64("seconds", time.Since(t.lastEventTime).Seconds()).Msg("Last block info")
			if time.Since(t.lastEventTime) > StaleTime {
				t.logger.Warn().Msg("No new blocks, reconnecting")
				loop = false
				break
			}
		case result := <-t.client.ResponsesCh:
			t.ProcessEvent(result)
		}
	}

	t.logger.Debug().Msg("Finished listening")
	t.Reconnect()
}

func (t *WebsocketClient) Reconnect() {
	if t.client == nil {
		t.logger.Debug().Msg("No client, return")
		return
	}

	t.logger.Debug().Msg("Reconnecting manually...")

	t.metricsManager.LogNodeReconnect(t.config.Name, t.url)

	if err := t.client.Stop(); err != nil {
		t.logger.Warn().Err(err).Msg("Error stopping the node")
		t.Channel <- &types.WSError{Error: err}
	}

	t.logger.Debug().Msg("Node disconnected, reconnecting...")

	time.Sleep(time.Second)
	t.ConnectAndListen()
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

	t.metricsManager.LogWSEvent(t.config.Name, t.url)
	t.lastEventTime = time.Now()

	var resultEvent types.EventResult
	if err := json.Unmarshal(event.Result, &resultEvent); err != nil {
		t.logger.Error().Err(err).Msg("Failed to parse event")
		t.Channel <- &types.WSError{Error: err}
		return
	}

	if resultEvent.Query == "" {
		t.logger.Debug().Msg("Event is empty, skipping")
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
		return
	}

	block, err := blockData.Block.ToBlock()
	if err != nil {
		t.logger.Error().Err(err).Msg("Failed to parse block")
		return
	}

	t.Channel <- block
}

func (t *WebsocketClient) SubscribeToUpdates() {
	t.logger.Info().Msg("Subscribing to updates...")

	queries := []string{
		constants.NewBlocksQuery,
	}

	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	for _, query := range queries {
		if err := t.client.Subscribe(ctxTimeout, query); err != nil {
			t.logger.Error().Err(err).Str("query", query).Msg("Failed to subscribe to query")
		} else {
			t.logger.Info().Str("query", query).Msg("Listening for incoming transactions")
		}
	}

	t.logger.Info().Msg("Subscribed to all updates")
}
