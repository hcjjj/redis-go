// Package cluster -----------------------------
// @file      : com.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/17 21:22
// -------------------------------------------
package cluster

import (
	"context"
	"errors"
	"redis-go/interface/resp"
	"redis-go/lib/utils"
	"redis-go/resp/client"
	"redis-go/resp/reply"
	"strconv"
)

// 获取一个 peer 的连接
func (cluster *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	pool, ok := cluster.peerConnection[peer]
	if !ok {
		return nil, errors.New(peer + "connection not found")
	}
	object, err := pool.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}
	c, ok := object.(*client.Client)
	if !ok {
		return nil, errors.New("wrong type")
	}
	return c, err
}

// 返回连接
func (cluster *ClusterDatabase) returnPeerClient(peer string, peerClient *client.Client) error {
	pool, ok := cluster.peerConnection[peer]
	if !ok {
		return errors.New("connection not found")
	}
	return pool.ReturnObject(context.Background(), peerClient)
}

// 指令的转发
func (cluster *ClusterDatabase) relay(peer string, c resp.Connection, args [][]byte) resp.Reply {
	if peer == cluster.self {
		return cluster.db.Exec(c, args)
	}
	peerClient, err := cluster.getPeerClient(peer)
	if err != nil {
		return reply.MakeErrReply(err.Error())
	}
	defer func() {
		// 避免连接耗尽
		_ = cluster.returnPeerClient(peer, peerClient)
	}()
	// 先切库 再发送具体指令
	peerClient.Send(utils.ToCmdLine("SELECT", strconv.Itoa(c.GetDBIndex())))
	return peerClient.Send(args)
}

// 指令的广播
func (cluster *ClusterDatabase) broadcast(c resp.Connection, args [][]byte) map[string]resp.Reply {
	results := make(map[string]resp.Reply)
	for _, node := range cluster.nodes {
		result := cluster.relay(node, c, args)
		results[node] = result
	}
	return results
}
