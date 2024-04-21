package lru

import "container/list"

/*
LRU算法 是一种介于FIFO算法和LFU算法之间的缓存算法

FIFO算法 顾名思义 先进先出 当缓存满时会自动删除最先进入缓存的数据
但是每次被删除的数据也有可能是最常访问的数据 这一模式并不能很好地匹配缓存的需求

LFU算法 Least Frequently Used 最少使用算法 每次缓存满时会删除最久未使用过的缓存
这种情况能够很好地满足缓存的需求 但是它所占用的开销算是很大的：不仅要存键值对缓存 还要存这个缓存对应的访问次数 而且每次删除缓存时都需要对访问次数排序
不仅如此 如果对缓存的访问形式突然发生改变 以前经常访问的缓存以后都不再访问 那么要淘汰以前访问次数多的缓存也需要很长的时间

LRU算法 Least Recently Used 最近最少使用 相对于仅考虑时间的FIFO和仅考虑频率的LFU 该算法认为 如果一个数据被访问了 那么它在未来的一段时间访问的概率就会很高

LRU算法由一个字典和一个队列组成 字典保存缓存具体数据 队列保存访问频率 每次删除数据时 默认删除队列出口的数据 当一个数据被访问时 就将其重新入队

具体到Go语言上 组成LRU算法的就是一个字典+由双向链表构成的队列
*/

// LRU 采用了LRU策略的缓存类
type LRU struct {
	maxMemoryBytes  int64                         // 允许使用的最大内存
	memoryUsedBytes int64                         // 当前已经使用的内存
	queue           *list.List                    // 队列 存储访问频率
	cacheMap        map[string]*list.Element      // 字典 存储具体的缓存的键值对 这里的值存的是双向链表的元素指针类型 具体的值在链表里
	OnEntryDeleted  func(key string, value Value) // 当缓存被删除时的回调函数 这种函数类型的默认值就是nil
}

// NewLRU LRU的构造函数
func NewLRU(maxMemoryBytes int64, onEntryDeleted func(string, Value)) (cache *LRU) {
	cache = &LRU{
		maxMemoryBytes: maxMemoryBytes,
		queue:          list.New(),
		cacheMap:       make(map[string]*list.Element),
		OnEntryDeleted: onEntryDeleted,
	}
	return
}

// Value 一条键值对缓存具体要存储的值的接口 该接口的目的是为了确保值对所有类型的通用性
type Value interface {
	GetMemoryUsed() int // 返回该缓存的值所占用的内存大小
}

// entry Cache中双向链表所存储的类型
type entry struct {
	key string
	// 之所以在双向链表中存储键 是为了删除更加方便 能够直接根据键去字典中删除元素
	value Value // 值必须是实现了Value接口的类型
}

// 开始实现功能

// GetValue 在缓存里查找值
func (cache *LRU) GetValue(key string) (value Value, isOk bool) {
	if element, isOk := cache.cacheMap[key]; isOk {
		cache.queue.MoveToFront(element) // 如果元素存在 则移动到队尾 注意双向链表的队尾是相对的 规定Front是队尾即可
		entry := element.Value.(*entry)  // 这个节点的值取出来 转换为entry类型
		return entry.value, true
	}
	value = nil
	isOk = false
	return
}

// RemoveOldestCache 淘汰一次缓存
func (cache *LRU) RemoveOldestCache() {
	element := cache.queue.Back() // 取出队首元素
	if element != nil {
		cache.queue.Remove(element) // 从双向链表中移除该元素
		cacheEntry := element.Value.(*entry)
		delete(cache.cacheMap, cacheEntry.key)                                                        // 从字典中移除该键值对
		cache.memoryUsedBytes -= int64(len(cacheEntry.key)) + int64(cacheEntry.value.GetMemoryUsed()) // 修改已使用的内存
		if cache.OnEntryDeleted != nil {                                                              // 如果回调函数不为空 则调用回调函数
			cache.OnEntryDeleted(cacheEntry.key, cacheEntry.value)
		}
	}
}

// SetValue 添加/修改缓存
func (cache *LRU) SetValue(key string, value Value) {
	if element, isOk := cache.cacheMap[key]; isOk { // 键存在
		cache.queue.MoveToFront(element)
		cacheEntry := element.Value.(*entry)
		cache.memoryUsedBytes += int64(value.GetMemoryUsed()) - int64(cacheEntry.value.GetMemoryUsed())
		cacheEntry.value = value
	} else { // 键不存在
		element = cache.queue.PushFront(&entry{key: key, value: value})
		cache.cacheMap[key] = element
		cache.memoryUsedBytes += int64(len(key)) + int64(value.GetMemoryUsed())
	}
	for cache.memoryUsedBytes > cache.maxMemoryBytes && cache.memoryUsedBytes != 0 { // 看已经使用的缓存内存有多大来淘汰旧缓存
		cache.RemoveOldestCache()
	}
}

// GetLRUEntryNumber 返回当前已缓存的键值对数量
func (cache *LRU) GetLRUEntryNumber() (len int) {
	len = cache.queue.Len()
	return
}
