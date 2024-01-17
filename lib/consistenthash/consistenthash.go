// Package consistenthash -----------------------------
// @file      : consistenthash.go
// @author    : hcjjj
// @contact   : hcjjj@foxmail.com
// @time      : 2024/1/16 17:21
// -------------------------------------------
package consistenthash

import (
	"hash/crc32"
	"sort"
)

type HashFunc func(data []byte) uint32

type NodeMap struct {
	hashFunc HashFunc
	// 各个节点的哈希值，需要排序所以用int
	// uint32 存入 int 在32位的机器上可能会溢出，64位的机器不会
	nodeHashs   []int
	nodehashMap map[int]string
}

func NewNodeMap(hf HashFunc) *NodeMap {
	m := &NodeMap{
		hashFunc:    hf,
		nodehashMap: make(map[int]string),
	}
	// 如果没指定哈希函数，就用自带的这个
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

func (m *NodeMap) IsEmpty() bool {
	return len(m.nodeHashs) == 0
}

func (m *NodeMap) AddNode(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		// string → []byte → int
		hash := int(m.hashFunc([]byte(key)))
		m.nodeHashs = append(m.nodeHashs, hash)
		// key: hash  val: string 的 key
		m.nodehashMap[hash] = key
	}
	// 没增加一个节点都需要排序一下
	sort.Ints(m.nodeHashs)
}

// PickNode 根据 key 搜索需要落在的节点
func (m *NodeMap) PickNode(key string) string {
	if m.IsEmpty() {
		return ""
	}
	hash := int(m.hashFunc([]byte(key)))
	idx := sort.Search(len(m.nodeHashs), func(i int) bool {
		// 传入的 hash 值要小于等于 节点的 hash 值
		return m.nodeHashs[i] >= hash
	})
	// 落在最后了，一个环，归为 0 号节点
	if idx == len(m.nodeHashs) {
		idx = 0
	}
	// 根据下标找到节点的hash值，在根据map找到对应的节点
	return m.nodehashMap[m.nodeHashs[idx]]
}
