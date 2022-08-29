package main

import (
	"fmt"
	pb "golang-cache/gocachepb"
	"golang-cache/singleflight"
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
	name string
	// 回调函数
	getter    Getter
	mainCache cache
	peers     PeerPicker
	// 使得key只会加载一次
	loader *singleflight.Group
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
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// RegisterPeers 将实现了 PeerPicker 接口的 HTTPPool 注入到 Group 中。
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// 使用回调函数加载数据
func (g *Group) load(key string) (value ByteView, err error) {
	res, err := g.loader.Do(key, func() (any, error) {
		if g.peers != nil {
			// 选择节点
			if peer, ok := g.peers.PickPeer(key); ok {
				// 从节点获取数据
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[tkgCache] Failed to get from peer", err)
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return res.(ByteView), nil
	}

	return
}

// 访问远程节点，获取缓存值。
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}

	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}

	return ByteView{b: res.Value}, nil
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
