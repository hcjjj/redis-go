// Package cluster -----------------------------
// @file      : cluster_database.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/17 14:21
// -------------------------------------------
package cluster

import (
	"context"
	database2 "redis-go/database"
	"redis-go/interface/database"
	"redis-go/interface/resp"
	"redis-go/lib/config"
	"redis-go/lib/consistenthash"
	"redis-go/lib/logger"
	"redis-go/resp/reply"
	"strings"

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
		self:       config.Properties.Self,
		db:         database2.NewStandaloneDatabase(),
		peerPicker: consistenthash.NewNodeMap(nil),
		// key是peer节点的地址
		peerConnection: make(map[string]*pool.ObjectPool),
	}
	// 初始化 nodes， self + peers
	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	// IP:PORT 作为 哈希的 key
	for _, peer := range config.Properties.Peers {
		nodes = append(nodes, peer)
	}
	nodes = append(nodes, config.Properties.Self)
	// 将节点加入一致性哈希的环
	cluster.peerPicker.AddNode(nodes...)
	// 初始化连接池 self 到每一个 peer
	ctx := context.Background()
	for _, peer := range config.Properties.Peers {
		cluster.peerConnection[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{
			Peer: peer,
		})
	}
	// 写到结构体的字段里面
	cluster.nodes = nodes
	return cluster
}

type CmdFunc func(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply

var router = makeRouter()

func (cluster *ClusterDatabase) Exec(client resp.Connection, args [][]byte) (result resp.Reply) {
	// 集群层的执行替代单机版的执行

	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			result = &reply.UnknowErrReply{}
		}
	}()

	cmdName := strings.ToLower(string(args[0]))
	cmdFunc, ok := router[cmdName]
	if !ok {
		reply.MakeErrReply(" cluster mode not supported cmd" + cmdName)
	}
	result = cmdFunc(cluster, client, args)
	//return result
	return
}

func (cluster *ClusterDatabase) Close() {
	cluster.db.Close()
}

func (cluster *ClusterDatabase) AfterClientClose(c resp.Connection) {
	cluster.db.AfterClientClose(c)
}
