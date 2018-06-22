package gateway

import (
	"io/ioutil"
	"encoding/json"
)

// 程序配置
type Config struct {
	WsPort int `json:"wsPort"`
	WsInChannelSize int `json:"wsInChannelSize"`
	WsOutChannelSize int `json:"wsOutChannelSize"`
	WsHeartbeatInterval int `json:"wsHeartbeatInterval"`
	MaxPushDelay int `json:"maxPushDelay"`
	MaxPushBatchSize int `json:"maxPushBatchSize"`
	AdminPort int `json:"adminPort"`
	BucketCount int `json:"bucketCount"`
	BucketWorkerCount int `json:"BucketWorkerCount"`
	DispatchChannelSize int `json:"dispatchChannelSize"`
	DispatchWorkerCount int `json:"dispatchWorkerCount"`
}

var (
	G_config *Config
)

func InitConfig(filename string) (err error) {
	var (
		content []byte
		conf Config
	)

	if content, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	if err = json.Unmarshal(content, &conf); err != nil {
		return
	}

	G_config = &conf
	return
}