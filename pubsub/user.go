package pubsub

import "time"

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/24 16:21
  @describe :
*/

// FriendInvite 订阅服务传输使用的好友请求结果
type FriendInvite struct {
	// ID 邀请ID
	ID int64 `json:"id"`

	// UserID 用户ID
	UserID int64 `json:"user_id"`

	// TargetID 目标ID
	TargetID int64 `json:"target_id"`

	// Status 状态
	Status int `json:"type"`

	// Note 备注
	Note string `json:"note"`

	// Reply 回复
	Reply string `json:"reply"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`
}
