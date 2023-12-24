// Package parser -----------------------------
// @file      : parser.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2023/12/23 20:43
// -------------------------------------------
package parser

import (
	"bufio"
	"errors"
	"io"
	"redis-go/interface/resp"
	"strconv"
)

type Payload struct {
	// 都是基于RESP协议的，数据格式一致，所以都用 Reply
	Date resp.Reply
	Err  error
}

type readState struct {
	readingMultiLine  bool
	expectedArgsCount int
	msgType           byte
	args              [][]byte
	bulkLen           int64
}

//finished 判断是否解析结束
func (s *readState) finished() bool {
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}

// ParseStream 异步解析，作为协议层对外的接口
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch)
	return ch
}

// parse0 解析器
func parse0(reader io.Reader, ch chan<- *Payload) {

}

// readLine 分割行
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n 表示数组 [SET, key, value]
	var msg []byte
	var err error
	// bulkLen 表示要读取字符的长度
	// 1. \r\n 切分
	if state.bulkLen == 0 {
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			// 出现io错误
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
	} else {
		// 2. 之前读到了$数字，严格读取字符个数（字符串中可能包含\r\n）
		// len("\r\n") == 2
		msg = make([]byte, state.bulkLen+2)
		_, err := io.ReadFull(bufReader, msg)
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		state.bulkLen = 0
	}
	return msg, false, nil
}

// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
// 3 SET key value
func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	// 取 * 后面的数字，表示成员个数
	var expectedLine uint64
	// *3\r\n 留下3
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = int(expectedLine)
		state.args = make([][]byte, 0, expectedLine)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

// parseBulkHeader 解析单行字符串
func parseBulkHeader(msg []byte, state *readState) error {
	// $4\r\nPING\r\n
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if state.bulkLen == -1 { // null bulk
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

// +OK -err
//func parserSingleLineReply(msg []byte) (resp.Reply, error) {
//str := strings.TrimSuffix(string(msg), "\r\n")
//}
