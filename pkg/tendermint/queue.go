package tendermint

import (
	"main/pkg/types"
	"sync"
)

type Queue struct {
	Data  []types.WebsocketEmittable
	Size  int
	Mutes sync.Mutex
}

func NewQueue(size int) Queue {
	return Queue{Data: make([]types.WebsocketEmittable, 0), Size: size}
}

func (q *Queue) Add(emittable types.WebsocketEmittable) {
	q.Mutes.Lock()

	if len(q.Data) >= q.Size {
		_, q.Data = q.Data[0], q.Data[1:]
	}

	q.Data = append(q.Data, emittable)
	q.Mutes.Unlock()
}

func (q *Queue) Has(emittable types.WebsocketEmittable) bool {
	for _, elem := range q.Data {
		if elem.Hash() == emittable.Hash() {
			return true
		}
	}

	return false
}
