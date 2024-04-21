package misakacache

import (
	pb "MisakaCache/src/misakacache/misakacachepb"
	"MisakaCache/src/misakacache/singleflight"
	"fmt"
	"log"
	"sync"
)

// Getter 接口 规定了一个Get方法 该方法用于规定缓存未命中时从哪里获得新的缓存
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 函数类型 专门用来实现Getter接口的函数类型
type GetterFunc func(key string) ([]byte, error)

// Get GetterFunc类下的 从Getter接口的Get函数实现而来的函数
func (function GetterFunc) Get(key string) ([]byte, error) {
	return function(key)
}

// Group 缓存对外交互的核心数据结构
type Group struct {
	name      string     // 该缓存的标识
	getter    Getter     // 缓存未能命中时的回调函数 类型是Getter接口
	mainCache cache      // 缓存主体 是具有并发保护的LRU缓存
	peers     PeerPicker // 这是实现了PeerPicker的HTTPPool
	// attention 为什么要将远程节点集成进HTTPPool 而不是节点本身？ 是否可以优化？

	loader *singleflight.Group // 非本地缓存的并发请求管理
}

// 全局变量
var (
	mu     sync.RWMutex              // 互斥锁的高级版本 读写锁 在原有的锁功能上 增加读写特性 当读锁锁定时 只会阻止写 同理当写锁锁定时 会同时阻止读写
	groups = make(map[string]*Group) // 保存多个Group缓存
)

// NewGroup 构造函数
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()
	group := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = group
	return group
}

// GetGroup 获取Group缓存
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// GetFromCache 从缓存中获取值
func (g *Group) GetFromCache(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, isOk := g.mainCache.get(key); isOk { // 缓存命中
		log.Println("[MisakaCache] hit")
		return v, nil
	}
	// 缓存未命中 调用load函数
	return g.load(key)
}

// load 缓存未命中时 从别的地方加载缓存
func (g *Group) load(key string) (value ByteView, err error) {
	viewi, err := g.loader.DoFunc(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok { // 先从存储着远程节点信息的HTTPPool中选出具体的远程节点
				if value, err := g.getFromPeer(peer, key); err != nil { // 再根据这个具体的远程节点开始请求
					return value, nil
				}
				log.Println("[MisakaCache] Failed to get from peer", err)
			}
		}
		return g.getFromLocal(key)
	})

	if err == nil {
		return viewi.(ByteView), err
	}
	return
}

// getFromLocal 从本地加载缓存 在这里调用Getter的Get函数 并且通过populateCache存入缓存
func (g *Group) getFromLocal(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{cacheBytes: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// populateCache 新的缓存值 存入缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// RegisterPeers 将初始化完成的HTTPPool注入到group中 仅一次
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// getFromPeer 从远程节点获得缓存
func (g *Group) getFromPeer(peer PeerCacheValueGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Key:   key,
		Group: g.name,
	}

	resp := &pb.Response{}

	err := peer.GetCacheFromPeer(req, resp)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{cacheBytes: resp.Value}, nil
}
