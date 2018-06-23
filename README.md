# go-push

用GO做推送

# arch

* BUCKET: 海量长连接按BUCKET打散, 减小推送遍历的锁粒度
* MERGE: 按广播/房间粒度的消息前置合并, 减少系统调用, 巨幅提升吞吐