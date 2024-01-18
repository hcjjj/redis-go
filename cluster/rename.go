// Package cluster -----------------------------
// @file      : rename.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/18 10:20
// -------------------------------------------
package cluster

import (
	"redis-go/interface/resp"
	"redis-go/resp/reply"
)

// rename k1 k2 值不变
func Rename(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	if len(cmdArgs) != 3 {
		reply.MakeErrReply("ERR Wrong number args")
	}
	src := string(cmdArgs[1])
	dest := string(cmdArgs[2])

	srcPeer := cluster.peerPicker.PickNode(src)
	destPeer := cluster.peerPicker.PickNode(dest)

	if srcPeer != destPeer {
		return reply.MakeErrReply("ERR rename must within on peer")
	}
	return cluster.relay(srcPeer, c, cmdArgs)
}
