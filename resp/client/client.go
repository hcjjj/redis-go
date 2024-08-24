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

// Client is a pipeline mode redis client
type Client struct {
	conn        net.Conn
	pendingReqs chan *request // wait to send
	waitingReqs chan *request // waiting response
	ticker      *time.Ticker
	addr        string

	working *sync.WaitGroup // its counter presents unfinished requests(pending and waiting)
}

// request is a message sends to redis server
type request struct {
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
	go client.handleWrite()
	go client.handleRead()
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
	for req := range client.pendingReqs {
		client.doRequest(req)
	}
}

// Send sends a request to redis server
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
	if timeout {
		return reply.MakeErrReply("server time out")
	}
	if request.err != nil {
		return reply.MakeErrReply("request failed " + request.err.Error())
	}
	return request.reply
}

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
	ch := parser.ParseStream(client.conn)
	for payload := range ch {
		if payload.Err != nil {
			client.reconnect()
			return
		}
		client.finishRequest(payload.Data)
	}
}

// 收到服务端的响应
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
	if request.waiting != nil {
		request.waiting.Done()
	}
}
