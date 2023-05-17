package tendermint

import (
	"main/pkg/metrics"
	"main/pkg/types"
	"sync"

	"github.com/rs/zerolog"
)

type RPCManager struct {
	rpc *RPC
}

func NewRPCManager(urls []string, logger zerolog.Logger, metricsManager *metrics.Manager) *RPCManager {
	rpc := NewRPC(urls, logger, metricsManager)
	return &RPCManager{rpc: rpc}
}

func (manager *RPCManager) GetBlock(height int64) (*types.SingleBlockResponse, error) {
	return manager.rpc.GetBlock(height)
}

func (manager *RPCManager) GetValidators() (types.Validators, error) {
	return manager.rpc.GetValidators()
}

func (manager *RPCManager) GetActiveSetAtBlock(height int64) (map[string]bool, error) {
	return manager.rpc.GetActiveSetAtBlock(height)
}

func (manager *RPCManager) GetActiveSetAtBlocks(blocks []int64) (map[int64]map[string]bool, []error) {
	activeSetsMap := make(map[int64]map[string]bool)
	errors := make([]error, 0)

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, height := range blocks {
		wg.Add(1)
		go func(height int64) {
			activeSet, err := manager.rpc.GetActiveSetAtBlock(height)
			mutex.Lock()
			defer mutex.Unlock()

			if err != nil {
				errors = append(errors, err)
			} else {
				activeSetsMap[height] = activeSet
			}

			wg.Done()
		}(height)
	}

	wg.Wait()
	return activeSetsMap, nil
}

func (manager *RPCManager) GetBlocksAtHeights(heights []int64) (map[int64]*types.SingleBlockResponse, []error) {
	blocksMap := make(map[int64]*types.SingleBlockResponse)
	errors := make([]error, 0)

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, height := range heights {
		wg.Add(1)
		go func(height int64) {
			block, err := manager.rpc.GetBlock(height)
			mutex.Lock()
			defer mutex.Unlock()

			if err != nil {
				errors = append(errors, err)
			} else {
				blocksMap[height] = block
			}

			wg.Done()
		}(height)
	}

	wg.Wait()
	return blocksMap, nil
}
