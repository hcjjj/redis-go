// Package reply -----------------------------
// @file      : consts.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2023/12/22 16:44
// -------------------------------------------
package reply

// 固定的一些回复

type PongReply struct {
}

var pongBytes = []byte("+PONG\r\n")

func (p PongReply) ToBytes() []byte {
	return pongBytes
}

func MakePongReply() *PongReply {
	return &PongReply{}
}

type OkReply struct {
}

var okBytes = []byte("+OK\r\n")

func (o OkReply) ToBytes() []byte {
	return okBytes
}

// 持有固定的一个，节约内存的一种方式
var theOkReply = new(OkReply)

func MakeOkReply() *OkReply {
	return theOkReply
}

// NullBulkReply 空的字符串回复
type NullBulkReply struct {
}

// 空回复，不是空字符串
var nullBulkBytes = []byte("$-1\r\n")

func (n NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

var emptyMultiBulkBytes = []byte("*0\r\n")

// EmptyMultiBulkReply is an empty list
type EmptyMultiBulkReply struct{}

func (e EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

type NoReply struct{}

var noBytes = []byte("")

func (n NoReply) ToBytes() []byte {
	return noBytes
}
