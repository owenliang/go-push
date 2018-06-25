package main

import (
	"net/http"
	"fmt"
	"os"
	"io/ioutil"
	"crypto/tls"
)

func main() {
	var (
		resp *http.Response
		err error
		buf []byte
		client *http.Client
	)

	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true,},	// 不校验服务端证书
		},
	}

	if resp, err = client.Get("https://localhost:7788/stats"); err != nil {
		goto ERR
	}

	defer resp.Body.Close()

	if buf, err = ioutil.ReadAll(resp.Body); err != nil {
		goto ERR
	}

	fmt.Println("返回值:", string(buf))
	return

ERR:
	fmt.Println(err)
	os.Exit(-1)
	return
}