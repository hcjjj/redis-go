// Package database -----------------------------
// @file      : standalone_database.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/3 16:46
// -------------------------------------------
package database

import "redis-go/interface/resp"

type CmdLine = [][]byte

type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	Close()
	AfterClientClose(c resp.Connection)
}

// DataEntity 指代 Redis 所有数据结构
type DataEntity struct {
	Data interface{}
}
