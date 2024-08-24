## 项目介绍

使用 Go 语言基于 Redis serialization protocol (RESP) 实现简易的 Redis，主要工作：TCP 服务器、协议解析器、内存数据库、持久化、集群

**编译运行：**

```shell
# redis.conf 设置服务器信息、数据库核心数、aof 持久化相关、集群相关
# 配置 peer 信息既开启集群模式，每个节点需要分别设置好各自配置文件的 self 和 peers
go build && ./redis-go
# 客户端 telnet 或者 网络调试助手，开启转义符指令解析
# https://www.cmsoft.cn/assistcenter/help/assistscript/
telnet 127.0.0.1 6379
```

## 实现逻辑

**TCP 服务器：**

main → ListenAndServeWithSignal → ListenAndServer🔁 → Handle🔁

**协议解析器：**

![](https://cdn.jsdelivr.net/gh/hcjjj/blog-img/resp.svg)

**内存数据库：**

![](https://cdn.jsdelivr.net/gh/hcjjj/blog-img/db.svg)

**持久化流程：**

![](https://cdn.jsdelivr.net/gh/hcjjj/blog-img/AOF.svg)

**集群架构：**

![](https://cdn.jsdelivr.net/gh/hcjjj/blog-img/cluster.svg)

**集群指令执行流程：**

![](https://cdn.jsdelivr.net/gh/hcjjj/blog-img/cluster0.svg)

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

## RESP 协议

Redis序列化协议规范，**[Redis serialization protocol specification](https://redis.io/docs/reference/protocol-spec/)**

* 正常回复（Redis → Client）
  * 以 "+" 开头，以 "\r\n" 结尾的字符串形式
  * 如：`+OK\r\n`
* 错误回复（Redis → Client）
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
* 多行字符串（数组）（Redis ⇄ Client）
  * 以 "*" 开头，后跟成员个数
  * 有3个成员的数组[SET, key, value]：`*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n`

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

**测试命令：**

* ping `$4\r\nping\r\n`
* set key value `*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n`
* set ke1 value `*3\r\n$3\r\nSET\r\n$3\r\nke1\r\n$5\r\nvalue\r\n`
* select 1 `*2\r\n$6\r\nselect\r\n$1\r\n1\r\n`
* get key `*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n`
* select 2 `*2\r\n$6\r\nselect\r\n$1\r\n1\r\n`

> telnet 需要逐条发送如 $4↩︎ping↩︎

## [相关文档](https://github.com/hcjjj/redis-go/blob/master/docs/redis.md)
