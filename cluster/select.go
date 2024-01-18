// Package cluster -----------------------------
// @file      : select.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/18 11:02
// -------------------------------------------
package cluster

import "redis-go/interface/resp"

func execSelect(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	// 转发的时候会补上 select 信息，所以这边只要自己执行一下就行了记录在本地就行
	return cluster.db.Exec(c, cmdArgs)
}
