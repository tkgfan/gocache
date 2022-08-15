package cache

import (
	"fmt"
	"log"
	"sync"
)

// Getter 定义获取数据接口
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc Getter接口函数
type GetterFunc func(key string) ([]byte, error)

// Get 实现接口函数
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group 是一个缓存命名空间，每个Group拥有一个唯一
// 的名称name。
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 创建一个Group实例
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("回调函数Getter不能为nil")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup 获取一个Group实例
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get 获取key的value，如果key不存在则会调用自定义函数去加载
// 数据同时将数据添加到缓存。
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[Golang-Cache] hit")
		return v, nil
	}
	return g.load(key)
}

// 使用回调函数加载数据
func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// 将key-value添加到缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
