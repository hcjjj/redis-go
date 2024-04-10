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
	// 记录协议层保持连接的用户信息
	activeConn sync.Map
	db         databseinterface.Database
	// 并发安全的 bool
	closing atomic.Boolean
}

func MakeHandler() *RespHandler {
	var db databseinterface.Database
	// 测试解析结果，直接反回解析结果给用户
	//db = database.NewEchoDatabase()
	// 单机版的
	db = database.NewStandaloneDatabase()
	// 判断是否启动集群版
	//if config.Properties.Self != "" && len(config.Properties.Peers) > 0 {
	//	db = cluster.MakeClusterDatabase()
	//} else {
	//	db = database.NewStandaloneDatabase()
	//}
	return &RespHandler{
		db: db,
	}
}

// 关闭一个客户端的连接
func (r *RespHandler) closeClient(client *connection.Connection) {
	// 关闭客户端
	_ = client.Close()
	// 客户端关闭后数据库需要做的一些善后操作
	r.db.AfterClientClose(client)
	// 移除服务的客户端信息
	r.activeConn.Delete(client)
}

// Handle 处理 TCP 连接
func (r *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if r.closing.Get() {
		_ = conn.Close()
	}
	// TCP 的 连接包装为 协议层的连接
	client := connection.NewConn(conn)
	// k 是 client  v 是空接口体  map → set
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
			// 将协议错误回写给客户端
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
		// 感觉这边需要判断 data 来反应直接回复客户端还是 需要和 db 打交道
		// 这边需要扩展~ PING 还不支持
		// 和 db 打交到
		switch payload.Data.(type) {
		case *reply.MultiBulkReply:
			result := r.db.Exec(client, payload.Data.(*reply.MultiBulkReply).Args)
			if result != nil {
				// 返回执行结果给客户端
				// ToBytes 结果再编码为 RESP 格式
				_ = client.Write(result.ToBytes())
			} else {
				_ = client.Write(unknownErrReplyBytes)
			}
		case *reply.BulkReply:
			cmd := payload.Data.(*reply.BulkReply).Arg
			if strings.ToLower(string(cmd)) == "ping" {
				args := make([][]byte, 1)
				args[0] = cmd
				result := r.db.Exec(client, args)
				if result != nil {
					// 返回执行结果给客户端
					// ToBytes 结果再编码为 RESP 格式
					_ = client.Write(result.ToBytes())
				} else {
					_ = client.Write(unknownErrReplyBytes)
				}
			}
		default:
			logger.Error("require multi bulk reply to exec")
			continue
		}

		//mReply, ok := payload.Data.(*reply.MultiBulkReply)
		//if !ok {
		//	logger.Error("require multi bulk reply to exec")
		//	continue
		//}
		//result := r.db.Exec(client, mReply.Args)
		//if result != nil {
		//	// 返回执行结果给客户端
		//	// ToBytes 结果再编码为 RESP 格式
		//	_ = client.Write(result.ToBytes())
		//} else {
		//	_ = client.Write(unknownErrReplyBytes)
		//}
	}
}

// Close 关闭整个 handler
func (r *RespHandler) Close() error {
	logger.Info("handler shutting down ...")
	r.closing.Set(true)
	// 逐步断开每个客户端的连接
	r.activeConn.Range(
		func(key interface{}, value interface{}) bool {
			client := key.(*connection.Connection)
			_ = client.Close()
			return true
		})
	r.db.Close()
	return nil
}
