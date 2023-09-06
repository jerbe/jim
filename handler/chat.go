package handler

import (
	"context"
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

/*
 {
	"action":"send_msg",
	"data":{}
}
*/

// SendChatMessageRequest 聊天发送消息请求参数
// @Description 聊天发送消息请求参数
type SendChatMessageRequest struct {
	// ActionID 行为ID,由前端生成
	ActionID string `json:"action_id" example:"8d7a3bcd72"`

	// ReceiverID 接收人; 可以是用户ID,也可以是群号
	ReceiverID int64 `json:"receiver_id" binding:"required" example:"1234"`

	// SessionType 会话类型; 1:私聊, 2:群聊
	SessionType int `json:"session_type" binding:"required" enums:"1,2" example:"1"`

	// Type 消息类型: 1-纯文本,2-图片,3-语音,4-视频, 5-位置
	Type int `json:"type" enums:"1,2,3,4,5" binding:"required" example:"1"`

	// Body 消息体;
	Body ChatMessageBody `json:"body" binding:"required"`
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

// SendChatMessageResponse 发送聊天返回数据
// @Description 发送聊天返回数据
type SendChatMessageResponse struct {
	// 发送人ID
	SenderID int64 `json:"sender_id" binding:"required" example:"1234456"`

	// MessageID 消息ID
	MessageID int64 `json:"message_id" binding:"required" example:"123"`

	// CreatedAt 创建
	CreatedAt int64 `json:"created_at" binding:"required" example:"12345678901234"`

	SendChatMessageRequest
}

// SendChatMessageHandler
// @Summary      发送聊天消息
// @Tags         聊天
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      SendChatMessageRequest  true  "请求JSON数据体"
// @Security 	 APIKeyHeader
// @Success      200  {object}  Response{data=SendChatMessageResponse}
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

	if req.ReceiverID <= 0 {
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
		sendChatMessageToWorld(ctx, req, currentUser)
		return
	}

}

// sendChatMessageToFriend 向好友发送聊天消息
func sendChatMessageToFriend(ctx *gin.Context, req *SendChatMessageRequest, currentUser *database.User) {
	targetID := req.ReceiverID
	if currentUser.ID == req.ReceiverID {
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

	// 私聊状态的房间号是按 用户ID排序分组
	roomID := utils.FormatPrivateRoomID(currentUser.ID, targetID)

	now := time.Now()
	// 插入消息数据库
	msg := &database.ChatMessage{
		RoomID:      roomID,
		Type:        req.Type,
		SessionType: database.ChatMessageSessionTypePrivate,
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

	err = database.AddChatMessage(msg)
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

	rsp := SendChatMessageResponse{
		SenderID:               currentUser.ID,
		MessageID:              msg.MessageID,
		CreatedAt:              msg.CreatedAt,
		SendChatMessageRequest: *req,
	}
	JSON(ctx, rsp)

	var psData = &pubsub.ChatMessage{
		ActionID:    rsp.ActionID,
		ReceiverID:  rsp.ReceiverID,
		SessionType: rsp.SessionType,
		Type:        rsp.Type,
		SenderID:    rsp.SenderID,
		MessageID:   rsp.MessageID,
		CreatedAt:   rsp.CreatedAt,
		Body: pubsub.ChatMessageBody{
			Text:          rsp.Body.Text,
			Src:           rsp.Body.Src,
			Format:        rsp.Body.Format,
			Size:          rsp.Body.Size,
			Longitude:     rsp.Body.Longitude,
			Latitude:      rsp.Body.Latitude,
			Scale:         rsp.Body.Scale,
			LocationLabel: rsp.Body.LocationLabel,
		},
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

// sendChatMessageToGroup 发送群聊信息
func sendChatMessageToGroup(ctx *gin.Context, req *SendChatMessageRequest, currentUser *database.User) {

	targetID := req.ReceiverID

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

	// 私聊状态的房间号是按 用户ID排序分组
	roomID := utils.FormatGroupRoomID(targetID)

	now := time.Now()
	// 插入消息数据库
	msg := &database.ChatMessage{
		RoomID:      roomID,
		Type:        req.Type,
		SessionType: database.ChatMessageSessionTypeGroup,
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

	err = database.AddChatMessage(msg)
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

	rsp := SendChatMessageResponse{
		SenderID:               currentUser.ID,
		MessageID:              msg.MessageID,
		CreatedAt:              msg.CreatedAt,
		SendChatMessageRequest: *req,
	}
	JSON(ctx, rsp)

	// 先查出所有群成员ID,这样订阅到的实例无需再次获取群成员信息
	memberStrIds, err := database.GetGroupMemberIDs(targetID)
	if err != nil && !errors.IsNoRecord(err) {
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).
			Int64("user_id", currentUser.ID).
			Int64("receiver_id", msg.ReceiverID).
			Int("session_type", msg.SessionType).
			Msg("获取群成员ID列表失败")
		return
	}

	var psData = &pubsub.ChatMessage{
		ActionID:    rsp.ActionID,
		ReceiverID:  rsp.ReceiverID,
		SessionType: rsp.SessionType,
		Type:        rsp.Type,
		SenderID:    rsp.SenderID,
		MessageID:   rsp.MessageID,
		CreatedAt:   rsp.CreatedAt,
		Body: pubsub.ChatMessageBody{
			Text:          rsp.Body.Text,
			Src:           rsp.Body.Src,
			Format:        rsp.Body.Format,
			Size:          rsp.Body.Size,
			Longitude:     rsp.Body.Longitude,
			Latitude:      rsp.Body.Latitude,
			Scale:         rsp.Body.Scale,
			LocationLabel: rsp.Body.LocationLabel,
		},
		PublishTargets: memberStrIds,
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

// sendChatMessageToWorld 发送世界信息
func sendChatMessageToWorld(ctx *gin.Context, req *SendChatMessageRequest, currentUser *database.User) {
	targetID := req.ReceiverID

	now := time.Now()
	// 插入消息数据库
	msg := &database.ChatMessage{
		RoomID:      fmt.Sprintf("world_%04x", targetID),
		Type:        req.Type,
		SessionType: database.ChatMessageSessionTypeWorld,
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

	rsp := SendChatMessageResponse{
		SenderID:               currentUser.ID,
		MessageID:              msg.MessageID,
		CreatedAt:              msg.CreatedAt,
		SendChatMessageRequest: *req,
	}
	JSON(ctx, rsp)

	var psData = &pubsub.ChatMessage{
		ActionID:    rsp.ActionID,
		ReceiverID:  rsp.ReceiverID,
		SessionType: rsp.SessionType,
		Type:        rsp.Type,
		SenderID:    rsp.SenderID,
		MessageID:   rsp.MessageID,
		CreatedAt:   rsp.CreatedAt,
		Body: pubsub.ChatMessageBody{
			Text:          rsp.Body.Text,
			Src:           rsp.Body.Src,
			Format:        rsp.Body.Format,
			Size:          rsp.Body.Size,
			Longitude:     rsp.Body.Longitude,
			Latitude:      rsp.Body.Latitude,
			Scale:         rsp.Body.Scale,
			LocationLabel: rsp.Body.LocationLabel,
		},
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

// RollbackChatMessageHandler 回滚聊天消息处理方法
func RollbackChatMessageHandler(ctx *gin.Context) {

}

// DeleteChatMessageHandler 删除聊天消息处理方法
func DeleteChatMessageHandler(ctx *gin.Context) {

}

// ========================================================================================
// ============================ SUBSCRIBE HANDLER =========================================
// ========================================================================================

// SubscribeChatMessageHandler 接收聊天消息
func SubscribeChatMessageHandler(ctx context.Context, payload *pubsub.Payload) {
	chatMsg := new(pubsub.ChatMessage)
	err := payload.UnmarshalData(chatMsg)
	if err != nil {
		log.Error().Err(err).
			Str("channel", payload.Channel).
			Str("payload.type", payload.Type).
			Msg("payload.data 不是 handler.SendChatMessageResponse 格式")
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
					Msg("payload.data 不是 handler.SendChatMessageResponse 格式")
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
