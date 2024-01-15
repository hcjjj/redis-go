// Package database -----------------------------
// @file      : db.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/10 20:55
// -------------------------------------------
package database

import (
	"redis-go/datastruct/dict"
	"redis-go/interface/database"
	"redis-go/interface/resp"
	"redis-go/resp/reply"
	"strings"
)

type DB struct {
	index  int
	data   dict.Dict
	addAof func(line CmdLine)
}

type ExecFunc func(db *DB, args [][]byte) resp.Reply
type CmdLine = [][]byte

func makeDB() *DB {
	db := &DB{
		data: dict.MakeSyncDict(),
		// 必须初始化，防止第一次运行出现错误
		addAof: func(line CmdLine) {},
	}
	return db
}

func (db *DB) Exec(c resp.Connection, cmdLine CmdLine) resp.Reply {
	//PING SET SETNX
	cmdName := strings.ToLower(string(cmdLine[0]))
	cmd, ok := cmdTable[cmdName]
	// 用户发送未知的命令
	if !ok {
		return reply.MakeErrReply("ERR unknown command " + cmdName)
	}
	if !validateArity(cmd.arity, cmdLine) {
		return reply.MakeArgNumErrReply(cmdName)
	}
	fun := cmd.executor
	// SET K V → K V
	return fun(db, cmdLine[1:])
}

// SET K V → arity = 3
// EXISTS k1 k2 k3 ... → arity = -2
// validateArity 判断用户发送的
func validateArity(arity int, cmdArgs [][]byte) bool {
	argNum := len(cmdArgs)
	// 固定长度
	if arity > 0 {
		return argNum == arity
	}
	// 变长
	return argNum >= -arity
}

// 常用的公共方法

func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {
	raw, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	entity, _ := raw.(*database.DataEntity)
	return entity, true
}

func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}
func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExists(key, entity)
}
func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}
func (db *DB) Remove(key string) {
	db.data.Remove(key)
}
func (db *DB) Removes(keys ...string) (deleted int) {
	deleted = 0
	for _, key := range keys {
		_, exists := db.data.Get(key)
		if exists {
			db.Remove(key)
			deleted++
		}
	}
	return deleted
}
func (db *DB) Flush() {
	db.data.Clear()
}
