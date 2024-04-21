package misakacache

// ByteView 只读数据结构 实现了Value接口 用于表示缓存的值 如果想要获取当前缓存的值 一律从GetByteCopy获取
type ByteView struct {
	cacheBytes []byte
}

// GetMemoryUsed 实现Value接口的方法 返回该缓存值的长度/占用内存多少
func (view ByteView) GetMemoryUsed() int {
	return len(view.cacheBytes)
}

// GetByteCopy 返回当前缓存值的一个拷贝 外部程序如果想要获取当前缓存的值 一律从该方法获取 用于防止缓存值被外部程序修改
func (view ByteView) GetByteCopy() []byte {
	return cloneBytes(view.cacheBytes)
}

// ToString 返回当前缓存的字符串表示
func (view ByteView) ToString() string {
	return string(view.cacheBytes)
}

func cloneBytes(b []byte) []byte {
	clone := make([]byte, len(b))
	copy(clone, b)
	return clone
}
