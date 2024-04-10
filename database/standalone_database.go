// Package database -----------------------------
// @file      : standalone_database.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/13 21:10
// -------------------------------------------
package database

import (
	"redis-go/aof"
	"redis-go/interface/resp"
	"redis-go/lib/config"
	"redis-go/lib/logger"
	"redis-go/resp/reply"
	"strconv"
	"strings"
)

type StandaloneDatabase struct {
	dbSet      []*DB
	aofHandler *aof.AofHandler
}

// NewStandaloneDatabase 创建 Redis 数据库的核心 默认为16个分数据库
func NewStandaloneDatabase() *StandaloneDatabase {
	database := &StandaloneDatabase{}
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}
	// 初始化 DB
	database.dbSet = make([]*DB, config.Properties.Databases)
	for i := range database.dbSet {
		db := makeDB()
		db.index = i
		database.dbSet[i] = db
	}
	// 初始化 aofHandler
	if config.Properties.AppendOnly {
		// 这边传递的是 database 指针
		// 因为 database 实现的接口的方式是通过结构体指针（指针接收者）
		aofHandler, err := aof.NewAofHandler(database)
		if err != nil {
			panic(err)
		}
		database.aofHandler = aofHandler
		// 初始化 db 中的 addAof 方法
		for _, db := range database.dbSet {
			// db 的值会变但是地址不会变
			// 这边是一个闭包 导致 db.index 写死了为 dbSet[15] 的 15
			// db 引用了 外面 for 的局部变量 db，其逃逸到堆上了
			// db = dbSet[0]
			// db = dbSet[1]
			// ...
			//db.addAof = func(line CmdLine) {
			//	database.aofHandler.AddAof(db.index, line)
			//}
			// sdb 的值和地址都会变
			sdb := db
			sdb.addAof = func(line CmdLine) {
				database.aofHandler.AddAof(sdb.index, line)
			}
		}
	}

	return database
}

// set k v
// get k
// select 2

func (database *StandaloneDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {

	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()

	cmdName := strings.ToLower(string(args[0]))
	if cmdName == "select" {
		if len(args) != 2 {
			return reply.MakeArgNumErrReply("select")
		}
		return execSelect(client, database, args[1:])
	}

	dbIndex := client.GetDBIndex()
	db := database.dbSet[dbIndex]
	return db.Exec(client, args)

}

func (database *StandaloneDatabase) Close() {

}

func (database *StandaloneDatabase) AfterClientClose(c resp.Connection) {

}

// select 2
// select a
// select 123123131231
func execSelect(c resp.Connection, database *StandaloneDatabase, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeErrReply("ERR invalid DB index")
	}
	if dbIndex >= len(database.dbSet) {
		return reply.MakeErrReply("ERR DB index is out of range")
	}
	c.SelectDB(dbIndex)
	return reply.MakeOkReply()
}
