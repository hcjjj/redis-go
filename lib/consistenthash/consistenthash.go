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
	// 使用什么哈希函数
	hashFunc HashFunc
	// 各个节点的哈希值，需要排序所以用int
	// uint32 存入 int 在32位的机器上可能会溢出，64位的机器不会
	nodeHashes []int
	// 根据哈希值找到对应节点
	nodeHashMap map[int]string
}

func NewNodeMap(hf HashFunc) *NodeMap {
	m := &NodeMap{
		hashFunc:    hf,
		nodeHashMap: make(map[int]string),
	}
	// 如果没指定哈希函数，就用自带的这个
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

func (m *NodeMap) IsEmpty() bool {
	return len(m.nodeHashes) == 0
}

func (m *NodeMap) AddNode(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		// string → []byte → int
		hash := int(m.hashFunc([]byte(key)))
		m.nodeHashes = append(m.nodeHashes, hash)
		// key: hash  val: string 的 key
		m.nodeHashMap[hash] = key
	}
	// 每增加一个节点都需要排序一下
	// 哈希环需要有序，以方便后续的查找
	sort.Ints(m.nodeHashes)
}

// PickNode 根据 key 搜索需要落在的节点
func (m *NodeMap) PickNode(key string) string {
	if m.IsEmpty() {
		return ""
	}
	hash := int(m.hashFunc([]byte(key)))
	// 找到第一个大于等于该哈希值的节点的下标
	idx := sort.Search(len(m.nodeHashes), func(i int) bool {
		// 传入的 hash 值要小于等于 节点的 hash 值
		return m.nodeHashes[i] >= hash
	})
	// 落在最后了，一个环，归为 0 号节点
	if idx == len(m.nodeHashes) {
		idx = 0
	}
	// 根据下标找到节点的hash值，在根据map找到对应的节点名称
	return m.nodeHashMap[m.nodeHashes[idx]]
}
