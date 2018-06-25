package logic

import (
	"net/http"
	"crypto/tls"
	"time"
	"net/url"
	"strconv"
)

// 与网关之间的通讯
type GateConn struct {
	schema string
	client *http.Client	// 内置长连接+并发连接数
}

func InitGateConn(gatewayConfig *GatewayConfig) (gateConn *GateConn, err error) {
	gateConn = &GateConn{
		schema: "https://" + gatewayConfig.Hostname + ":" + strconv.Itoa(gatewayConfig.Port),
	}

	// HTTP/2 客户端
	gateConn.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true,},	// 不校验服务端证书
			MaxIdleConns: G_config.GatewayMaxConnection,
			MaxIdleConnsPerHost: G_config.GatewayMaxConnection,
			IdleConnTimeout: time.Duration(G_config.GatewayIdleTimeout) * time.Second,	// 连接空闲超时
		},
		Timeout: time.Duration(G_config.GatewayTimeout) * time.Millisecond, // 请求超时
	}
	return
}

// 出于性能考虑, 消息数组在此前已经编码成json
func (gateConn *GateConn) PushAll(itemsJson []byte) (err error) {
	var (
		apiUrl string
		form url.Values
		resp *http.Response
	)

	form = url.Values{}
	form.Set("items", string(itemsJson))

	apiUrl = gateConn.schema + "/push/all"
	if resp, err = gateConn.client.PostForm(apiUrl, form); err != nil {
		return
	}
	defer resp.Body.Close()
	return
}

// 出于性能考虑, 消息数组在此前已经编码成json
func (gateConn *GateConn) PushRoom(room string, itemsJson []byte) (err error) {
	var (
		apiUrl string
		form url.Values
		resp *http.Response
	)

	form = url.Values{}
	form.Set("room", room)
	form.Set("items", string(itemsJson))

	apiUrl = gateConn.schema + "/push/room"
	if resp, err = gateConn.client.PostForm(apiUrl, form); err != nil {
		return
	}
	defer resp.Body.Close()
	return
}