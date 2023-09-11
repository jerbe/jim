package pubsub

import (
	"context"
	"sync"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/24 16:19
  @describe :
*/

var chatMessagePool = &sync.Pool{
	New: func() any {
		return new(ChatMessage)
	},
}

var chatMessageBodyPool = &sync.Pool{
	New: func() any {
		return new(ChatMessageBody)
	},
}

func NewChatMessage() *ChatMessage {
	msg := chatMessagePool.Get().(*ChatMessage)
	msg.ActionID = ""
	msg.ReceiverID = 0
	msg.SessionType = 0
	msg.Type = 0
	msg.SenderID = 0
	msg.MessageID = 0
	msg.CreatedAt = 0
	msg.Body = nil
	msg.PublishTargets = nil
	return msg
}

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
	Body *ChatMessageBody `json:"body"`

	// PublishTargets 推送目标列表
	// 为什么增加 PublishTargets 这个参数?
	// 因为分布式中,会多个服务实例都订阅到该方法,将导致多个服务实例再去查询数据库,比方说群成员列表等,所以预先加入 PublishTargets .
	// 订阅者可以直接从传参拿需要推送的目标 ,能尽量少请求数据库就尽量少请求
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

func NewChatMessageBody() *ChatMessageBody {
	body := chatMessageBodyPool.Get().(*ChatMessageBody)
	body.Text = ""
	body.Src = ""
	body.Format = ""
	body.Size = ""
	body.Longitude = ""
	body.Latitude = ""
	body.Scale = 0
	body.LocationLabel = ""
	return body
}

// PublishChatMessage 发布聊天消息到其他服务器上
func PublishChatMessage(ctx context.Context, data *ChatMessage) error {
	err := PublishWithPayload(ctx, ChannelChatMessage, PayloadTypeChatMessage, data)
	chatMessagePool.Put(data)
	return err
}
