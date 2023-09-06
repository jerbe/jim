package pubsub

import (
	"context"
	"fmt"
	"github.com/jerbe/jim/log"
	"github.com/redis/go-redis/v9"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/25 10:35
  @describe :
*/

// SubscribeHandlerFunc 订阅处理方法
type SubscribeHandlerFunc func(context.Context, *Payload)

// subscriber 订阅器
type subscriber struct {
	// m
	m map[string]SubscribeHandlerFunc

	// chs 通道名称数据
	chs map[string]any
}

func (s *subscriber) genKey(channel, typ string) string {
	return fmt.Sprintf("%s:%s", channel, typ)
}

// receiveHandler 订阅接收处理中转站
func (s *subscriber) receiveHandler(ctx context.Context, msg *redis.Message) {
	payload := &Payload{}
	payloadBinary := []byte(msg.Payload)
	err := payload.UnmarshalBinary(payloadBinary)
	if err != nil {
		log.Error().Err(err).
			Str("channel", msg.Channel).
			Str("payload", msg.Payload). // @todo 此处为敏感数据,上线前删除
			Msg("解码payload失败")
		return
	}
	payload.Channel = msg.Channel
	go s.Do(ctx, payload)
}

func (s *subscriber) beforeSubscribe(channel, typ string, fn SubscribeHandlerFunc) {
	if _, ok := s.chs[channel]; ok {
		return
	}
	DefaultPubsuber().Subscribe(context.Background(), channel, s.receiveHandler)
	s.chs[channel] = struct{}{}
}

// Subscribe 订阅
func (s *subscriber) Subscribe(channel, typ string, fn SubscribeHandlerFunc) {
	s.beforeSubscribe(channel, typ, fn)
	subKey := s.genKey(channel, typ)
	if _, ok := s.m[subKey]; ok {
		log.Warn().Msgf("already subscribe %s", subKey)
		return
	}
	s.m[subKey] = fn
	return
}

// Do 执行
func (s *subscriber) Do(ctx context.Context, payload *Payload) {
	subKey := s.genKey(payload.Channel, payload.Type)
	fn, ok := s.m[subKey]
	if !ok {
		log.Warn().Msgf("never goSubscribe %s", subKey)
		return
	}
	defer func() {
		if obj := recover(); obj != nil {
			log.Error().Str("obj", fmt.Sprintf("%+v", obj)).Msg("recover")
		}
	}()
	fn(ctx, payload)

	return
}

func NewSubscriber() *subscriber {
	sm := &subscriber{
		m:   make(map[string]SubscribeHandlerFunc),
		chs: make(map[string]any),
	}
	return sm
}
