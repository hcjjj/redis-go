// Package cluster  -----------------------------
// @file      : client_pool.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/17 14:39
// -------------------------------------------
package cluster

import (
	"context"
	"errors"
	"redis-go/resp/client"

	pool "github.com/jolestar/go-commons-pool/v2"
)

type connectionFactory struct {
	// 保存所连接节点的地址
	Peer string
}

func (f connectionFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	// 建立连接
	c, err := client.MakeClient(f.Peer)
	if err != nil {
		return nil, err
	}
	c.Start()
	// 丢到池子里面
	return pool.NewPooledObject(c), nil
}

func (f connectionFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	c, ok := object.Object.(*client.Client)
	if !ok {
		return errors.New("池化对象 type mismatch")
	}
	// 关闭连接
	c.Close()
	return nil
}

func (f connectionFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	return true
}

func (f connectionFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}

func (f connectionFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}
