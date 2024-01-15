// Package database -----------------------------
// @file      : keys.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/12 20:22
// -------------------------------------------
package database

import (
	"redis-go/interface/resp"
	"redis-go/lib/utils"
	"redis-go/lib/wildcard"
	"redis-go/resp/reply"
)

// DEL k1 k2 k3 ...
func execDel(db *DB, args [][]byte) resp.Reply {
	// 传入的时候第一个关键字已经被去除了
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}
	deleted := db.Removes(keys...)

	// 落盘
	if deleted > 0 {
		db.addAof(utils.ToCmdLine2("del", args...))
	}

	return reply.MakeIntReply(int64(deleted))
}

// EXISTS k1 k2 k3 ...
func execExists(db *DB, args [][]byte) resp.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		_, exists := db.GetEntity(key)
		if exists {
			result++
		}
	}
	return reply.MakeIntReply(result)
}

// FLUSHDB
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()

	db.addAof(utils.ToCmdLine2("flushdb", args...))

	return reply.MakeOkReply()
}

// TYPE k1
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		reply.MakeStatusReply("none")
	}
	switch entity.Data.(type) {
	case []byte:
		reply.MakeStatusReply("string")
	}
	// TODO:
	return &reply.UnknowErrReply{}
}

// RENAME k1 k2  k1:v → k2:v
func execRename(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])
	entity, exists := db.GetEntity(src)
	if !exists {
		reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)

	db.addAof(utils.ToCmdLine2("rename", args...))

	return reply.MakeOkReply()
}

// RENAMNX
func execRenamenx(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])

	_, ok := db.GetEntity(dest)
	if ok {
		return reply.MakeIntReply(0)
	}

	entity, exists := db.GetEntity(src)
	if !exists {
		reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)

	db.addAof(utils.ToCmdLine2("renamenx", args...))

	return reply.MakeIntReply(1)
}

// KEYS *
// 含有 * 通配符
func execKeys(db *DB, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0]))
	result := make([][]byte, 0)
	db.data.ForEach(func(key string, val interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}

func init() {
	RegisterCommand("DEL", execDel, -2)
	RegisterCommand("EXISTS", execExists, -2)
	// 忽略掉 FULSHDB 后续的一些参数
	// FLUSHDB a b c
	RegisterCommand("FlushDB", execFlushDB, -1)
	// TYPE k1
	RegisterCommand("TYPE", execType, 2)
	// RENAME k1 k2
	RegisterCommand("RENAME", execRename, 3)
	RegisterCommand("RENAMENX", execRenamenx, 3)
	// KEYS *
	RegisterCommand("KEYS", execKeys, 2)
}
