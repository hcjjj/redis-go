// Package cluster -----------------------------
// @file      : flushdb.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/18 10:40
// -------------------------------------------
package cluster

import (
	"redis-go/interface/resp"
	"redis-go/resp/reply"
)

func flushdb(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	// 所有节点 ok 才 ok
	replies := cluster.broadcast(c, cmdArgs)
	var errReply reply.ErrorReply
	for _, r := range replies {
		if reply.IsErrReply(r) {
			errReply = r.(reply.ErrorReply)
			break
		}
	}
	if errReply == nil {
		return reply.MakeOkReply()
	}
	return reply.MakeErrReply("error: " + errReply.Error())
}
