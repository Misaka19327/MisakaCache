package singleflight

import "sync"

// call 含义为正在进行中 或者是已经结束的请求
type call struct {
	wg  sync.WaitGroup // 避免重入
	val interface{}
	err error
}

// Group 对call进行管理
type Group struct {
	mu sync.Mutex // 互斥锁 这个锁是保护下面的m的
	m  map[string]*call
}

// DoFunc 要保护的方法调用入口 针对同样的key 传入的fn只会被调用一次 直到第一次fn的调用完成
func (g *Group) DoFunc(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock() // 加锁

	if g.m == nil { // 如果没初始化的话先初始化
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok { // 已经存在之前的请求 则等待之前的请求完成 直接返回之前的请求结果
		g.mu.Unlock() // 读m完成 解锁
		c.wg.Wait()   // 等待信号量归零
		return c.val, c.err
	}

	// 如果不存在相同的请求 则新增请求
	c := new(call)
	c.wg.Add(1)   // 开始调用 信号量+1
	g.m[key] = c  // 写入m
	g.mu.Unlock() // 读写m完成 解锁

	c.val, c.err = fn() // 实际调用
	c.wg.Done()         // 调用已完成 信号量-1

	g.mu.Lock()      // 加锁
	delete(g.m, key) // 请求结束 删除
	g.mu.Unlock()    // 写入m完成 解锁

	return c.val, c.err
}
