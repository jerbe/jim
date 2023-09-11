package websocket

import (
	"encoding"
	"encoding/json"
	"fmt"
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
	keyCnt  uint64
	connCnt uint64
}

// shardCount 分片数量
var shardCount = 16 // runtime.NumCPU()

// simpleLoadBalancingIndex 简单的负载均衡获取方式
func simpleLoadBalancingIndex(key any) int {
	if key == nil {
		panic("不能空指针")
	}

	var id int
	switch k := key.(type) {
	case *uint:
		id = int(*k)
	case *uint8:
		id = int(*k)
	case *uint16:
		id = int(*k)
	case *uint32:
		id = int(*k)
	case *uint64:
		id = int(*k)
	case *int:
		id = *k
	case *int8:
		id = int(*k)
	case *int16:
		id = int(*k)
	case *int32:
		id = int(*k)
	case *int64:
		id = int(*k)

	case uint:
		id = int(k)
	case uint8:
		id = int(k)
	case uint16:
		id = int(k)
	case uint32:
		id = int(k)
	case uint64:
		id = int(k)
	case int:
		id = k
	case int8:
		id = int(k)
	case int16:
		id = int(k)
	case int32:
		id = int(k)
	case int64:
		id = int(k)

	case *string:
		id = int((*k)[0] + (*k)[len(*k)-1])
	case string:
		id = int(k[0] + k[len(k)-1])
	case fmt.Stringer:
		str := k.String()
		id = int(str[0] + str[len(str)-1])
	default:
		panic("无效的key类型")
	}

	return id % shardCount
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
		atomic.AddUint64(&wm.keyCnt, 1)
		s = mCA{
			conn: struct{}{},
		}
		wm.sm[i][key] = s
	}

	atomic.AddUint64(&wm.connCnt, 1)
}

// RemoveConnect 删除一条网络连接
func (wm *Manager) RemoveConnect(key string, conn *websocket.Conn) {
	wm.rwMux.Lock()
	defer wm.rwMux.Unlock()
	i := simpleLoadBalancingIndex(key)
	s, ok := wm.sm[i][key]
	if ok {
		atomic.StoreUint64(&wm.connCnt, atomic.LoadUint64(&wm.connCnt)-1)
		delete(s, conn)
		if len(s) == 0 {
			atomic.StoreUint64(&wm.keyCnt, atomic.LoadUint64(&wm.keyCnt)-1)
			delete(wm.sm[i], key)
		}
	}
}

// PushMessage 推送消息到网络连接上
func (wm *Manager) PushMessage(msg []byte, keys ...any) {
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

		// 预先消耗余数
		writeFun(keys[0:mod])

		// 这步很重要
		steps--
		// 再消耗剩余的倍数
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

// PushData 推送传入的对象进行解析,并发到各个ws连接上
func (wm *Manager) PushData(data any, keys ...any) {
	var msg []byte
	var err error
	switch v := data.(type) {
	case encoding.BinaryMarshaler:
		msg, err = v.MarshalBinary()
	case encoding.TextMarshaler:
		msg, err = v.MarshalText()
	case json.Marshaler:
		msg, err = v.MarshalJSON()
	default:
		msg, err = json.Marshal(v)
	}

	if err != nil {
		log.Error().Err(err).Str("function", "Manager.PushData").Str("data_type", fmt.Sprintf("%T", data)).Msg("编码数据报错")
		return
	}

	wm.PushMessage(msg, keys...)
}

// KeysCount 获取所有键数量
func (wm *Manager) KeysCount() uint64 {
	return wm.keyCnt
}

// ConnectCount 所以网络连接数量
func (wm *Manager) ConnectCount() uint64 {
	return wm.connCnt
}

func NewManager() *Manager {
	return &Manager{
		sm: func() []mSmCA {
			slice := make([]mSmCA, shardCount)
			for i := 0; i < shardCount; i++ {
				slice[i] = make(mSmCA)
			}
			return slice
		}(),
		rwMux: new(sync.RWMutex),
	}
}

// DefaultManager 默认管理器
var DefaultManager = NewManager()
