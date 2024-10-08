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
	// 一组DB的指针
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
	// 初始化 aofHandler 先查看有没有开启这个功能
	if config.Properties.AppendOnly {
		// 这边传递的是 database 指针
		// 因为 database 实现的接口的方式是通过结构体指针（指针接收者）
		// new 的时候就会恢复数据了
		aofHandler, err := aof.NewAofHandler(database)
		if err != nil {
			panic(err)
			logger.Error("AOF启动失败")
		}
		database.aofHandler = aofHandler
		// 初始化 db 中的 addAof 方法
		for _, db := range database.dbSet {

			// for range 使用闭包 坑
			// 在没有将变量 db 的拷贝值传进匿名函数之前，只能获取最后一次循环的值

			//db.addAof = func(line CmdLine) {
			//	// 这个 db.index 都是15
			//	fmt.Println(db.index)
			//	database.aofHandler.AddAof(db.index, line)
			//}

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
	// select 是一个特例 他是操作数据库的 不是分db
	if cmdName == "select" {
		if len(args) != 2 {
			return reply.MakeArgNumErrReply("select")
		}
		return execSelect(client, database, args[1:])
	}
	//  require multi bulk reply to exec
	//if cmdName == "ping" {
	//	return reply.MakePongReply()
	//}

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
	// byte → string → int
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
