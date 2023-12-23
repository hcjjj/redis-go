# 建设中 ... ⛏
## 实现 TCP 服务器
基于net包开发TCP服务器，可以同时处理多个客户端的连接、通信、正常/异常结束

## 实现 Redis 协议解析器
Redis 网络协议，**REdis SeriaIization ProtocoI (RESP)** ：

* 正常回复（Redis → Client）
    * 以 "+" 开头，以 "\r\n" 结尾的字符串形式
    * 如：`+OK\r\n`
* 错误回复（Redis → Client）
    * 以 "-" 开头，以 "\r\n" 结尾的字符串形式
    * 如：`-Error message\r\n`
* 整数（Redis ⇄ Client）
    * 以 ":" 开头，以 "\r\n" 结尾的字符串形式
    * 如：`:123456\r\n`
* 多行字符串（Redis ⇄ Client）
    * 以 "$" 开头，后跟实际发送字节数，以 "\r\n" 结尾
    * "Redis"：`$5\r\nRedis\r\n`
    *  ""：`$0\r\n\r\n`
    * "Redis\r\ngo"：`$11\r\nRedis\r\ngo\r\n`
* 数组（Redis ⇄ Client）
    * 以 "*" 开头，后跟成员个数
    * 有3个成员的数组 [SET, key, value]：`*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n` 
 

## 实现内存数据库

## 实现 Redis 持久化

## 实现 Redis 集群