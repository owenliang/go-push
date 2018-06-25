package logic

import (
	"io/ioutil"
	"encoding/json"
)

type GatewayConfig struct {
	Hostname string `json:"hostname"`
	Port int `json:"port"`
}

// 程序配置
type Config struct {
	ServicePort int `json:"servicePort"`
	ServiceReadTimeout int `json:"serviceReadTimeout"`
	ServiceWriteTimeout int `json:"serviceWriteTimeout"`
	GatewayList []GatewayConfig`json:"gatewayList"`
	GatewayMaxConnection int `json:"gatewayMaxConnection"`
	GatewayTimeout int `json:"gatewayTimeout"`
	GatewayIdleTimeout int `json:"gatewayIdleTimeout"`
	GatewayDispatchWorkerCount int `json:"gatewayDispatchWorkerCount"`
	GatewayDispatchChannelSize int `json:"gatewayDispatchChannelSize"`
	GatewayMaxPendingCount int `json:"gatewayMaxPendingCount"`
	GatewayPushRetry int `json:"gatewayPushRetry"`
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