package tendermint

import (
	"main/pkg/types"
	"sync"

	"github.com/rs/zerolog"
)

type RPCManager struct {
	rpc *RPC
}

func NewRPCManager(urls []string, logger zerolog.Logger) *RPCManager {
	rpc := NewRPC(urls, logger)
	return &RPCManager{rpc: rpc}
}

func (manager *RPCManager) GetLatestBlock() (*types.SingleBlockResponse, error) {
	return manager.rpc.GetLatestBlock()
}

func (manager *RPCManager) GetBlocksFromTo(from, to, limit int64) (*types.BlockSearchResponse, error) {
	return manager.rpc.GetBlocksFromTo(from, to, limit)
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
