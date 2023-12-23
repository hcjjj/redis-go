// Package resp -----------------------------
// @file      : reply.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2023/12/22 16:42
// -------------------------------------------
package resp

// Reply 服务器对客户端的回复
type Reply interface {
	ToBytes() []byte
}
