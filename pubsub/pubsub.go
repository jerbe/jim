package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jerbe/jim/errors"
	"github.com/jerbe/jim/log"
	"github.com/redis/go-redis/v9"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/18 09:58
  @describe :
*/

var defaultPubsuber *pubsuber

func DefaultPubsuber() *pubsuber {
	return defaultPubsuber
}

type pubsuber struct {
	redisCli redis.UniversalClient
	channels map[string]any
}

// goSubscribe 协程用的订阅方法
func (p *pubsuber) goSubscribe(ctx context.Context, channel string, fn func(context.Context, *redis.Message)) {
	if p.redisCli == nil {
		return
	}
	suber := p.redisCli.Subscribe(ctx, channel)
	defer func() {
		if obj := recover(); obj != nil {
			log.Error().Str("recover", fmt.Sprintf("%+v", obj)).Str("channel", channel).Msg("recover")
			go p.goSubscribe(ctx, channel, fn)
		}
		err := suber.Close()
		if err != nil {
			log.Error().Err(err).Str("channel", channel).Msg("关闭redis订阅失败")
			return
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-suber.Channel():
			go func(c context.Context, m *redis.Message) {
				defer func() {
					if obj := recover(); obj != nil {
						log.Error().Str("obj", fmt.Sprintf("%+v", obj)).Str("channel", channel).Msg("recover")
					}
				}()
				fn(c, m)
			}(ctx, msg)
		}
	}

}

// Subscribe 订阅频道
func (p *pubsuber) Subscribe(ctx context.Context, channel string, fn func(context.Context, *redis.Message)) {
	if p.redisCli == nil {
		return
	}
	if _, ok := p.channels[channel]; ok {
		return
	}
	p.channels[channel] = struct{}{}

	go p.goSubscribe(ctx, channel, fn)
}

// Publish 往频道内推送消息
func (p *pubsuber) Publish(ctx context.Context, chanel string, message any) error {
	if p.redisCli == nil {
		return errors.New("redis client is nil")
	}
	cmd := p.redisCli.Publish(ctx, chanel, message)
	return cmd.Err()
}

// Payload 推送订阅的有效谁
type Payload struct {
	// Channel 通道
	Channel string `json:"channel,omitempty"`

	// Type 类型
	Type string `json:"type"`

	// Data 具体数据,这里为什么不使用any,因为设置成any最后还需要映射成原来的struct,比较麻烦
	Data string `json:"data"`
}

// MarshalBinary 实现 encoding.BinaryMarshaler 接口
func (p *Payload) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

// UnmarshalBinary 实现 encoding.UnmarshalBinary 接口
func (p *Payload) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}

// UnmarshalData 解码data里面的数据
func (p *Payload) UnmarshalData(dest any) error {
	return json.Unmarshal([]byte(p.Data), dest)
}

// PublishWithPayload 发布推送数据
func PublishWithPayload(ctx context.Context, channel, typ string, data any) error {
	if defaultPubsuber == nil {
		return errors.New("default pubsuber is nil")
	}
	binary, err := json.Marshal(data)
	if err != nil {
		return err
	}

	payload := &Payload{
		Type: typ,
		Data: string(binary),
	}
	return defaultPubsuber.Publish(ctx, channel, payload)
}
