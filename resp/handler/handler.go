// Package handler -----------------------------
// @file      : handler.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/3 11:20
// -------------------------------------------
package handler

import (
	"context"
	"io"
	"net"
	"redis-go/database"
	databseinterface "redis-go/interface/database"
	"redis-go/lib/logger"
	"redis-go/lib/sync/atomic"
	"redis-go/resp/connection"
	"redis-go/resp/parser"
	"redis-go/resp/reply"
	"strings"
	"sync"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

type RespHandler struct {
	activeConn sync.Map
	db         databseinterface.Database
	closing    atomic.Boolean
}

func MakeHandler() *RespHandler {
	var db databseinterface.Database
	db = database.NewEchoDatabase()
	return &RespHandler{
		db: db,
	}
}

// 关闭一个客户端的连接
func (r *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	r.db.AfterClientClose(client)
	r.activeConn.Delete(client)
}

// Handle 处理 TCP 连接
func (r *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if r.closing.Get() {
		_ = conn.Close()
	}
	client := connection.NewConn(conn)
	r.activeConn.Store(client, struct{}{})
	ch := parser.ParseStream(conn)
	// 监听管道
	for payload := range ch {
		// 异常逻辑
		if payload.Err != nil {
			// 客户端关闭
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				r.closeClient(client)
				logger.Info("Connection closed: " + client.RemoteAddr().String())
				return
			}
			// protocol error
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				r.closeClient(client)
				logger.Info("Connection closed: " + client.RemoteAddr().String())
				return
			}
			continue
		}
		// 正常解析逻辑
		if payload.Data == nil {
			continue
		}
		mReply, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		result := r.db.Exec(client, mReply.Args)
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}
	}
}

// Close 关闭整个 handler
func (r *RespHandler) Close() error {
	logger.Info("handler shutting down")
	r.closing.Set(true)
	r.activeConn.Range(
		func(key interface{}, value interface{}) bool {
			client := key.(*connection.Connection)
			_ = client.Close()
			return true
		})
	r.db.Close()
	return nil
}
