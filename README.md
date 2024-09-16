## 项目介绍

使用 Golang 基于 RESP （Redis serialization protocol） 实现的简易 Redis，主要包括 TCP 服务器、协议解析器、内存数据库、持久化和切集群

**编译运行：**

```shell
# redis.conf 设置服务器信息、数据库核心数、aof 持久化相关、集群相关
# 配置 peer 信息既开启集群模式，每个节点需要分别设置好各自配置文件的 self 和 peers
go build && ./redis-go
# 客户端： redis-cli/telnet/网络调试助手（开启转义符指令解析）
redis-cli -h 127.0.0.1 -p 6379
# telnet 127.0.0.1 6379
```

> **[Redis 知识体系整理](https://hcjjj.github.io/2024/05/22/redis/)**

## 实现逻辑

**TCP 服务器：**

![](https://cdn.jsdelivr.net/gh/hcjjj/blog-img/tcp.svg)

**协议解析器：**

![](https://cdn.jsdelivr.net/gh/hcjjj/blog-img/resp4.svg)

**内存数据库：**

![](https://cdn.jsdelivr.net/gh/hcjjj/blog-img/db.svg)

**AOF 持久化：**

![](https://cdn.jsdelivr.net/gh/hcjjj/blog-img/AOF.svg)

**集群架构：**

![](https://cdn.jsdelivr.net/gh/hcjjj/blog-img/cluster.svg)

**集群指令执行流程：**

![](https://cdn.jsdelivr.net/gh/hcjjj/blog-img/cluster0.svg)

## 要点概览

* **TCP 服务器**
  * TCP 服务器：应用服务器的基础，包含服务的监听、处理每个连接的逻辑等
  * 优雅关闭：保证服务关闭前完成必要的清理工作，如完成正在进行的数据传输，关闭 TCP 连接等
  * 拆包与粘包：采用 RESP 从 TCP 提供的字节流中正确地解析出应用层消息
* **协议解析器**
  * RESP 规范：RESP 是一个二进制安全的文本协议，工作于 TCP 协议上
  * 协议解析器：按照 RESP 规范解析 Socket 数据，基于 TCP 服务器搭建应用服务器
* **内存数据库**
  * 底层采用 `sync.map`，官方提供的并发安全哈希表, 适合读多写少的场景
  * 构建指令名称及其对应执行方法的映射表 cmdTable

* **AOF 持久化**
  * 
* **分布式集群**

## 目录结构

```shell
├── aof # AOF 持久化
├── cluster # 集群层
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
├── resp # RESP 解析
│   ├── client # 客户端
│   ├── connection
│   ├── handler
│   ├── parser # 解析客户端发来的数据
│   └── reply # 封装服务器对客户端的回复
└── tcp # TCP 服务器
```

## RESP

Redis 序列化协议规范，**[Redis serialization protocol specification](https://redis.io/docs/reference/protocol-spec/)**

RESP 是一个二进制安全的文本协议，以行作为单位，客户端和服务器发送的命令或数据一律以 `\r\n`（CRLF）作为换行符，RESP 的二进制安全性允许在 key 或者 value 中包含 `\r` 或者 `\n` 这样的特殊字符。

> 二进制安全是指允许协议中出现任意字符而不会导致故障

* 正确回复（Redis → Client）
  * 以 **`+`** 开头，以 "\r\n" 结尾的字符串形式
  * 如：`+OK\r\n`
* 错误回复（Redis → Client）
  * 以 **`-`** 开头，以 "\r\n" 结尾的字符串形式
  * 如：`-Error message\r\n`
* 整数（Redis ⇄ Client）
  * 以 **`:`** 开头，以 "\r\n" 结尾的字符串形式
  * 如：`:123456\r\n`
* 单行字符串（Redis ⇄ Client）
  * 以 **`$`** 开头，后跟实际发送字节数，以 "\r\n " 结尾
  * "Redis"：`$5\r\nRedis\r\n`
  * ""：`$0\r\n\r\n`
  * "Redis\r\ngo"：`$11\r\nRedis\r\ngo\r\n`
  * nil：`$-1`
* 多行字符串（数组）（Redis ⇄ Client）
  * 以 **`*`** 开头，后跟成员个数
  * 有 3 个成员的数组 [SET, key, value]：`*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n`

## 支持命令

* PING
* SELECT
* Key 命令集
  * DEL
  * EXISTS
  * FlushDB
  * TYPE
  * RENAME
  * RENAMENX
  * KEYS
* String 命令集
  * GET
  * SET
  * SETNX
  * GETSET
  * STRLEN
* ...

![](https://cdn.jsdelivr.net/gh/hcjjj/blog-img/20240411200044.png)

**测试命令**

* ping `$4\r\nping\r\n`
* set key value `*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n`
* set ke1 value `*3\r\n$3\r\nSET\r\n$3\r\nke1\r\n$5\r\nvalue\r\n`
* select 1 `*2\r\n$6\r\nselect\r\n$1\r\n1\r\n`
* get key `*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n`
* select 2 `*2\r\n$6\r\nselect\r\n$1\r\n1\r\n`

> telnet 需要逐条发送如 $4↩︎ping↩︎

## 性能测试

```shell
OS:  Windows 11 💻 / Ubuntu 22.04.4 LTS on Windows 10 x86_64 🐧
CPU: AMD Ryzen 7 6800H with Radeon Graphics (16) @ 3.200GHz
Memory: 61159MiB
```

**redis-go**

```shell
❯ redis-benchmark -h 127.0.0.1 -p 6379 -t set,get -n 10000 -q
ERROR: ERR unknown command config
ERROR: failed to fetch CONFIG from 127.0.0.1:6379
WARN: could not fetch server CONFIG
====== SET ======
  10000 requests completed in 0.17 seconds
  50 parallel clients
  3 bytes payload
  keep alive: 1
  multi-thread: no
0.01% <= 0.1 milliseconds
...
SET: 58823.53 requests per second
...
GET: 62500.00 requests per second
```

**redis**

```shell
❯ redis-benchmark -h 127.0.0.1 -p 6379 -t set,get -n 10000
====== SET ======
  10000 requests completed in 0.15 seconds
  50 parallel clients
  3 bytes payload
  keep alive: 1
  host configuration "save": 900 1 300 10 60 10000
  host configuration "appendonly": no
  multi-thread: no
0.01% <= 0.1 milliseconds
...
66666.66 requests per second
...
71942.45 requests per second
```
