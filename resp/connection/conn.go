// Package connection -----------------------------
// @file      : conn.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/3 11:04
// -------------------------------------------
package connection

import (
	"net"
	"redis-go/lib/sync/wait"
	"sync"
	"time"
)

type Connection struct {
	// TCP 连接信息
	conn net.Conn
	// 关闭服务器之前需要将还没处理的命令处理完
	waitingReply wait.Wait
	// 避免并发问题
	mu sync.Mutex
	// 选择哪个 db
	selectedDB int
}

func NewConn(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) Close() error {
	// 防止客户端关闭引起服务端的异常
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	c.conn.Close()
	return nil
}

// 给客户端发送数据
func (c *Connection) Write(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}
	c.mu.Lock()
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.mu.Unlock()
	}()
	_, err := c.conn.Write(bytes)
	return err
}

func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

func (c *Connection) SelectDB(dbNum int) {
	c.selectedDB = dbNum
}
