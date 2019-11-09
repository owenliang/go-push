# go-push

golang实现的、可扩展的通用消息推送原型。

# 安装

已升级到golang1.13，基于gomod管理依赖。

* 下载go-push

```
go get github.com/owenliang/go-push
```

* 安装依赖

```
export GOPROXY=goproxy.io
go mod download
```

* 编译gateway服务

```
cd gateway/cli && go build && cd -
```

* 编译logic服务

```
cd logic/cli && go build && cd -
```

# 架构

* gateway: 长连接网关
    * 海量长连接按BUCKET打散, 减小推送遍历的锁粒度
    * 按广播/房间粒度的消息前置合并, 减少编码CPU损耗, 减少系统网络调用, 巨幅提升吞吐
* logic: 逻辑服务器
    * 本身无状态, 负责将推送消息分发到所有gateway节点
    * 对调用方暴露HTTP/1接口, 方便业务对接
    * 采用HTTP/2长连接RPC向gateway集群分发消息

# 潜在问题

* 推送主要瓶颈是gateway层而不是内部通讯, 所以gateway和logic之间仍旧采用了小包通讯(对网卡有PPS压力), 同时logic为业务提供了批量推送接口来缓解特殊需求.

# 压测

## 环境

* 16 vcore
* client, logic, gateway deployed together

## 带宽

![bandwidth](https://github.com/owenliang/go-push/blob/master/bandwidth.png?raw=true)

## CPU占用

![cpu usage](https://github.com/owenliang/go-push/blob/master/cpu.png?raw=true)

# logic的推送API

* 全员广播

```
curl http://localhost:7799/push/all -d 'items=[{"msg": "hi"},{"msg": "bye"}]'
```

* 房间广播

```
curl http://localhost:7799/push/room -d 'room=default&items=[{"msg": "hi"},{"msg": "bye"}]'
```

## gateway的websocekt协议

* PING(客户端->服务端)

```
{"type": "PING", "data": {}}
```

* PONG(服务端->客户端)

```
{"type": "PONG", "data": {}}
```

* JOIN(客户端->服务端)

```
{"type": "JOIN", "data": {"room": "fengtimo"}}
```

* LEAVE(客户端->服务端)

```
{"type": "LEAVE", "data": {"room": "fengtimo"}}
```

* PUSH(服务端->客户端)

```
{"type": "PUSH", "data": {"items": [{"name": "go-push"}, {"age": "1"}]}}
```

# 加入社群

我的博客：

[鱼儿的博客](https://yuerblog.cc)

Go微信群：

![wechat](https://github.com/owenliang/go-push/blob/master/E0B82999-82BB-4963-BC2D-FDEAF42230FC.jpeg?raw=true)

知识星球（独家知识）：
![xingqiu](https://github.com/owenliang/go-push/blob/master/xingqiu.png?raw=true)