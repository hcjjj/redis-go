// Package cluster -----------------------------
// @file      : cluster_database.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/17 14:21
// -------------------------------------------
package cluster

import (
	"context"
	"redis-go/config"
	database2 "redis-go/database"
	"redis-go/interface/database"
	"redis-go/interface/resp"
	"redis-go/lib/consistenthash"

	pool "github.com/jolestar/go-commons-pool/v2"
)

type ClusterDatabase struct {
	// 自己的信息
	self string
	// 集群的信息
	nodes []string
	// 一致性哈希
	peerPicker *consistenthash.NodeMap
	// 客户端连接池
	// 如：3 个 节点 需要 2 个池子
	// 连接池需要用到工厂 connectionFactory
	peerConnection map[string]*pool.ObjectPool
	// standalone_database
	db database.Database
}

func MakeClusterDatabase() *ClusterDatabase {
	// 一堆的初始化工作
	cluster := &ClusterDatabase{
		self:           config.Properties.Self,
		db:             database2.NewStandaloneDatabase(),
		peerPicker:     consistenthash.NewNodeMap(nil),
		peerConnection: make(map[string]*pool.ObjectPool),
	}
	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	for _, peer := range config.Properties.Peers {
		nodes = append(nodes, peer)
	}
	nodes = append(nodes, config.Properties.Self)
	cluster.peerPicker.AddNode(nodes...)
	ctx := context.Background()
	for _, peer := range config.Properties.Peers {
		pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{
			Peer: peer,
		})
	}
	cluster.nodes = nodes
	return cluster

}

func (c *ClusterDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	//TODO implement me
	panic("implement me")
}

func (c *ClusterDatabase) Close() {
	//TODO implement me
	panic("implement me")
}

func (c *ClusterDatabase) AfterClientClose(resp.Connection) {
	//TODO implement me
	panic("implement me")
}
