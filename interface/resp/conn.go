// Package resp -----------------------------
// @file      : conn.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2023/12/17 22:02
// -------------------------------------------
package resp

// Connection 代表协议层的一个客户端的连接
type Connection interface {
	Write([]byte) error
	GetDBIndex() int
	SelectDB(int)
}
