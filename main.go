// Package main -----------------------------
// @file      : main.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2023/12/15 20:23
// -------------------------------------------
package main

import (
	"fmt"
	"os"
	"redis-go/lib/config"
	"redis-go/lib/logger"
	"redis-go/resp/handler"
	"redis-go/tcp"
)

const configFile string = "redis.conf"

var defaultProperties = &config.ServerProperties{
	Bind: "0.0.0.0",
	Port: 6379,
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func main() {
	// 初始化工作
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "redis-go",
		Ext:        "log",
		TimeFormat: "2023-12-15",
	})
	if fileExists(configFile) {
		config.SetupConfig(configFile)
	} else {
		// 如果没有配置文件
		// 这样子就是单机模式了
		config.Properties = defaultProperties
	}
	logger.Info(fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port))
	// 业务
	err := tcp.ListenAndServeWithSignal(
		&tcp.Config{
			// IP:PORT
			Address: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
		},
		//tcp.MakeEchoHandler())
		handler.MakeHandler())
	if err != nil {
		logger.Error(err)
	}
}
