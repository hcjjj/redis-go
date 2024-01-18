// Package cluster -----------------------------
// @file      : del.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/18 10:49
// -------------------------------------------
package cluster

import (
	"redis-go/interface/resp"
	"redis-go/resp/reply"
)

// del k1 k2 k3 ...
func Del(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	replies := cluster.broadcast(c, cmdArgs)
	var errReply reply.ErrorReply
	var deleted int64 = 0

	for _, r := range replies {
		if reply.IsErrReply(r) {
			errReply = r.(reply.ErrorReply)
			break
		}
		intReply, ok := r.(*reply.IntReply)
		if !ok {
			errReply = reply.MakeErrReply("error")
		}
		deleted += intReply.Code
	}
	if errReply == nil {
		return reply.MakeIntReply(deleted)
	}

	return reply.MakeErrReply("error: " + errReply.Error())
}
