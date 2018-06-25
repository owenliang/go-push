# go-push

用GO做推送

# Dependency

* golang.org/x/net/http2 (注: GFW已墙, 请到海外服务器下载)
* github.com/gorilla/websocket

# Arch

* gateway: 长连接网关
    * 海量长连接按BUCKET打散, 减小推送遍历的锁粒度
    * 按广播/房间粒度的消息前置合并, 减少编码CPU损耗, 减少系统网络调用, 巨幅提升吞吐
* logic: 逻辑服务器
    * 本身无状态, 负责将推送消息分发到所有gateway节点
    * 对调用方暴露HTTP/1接口, 方便业务对接
    * 采用HTTP/2长连接RPC向gateway分发消息

# May be a problem

* 推送主要瓶颈是gateway层而不是内部通讯, 所以gateway和logic之间仍旧采用了小包通讯(对网卡有PPS压力), 同时logic为业务提供了批量推送接口来缓解特殊需求.

# Benchmark

## environment

* 16 vcore
* client, logic, gateway deployed together

## Bandwidth

![bandwidth](https://github.com/owenliang/go-push/blob/master/bandwidth.png?raw=true)

## Cpu Usage

![cpu usage](https://github.com/owenliang/go-push/blob/master/cpu.png?raw=true)

# API

## 全员广播

```
curl http://localhost:7799/push/all -d 'items=[{"msg": "hi"},{"msg": "bye"}]'
```

## 房间广播

```
curl http://localhost:7799/push/room -d 'room=default&items=[{"msg": "hi"},{"msg": "bye"}]'
```

## Protocol

* PING(upstream)

```
{"type": "PING", "data": {}}
```

* PONG(downstream)

```
{"type": "PONG", "data": {}}
```

* PUSH(downstream)

```
{"type": "PUSH", "data": {"items": [{"name": "go-push"}, {"age": "1"}]}}
```