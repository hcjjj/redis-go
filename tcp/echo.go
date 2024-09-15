// Package tcp -------------------------------
// @file      : echo.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2023/12/15 19:44
// -------------------------------------------
package tcp

import (
	"bufio"
	"context"
	"io"
	"net"
	"redis-go/lib/logger"
	"redis-go/lib/sync/atomic"
	"redis-go/lib/sync/wait"
	"sync"
	"time"
)

type EchoClient struct {
	Conn net.Conn
	// 用自己包装的 WaitGroup 加入超时的功能
	Waiting wait.Wait
}

func (e *EchoClient) Close() error {
	// 关闭客户端前有个等待超时的时间
	e.Waiting.WaitWithTimeout(10 * time.Second)
	_ = e.Conn.Close()
	return nil
}

type EchoHandler struct {
	// 记录有多少个连接
	activeConn sync.Map
	// × closing boll 防止并发问题
	// 原子的 bool
	closing atomic.Boolean
}

func MakeEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (handler *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	// 如果客户端是正在关闭中的
	if handler.closing.Get() {
		_ = conn.Close()
	}
	// 包装为自己定义的，代表一个连接上来的客户端
	client := &EchoClient{
		Conn: conn,
	}
	// 记录所有连接的客户端，只需要 key 不需要 val
	handler.activeConn.Store(client, struct{}{})
	// 缓冲区接收用户发来的数据
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("Connecting close: " + client.Conn.RemoteAddr().String())
				handler.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		// 正在进行业务 不要关闭 除非是超时了
		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		// 业务结束
		client.Waiting.Done()
	}
}

func (handler *EchoHandler) Close() error {
	logger.Info("EchoHandler shutting down ...")
	handler.closing.Set(true)
	// sync 包下 map 的遍历方式
	handler.activeConn.Range(func(key, value interface{}) bool {
		// 将空接口转换为 EchoClient
		client := key.(*EchoClient)
		// 对每个客户端都进行关闭连接的操作
		_ = client.Conn.Close()
		// 表示方法施加到所有 key value
		return true
	})
	return nil
}
