# MisakaCache

本项目是以极客兔兔的[GeeCache](https://geektutu.com/post/geecache.html)项目为原型，使用Go语言实现的一个分布式缓存系统，目前使用LRU作为缓存淘汰策略，节点间使用HTTP+ProtocolBuffers进行通信，并且有并发保护。

本项目即将进行的改进如下：

- [ ] 新建一个对外的API节点 专用于和缓存节点进行通信
- [ ] 多个淘汰策略，比如LFU、ARC
- [ ] HTTP通信改为RPC通信
- [ ] 细化锁的粒度来提高并发性能
- [ ] 实现热点互备来避免热点数据频繁请求影响性能
- [ ] 加入etcd
- [ ] 加入缓存过期机制