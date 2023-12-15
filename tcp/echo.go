// Package tcp -----------------------------
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
	// 用自己包装的 WaitGroup
	Waiting wait.Wait
}

func (e *EchoClient) Close() error {
	// 关闭客户端前有个等待超时的时间
	e.Waiting.WaitWithTimeout(10 * time.Second)
	_ = e.Conn.Close()
	return nil
}

type EchoHandler struct {
	activeConn sync.Map
	// × closing boll 防止并发问题
	closing atomic.Boolean
}

func MakeHandler() *EchoHandler {
	return &EchoHandler{}
}

func (handler *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	// 如果客户端是正在关闭中的
	if handler.closing.Get() {
		_ = conn.Close()
	}
	// 包装为自己定义的
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
				logger.Info("Connecting close")
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
		client.Waiting.Done()
	}
}

func (handler *EchoHandler) Close() error {
	logger.Info("handler shutting down")
	handler.closing.Set(true)
	handler.activeConn.Range(func(key, value interface{}) bool {
		// 空接口转换为 EchoClient
		client := key.(*EchoClient)
		_ = client.Conn.Close()
		// 表示方法施加到所有 key value
		return true
	})
	return nil
}
