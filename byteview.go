package main

// ByteView 用来表示缓存值
type ByteView struct {
	// b将会缓存真实数据
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

// ByteSlice 返回一个深度复制的切片
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// 深度复制
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
