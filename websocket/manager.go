package websocket

import (
	"encoding/json"
	syslog "log"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/jerbe/jim/log"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/16 10:38
  @describe :
*/

type Key interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~string
}

type mCA map[*websocket.Conn]any

type mSmCA map[any]mCA

// Manager websocket管理器
type Manager struct {
	sm      []mSmCA
	rwMux   *sync.RWMutex
	keyCnt  int64
	connCnt int64
}

// shardCount 分片数量
var shardCount = 16 // runtime.NumCPU()

// simpleLoadBalancingIndex 简单的负载均衡获取方式
func simpleLoadBalancingIndex(key any) int {
	var id int64
	switch key.(type) {
	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64:
		id = key.(int64)
	case string:
		id = int64(key.(string)[0] + key.(string)[len(key.(string))-1]) // 简单的负载分类法
	}

	return int(id) % shardCount
}

// AddConnect 添加一条网络连接
func (wm *Manager) AddConnect(key any, conn *websocket.Conn) {
	wm.rwMux.Lock()
	defer wm.rwMux.Unlock()

	i := simpleLoadBalancingIndex(key)
	s, ok := wm.sm[i][key]
	if ok {
		s[conn] = struct{}{}
	} else {
		atomic.AddInt64(&wm.keyCnt, 1)
		s = mCA{
			conn: struct{}{},
		}
		wm.sm[i][key] = s
	}

	atomic.AddInt64(&wm.connCnt, 1)

	syslog.Println("AddConnect => 现在连接数有:", wm.connCnt)
}

// RemoveConnect 删除一条网络连接
func (wm *Manager) RemoveConnect(key string, conn *websocket.Conn) {
	wm.rwMux.Lock()
	defer wm.rwMux.Unlock()
	i := simpleLoadBalancingIndex(key)
	s, ok := wm.sm[i][key]
	if ok {
		atomic.AddInt64(&wm.connCnt, -1)
		delete(s, conn)
		if len(s) == 0 {
			atomic.AddInt64(&wm.keyCnt, -1)
			delete(wm.sm[i], key)
		}
	}
	syslog.Println("RemoveConnect:= => 现在连接数有:", wm.connCnt)
}

// PushMessage 推送消息到网络连接上
func (wm *Manager) PushMessage(msg []byte, keys ...string) {
	wm.rwMux.RLock()
	defer wm.rwMux.RUnlock()

	writeMessage := func(conn *websocket.Conn) {
		defer func() {
			if obj := recover(); obj != nil {
				log.Error().Str("function", "Manager.PushMessage.writeMessage").Any("panic", obj).Str("remote_addr", conn.RemoteAddr().String()).Msg("推送消息到客户端发生panic")
			}
		}()

		err := conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Error().Err(err).Str("function", "Manager.PushMessage.writeMessage").Str("remote_addr", conn.RemoteAddr().String()).Msg("推送消息到客户端发生错误")
			return
		}
	}

	// 设置了IDS
	if len(keys) > 0 {
		const step = 200
		steps := len(keys) / step
		mod := len(keys) % step
		if mod > 0 {
			steps++
		}

		writeFun := func(ks []string) {
			for j := 0; j < len(ks); j++ {
				key := ks[j]
				idx := simpleLoadBalancingIndex(key)
				mca, ok := wm.sm[idx][key]
				if ok {
					for conn, _ := range mca {
						go writeMessage(conn)
					}
				}
			}
		}

		// 预先消耗掉一些
		writeFun(keys[0:mod])

		// 这步很重要
		steps--

		// 消耗剩余的
		for i := 0; i < steps; i++ {
			start := i + mod
			end := (i+1)*steps + mod
			writeFun(keys[start:end])
		}

		return
	}

	// 未设置ids
	if l := len(keys); l == 0 {
		for _, msmca := range wm.sm {
			go func(msmca mSmCA) {
				for _, mca := range msmca {
					for conn, _ := range mca {
						go writeMessage(conn)
					}
				}
			}(msmca)
		}
	}
}

// PushJson 推送json格式的数据到各个ws连接上
func (wm *Manager) PushJson(data any, keys ...any) {
	wm.rwMux.RLock()
	defer wm.rwMux.RUnlock()

	msg, _ := json.Marshal(data)

	writeMessage := func(conn *websocket.Conn) {
		defer func() {
			if obj := recover(); obj != nil {
				log.Error().Str("function", "Manager.PushMessage.writeMessage").Any("panic", obj).Str("remote_addr", conn.RemoteAddr().String()).Msg("推送消息到客户端发生panic")
			}
		}()

		err := conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Error().Err(err).Str("function", "Manager.PushMessage.writeMessage").Str("remote_addr", conn.RemoteAddr().String()).Msg("推送消息到客户端发生错误")
			return
		}
	}

	// 设置了IDS
	if len(keys) > 0 {
		const step = 200
		steps := len(keys) / step
		mod := len(keys) % step
		if mod > 0 {
			steps++
		}

		writeFun := func(ks []any) {
			for j := 0; j < len(ks); j++ {
				key := ks[j]
				idx := simpleLoadBalancingIndex(key)
				mca, ok := wm.sm[idx][key]
				if ok {
					for conn, _ := range mca {
						go writeMessage(conn)
					}
				}
			}
		}

		// 预先消耗掉一些
		writeFun(keys[0:mod])

		// 这步很重要
		steps--

		// 消耗剩余的
		for i := 0; i < steps; i++ {
			start := i + mod
			end := (i+1)*steps + mod
			writeFun(keys[start:end])
		}

		return
	}

	// 未设置ids
	if l := len(keys); l == 0 {
		for _, msmca := range wm.sm {
			go func(msmca mSmCA) {
				for _, mca := range msmca {
					for conn, _ := range mca {
						go writeMessage(conn)
					}
				}
			}(msmca)
		}
	}
}

// KeysCount 获取所有键数量
func (wm *Manager) KeysCount() int64 {
	return wm.keyCnt
}

// ConnectCount 所以网络连接数量
func (wm *Manager) ConnectCount() int64 {
	return wm.connCnt
}

func NewManager() *Manager {
	return &Manager{
		sm: func() []mSmCA {
			numCpu := shardCount
			slice := make([]mSmCA, numCpu, numCpu)
			for i := 0; i < numCpu; i++ {
				slice[i] = make(mSmCA)
			}
			return slice
		}(),
		rwMux: new(sync.RWMutex),
	}
}

// DefaultManager 默认管理器
var DefaultManager = NewManager()
