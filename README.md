# go-push

用GO做推送

# arch

* gateway: 长连接网关
    * 海量长连接按BUCKET打散, 减小推送遍历的锁粒度
    * 按广播/房间粒度的消息前置合并, 减少编码CPU损耗, 减少系统网络调用, 巨幅提升吞吐
* logic: 逻辑服务器
    * 本身无状态, 负责将推送消息分发到所有gateway节点
    * 对调用方暴露HTTP/1接口, 方便业务对接
    * 采用HTTP/2长连接RPC向gateway分发消息

# problem

* 推送主要瓶颈是gateway层而不是内部通讯, 所以gateway和logic之间仍旧采用了小包通讯(对网卡有PPS压力), 同时logic为业务提供了批量推送接口来缓解特殊需求.

# benchmark

16 vcore

## bandwidth

![bandwidth]()

## cpu usage

![cpu usage]()