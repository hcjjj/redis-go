// Package database -----------------------------
// @file      : ping.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/12 20:11
// -------------------------------------------
package database

import (
	"redis-go/interface/resp"
	"redis-go/resp/reply"
)

func Ping(db *DB, args [][]byte) resp.Reply {
	return reply.MakePongReply()
}

// PING
func init() {
	RegisterCommand("ping", Ping, 1)
}
