package misakacache

import (
	"MisakaCache/src/misakacache/consistenthash"
	pb "MisakaCache/src/misakacache/misakacachepb"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/" // 默认资源地址
	defaultReplicas = 50            // 默认真实节点和虚拟节点倍数
)

// HTTPPool 一个缓存对应一个HTTP池 记录自身的地址和URL
type HTTPPool struct {
	selfAddr    string // 记录缓存自身的地址 包括端口
	basePath    string // 记录URL
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*httpClient // 集成一个HTTP客户端
}

// NewHTTPPool HTTPPool的构造方法
func NewHTTPPool(selfAddr string) (result *HTTPPool) {
	result = &HTTPPool{
		selfAddr: selfAddr,
		basePath: defaultBasePath,
	}
	return
}

// Log 记录信息 参数v可传多个值 这些值会按format来进行格式化 再进入log
func (pool *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", pool.selfAddr, fmt.Sprintf(format, v...))
}

// ServeHTTP HTTP服务端的Handler方法 对有效的缓存请求进行响应
func (pool *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, pool.basePath) { // 检查请求是否有效
		panic("HTTPPool is serving unexpected path: " + r.URL.Path)
	}
	pool.Log("%s %s", r.Method, r.URL.Path) // log记录该次请求的信息

	parts := strings.SplitN(r.URL.Path[len(pool.basePath):], "/", 2)
	if len(parts) != 2 { // 检查请求是否有效
		http.Error(w, "bad request", http.StatusBadRequest) // 400
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil { // 请求的缓存不存在
		http.Error(w, "no such group:"+groupName, http.StatusNotFound) // 404
		return
	}

	view, err := group.GetFromCache(key) // fixme
	if err != nil {                      // 缓存请求失败
		http.Error(w, err.Error(), http.StatusInternalServerError) // 500
		return
	}

	body, err := proto.Marshal(&pb.Response{Value: view.GetByteCopy()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = w.Write(body)
	if err != nil { // 写入响应失败
		http.Error(w, "write into responseWriter error: "+err.Error(), http.StatusInternalServerError) // 500
		return
	}
}

// SetNewPeer 在本节点初始化远程节点信息
func (p *HTTPPool) SetNewPeer(peers ...string) {
	p.mu.Lock() // attention 加锁必要性?
	defer p.mu.Unlock()

	p.peers = consistenthash.NewMap(nil, defaultReplicas)
	p.peers.AddRealNode(peers...)
	p.httpGetters = make(map[string]*httpClient, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpClient{baseURL: peer + p.basePath} // attention 这里的路径构建可能会有问题
	}
}

// PickPeer 根据一致性哈希挑选合适的远程节点
func (p *HTTPPool) PickPeer(key string) (PeerCacheValueGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	peer := p.peers.GetRealNodeByKey(key)
	if peer != "" && peer != p.selfAddr { // 注意这里 这里排除了自身节点
		p.Log("PickPeer picked %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)

// httpClient HTTP客户端 向远程节点发送请求 一个远程节点对应一个HTTP客户端
type httpClient struct {
	baseURL string
}

// GetCacheFromPeer 实现PeerCacheValueGetter接口 从远程节点获得缓存
func (h *httpClient) GetCacheFromPeer(in *pb.Request, out *pb.Response) error {
	URL := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(in.GetGroup()), url.QueryEscape(in.GetKey()))

	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("server returned error: %v", resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body) // 读取响应体 原有的ioutil.ReadAll方法被弃用
	if err != nil {
		err = fmt.Errorf("reading response body error: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}

var _ PeerCacheValueGetter = (*httpClient)(nil) // 检查接口是否被完整实现
