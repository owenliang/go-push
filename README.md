# go-push

用GO做推送

# arch

![arch](https://github.com/owenliang/go-push/blob/master/gateway/GO%20push.png?raw=true)

# millstones

* 一期
  * gateway
    * websocket长连网关
    * 支持PING,PONG健康心跳
    * 支持JOIN,LEAVE房间, 无房间合法性校验, 暂时限制1个连接最多同时在线N个ROOM
    * 支持PUSH推送, 支持牺牲一定延迟, 对窗口内消息做合并打包发送, 换取更高吞吐
    * 基于HTTP实现推送提交接口
    * 暂不引入uid用户概念
* 二期
  * gateway: 
    * 基于HTTP/2实现内部服务间通讯, 原有HTTP推送接口迁移到HTTP/2
  * logic:
    * 基于HTTP实现推送提交接口，即取代原先gateway的职责
    * 基于HTTP/2实现推送消息向gateway的分发
* 三期
  * passport:
    * 基于HTTP协议提供登录服务
    * 基于用户账户系统, 提供帐号密码登录换取JWT
  * gateway:
    * 增加AUTH命令，原地校验JWT完成登录
* 四期
  * logic:
    * 接收gateway发来的会话心跳，向redis cluster存储uid所在的gateway信息
    * 支持向uid定向推送，从redis cluter找到gateway信息，利用HTTP/2直连推送
    
    
    
