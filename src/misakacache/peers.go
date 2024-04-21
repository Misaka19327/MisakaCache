package misakacache

import pb "MisakaCache/src/misakacache/misakacachepb"

// PeerPicker 接口 根据key挑选远程节点
type PeerPicker interface {
	PickPeer(key string) (peerGetter PeerCacheValueGetter, ok bool)
}

// PeerCacheValueGetter 接口 根据key和给定的group获取缓存值
type PeerCacheValueGetter interface {
	GetCacheFromPeer(in *pb.Request, out *pb.Response) error
}
