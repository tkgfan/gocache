// Package consistenthash
// author tkg
// date 2022/8/20
package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 散列函数变量
type Hash func(data []byte) uint32

type Map struct {
	// 哈希函数
	hash Hash
	// 虚拟节点倍数
	replicas int
	// 哈希环keys
	keys []int
	// 虚拟节点与真实节点映射表
	hashMap map[int]string
}

func New(replicas int, fn Hash) *Map {
	if fn == nil {
		fn = crc32.ChecksumIEEE
	}

	return &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
}

// Add 添加节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get 根据key顺时针选择最近的节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// idx可能为len(m.keys)
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
