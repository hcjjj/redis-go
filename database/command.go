// Package database -----------------------------
// @file      : command.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/12 18:28
// -------------------------------------------
package database

import "strings"

// 支持的 指令表
var cmdTable = make(map[string]*command)

type command struct {
	executor ExecFunc
	arity    int
}

func RegisterCommand(name string, executor ExecFunc, arity int) {
	name = strings.ToLower(name)
	cmdTable[name] = &command{
		executor: executor,
		arity:    arity,
	}
}
