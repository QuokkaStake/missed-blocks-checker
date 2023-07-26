package tendermint

import (
	"main/pkg/metrics"
	"sync"

	"main/pkg/config"
	"main/pkg/types"

	"github.com/rs/zerolog"
)

type WebsocketManager struct {
	logger         zerolog.Logger
	nodes          []*WebsocketClient
	metricsManager *metrics.Manager
	queue          Queue
	mutex          sync.Mutex

	Channel chan types.WebsocketEmittable
}

func NewWebsocketManager(
	logger zerolog.Logger,
	config *config.ChainConfig,
	metricsManager *metrics.Manager,
) *WebsocketManager {
	nodes := make([]*WebsocketClient, len(config.RPCEndpoints))

	for index, url := range config.RPCEndpoints {
		nodes[index] = NewWebsocketClient(logger, url, config, metricsManager)
	}

	return &WebsocketManager{
		logger:         logger.With().Str("component", "websocket_manager").Logger(),
		nodes:          nodes,
		metricsManager: metricsManager,
		queue:          NewQueue(100),
		Channel:        make(chan types.WebsocketEmittable),
	}
}

func (m *WebsocketManager) Listen() {
	for _, node := range m.nodes {
		go node.Listen()
	}

	for _, node := range m.nodes {
		go m.ProcessNode(node)
	}
}

func (m *WebsocketManager) ProcessNode(node *WebsocketClient) {
	for msg := range node.Channel {
		m.mutex.Lock()

		if m.queue.Has(msg) {
			m.logger.Trace().
				Str("hash", msg.Hash()).
				Msg("Message already received, not sending again")
			m.mutex.Unlock()
			continue
		}

		m.Channel <- msg
		m.queue.Add(msg)
		m.mutex.Unlock()
	}
}
