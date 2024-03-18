# 建设中 ... ⛏
## 实现 TCP 服务器
基于 net 包开发 TCP 服务器，支持同时处理多个客户端的连接、业务、正常和异常结束等

## 实现 Redis 协议解析器

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

## 实现内存数据库

KV 内存数据库的核心是并发安全的哈希表 **sync.map**

## 实现 Redis 持久化

AOF 持久化是典型的异步任务，主协程 (goroutine) 可以使用 channel 将数据发送到异步协程由异步协程执行持久化操作

## 实现 Redis 集群

单台服务器的CPU和内存等资源是有限的，利用多台机器建立分布式系统，分工处理来提高系统容量和吞吐量

## 测试命令

* set key value `*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n`
* select 2 `*2\r\n$6\r\nselect\r\n$1\r\n1\r\n`
* get key `*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n`

## 问题记录

* 问题一
  * 当客户端主动断开连接的时候服务器报错，`panic: sync: negative WaitGroup counter`
  * `waitDone.Add(1)` 不小心写成 `waitDone.Add(0)`，导致后续的 `waitDone.Done()` 出现 panic
* 问题二
  * `imports redis-go/database: import cycle not allowed`
  * aof.go 文件导包错误，需要的是 "redis-go/interface/database"，而不是 "redis-go/database"
* 问题三
  * `[ERROR][database.go:76] runtime error: index out of range [1] with length 1`
  * execSelect 方法中的 `strconv.Atoi(string(args[0]))` 写成了 1

## 代码文件总览

* 项目基础配置
  * config
  * lib/logger 日志记录
  * lib/sync 同步工具
  * lib/wildcard 通配符
  * lib/utils 格式转换工具
* 实现 TCP 服务器
  * interface/tcp
  * tcp/server.go 管理对多个客户端的连接
  * tcp/echo.go 回发的TCP层测试
* 实现 Redis 协议解析器
  * interface/resp
  * resp/reply 定义服务端对客户端静态/动态回复
  * resp/parser 对客户端发来的字节数据进行解析
  * resp/connection 处理客户端的请求（发送数据）
  * resp/handler 处理客户端的请求（解析数据为指令）
  * database/echo_database.go 回发的内核层测试
* 实现内存数据库
  * datastruct/database.go 数据库核心，定义 Redis 底层数据结构的接口与实现
  * datastruct/sync_dict 对底层并发map的包装，方便更换实现
  * database/db.go 定义分数据库、底层执行逻辑
  * database/command.go 注册命令的方法
  * database/keys.go ping.go string.go ... 实现相关命令
* 实现 Redis 持久化
  * Append Only File
  * aof/aof.go 实现指令落盘和加载恢复数据
* 实现 Redis 集群
  * 一致性哈希（减少传统哈希增加节点时数据哈希存储不一致调整的开销）
  * lib/consistenthash 一致性哈希
  * cluster/cluster_database.go 集群
  * resp/client 客户端
  * go-commos-pool 开源连接池工具
  * cluster/client_pool.go 连接工厂，给连接池用的
  * cluster/com.go 节点之间的通信
  * cluster/router.go 指令路由
  * `go build` 生成 exe
  * `go mod tidy`  整理依赖