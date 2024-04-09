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
	"redis-go/lib/logger"
	"redis-go/resp/reply"
	"runtime/debug"
	"strconv"
	"strings"
)

type Payload struct {
	// 表示客户端给服务器的数据
	// 都是基于RESP协议的，数据格式一致，所以都用 Reply
	Data resp.Reply
	Err  error
}

// readState 解析状态参数
type readState struct {
	readingMultiLine bool
	// 参数个数
	expectedArgsCount int
	msgType           byte
	// 实际解析的参数内容
	args [][]byte
	// 预期要读取的字节数
	bulkLen int64
}

// finished 判断是否解析结束
func (s *readState) finished() bool {
	// 已解析参数和预期参数数量一致了
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}

// ParseStream 异步解析，作为协议层对外的接口
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	// 这个parse0 解析器协程是一个用户一个
	go parse0(reader, ch)
	return ch
}

// parse0 解析器
func parse0(reader io.Reader, ch chan<- *Payload) {
	// 防止出现异常 循环中断
	defer func() {
		if err := recover(); err != nil {
			logger.Error(string(debug.Stack()))
		}
	}()
	bufReader := bufio.NewReader(reader)
	// 解析器状态
	var state readState
	var err error
	var msg []byte
	for true {
		var ioErr bool
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil {
			// 出现 io 错误 解析直接结束
			if ioErr {
				ch <- &Payload{
					Err: err,
				}
				close(ch)
				return
			}
			// 如果是协议错误
			ch <- &Payload{
				Err: err,
			}
			// 重置 state
			state = readState{}
			// 继续解析用户发来的下一条数据
			continue
		}
		// 判断是否未多行解析模式
		// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
		// *3 即第一次解
		// $3\r\n
		//...
		if !state.readingMultiLine {
			//fmt.Printf("%s", string(msg))
			// *3/r/n
			if msg[0] == '*' { //*3
				err := parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error [parseMultiBulkHeader]: " + string(msg)),
					}
					state = readState{}
					continue
				}
				if state.expectedArgsCount == 0 {
					// 如果用发的是 *0...
					// 这个 Payload 是通过 ch 传送给 Redis 核心层的
					ch <- &Payload{
						Data: &reply.EmptyMultiBulkReply{},
					}
					state = readState{}
					continue
				}
				// $4\r\nPING\r\n
			} else if msg[0] == '$' {
				err := parseBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{}
					continue
				}
				if state.bulkLen == -1 {
					// $-1/r/n
					ch <- &Payload{
						Data: &reply.NullBulkReply{},
					}
					state = readState{}
					continue
				}
			} else {
				// + 或 - 或 :
				result, err := parseSingleLineReply(msg)
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
				continue
			}
		} else {
			// 进入多行模式
			err := readBody(msg, &state)
			if err != nil {
				ch <- &Payload{
					Err: errors.New("protocol error: " + string(msg)),
				}
				state = readState{}
				continue
			}
			// 解析完成
			if state.finished() {
				var result resp.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0])
				}
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				// 重置解析器状态
				state = readState{}
			}
		}
	}
}

// readLine 分割行
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n 表示数组 [SET, key, value]
	// msg 表示读取的一行内容
	var msg []byte
	var err error
	// bulkLen 表示要读取字符的长度
	// 1. \r\n 切分
	// 没有预设的长度
	if state.bulkLen == 0 {
		// xxxx\r\n
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			// 出现io错误
			return nil, true, err
		}
		// 如果 \n 前面不是 \r 表示协议错误
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
	} else {
		// 2. 之前读到了$数字，严格读取字符个数（字符串中可能包含\r\n）
		// len("\r\n") == 2
		msg = make([]byte, state.bulkLen+2)
		// 塞满msg，截至特定字节数量
		_, err := io.ReadFull(bufReader, msg)
		if err != nil {
			// 出现io错误
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
			// 判断用户发送的数据是否符合协议
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		// 预期要读取的字节长度置为 0
		state.bulkLen = 0
	}
	return msg, false, nil
}

// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
// 3 SET key value
// *2\r\n$6\r\nselect\r\n$1\r\n1\r\n
// select 1
func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	// 取 * 后面的数字，表示成员个数
	var expectedLine uint64
	// *3\r\n 留下3
	// string → int
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		// * 表示真正读的是数组
		state.msgType = msg[0]
		// 标记是读取多行的
		state.readingMultiLine = true
		state.expectedArgsCount = int(expectedLine)
		// set key value 为三个参数
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
		// 多行模式，$4\r\nPING\r\n 是两行的，所以要打开
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

// +OK\r\n -err\r\n :5\r\n
func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	// 去掉末尾的 "\r\n"
	str := strings.TrimSuffix(string(msg), "\r\n")
	var result resp.Reply
	switch msg[0] {
	case '+':
		result = reply.MakeStatusReply(str[1:])
	case '-':
		result = reply.MakeErrReply(str[1:])
	case ':':
		// str -> int
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(val)
	}
	return result, nil
}

// 解析 *3 $3 后面的内容
// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
// $4\r\nPING\r\n
// 情况一：$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
// 情况二：PING\r\n
func readBody(msg []byte, state *readState) error {
	// 去掉末尾的 \r\n
	line := msg[0 : len(msg)-2]
	var err error
	// $3
	if line[0] == '$' {
		// 去掉 $ 然后 str -> int
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		// $0\r\n
		if state.bulkLen <= 0 {
			// 塞个空的参数
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return nil
}
