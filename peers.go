// Package gocache
// author tkg
// date 2022/8/21
package main

import (
	pb "golang-cache/gocachepb"
)

type PeerPicker interface {
	// PickPeer 根据key选择相应节点PeerGetter
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	// Get 从对应group查找缓存值
	Get(in *pb.Request, out *pb.Response) error
}
