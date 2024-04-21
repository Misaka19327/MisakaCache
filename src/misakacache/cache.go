package misakacache

import (
	"MisakaCache/src/misakacache/lru"
	"sync"
)

// cache 对LRU的一次封装 并且追加并发保护
type cache struct {
	mutex      sync.Mutex // 互斥锁
	lru        *lru.LRU   // 封装的LRU
	cacheBytes int64
}

// add 对LRU.SetValue的封装
func (c *cache) add(key string, value ByteView) {
	c.mutex.Lock() // 互斥锁加锁
	defer c.mutex.Unlock()
	if c.lru == nil {
		c.lru = lru.NewLRU(c.cacheBytes, nil) // 懒加载
	}
	c.lru.SetValue(key, value)
}

// get 对LRU.GetValue的封装
func (c *cache) get(key string) (value ByteView, isOk bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil {
		return
	}

	if v, isOk := c.lru.GetValue(key); isOk {
		return v.(ByteView), isOk
	}
	return
}
