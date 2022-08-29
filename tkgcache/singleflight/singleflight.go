// Package singleflight
// author tkg
// date 2022/8/27
package singleflight

import "sync"

// 代表请求
type call struct {
	// 避免重入
	wg  sync.WaitGroup
	val any
	err error
}

// Group 管理不同的key请求
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()

	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		// 有可能map存在key但是还没有数据故需要加锁
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
