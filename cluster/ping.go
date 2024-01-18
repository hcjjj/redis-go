// Package cluster -----------------------------
// @file      : ping.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/18 10:18
// -------------------------------------------
package cluster

import "redis-go/interface/resp"

func ping(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	return cluster.db.Exec(c, cmdArgs)
}
