package pubsub

import (
	"context"

	"github.com/jerbe/jim/log"

	"github.com/redis/go-redis/v9"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/24 16:21
  @describe :
*/

// PublishNotifyMessage 发布聊天消息到其他服务器上
func PublishNotifyMessage(ctx context.Context, typ string, data any) error {
	return PublishWithPayload(ctx, ChannelNotify, typ, data)
}

// notifyMessageHandler 接收通知消息
func notifyMessageHandler(ctx context.Context, msg *redis.Message) {
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

	switch payload.Type {
	case PayloadTypeFriendInvite:

	}

	log.Info().Any("payload", payload).Send()
}
