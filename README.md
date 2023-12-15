# 建设中 ... ⛏

## 实现 TCP 服务器
**实现逻辑：**

* 主要是基于 net 包开发支持 TCP 连接服务器，可以同时处理多个客户端的连接、业务、正常和异常结束等

**Tips：**

* `ctrl + i ` 快捷实现的接口

**问题一：** 当客户端主动断开连接的时候服务器报错

`panic: sync: negative WaitGroup counter`

`waitDone.Add(1)` 不小心写成 `waitDone.Add(0)`，导致后续的 `waitDone.Done()` 出现 panic

## Redis 背景知识

## 实现 Redis 协议解析器

## 实现内存数据库

## 实现 Redis 持久化

## 实现 Redis 集群