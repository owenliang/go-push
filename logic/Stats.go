package logic

import (
	"sync/atomic"
	"encoding/json"
)

type Stats struct {
	// 分发总消息数
	DispatchTotal int64 `json:"DispatchTotal"`
	// 分发丢弃消息数
	DispatchFail int64 `json:"DispatchFail"`
	// 推送失败次数
	PushFail int64 `json:"PushFail"`
}

var (
	G_stats *Stats
)

func InitStats() (err error) {
	G_stats = &Stats{}
	return
}

func DispatchTotal_INCR(batchSize int64) {
	atomic.AddInt64(&G_stats.DispatchTotal, batchSize)
}

func DispatchFail_INCR(batchSize int64) {
	atomic.AddInt64(&G_stats.DispatchFail, batchSize)
}

func PushFail_INCR() {
	atomic.AddInt64(&G_stats.PushFail, 1)
}

func (stats *Stats) Dump() (data []byte, err error){
	return json.Marshal(G_stats)
}