package handler

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jerbe/jim/database"
	"github.com/jerbe/jim/errors"
	"github.com/jerbe/jim/log"
	"github.com/jerbe/jim/pubsub"
	"github.com/jerbe/jim/utils"
	"github.com/jerbe/jim/websocket"
	"strconv"
	"time"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/18 15:11
  @describe :
*/
// ========================================================================================
// ================================= HTTP HANDLER =========================================
// ========================================================================================

// User
// @Description 用户信息结构
type User struct {
	// ID 用户ID
	ID int64 `json:"id" example:"10096"`

	// Username 用户名
	Username string `json:"username,omitempty" example:"admin"`

	// Nickname 用户昵称
	Nickname string `json:"nickname" example:"昵称"`

	// BirthDate 出生日期
	BirthDate *time.Time `json:"birth_date,omitempty" example:"2018-01-02"`

	// Avatar 头像地址
	Avatar string `json:"avatar,omitempty" format:"url" example:"https://www.baidu.com/logo.png"`

	// OnlineStatus 在线状态
	OnlineStatus int `json:"online_status,omitempty" enums:"0,1" example:"1"`
}

// FindFriendRequest 查找好友请求参数
// @Description 查找好友请求参数
type FindFriendRequest struct {
	// UserID 用户ID
	UserID *int64 `json:"user_id"  example:"1"`

	// Nickname 昵称
	Nickname *string `json:"nickname" example:"昵称"`

	// StartID 开始搜索ID,下次搜索用上返回的最大ID
	StartID *int64 `json:"start_id" example:"0"`
}

// FindFriendResponse 查找好友返回参数
// @Description 查找好友返回参数
type FindFriendResponse struct {
	Users []*User `json:"users" `
}

// FindFriendHandler
// @Summary      查找好友
// @Tags         朋友
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      FindFriendRequest  true  "请求JSON数据体"
// @Security 	 APIKeyHeader
// @Success      200  {object}  Response{data=FindFriendResponse}
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/friend/find [post]
func FindFriendHandler(ctx *gin.Context) {
	req := new(FindFriendRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	if req.UserID == nil && req.Nickname == nil {
		JSONError(ctx, StatusError, "必要参数未填写")
		return
	}

	if req.UserID != nil && *req.UserID <= 0 {
		JSONError(ctx, StatusError, "'user_id'必须大于0")
		return
	}

	if req.Nickname != nil && utils.StringLen(*req.Nickname) < 2 {
		JSONError(ctx, StatusError, "'nickname'必须大于2个字符")
		return
	}

	filter := &database.SearchUsersFilter{}
	if req.UserID != nil && *req.UserID > 0 {
		filter.UserID = req.UserID
	}

	if req.Nickname != nil && *req.Nickname != "" {
		filter.Nickname = req.Nickname
	}

	if req.StartID == nil {
		req.StartID = new(int64)
	}

	filter.StartID = *req.StartID
	filter.Limit = 20

	users, err := database.SearchUsers(filter)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, "找不到符合的用户")
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).
			Any("filter", filter).Msg("搜索用户消息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}
	rspUsers := make([]*User, len(users), len(users))
	for i := 0; i < len(users); i++ {
		rspUsers[i] = &User{
			ID:           users[i].ID,
			Username:     users[i].Username,
			Nickname:     users[i].Nickname,
			BirthDate:    users[i].BirthDate,
			Avatar:       users[i].Avatar,
			OnlineStatus: users[i].OnlineStatus,
		}
	}
	rsp := FindFriendResponse{Users: rspUsers}
	JSON(ctx, rsp)
}

// AddFriendInviteRequest 邀请用户成为好友操作请求参数
// @Description 邀请用户成为好友操作请求参数
type AddFriendInviteRequest struct {
	// UserID 目标用户ID
	UserID int64 `json:"user_id" binding:"required" min:"1" example:"1"`

	// Note 添加好友时,给对方看的备注
	Note string `json:"note" example:"你好,我是Jerbe"`
}

// AddFriendInviteHandler
// @Summary      邀请加为好友
// @Tags         朋友
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      AddFriendInviteRequest  true  "请求JSON数据体"
// @Security 	 APIKeyHeader
// @Success      200  {object}  Response
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/friend/invite/add [post]
func AddFriendInviteHandler(ctx *gin.Context) {
	req := new(AddFriendInviteRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	if req.UserID <= 0 {
		JSONError(ctx, StatusError, MessageInvalidUserID)
		return
	}

	currentUser := LoginUserFromContext(ctx)
	target, err := database.GetUser(req.UserID)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, MessageTargetNotExists)
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("target_id", req.UserID).Msg("获取用户信息异常")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	// 对方要是被禁用,或者删除,禁止加好友
	if target.Status == 0 {
		JSONError(ctx, StatusError, MessageTargetDisable)
		return
	}

	// 对方要是被禁用,或者删除,禁止加好友
	if target.Status == 2 {
		JSONError(ctx, StatusError, MessageTargetDeleted)
		return
	}

	// 1 先获取关系数据
	// 因为可能以前加过好友,但是后来单方面删除了
	relation, err := database.GetUserRelationByUsersID(currentUser.ID, target.ID, database.NewGetOptions().SetUseCache(false))
	if err != nil && !errors.IsNoRecord(err) {
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("target_id", req.UserID).Msg("获取用户关系信息异常")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	now := time.Now()

	// 1.1 关系记录存在,说明以前加过好友
	if relation != nil {
		// 1.1.1 存在记录并且已是好友关系,无需再次申请
		if relation != nil && relation.Status == 3 {
			JSONError(ctx, StatusError, MessageAlreadyfriends)
			return
		}

		// 1.1.2 如果对方还未删除本人,则直接进行添加,无需进行发起申请加好友请求,直接设置成加好友
		if (relation.UserAID == currentUser.ID && relation.Status&0b01 != 0) ||
			(relation.UserBID == currentUser.ID && relation.Status&0b10 != 0) {
			now := time.Now()
			updateFilter := &database.UpdateUserRelationFilter{
				ID: relation.ID,
			}
			updateData := (&database.UpdateUserRelationData{}).SetStatus(0b11).SetUpdatedAt(now)
			_, err := database.UpdateUserRelationTx(updateFilter, updateData)
			if err != nil {
				log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("target_id", req.UserID).Msg("添加用户关系失败")
				JSONError(ctx, StatusError, MessageInternalServerError)
				return
			}

			JSON(ctx)

			// 发送一条 say hello 的聊天消息
			go sayHelloFn(ctx, currentUser.ID, target.ID, &now)

			return
		}
	}

	// 2 获取是否存在最后一条邀请记录
	invite, err := database.GetLastUserRelationInvite(currentUser.ID, req.UserID)
	if err != nil && !errors.IsNoRecord(err) {
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("target_id", req.UserID).Msg("添加建立关系请求异常")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	// 2.1 存在邀请记录
	if invite != nil {
		// 2.1.1 该邀请是本人发起的
		// 直接再次推送消息
		if currentUser.ID == invite.UserID {
			JSONError(ctx, StatusError, "已发起请求待对方确认")

			// 推送一条邀请记录到对方
			go publishInvite(ctx, invite)
			return
		}

		// 2.1.2 该邀请是对方发起的,则可以直接加为好友
		if currentUser.ID == invite.TargetID {
			// 1) 原先不存在关系则添加关系成加为好友
			if relation == nil {
				relation = &database.UserRelation{
					UserAID:     currentUser.ID,
					UserBID:     target.ID,
					Status:      0b11,
					BlockStatus: 0b11,
					UpdatedAt:   now,
					CreatedAt:   now,
				}
				_, err := database.CreateUserRelationTx(relation)
				if err != nil {
					log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("target_id", req.UserID).Msg("添加用户关系失败")
					JSONError(ctx, StatusError, MessageInternalServerError)
					return
				}
			} else { // 2) 原先存在关系则更新关系成加为好友
				updateFilter := &database.UpdateUserRelationFilter{UserAID: currentUser.ID, UserBID: target.ID}
				updateData := (&database.UpdateUserRelationData{}).SetStatus(0b11)
				_, err = database.UpdateUserRelationTx(updateFilter, updateData)
				if err != nil {
					log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("target_id", req.UserID).Msg("添加用户关系失败")
					JSONError(ctx, StatusError, MessageInternalServerError)
					return
				}
			}

			JSON(ctx)

			// 直接发送聊天消息到对方上
			go sayHelloFn(ctx, currentUser.ID, target.ID, &now)

			return
		}

		JSONError(ctx, StatusError, "有获取到邀请信息,但都不是双方的记录")

		log.InfoFromGinContext(ctx).Int64("invite_id", invite.ID).Int64("target_id", req.UserID).Msg("有获取到邀请信息,但都不是双方的记录")
		return
	}

	// 3 发送建立关系邀请
	invite = &database.UserRelationInvite{
		UserID:    currentUser.ID,
		TargetID:  req.UserID,
		Note:      req.Note,
		Status:    database.UserRelationInviteStatusPending,
		UpdatedAt: now,
		CreatedAt: now,
	}
	err = database.AddUserRelationInvite(invite)
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("target_id", req.UserID).Msg("添加建立关系请求异常")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	// 推送一条邀请记录到对方
	go publishInvite(ctx, invite)

	JSON(ctx)
}

// UpdateFriendInviteRequest 更新邀请好友请求
// @Description 更新邀请好友请求
type UpdateFriendInviteRequest struct {
	// ID 邀请记录ID号
	ID int64 `json:"id" binding:"required" min:"1" example:"1"`

	// Status 设置状态
	// 1-确认加为好友关系,2-拒绝加为好友
	Status int `json:"status" binding:"required" enums:"1,2" example:"1"`

	// Reply 回复信息
	// status:1 为打招呼
	// status:2 为拒绝理由
	Reply string `json:"reply" example:"我拒绝"`
}

// UpdateFriendInviteHandler
// @Summary      处理好友邀请
// @Tags         朋友
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      UpdateFriendInviteRequest  true  "请求JSON数据体"
// @Security 	 APIKeyHeader
// @Success      200  {object}  Response
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/friend/invite/update [post]
func UpdateFriendInviteHandler(ctx *gin.Context) {
	req := new(UpdateFriendInviteRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	// 无效的ID
	if req.ID <= 0 {
		JSONError(ctx, StatusError, "'id'不能为空")
		return
	}

	// 无效的状态
	if req.Status <= 0 || req.Status > 2 {
		JSONError(ctx, StatusError, "'status'无效")
		return
	}

	// 限定reply长度
	if l := len([]rune(req.Reply)); l > 20 {
		JSONError(ctx, StatusError, "'reply'不可以超过20个字符")
		return
	}

	invite, err := database.GetUserRelationInvite(req.ID)
	if err != nil {
		if errors.IsNoRecord(err) {
			log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("invite_id", req.ID).Msg("找不到邀请记录")
			JSONError(ctx, StatusError, errors.NoRecords.Error())
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("invite_id", req.ID).Msg("获取用户邀请记录失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	if utils.In(invite.Status, 1, 2) {
		JSONError(ctx, StatusError, "该邀请已处理,无法再次处理")
		return
	}

	currentUser := LoginUserFromContext(ctx)
	if invite.TargetID != currentUser.ID {
		JSONError(ctx, StatusError, "没有权限")
		return
	}

	now := time.Now()
	updateFilter := &database.UpdateUserRelationInviteFilter{
		ID: invite.ID,
	}
	updateData := &database.UpdateUserRelationInviteData{
		UpdatedAt: now,
		Reply:     req.Reply,
		Status:    req.Status,
	}

	// 1. 如果是接受,建立/更新用户之间的关系并say hello
	if req.Status == 1 {
		// database.CreateUserRelationTx 已经包含了重复插入的检测,所以不需要判断关系记录是否存在
		relation := &database.UserRelation{
			UserAID:     invite.UserID,
			UserBID:     invite.TargetID,
			Status:      0b11,
			BlockStatus: 0b11,
			UpdatedAt:   now,
			CreatedAt:   now,
		}
		_, err = database.CreateUserRelationTx(relation)
		if err != nil {
			log.ErrorFromGinContext(ctx).Err(err).Int64("user_a_id", invite.UserID).Int64("user_b_id", invite.TargetID).Msgf("添加用户关系失败\n%+v", err)
			JSONError(ctx, StatusError, MessageInternalServerError)
			return
		}

		go sayHelloFn(ctx, currentUser.ID, invite.UserID, &now)

		JSON(ctx)
		return
	}

	// 2. 如果是拒绝,直接更新邀请记录并推送
	_, err = database.UpdateUserRelationInvite(updateFilter, updateData)
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Msg("更新用户邀请记录失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	// 2.1 推送拒绝的通知到对方
	go publishInvite(ctx, invite)
	JSON(ctx)
}

// UpdateFriendRequest 更新朋友信息请求参数
// @Description 更新朋友信息请求参数
type UpdateFriendRequest struct {
	// UserID 对方的用户ID
	UserID int64 `json:"user_id" binding:"required" min:"1" example:"1"`

	// Status 好友状态, 0-删除对方
	Status *int `json:"status" enums:"0" example:"0"`

	// BlockStatus 黑名单状态, 0-把对方加入黑名单, 1-将对方从黑名单中移除
	BlockStatus *int `json:"block_status" enums:"0" example:"1"`

	// Remark 备注
	Remark *string `json:"remark" example:"这个是我基友"`
}

// UpdateFriendHandle
// @Summary      更新朋友信息
// // 可进行删除好友,拉黑好友
// @Tags         朋友
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      UpdateFriendInviteRequest  true  "请求JSON数据体"
// @Security 	 APIKeyHeader
// @Success      200  {object}  Response
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/friend/update [post]
func UpdateFriendHandle(ctx *gin.Context) {
	req := new(UpdateFriendRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}
	currentUser := LoginUserFromContext(ctx)

	if req.UserID <= 0 {
		JSONError(ctx, StatusError, "'target_id' 验证无效")
		return
	}

	if req.UserID == currentUser.ID {
		JSONError(ctx, StatusError, "自己是自己,不能修改自己与自己的关系")
		return
	}

	if req.Status == nil && req.BlockStatus == nil && req.Remark == nil {
		JSONError(ctx, StatusError, "参数未设置")
		return
	}

	if req.Status != nil && req.BlockStatus != nil {
		JSONError(ctx, StatusError, "'status'或'block_status'只能选择一样")
		return
	}

	// 只能进行删除,0为删除好友
	if req.Status != nil && *req.Status != 0 {
		JSONError(ctx, StatusError, "'status'必须为'0'")
		return
	}

	// 只能进行删除,0为拉黑好友,1为取消拉黑
	if req.BlockStatus != nil && !utils.In(*req.BlockStatus, 0, 1) {
		JSONError(ctx, StatusError, "'block_status'必须为'0'或'1'")
		return
	}

	relation, err := database.GetUserRelationByUsersID(currentUser.ID, req.UserID)
	if errors.IsNoRecord(err) || relation == nil {
		JSONError(ctx, StatusError, "你们并不是好友关系")
		return
	}
	if err != nil {
		JSONError(ctx, StatusError, MessageInternalServerError)
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).
			Int64("target_id", req.UserID).
			Msg("获取好友关系失败")
		return
	}

	change := false
	// 这里为什么一直判断用户ID大小?
	// 因为我们入库时候是按左小右大的顺序添加,比较特殊,所以需要进行验证
	data := &database.UpdateUserRelationData{}
	// 判断好友状态
	if req.Status != nil {
		data.Status = new(int)
		if currentUser.ID < req.UserID {
			*data.Status = relation.Status &^ 0b10 // 清空左边的位值成0
		} else {
			*data.Status = relation.Status &^ 0b01 // 清空右边边的位值成0
		}
		change = *data.Status != relation.Status
	}

	// 判断拉黑状态
	if req.BlockStatus != nil {
		data.BlockStatus = new(int)
		if *req.BlockStatus == 0 {
			if currentUser.ID < req.UserID {
				*data.BlockStatus = relation.BlockStatus &^ 0b10 // 清空左边的位值成0
			} else {
				*data.BlockStatus = relation.BlockStatus &^ 0b01 // 清空右边边的位值成0
			}
		}

		if *req.BlockStatus == 1 {
			if currentUser.ID < req.UserID {
				*data.BlockStatus = relation.BlockStatus | 0b10 // 填充左边的位值成1
			} else {
				*data.BlockStatus = relation.BlockStatus | 0b01 // 填充右边边的位值成1
			}
		}

		change = !change && *data.BlockStatus != relation.BlockStatus
	}

	// 判断备注状态
	if req.Remark != nil {
		if currentUser.ID < req.UserID {
			data.RemarkOnB = req.Remark

			change = !change && *data.RemarkOnB != relation.RemarkOnB
		} else {
			data.RemarkOnA = req.Remark
			change = !change && *data.RemarkOnA != relation.RemarkOnA
		}
	}

	// 为什么做这步? 高并发情况下,能少一次数据库请求就尽量少请求
	if !change {
		JSONError(ctx, StatusError, "数据未更新")
		return
	}

	filter := &database.UpdateUserRelationFilter{ID: relation.ID}
	cnt, err := database.UpdateUserRelation(filter, data)
	if err != nil {
		JSONError(ctx, StatusError, MessageInternalServerError)
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).
			Int64("target_id", req.UserID).
			Msg("更新好友关系失败")
		return
	}

	if cnt == 0 {
		JSONError(ctx, StatusError, "数据未更新")
		return
	}
	JSON(ctx)
}

// sayHelloFn 打招呼的公用方法,后续可能会重用
func sayHelloFn(ctx *gin.Context, userID, targetID int64, now *time.Time) error {
	if now == nil {
		x := time.Now()
		now = &x
	}
	defer func() {
		if obj := recover(); obj != nil {
			l := log.ErrorFromGinContext(ctx)
			if e, ok := obj.(error); ok {
				l.Err(e).Str("recover", fmt.Sprintf("%+v", e))
			}
			l.Str("panic", "true").Send()
		}
	}()
	// 直接发送聊天消息到对方上
	// 私聊状态的房间号是按 用户ID排序分组
	roomID := utils.FormatPrivateRoomID(userID, targetID)

	// 插入消息数据库
	msg := &database.ChatMessage{
		RoomID:      roomID,
		Type:        database.ChatMessageTypePlainText,
		SessionType: database.ChatMessageSessionTypePrivate,
		SenderID:    userID,
		ReceiverID:  targetID,
		SendStatus:  1,
		ReadStatus:  0,
		Status:      1,
		CreatedAt:   now.UnixMilli(),
		UpdatedAt:   now.UnixMilli(),
		Body: database.ChatMessageBody{
			Text: "我们已经是好友了,可以快乐得聊天了",
		},
	}
	err := database.AddChatMessage(msg)

	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).
			Int("message_type", msg.Type).
			Int64("receiver_id", msg.ReceiverID).
			Int("session_type", msg.SessionType).
			Msg("添加聊天消息失败")
		return errors.Wrap(err)
	}

	// 进行多服务器订阅推送
	var psData = &pubsub.ChatMessage{
		ReceiverID:  msg.ReceiverID,
		SessionType: msg.SessionType,
		Type:        msg.Type,
		SenderID:    msg.SenderID,
		MessageID:   msg.MessageID,
		CreatedAt:   msg.CreatedAt,
		Body: pubsub.ChatMessageBody{
			Text:          msg.Body.Text,
			Src:           msg.Body.Src,
			Format:        msg.Body.Format,
			Size:          msg.Body.Size,
			Longitude:     msg.Body.Longitude,
			Latitude:      msg.Body.Latitude,
			Scale:         msg.Body.Scale,
			LocationLabel: msg.Body.LocationLabel,
		},
	}
	err = pubsub.PublishChatMessage(ctx, psData)
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).
			Int64("user_id", userID).
			Int64("receiver_id", msg.ReceiverID).
			Int("session_type", msg.SessionType).
			Msg("推送聊天消息到订阅服务失败")

		return errors.Wrap(err)
		//@ todo 需要重做推送!
	}
	return nil
}

// publishInvite 推送邀请通知
func publishInvite(ctx *gin.Context, invite *database.UserRelationInvite) error {
	defer func() {
		if obj := recover(); obj != nil {
			l := log.ErrorFromGinContext(ctx)
			if e, ok := obj.(error); ok {
				l.Err(e).Str("recover", fmt.Sprintf("%+v", e))
			}
			l.Str("panic", "true").Send()
		}
	}()
	// @todo 推送邀请加好友的消息到对方ws上
	psData := &pubsub.FriendInvite{
		ID:        invite.ID,
		UserID:    invite.UserID,
		TargetID:  invite.TargetID,
		Status:    database.UserRelationInviteStatusPending,
		Note:      invite.Note,
		Reply:     invite.Reply,
		CreatedAt: invite.CreatedAt,
	}
	err := pubsub.PublishNotifyMessage(context.Background(), pubsub.PayloadTypeFriendInvite, psData)
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).
			Int64("invite_id", invite.ID).
			Int64("target_id", invite.UserID).
			Msgf("推送好友邀请信息到对方失败\n%+v", err)
		return errors.Wrap(err)
	}
	return nil
}

// ========================================================================================
// ============================ SUBSCRIBE HANDLER =========================================
// ========================================================================================

// SubscribeFriendInviteHandler 订阅好友邀请控制器
func SubscribeFriendInviteHandler(ctx context.Context, payload *pubsub.Payload) {
	fi := new(pubsub.FriendInvite)
	err := payload.UnmarshalData(fi)
	if err != nil {
		log.Error().Err(err).Str("payload.channel", payload.Channel).Str("payload.type", payload.Type).Send()
		return
	}

	wsPayload := websocket.Payload{
		Type: payload.Type,
		Data: fi,
	}

	// 如果邀请记录是进行中,则直接发给目标
	if fi.Status == database.UserRelationInviteStatusPending {
		websocketManager.PushJson(wsPayload, strconv.FormatInt(fi.TargetID, 10))
		return
	}

	// 如果邀请记录是拒绝,则直接发给用户
	// 因为同意的时候是直接say hello了
	if fi.Status == database.UserRelationInviteStatusReject {
		websocketManager.PushJson(wsPayload, strconv.FormatInt(fi.UserID, 10))
		return
	}
}
