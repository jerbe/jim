package pubsub

import (
	"context"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/24 16:19
  @describe :
*/

// ChatMessage 订阅传输用的聊天消息结构
type ChatMessage struct {
	// ActionID 行为ID
	ActionID string `json:"action_id"`

	// ReceiverID 接收人; 可以是用户ID,也可以是房间号
	ReceiverID int64 `json:"receiver_id"`

	// SessionType 会话类型; 1:私聊, 2:群聊
	SessionType int `json:"session_type"`

	// Type 消息类型: 1-纯文本,2-图片,3-语音,4-视频, 5-位置
	Type int `json:"type"`

	// 发送人ID
	SenderID int64 `json:"sender_id"`

	// MessageID 消息ID
	MessageID int64 `json:"message_id"`

	// CreatedAt 创建
	CreatedAt int64 `json:"created_at"`

	// Body 消息体;
	Body ChatMessageBody `json:"body"`

	// PublishTargets 群成员ID列表
	PublishTargets []int64 `json:"publish_targets,omitempty"`
}

// ChatMessageBody 消息主体
type ChatMessageBody struct {
	// 文本信息。适用消息类型: 1
	Text string `bson:"text,omitempty" json:"text,omitempty"`

	// 来源地址。通用字段，适用消息类型: 2,3,4
	Src string `bson:"src,omitempty" json:"src,omitempty"`

	// 文件格式。适用消息类型: 2,3,4
	Format string `bson:"format,omitempty" json:"format,omitempty"`

	// 文件大小。适用消息类型: 2,3,4
	Size string `bson:"size,omitempty" json:"size,omitempty"`

	// 位置信息-经度。 适用消息类型: 5
	Longitude string `bson:"longitude,omitempty" json:"longitude,omitempty"`

	// 位置信息-纬度。 适用消息类型: 5
	Latitude string `bson:"latitude,omitempty" json:"latitude,omitempty"`

	// 位置信息-地图缩放级别。 适用消息类型: 5
	Scale float64 `bson:"scale,omitempty" json:"scale,omitempty"`

	// 位置信息标签。适用消息类型: 5
	LocationLabel string `bson:"location_label,omitempty" json:"location_label,omitempty"`
}

// PublishChatMessage 发布聊天消息到其他服务器上
func PublishChatMessage(ctx context.Context, data *ChatMessage) error {
	return PublishWithPayload(ctx, ChannelChatMessage, PayloadTypeChatMessage, data)
}
