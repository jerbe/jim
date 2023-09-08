package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jerbe/jim/errors"
	"github.com/jerbe/jim/pubsub"
	"github.com/jerbe/jim/websocket"
	"time"

	"github.com/jerbe/jim/database"
	"github.com/jerbe/jim/log"
	"github.com/jerbe/jim/utils"

	"github.com/gin-gonic/gin"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/15 16:41
  @describe :
*/

// ChatMessage 消息体
// @Description 消息体
type ChatMessage struct {
	// ID 消息ID
	ID string `json:"id" example:"9d7a3bcd72"`

	// ActionID 行为ID,由前端生成
	ActionID string `json:"action_id,omitempty" example:"8d7a3bcd72"`

	// SessionType 会话类型; 1:私聊, 2:群聊
	SessionType int `json:"session_type" binding:"required" enums:"1,2" example:"1"`

	// Type 消息类型; 1-纯文本,2-图片,3-语音,4-视频, 5-位置
	Type int `json:"type" enums:"1,2,3,4,5" binding:"required" example:"1"`

	// SenderID 发送方ID
	SenderID int64 `json:"sender_id" example:"1234456"`

	// ReceiverID 接收方ID
	ReceiverID int64 `json:"receiver_id" example:"1"`

	// MessageID 消息ID
	MessageID int64 `json:"message_id" example:"123"`

	// CreatedAt 创建
	CreatedAt int64 `json:"created_at"example:"12345678901234"`

	// Body 消息体;
	Body ChatMessageBody `json:"body" binding:"required"`
}

func (cm *ChatMessage) MarshalBinary() ([]byte, error) {
	return json.Marshal(cm)
}

func (cm *ChatMessage) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, cm)
}

// ChatMessageBody 消息主体
// @Description 消息主体
type ChatMessageBody struct {
	// 文本信息。适用消息类型: 1
	Text string `json:"text,omitempty" example:"这是一条聊天文案"`

	// 来源地址。通用字段，适用消息类型: 2,3,4
	Src string `json:"src,omitempty" example:"https://www.baidu.com/logo.png"`

	// 文件格式。适用消息类型: 2,3,4
	Format string `json:"format,omitempty" example:"jpeg"`

	// 文件大小。适用消息类型: 2,3,4
	Size string `json:"size,omitempty" example:"1234567890"`

	// 位置信息-经度。 适用消息类型: 5
	Longitude string `json:"longitude,omitempty" example:"0.213124212313"`

	// 位置信息-纬度。 适用消息类型: 5
	Latitude string `json:"latitude,omitempty" example:"0.913124212313"`

	// 位置信息-地图缩放级别。 适用消息类型: 5
	Scale float64 `json:"scale,omitempty" example:"0.22"`

	// 位置信息标签。适用消息类型: 5
	LocationLabel string `json:"location_label,omitempty" example:"成人影视学院"`
}

// SendChatMessageRequest 聊天发送消息请求参数
// @Description 聊天发送消息请求参数
type SendChatMessageRequest struct {
	// ActionID 行为ID,由前端生成
	ActionID string `json:"action_id" example:"8d7a3bcd72"`

	// SessionType 会话类型; 1:私聊, 2:群聊
	SessionType int `json:"session_type" binding:"required" enums:"1,2" example:"1"`

	// Type 消息类型; 1-纯文本,2-图片,3-语音,4-视频, 5-位置
	Type int `json:"type" enums:"1,2,3,4,5" binding:"required" example:"1"`

	// TargetID 目标ID; 可以是用户ID,也可以是群ID,也可以是世界频道ID
	TargetID int64 `json:"target_id" binding:"required" example:"1234"`

	// Body 消息体;
	Body ChatMessageBody `json:"body" binding:"required"`
}

// SendChatMessageHandler
// @Summary      发送聊天消息
// @Tags         聊天
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      SendChatMessageRequest  true  "请求JSON数据体"
// @Security 	 APIKeyHeader
// @Success      200  {object}  Response{data=[]ChatMessage}
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/chat/message/send [post]
func SendChatMessageHandler(ctx *gin.Context) {
	currentUser := LoginUserFromContext(ctx)

	req := &SendChatMessageRequest{}
	err := ctx.BindJSON(req)
	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	if !utils.In(req.SessionType, database.ChatMessageSessionTypePrivate, database.ChatMessageSessionTypeGroup, database.ChatMessageSessionTypeWorld) {
		JSONError(ctx, StatusError, MessageInvalidSessionType)
		return
	}

	if req.TargetID <= 0 {
		JSONError(ctx, StatusError, MessageInvalidReceiverID)
		return
	}

	if !(req.Type > 0 && req.Type <= 5) {
		JSONError(ctx, StatusError, MessageInvalidType)
		return
	}

	// 检验各个字段是否正确

	// 私聊状态
	if req.SessionType == database.ChatMessageSessionTypePrivate { //私聊
		sendChatMessageToFriend(ctx, req, currentUser)
		return
	}

	// 群聊
	if req.SessionType == database.ChatMessageSessionTypeGroup {
		sendChatMessageToGroup(ctx, req, currentUser)
		return
	}

	if req.SessionType == database.ChatMessageSessionTypeWorld {
		sendChatMessage(ctx, req, nil)
		return
	}

}

// sendChatMessage 发送聊天消息
func sendChatMessage(ctx *gin.Context, req *SendChatMessageRequest, pubSubMsgFn func(*pubsub.ChatMessage) error) {
	currentUser := LoginUserFromContext(ctx)
	targetID := req.TargetID

	roomID := ""
	switch req.SessionType {
	case database.ChatMessageSessionTypePrivate:
		// 私聊状态的房间号是按 用户ID排序分组
		roomID = utils.FormatPrivateRoomID(currentUser.ID, targetID)
	case database.ChatMessageSessionTypeGroup:
		roomID = utils.FormatGroupRoomID(targetID)
	case database.ChatMessageSessionTypeWorld:
		roomID = utils.FormatWorldRoomID(targetID)
	}

	now := time.Now()
	// 插入消息数据库
	msg := &database.ChatMessage{
		RoomID:      roomID,
		Type:        req.Type,
		SessionType: req.SessionType,
		SenderID:    currentUser.ID,
		ReceiverID:  targetID,
		SendStatus:  1,
		ReadStatus:  0,
		Status:      1,
		CreatedAt:   now.UnixMilli(),
		UpdatedAt:   now.UnixMilli(),
		Body: database.ChatMessageBody{
			Text:          req.Body.Text,
			Src:           req.Body.Src,
			Format:        req.Body.Format,
			Size:          req.Body.Size,
			Longitude:     req.Body.Longitude,
			Latitude:      req.Body.Latitude,
			Scale:         req.Body.Scale,
			LocationLabel: req.Body.LocationLabel,
		},
	}

	err := database.AddChatMessage(msg)
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).
			Int64("user_id", currentUser.ID).
			Int64("receiver_id", msg.ReceiverID).
			Int("session_type", msg.SessionType).
			Msg("添加聊天消息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	rsp := &ChatMessage{
		ID:          msg.ID.Hex(),
		ActionID:    req.ActionID,
		SessionType: msg.SessionType,
		Type:        msg.Type,
		SenderID:    currentUser.ID,
		ReceiverID:  msg.ReceiverID,
		MessageID:   msg.MessageID,
		CreatedAt:   msg.CreatedAt,
	}
	JSON(ctx, rsp)

	psData := fillChatMessageForPublish(rsp)

	if pubSubMsgFn != nil {
		if err = pubSubMsgFn(psData); err != nil {
			log.ErrorFromGinContext(ctx).Err(err).
				Str("err_format", fmt.Sprintf("%+v", err)).
				Int64("user_id", currentUser.ID).
				Int64("receiver_id", msg.ReceiverID).
				Int("session_type", msg.SessionType).
				Msg("执行推送消息操作方法失败")
			return
		}
	}

	err = pubsub.PublishChatMessage(ctx, psData)
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).
			Int64("user_id", currentUser.ID).
			Int64("receiver_id", msg.ReceiverID).
			Int("session_type", msg.SessionType).
			Msg("推送聊天消息到管道失败")

		//@ todo 需要重做推送
	}

	return
}

// sendChatMessageToFriend 向好友发送聊天消息
func sendChatMessageToFriend(ctx *gin.Context, req *SendChatMessageRequest, currentUser *database.User) {
	targetID := req.TargetID
	if currentUser.ID == req.TargetID {
		JSONError(ctx, StatusError, MessageChatYourself)
		return
	}

	// 检测与对方的关系
	relation, err := database.GetUserRelationByUsersID(currentUser.ID, targetID)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, MessageNotFriends)
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Msg("获取好友关系失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	// 把对方拉黑的
	/*
		if (relation.UserAID == currentUser.ID && relation.Status&2 == 0) ||
			(relation.FriendID == currentUser.ID && relation.Status&1 == 0) {
			JSONError(ctx, StatusError, "对方已被你拉黑")
			return
		}
	*/

	// 被对方删除
	if relation.Status != 0b11 {
		JSONError(ctx, StatusError, MessageNotFriends)
		return
	}

	// 被对方拉黑的
	if (relation.UserAID == currentUser.ID && relation.BlockStatus&0b01 == 0) ||
		(relation.UserBID == currentUser.ID && relation.BlockStatus&0b10 == 0) {
		JSONError(ctx, StatusError, MessageBlockYou)
		return
	}

	// 发送聊天消息
	sendChatMessage(ctx, req, nil)
	return
}

// sendChatMessageToGroup 发送群聊信息
func sendChatMessageToGroup(ctx *gin.Context, req *SendChatMessageRequest, currentUser *database.User) {

	targetID := req.TargetID

	// 获取群消息
	group, err := database.GetGroup(targetID)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, "找不到该群")
			return
		}

		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Msg("获取群信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	// 判断当前用户是否在群内
	member, err := database.GetGroupMember(targetID, currentUser.ID)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, "您不是该群成员")
			return
		}

		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Msg("获取群成员失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	// 不是群管理以上的成员就或提示
	if group.SpeakStatus == 0 && member.Role > 0 {
		JSONError(ctx, StatusError, "已全员禁言")
		return
	}

	if member.SpeakStatus == 0 {
		JSONError(ctx, StatusError, "您已经被禁言")
		return
	}

	sendChatMessage(ctx, req, func(message *pubsub.ChatMessage) error {
		// 先查出所有群成员ID,这样订阅到的实例无需再次获取群成员信息
		memberIDs, err := database.GetGroupMemberIDs(targetID)
		if err != nil && !errors.IsNoRecord(err) {
			return errors.Wrap(err)
		}
		message.PublishTargets = memberIDs
		return nil
	})
	return

}

// fillChatMessageForPublish 填充推送用的聊天消息
func fillChatMessageForPublish(rsp *ChatMessage) *pubsub.ChatMessage {
	msg := pubsub.NewChatMessage()
	msg.ActionID = rsp.ActionID
	msg.ReceiverID = rsp.ReceiverID
	msg.SessionType = rsp.SessionType
	msg.Type = rsp.Type
	msg.SenderID = rsp.SenderID
	msg.MessageID = rsp.MessageID
	msg.CreatedAt = rsp.CreatedAt

	msgBody := pubsub.NewChatMessageBody()
	msgBody.Text = rsp.Body.Text
	msgBody.Src = rsp.Body.Src
	msgBody.Format = rsp.Body.Format
	msgBody.Size = rsp.Body.Size
	msgBody.Longitude = rsp.Body.Longitude
	msgBody.Latitude = rsp.Body.Latitude
	msgBody.Scale = rsp.Body.Scale
	msgBody.LocationLabel = rsp.Body.LocationLabel

	msg.Body = msgBody
	return msg
}

// RollbackChatMessageHandler 回滚聊天消息处理方法
func RollbackChatMessageHandler(ctx *gin.Context) {

}

// DeleteChatMessageHandler 删除聊天消息处理方法
func DeleteChatMessageHandler(ctx *gin.Context) {

}

// GetLastChatMessagesRequest
// @Description 获取最后聊天消息列表请求参数
type GetLastChatMessagesRequest struct {
	// TargetID 目标ID; 朋友ID/群ID/世界频道ID
	TargetID int64 `form:"target_id" json:"target_id"`

	// SessionType 会话类型; 1-私人会话;2-群聊会话;99-世界频道会话
	SessionType int `form:"session_type" json:"session_type"`
}

// GetLastChatMessagesHandler
// @Summary      获取最近的聊天消息
// @Tags         聊天
// @Accept       json
// @Produce      json
// @Param        target_id    query      int  true  "目标ID; 朋友ID/群ID/世界频道ID"
// @Param        session_type    query      int  true  "会话类型; 1-私人会话;2-群聊会话;99-世界频道会话"
// @Security 	 APIKeyQuery
// @Success      200  {object}  Response{data=[]ChatMessage}
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/chat/message/last [get]
func GetLastChatMessagesHandler(ctx *gin.Context) {
	var req = new(GetLastChatMessagesRequest)
	err := ctx.BindQuery(req)
	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	if req.TargetID <= 0 {
		JSONError(ctx, StatusError, MessageInvalidTargetID)
		return
	}

	if !utils.In(req.SessionType, database.ChatMessageSessionTypePrivate, database.ChatMessageSessionTypePrivate, database.ChatMessageSessionTypePrivate) {
		JSONError(ctx, StatusError, MessageInvalidSessionType)
		return
	}

	currentUser := LoginUserFromContext(ctx)
	roomID := ""

	switch req.SessionType {
	case database.ChatMessageSessionTypePrivate:
		roomID = utils.FormatPrivateRoomID(currentUser.ID, req.TargetID)
	case database.ChatMessageSessionTypeGroup:
		roomID = utils.FormatGroupRoomID(req.TargetID)
	case database.ChatMessageSessionTypeWorld:
		roomID = utils.FormatWorldRoomID(req.TargetID)
	}

	list, err := database.GetLastChatMessageList(roomID, req.SessionType)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, MessageNotFound)
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Str("room_id", roomID).Msg("获取最近消息列表失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	rsps := make([]*ChatMessage, len(list))
	for i := 0; i < len(list); i++ {
		item := list[i]
		msg := &ChatMessage{
			ID:          item.ID.Hex(),
			SessionType: item.SessionType,
			Type:        item.Type,
			SenderID:    item.SenderID,
			ReceiverID:  item.ReceiverID,
			MessageID:   item.MessageID,
			CreatedAt:   item.CreatedAt,
			Body: ChatMessageBody{
				Text:          item.Body.Text,
				Src:           item.Body.Src,
				Format:        item.Body.Format,
				Size:          item.Body.Size,
				Longitude:     item.Body.Longitude,
				Latitude:      item.Body.Latitude,
				Scale:         item.Body.Scale,
				LocationLabel: item.Body.LocationLabel,
			},
		}
		rsps[i] = msg
	}
	JSON(ctx, rsps)

}

// ========================================================================================
// ============================ SUBSCRIBE HANDLER =========================================
// ========================================================================================

// SubscribeChatMessageHandler 接收聊天消息
func SubscribeChatMessageHandler(ctx context.Context, payload *pubsub.Payload) {
	chatMsg := pubsub.NewChatMessage()
	err := payload.UnmarshalData(chatMsg)
	if err != nil {
		log.Error().Err(err).
			Str("channel", payload.Channel).
			Str("payload.type", payload.Type).
			Msg("payload.data 不是 handler.ChatMessage 格式")
		return
	}

	wsPayload := websocket.Payload{
		Type: payload.Type,
		Data: chatMsg,
	}

	switch chatMsg.SessionType {
	case database.ChatMessageSessionTypePrivate: // 处理私聊会话
		websocketManager.PushJson(wsPayload, chatMsg.SenderID, chatMsg.ReceiverID)
	case database.ChatMessageSessionTypeGroup: // 处理群聊会话
		// 为什么增加GroupMembers这个参数?
		// 因为分布式中,会多个实例都订阅到该方法,将导致多个实例都会执行 database.GetGroupMemberIDsString 的方法,所以 加入 PublishTargets,直接从数据中拿 ,能尽量少请求数据库就尽量少请求
		memberStrIds := chatMsg.PublishTargets
		if len(memberStrIds) == 0 {
			// 找出群成员的ID列表
			groupID := chatMsg.ReceiverID
			memberStrIds, err = database.GetGroupMemberIDs(groupID)
			if err != nil {
				log.Error().Err(err).
					Str("channel", payload.Channel).
					Str("payload.type", payload.Type).
					Msg("payload.data 不是 handler.ChatMessage 格式")
				return
			}
		}

		var anySlice = make([]any, len(memberStrIds))
		for i := 0; i < len(memberStrIds); i++ {
			anySlice[i] = memberStrIds[i]
		}

		if len(anySlice) == 0 {
			log.Error().Err(err).
				Str("channel", payload.Channel).
				Str("payload.type", payload.Type).
				Msg("群成员数量为空")
			return
		}

		websocketManager.PushJson(wsPayload, anySlice...)
	case database.ChatMessageSessionTypeWorld: // 处理世界会话
		websocketManager.PushJson(wsPayload)
	}
}
