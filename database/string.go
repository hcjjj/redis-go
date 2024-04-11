// Package database -----------------------------
// @file      : string.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/13 20:37
// -------------------------------------------
package database

import (
	"redis-go/interface/database"
	"redis-go/interface/resp"
	"redis-go/lib/utils"
	"redis-go/resp/reply"
)

// GET k1
func execGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeNullBulkReply()
	}
	// 这边只有string 如果是其他类型需要判断转化是否成功
	bytes := entity.Data.([]byte)
	return reply.MakeBulkReply(bytes)
}

// SET k1 v
func execSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity := &database.DataEntity{
		Data: value,
	}
	db.PutEntity(key, entity)

	// 往文件里面写指令
	db.addAof(utils.ToCmdLine2("set", args...))

	return reply.MakeOkReply()
}

// SETNX k1 v1
func execSetnx(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity := &database.DataEntity{
		Data: value,
	}
	result := db.PutIfAbsent(key, entity)

	db.addAof(utils.ToCmdLine2("setnx", args...))

	return reply.MakeIntReply(int64(result))
}

// GETSET k1 v1
func execGetSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := string(args[1])
	// 读取原来的值 返回用
	entity, exists := db.GetEntity(key)
	// 设置新的值
	db.PutEntity(key, &database.DataEntity{
		Data: value,
	})

	db.addAof(utils.ToCmdLine2("getset", args...))

	if !exists {
		return reply.MakeNullBulkReply()
	}
	return reply.MakeBulkReply(entity.Data.([]byte))
}

// STRLEN
func execStrLen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeNullBulkReply()
	}
	bytes := entity.Data.([]byte)
	return reply.MakeIntReply(int64(len(bytes)))
}

func init() {
	RegisterCommand("Get", execGet, 2)
	RegisterCommand("Set", execSet, 3)
	RegisterCommand("SetNx", execSetnx, 3)
	RegisterCommand("GetSet", execGetSet, 3)
	RegisterCommand("StrLen", execStrLen, 2)
}
