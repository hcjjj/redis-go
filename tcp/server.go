// Package tcp -----------------------------
// @file      : server.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2023/12/15 19:34
// -------------------------------------------
package tcp

import (
	"context"
	"net"
	"os"
	"os/signal"
	"redis-go/interface/tcp"
	"redis-go/lib/logger"
	"sync"
	"syscall"
)

// Config tcp连接配置信息
type Config struct {
	Address string
}

var ClientCounter int32

func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	// 获取操作系统给程序发送的信号
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	// 转发信号到自定义的 closeChan
	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info("Start listen ...")
	logger.Info(
		"\n" +
			"                _ _                       \n" +
			"               | (_)                      \n" +
			"   _ __ ___  __| |_ ___ ______ __ _  ___  \n" +
			"  | '__/ _ \\/ _` | / __|______/ _` |/ _ \\ \n" +
			"  | | |  __/ (_| | \\__ \\     | (_| | (_) |\n" +
			"  |_|  \\___|\\__,_|_|___/      \\__, |\\___/ \n" +
			"                               __/ |      \n" +
			"                              |___/       \n")
	ListenAndServe(listener, handler, closeChan)
	return nil
}

func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	// 监听应用程序被关闭的系统信号
	go func() {
		<-closeChan
		logger.Info("Shutting down")
		_ = listener.Close()
		_ = handler.Close()
	}()
	// 结束时需要关闭
	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()
	ctx := context.Background()
	var waitDone sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("Accepted link: " + conn.RemoteAddr().String())
		waitDone.Add(1)
		// 一个协程处理一个连接
		go func() {
			// 防止连接出现 panic 导致没 Done()
			defer func() {
				waitDone.Done()
			}()
			// 这个 Handle 也是一直在 for 的
			handler.Handle(ctx, conn)
		}()
	}
	// 出现错误跳出循环时需要等待已存在的连接结束
	waitDone.Wait()
}
