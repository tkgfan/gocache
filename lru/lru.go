package lru

import (
	"container/list"
)

type Cache struct {
	// 允许使用的最大字节数
	maxBytes int64
	// 当前已使用的字节数
	nbytes int64
	ll     *list.List
	cache  map[string]*list.Element
	// 某条记录被移除的回调函数
	OnEvicted func(key string, value Value)
}

// 双向链表的节点值，之所以保存key是为了方便字典查找
type entry struct {
	key   string
	value Value
}

// Value 此接口只有一个方法用于返回值所占用的内存大小
type Value interface {
	Len() int
}

// New 初始化函数
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get 获取数据同时将该数据移动到链表头部
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOlds 淘汰最近最久未使用的节点（队尾元素）
// 如果存在OnEvicted钩子函数则执行钩子函数
func (c *Cache) RemoveOlds() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add 无则新增，有则修改。这里是先新增再检查nbytes是否
// 大于最大存储缓存。需要注意的是如果maxBytes是0，那么将
// 不会移除任何元素
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOlds()
	}
}

func (c Cache) Len() int {
	return c.ll.Len()
}
