package gateway

import (
	"sync/atomic"
	"encoding/json"
)

type Stats struct {
	// 反馈在线长连接的数量
	OnlineConnections int64 `json:"onlineConnections"`

	// 反馈客户端的推送压力
	SendMessageTotal int64 `json:"sendMessageTotal"`
	SendMessageFail int64 `json:"sendMessageFail"`

	// 反馈ConnMgr消息分发模块的压力
	DispatchPending int64 `json:"dispatchPending"`
	PushJobPending int64 `json:"pushJobPending"`
	DispatchFail int64 `json:"dispatchFail"`

	// 返回出在线的房间总数, 有利于分析内存上涨的原因
	RoomCount int64 `json:"roomCount"`

	// Merger模块处理队列, 反馈出消息合并的压力情况
	MergerPending int64 `json:"mergerPending"`

	// Merger模块合并发送的消息总数与失败总数
	MergerRoomTotal int64 `json:"mergerRoomTotal"`
	MergerAllTotal int64 `json:"mergerAllTotal"`
	MergerRoomFail int64 `json:"mergerRoomFail"`
	MergerAllFail int64 `json:"mergerAllFail"`
}

var (
	G_stats *Stats
)

func InitStats() (err error) {
	G_stats = &Stats{}
	return
}

func DispatchPending_INCR() {
	atomic.AddInt64(&G_stats.DispatchPending, 1)
}

func DispatchPending_DESC() {
	atomic.AddInt64(&G_stats.DispatchPending, -1)
}

func PushJobPending_INCR() {
	atomic.AddInt64(&G_stats.PushJobPending, 1)
}

func PushJobPending_DESC() {
	atomic.AddInt64(&G_stats.PushJobPending, -1)
}

func OnlineConnections_INCR() {
	atomic.AddInt64(&G_stats.OnlineConnections, 1)
}

func OnlineConnections_DESC() {
	atomic.AddInt64(&G_stats.OnlineConnections, -1)
}

func RoomCount_INCR() {
	atomic.AddInt64(&G_stats.RoomCount, 1)
}

func RoomCount_DESC() {
	atomic.AddInt64(&G_stats.RoomCount, -1)
}

func MergerPending_INCR() {
	atomic.AddInt64(&G_stats.MergerPending, 1)
}

func MergerPending_DESC() {
	atomic.AddInt64(&G_stats.MergerPending, -1)
}

func MergerRoomTotal_INCR(batchSize int64) {
	atomic.AddInt64(&G_stats.MergerRoomTotal, batchSize)
}

func MergerAllTotal_INCR(batchSize int64) {
	atomic.AddInt64(&G_stats.MergerAllTotal, batchSize)
}

func MergerRoomFail_INCR(batchSize int64) {
	atomic.AddInt64(&G_stats.MergerRoomFail, batchSize)
}

func MergerAllFail_INCR(batchSize int64) {
	atomic.AddInt64(&G_stats.MergerAllFail, batchSize)
}

func DispatchFail_INCR() {
	atomic.AddInt64(&G_stats.DispatchFail, 1)
}

func SendMessageFail_INCR() {
	atomic.AddInt64(&G_stats.SendMessageFail, 1)
}

func SendMessageTotal_INCR() {
	atomic.AddInt64(&G_stats.SendMessageTotal, 1)
}

func (stats *Stats) Dump() (data []byte, err error){
	return json.Marshal(G_stats)
}