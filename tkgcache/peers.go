// Package tkgcache
// author tkg
// date 2022/8/21
package tkgcache

type PeerPicker interface {
	// PickPeer 根据key选择相应节点PeerGetter
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	// Get 从对应group查找缓存值
	Get(group string, key string) ([]byte, error)
}
