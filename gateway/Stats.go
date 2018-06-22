package gateway

import (
	"sync/atomic"
	"encoding/json"
)

type Stats struct {
	PushAllPending int64 `json:"pushAllPending"`
	PushJobPending int64 `json:"pushJobPending"`
}

var (
	G_stats *Stats
)

func InitStats() (err error) {
	G_stats = &Stats{}
	return
}

func PushAllPending_INCR() {
	atomic.AddInt64(&G_stats.PushAllPending, 1)
}

func PushAllPending_DESC() {
	atomic.AddInt64(&G_stats.PushAllPending, -1)
}


func PushJobPending_INCR() {
	atomic.AddInt64(&G_stats.PushJobPending, 1)
}

func PushJobPending_DESC() {
	atomic.AddInt64(&G_stats.PushJobPending, -1)
}

func (stats *Stats) Dump() (data []byte, err error){
	return json.Marshal(G_stats)
}