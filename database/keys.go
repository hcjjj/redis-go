// Package database -----------------------------
// @file      : keys.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/12 20:22
// -------------------------------------------
package database

import (
	"redis-go/interface/resp"
	"redis-go/resp/reply"
)

// DEL k1 k2 k3 ...
func execDel(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}
	deleted := db.Removes(keys...)
	return reply.MakeIntReply(int64(deleted))
}

// EXISTS k1 k2 k3 ...
//func execExists(db *DB, arg [][]byte) resp.Reply {
//
//}

// KEYS
// FLUSHDB
// TYPE
// RENAME
// RENAMNX

func init() {
	RegisterCommand("DEL", execDel, -2)

}
