// Package database -----------------------------
// @file      : command.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/12 18:28
// -------------------------------------------
package database

import "strings"

// 支持的 指令表
// 每个指令对应一个 command 结构体
var cmdTable = make(map[string]*command)

type command struct {
	// 对应的执行的方法
	executor ExecFunc
	// 参数的数量
	arity int
}

func RegisterCommand(name string, executor ExecFunc, arity int) {
	name = strings.ToLower(name)
	cmdTable[name] = &command{
		executor: executor,
		arity:    arity,
	}
}
