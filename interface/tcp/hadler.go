package tcp

import (
	"context"
	"net"
)

// Handler redis的业务逻辑处理
type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}
