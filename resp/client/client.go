// Package client -----------------------------
// @file      : client.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/17 14:10
// -------------------------------------------
package client

import (
	"errors"
	"net"
	"redis-go/interface/resp"
	"redis-go/lib/logger"
	"redis-go/lib/sync/wait"
	"redis-go/resp/parser"
	"redis-go/resp/reply"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// Client 客户端的核心，它包含了管理请求、连接和状态的主要字段
type Client struct {
	conn        net.Conn
	pendingReqs chan *request // wait to send
	waitingReqs chan *request // waiting response
	ticker      *time.Ticker
	addr        string
	working     *sync.WaitGroup // its counter presents unfinished requests(pending and waiting)
}

// request is a message sends to redis server
type request struct {
	// 引入请求 id 并使用 map 进行请求-响应匹配会是一个更稳健的设计
	// 暂时没用到 id
	id        uint64
	args      [][]byte
	reply     resp.Reply
	heartbeat bool
	waiting   *wait.Wait
	err       error
}

const (
	chanSize = 256
	maxWait  = 3 * time.Second
)

// MakeClient creates a new client
func MakeClient(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Client{
		addr:        addr,
		conn:        conn,
		pendingReqs: make(chan *request, chanSize),
		waitingReqs: make(chan *request, chanSize),
		working:     &sync.WaitGroup{},
	}, nil
}

// Start starts asynchronous goroutines
func (client *Client) Start() {
	client.ticker = time.NewTicker(10 * time.Second)
	// 用于将请求从 pendingReqs 中取出并发送给服务器
	go client.handleWrite()
	// 从服务器读取响应并将其与相应的请求匹配
	go client.handleRead()
	// 每隔一段时间发送心跳包，确保连接的活跃状态
	go client.heartbeat()
}

// Close stops asynchronous goroutines and close connection
func (client *Client) Close() {
	client.ticker.Stop()
	// stop new request
	close(client.pendingReqs)

	// wait stop process
	client.working.Wait()

	// 关闭与服务端的连接，连接关闭后读协程会退出
	_ = client.conn.Close()
	// 关闭队列
	close(client.waitingReqs)
}

// 用于在连接断开时重新连接到 Redis 服务器。它会进行最多三次的重试。如果重试失败，则关闭客户端
func (client *Client) reconnect() {
	logger.Info("reconnect with: " + client.addr)
	_ = client.conn.Close() // ignore possible errors from repeated closes

	var conn net.Conn
	for i := 0; i < 3; i++ {
		var err error
		conn, err = net.Dial("tcp", client.addr)
		if err != nil {
			logger.Error("reconnect error: " + err.Error())
			time.Sleep(time.Second)
			continue
		} else {
			break
		}
	}
	if conn == nil { // reach max retry, abort
		client.Close()
		return
	}
	client.conn = conn

	close(client.waitingReqs)
	for req := range client.waitingReqs {
		req.err = errors.New("connection closed")
		req.waiting.Done()
	}
	client.waitingReqs = make(chan *request, chanSize)
	// restart handle read
	go client.handleRead()
}

func (client *Client) heartbeat() {
	for range client.ticker.C {
		client.doHeartbeat()
	}
}

// 写协程入口
func (client *Client) handleWrite() {
	// 从 pendingReqs 通道中取出请求并将其发送给服务器
	for req := range client.pendingReqs {
		client.doRequest(req)
	}
}

// Send 用于发送请求并等待响应
func (client *Client) Send(args [][]byte) resp.Reply {
	request := &request{
		args:      args,
		heartbeat: false,
		waiting:   &wait.Wait{},
	}
	request.waiting.Add(1)
	client.working.Add(1)
	defer client.working.Done()
	// 请求入队
	client.pendingReqs <- request
	// 等待响应或者超时
	timeout := request.waiting.WaitWithTimeout(maxWait)
	// 对应着 finishRequest 那边的 request.waiting.Done()
	if timeout {
		return reply.MakeErrReply("server time out")
	}
	if request.err != nil {
		return reply.MakeErrReply("request failed " + request.err.Error())
	}
	return request.reply
}

// 定时向 Redis 服务器发送 PING 请求，确保连接的稳定性
func (client *Client) doHeartbeat() {
	request := &request{
		args:      [][]byte{[]byte("PING")},
		heartbeat: true,
		waiting:   &wait.Wait{},
	}
	request.waiting.Add(1)
	client.working.Add(1)
	defer client.working.Done()
	client.pendingReqs <- request
	request.waiting.WaitWithTimeout(maxWait)
}

// 发送请求
func (client *Client) doRequest(req *request) {
	if req == nil || len(req.args) == 0 {
		return
	}
	// 序列化请求
	re := reply.MakeMultiBulkReply(req.args)
	bytes := re.ToBytes()
	var err error
	// 失败重试
	for i := 0; i < 3; i++ { // only retry, waiting for handleRead
		_, err = client.conn.Write(bytes)
		if err == nil ||
			(!strings.Contains(err.Error(), "timeout") && // only retry timeout
				!strings.Contains(err.Error(), "deadline exceeded")) {
			break
		}
	}
	if err == nil {
		// 发送成功等待服务器响应
		client.waitingReqs <- req
	} else {
		req.err = err
		req.waiting.Done()
	}
}

// 读协程是个 RESP 协议解析器
func (client *Client) handleRead() {
	// 持续监听服务器的响应
	ch := parser.ParseStream(client.conn)
	for payload := range ch {
		if payload.Err != nil {
			client.reconnect()
			return
		}
		// 匹配请求并完成
		client.finishRequest(payload.Data)
	}
}

// 将该响应与之前发送的请求进行匹配
func (client *Client) finishRequest(reply resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			logger.Error(err)
		}
	}()
	request := <-client.waitingReqs
	if request == nil {
		return
	}
	request.reply = reply
	//  解除阻塞，表示该请求已处理完成
	if request.waiting != nil {
		request.waiting.Done()
	}
}
