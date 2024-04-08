# 建设中 ... ⛏
## 项目介绍

使用 Go 语言基于 Redis serialization protocol (RESP) 实现简易的 Redis，主要工作如下：

* TCP 服务器
* Redis 协议解析器
* 内存数据库
* Redis 持久化
* Redis 集群

## 运行方式

**单机版：**

```shell
# 服务器
go build main.go
./main
# 客户端 telnet
telnet 127.0.0.1 6379
# redis-cli
redis-cli -h 127.0.0.1
```

**集群版：**

```shell

```

## 测试命令

* set key value `*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n`
* select 2 `*2\r\n$6\r\nselect\r\n$1\r\n1\r\n`
* get key `*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n`

## 实现逻辑

### 目录结构

```shell
├── aof # AOF 持久化
├── cluster # 集群
├── config # 解析配置文件 redis.conf
├── database # 内存数据库
├── datastruct # 支持的数据结构
│   └── dict
├── interface # 接口定义
│   ├── database
│   ├── resp
│   └── tcp
├── lib # 基础工具
│   ├── consistenthash # 一致性哈希
│   ├── logger # 日志记录
│   ├── sync # 同步工具
│   │   ├── atomic
│   │   └── wait
│   ├── utils # 格式转换
│   └── wildcard # 通配符
├── resp # RESP 协议
│   ├── client
│   ├── connection
│   ├── handler
│   ├── parser
│   └── reply
└── tcp # TCP 服务器
```

### TCP 服务器

基于 net 包开发 TCP 服务器，支持同时处理多个客户端的连接、业务、正常和异常结束等

![](https://cdn.jsdelivr.net/gh/hcjjj/blog-img/TCP.png)

### Redis 协议解析器

Redis 网络协议，**REdis SeriaIization ProtocoI (RESP)** ：
* 正常回复（Redis ⇄ Client）
  * 以 "+" 开头，以 "\r\n" 结尾的字符串形式
  * 如：`+OK\r\n`
* 错误回复（Redis ⇄ Client）
  * 以 "-" 开头，以 "\r\n" 结尾的字符串形式
  * 如：`-Error message\r\n`
* 整数（Redis ⇄ Client）
  * 以 ":" 开头，以 "\r\n" 结尾的字符串形式
  * 如：`:123456\r\n`
* 单行字符串（Redis ⇄ Client）
  * 以 "$" 开头，后跟实际发送字节数，以 "\r\n" 结尾
  * "Redis"：`$5\r\nRedis\r\n`
  * ""：`$0\r\n\r\n`
  * "Redis\r\ngo"：`$11\r\nRedis\r\ngo\r\n`
* 数组（Redis ⇄ Client）
  * 以 "*" 开头，后跟成员个数
  * 有3个成员的数组[SET, key, value]：`*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n`

### 内存数据库

KV 内存数据库的核心是并发安全的哈希表 **sync.map**

### Redis 持久化

AOF 持久化是典型的异步任务，主协程 (goroutine) 可以使用 channel 将数据发送到异步协程由异步协程执行持久化操作

### Redis 集群

单台服务器的CPU和内存等资源是有限的，利用多台机器建立分布式系统，分工处理来提高系统容量和吞吐量
