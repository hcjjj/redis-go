// Package cluster -----------------------------
// @file      : router.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/18 10:08
// -------------------------------------------
package cluster

import "redis-go/interface/resp"

func makeRouter() map[string]CmdFunc {
	routerMap := make(map[string]CmdFunc)
	// exists k1 ...
	// 可以直接转发的指令
	routerMap["exists"] = defaultFunc
	routerMap["type"] = defaultFunc
	routerMap["set"] = defaultFunc
	routerMap["setnx"] = defaultFunc
	routerMap["get"] = defaultFunc
	routerMap["getset"] = defaultFunc
	// 特殊模式的指令
	routerMap["ping"] = ping
	routerMap["rename"] = Rename
	routerMap["renamenx"] = Rename
	routerMap["flushdb"] = flushdb
	routerMap["del"] = Del
	routerMap["select"] = execSelect
	return routerMap
}

// GET Key
// SET k1 v1
func defaultFunc(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	// 根据数据的 key 来选择要执行的节点
	key := string(cmdArgs[1])
	// peer 是节点的地址
	peer := cluster.peerPicker.PickNode(key)
	return cluster.relay(peer, c, cmdArgs)
}
