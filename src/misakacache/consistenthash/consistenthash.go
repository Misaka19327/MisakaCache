package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// HashFunc 哈希函数类型
type HashFunc func(data []byte) uint32

// Map 一致性哈希的主要数据结构 这里的一致性哈希维护的节点仅为节点名
type Map struct {
	hash           HashFunc
	replicasNumber int            // 真实节点和虚拟节点的映射倍数
	keys           []int          // 一致性哈希的环的抽象版
	hashmap        map[int]string // 虚拟节点到真实节点的映射
}

// NewMap 一致性哈希的构造函数 默认情况下选择CRC32校验和作为哈希值
func NewMap(hashFunc HashFunc, replicasNumber int) (result *Map) {
	result = &Map{
		hash:           hashFunc,
		replicasNumber: replicasNumber,
		hashmap:        make(map[int]string),
	}
	if hashFunc == nil {
		result.hash = crc32.ChecksumIEEE
	}
	return
}

// AddRealNode 为一致性哈希添加真实节点（可以一次添加多个真实节点）
func (m *Map) AddRealNode(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicasNumber; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key))) // 这个strconv.Itoa等效FormatInt 从整型转字符串
			m.keys = append(m.keys, hash)
			m.hashmap[hash] = key // 添加虚拟节点到真实节点的映射
		}
	}
	sort.Ints(m.keys) // 排序
}

// GetRealNodeByKey 根据key来获得真实节点
func (m *Map) GetRealNodeByKey(key string) (result string) {
	if len(m.keys) == 0 { // key检查是否有效
		return ""
	}

	keyHash := int(m.hash([]byte(key)))
	index := sort.Search(len(m.keys), func(i int) bool { // 二分查找
		return m.keys[i] >= keyHash
	})

	return m.hashmap[m.keys[index%len(m.keys)]] // 去映射里查找真实节点
}
