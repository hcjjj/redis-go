// Package aof -----------------------------
// @file      : aof.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/14 16:33
// -------------------------------------------
package aof

import (
	"os"
	"redis-go/config"
	"redis-go/database"
	"redis-go/lib/logger"
	"redis-go/lib/utils"
	"redis-go/resp/reply"
	"strconv"
)

const (
	aofQueueSize = 1 << 16
)

// CmdLine is alias for [][]byte, represents a command line
type CmdLine = [][]byte

type payload struct {
	cmdLine CmdLine
	dbIndex int
}

// AofHandler receive msgs from channel and write to AOF file
type AofHandler struct {
	database    database.Database
	aofChan     chan *payload //写aof文件的缓存池
	aofFile     *os.File      // .aof文件
	aofFilename string        // 文件名
	currentDB   int           // 记录上一条指令工作的 db
}

// NewAOFHandler creates a new aof.AofHandler
func NewAOFHandler(db database.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFilename = config.Properties.AppendFilename
	handler.database = db
	//加载已有的数据
	handler.LoadAof()
	aofFile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600) //读写方式打开文件
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile
	// channel缓冲
	handler.aofChan = make(chan *payload, aofQueueSize)
	// 异步的
	go func() {
		handler.handleAof()
	}()
	return handler, nil
}

// AddAof send command to aof goroutine through channel
// Add payload(set k v) -> aofChan
func (handler *AofHandler) AddAof(dbIndex int, cmdLine CmdLine) {
	// 可以append且aofChan已经初始化
	if config.Properties.AppendOnly && handler.aofChan != nil {
		handler.aofChan <- &payload{
			cmdLine: cmdLine,
			dbIndex: dbIndex,
		}
	}
}

// handleAof listen aof channel and write into file
// payload(set k v) <- aofChan 落盘
func (handler *AofHandler) handleAof() {
	// serialized execution
	handler.currentDB = 0
	for p := range handler.aofChan {
		if p.dbIndex != handler.currentDB {
			// 不一致 插入 select db
			data := reply.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(p.dbIndex))).ToBytes()
			// 写入文件
			_, err := handler.aofFile.Write(data)
			if err != nil {
				logger.Warn(err)
				continue // skip this command
			}
			handler.currentDB = p.dbIndex
		}
		data := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Warn(err)
		}
	}
}

// LoadAof read aof file
func (handler *AofHandler) LoadAof() {

}
