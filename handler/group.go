package handler

import (
	"fmt"
	"time"

	"github.com/jerbe/jim/database"
	"github.com/jerbe/jim/errors"
	"github.com/jerbe/jim/log"
	"github.com/jerbe/jim/utils"

	goutils "github.com/jerbe/go-utils"

	"github.com/gin-gonic/gin"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/29 18:53
  @describe :
*/

// CreateGroupRequest 创建群请求参数
// @Description 创建群请求参数
type CreateGroupRequest struct {
	// MemberIDs 其他成员的用户ID
	MemberIDs []int64 `json:"member_ids" binding:"required" minLength:"1" maxLength:"50" example:"1,2,3,4"`
}

// CreateGroupResponse 创建群返回参数
// @Description 创建群返回参数
type CreateGroupResponse struct {
	// GroupID 群ID
	GroupID int64 `json:"group_id" binding:"required" example:"1098"`

	// GroupName 群名称
	GroupName string `json:"group_name"  binding:"required" example:"群聊1098"`
}

// CreateGroupHandler
// @Summary      创建群
// @Tags         群组
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      CreateGroupRequest  true  "请求JSON数据体"
// @Security 	 APIKeyHeader
// @Success      200  {object}  Response{data=CreateGroupResponse}
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/group/create [post]
func CreateGroupHandler(ctx *gin.Context) {
	req := new(CreateGroupRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	if len(req.MemberIDs) == 0 {
		JSONError(ctx, StatusError, "成员数量不能为0")
		return
	}

	var members []int64
	err = goutils.SliceUnique(req.MemberIDs, &members)
	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	if len(members)+1 >= database.GroupMaxMember {
		JSONError(ctx, StatusError, fmt.Sprintf("成员数量不能超过%d", database.GroupMaxMember))
		return
	}

	currentUser := LoginUserFromContext(ctx)

	group, err := database.CreateGroupTx(currentUser.ID, members)
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Ints64("members", members).Msg("创建群失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	rsp := CreateGroupResponse{
		GroupID:   group.ID,
		GroupName: group.Name,
	}
	JSON(ctx, rsp)

}

// JoinGroupRequest 入群请求参数
// @Description 入群请求参数
type JoinGroupRequest struct {
	// GroupID 群ID
	GroupID int64 `json:"group_id" binding:"required" example:"1098"`
}

// JoinGroupHandler
// @Summary      加入群
// @Tags         群组
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      JoinGroupRequest  true  "请求JSON数据体"
// @Security 	 APIKeyHeader
// @Success      200  {object}  Response
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/group/join [post]
func JoinGroupHandler(ctx *gin.Context) {
	req := new(JoinGroupRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	currentUser := LoginUserFromContext(ctx)

	member, err := database.GetGroupMember(req.GroupID, currentUser.ID)
	if err == nil && member != nil {
		JSONError(ctx, StatusError, "您已经是该群成员")
		return
	}

	if err != nil && !errors.IsNoRecord(err) {
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Msg("获取群成员信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
	}

	// @todo 判断群是否允许无条件加入

	// 判断群成员是否已经超过限制大小
	group, err := database.GetGroup(req.GroupID)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, "该群不存在")
			return
		}

		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Msg("获取群信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	groupMemberCnt, err := database.GetGroupMemberCount(req.GroupID)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, "该群已解散")
			return
		}

		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Msg("获取群成员数量失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	if groupMemberCnt+1 > int64(group.MaxMember) {
		JSONError(ctx, StatusError, fmt.Sprintf("群成员超过上限."), map[string]any{"max": group.MaxMember, "total": groupMemberCnt})
		return
	}

	// 入库
	addData := database.AddGroupMembersData{UserIDs: []int64{currentUser.ID}, CreatorID: currentUser.ID, CreatedAt: time.Now()}
	_, err = database.AddGroupMembers(req.GroupID, &addData)
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Msg("加入群组失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}
	JSON(ctx)
}

// LeaveGroupRequest 离群请求参数
// @Description 离群请求参数
type LeaveGroupRequest struct {
	// GroupID 群ID
	GroupID int64 `json:"group_id" binding:"required" example:"1098"`
}

// LeaveGroupHandler
// @Summary      离开群
// @Tags         群组
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      LeaveGroupRequest  true  "请求JSON数据体"
// @Security 	 APIKeyHeader
// @Success      200  {object}  Response
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/group/leave [post]
func LeaveGroupHandler(ctx *gin.Context) {
	req := new(JoinGroupRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	currentUser := LoginUserFromContext(ctx)

	member, err := database.GetGroupMember(req.GroupID, currentUser.ID)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, "您不是该群成员")
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Int64("member_id", currentUser.ID).Msg("获取群成员信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
	}

	if member.Role == 1 {
		JSONError(ctx, StatusError, "您是群主,无法离群")
		return
	}

	// 如果是管理员也可以主动退出
	var rgmf = database.RemoveGroupMembersFilter{GroupID: req.GroupID, UserIDs: []int64{currentUser.ID}, Roles: []int{2}}
	_, err = database.RemoveGroupMembers(&rgmf)
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Int64("member_id", currentUser.ID).Msg("移除群成员失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	JSON(ctx)
}

// UpdateGroupRequest 更新群组信息请求
// @Description 更新群组信息请求参数
type UpdateGroupRequest struct {
	// GroupID 群组ID
	GroupID int64 `json:"group_id" binding:"required" example:"1098"`

	// Name 群名称
	Name *string `json:"name,omitempty" minLength:"1" maxLength:"50" example:"昵称"`

	// SpeakStatus 发言状态, 必须为管理员以上级别
	SpeakStatus *int `json:"speak_status,omitempty" enums:"0,1" example:"1"`

	// OwnerID 新群主
	OwnerID *int64 `json:"owner_id,omitempty" example:"1"`
}

// UpdateGroupHandler
// @Summary      更新群信息
// @Tags         群组
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      UpdateGroupRequest  true  "请求JSON数据体"
// @Security 	 APIKeyHeader
// @Success      200  {object}  Response{}
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/group/update [post]
func UpdateGroupHandler(ctx *gin.Context) {
	req := new(UpdateGroupRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	if req.GroupID <= 0 {
		JSONError(ctx, StatusError, "'group_id'无效")
		return
	}

	if goutils.EqualAll(nil, req.Name, req.OwnerID, req.SpeakStatus) {
		JSONError(ctx, StatusError, "更改的项未填写")
		return
	}

	currentUser := LoginUserFromContext(ctx)

	group, err := database.GetGroup(req.GroupID)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, "该群不存在")
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Msg("获取群信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	member, err := database.GetGroupMember(req.GroupID, currentUser.ID)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, "您不是该群成员")
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Int64("member_id", currentUser.ID).Msg("获取群成员信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	updateData := &database.UpdateGroupData{}

	if req.Name != nil {
		if *req.Name == "" {
			JSONError(ctx, StatusError, "群名不能为空")
			return
		}

		if utils.StringLen(*req.Name) > 50 {
			JSONError(ctx, StatusError, "群名长度不能大于50")
			return
		}
		updateData.Name = req.Name
	}

	if req.SpeakStatus != nil {
		if member.Role == 0 {
			JSONError(ctx, StatusError, "无权限修改发言权限")
			return
		}

		if !goutils.In(*req.SpeakStatus, 0, 1) {
			JSONError(ctx, StatusError, "修改发言权限失败,必须是0/1")
			return
		}

		updateData.SpeakStatus = req.SpeakStatus
	}

	if req.OwnerID != nil {
		if group.OwnerID != currentUser.ID {
			JSONError(ctx, StatusError, "您不是该群的群主,禁止转移群群主")
			return
		}

		_, err = database.GetGroupMember(req.GroupID, *req.OwnerID)
		if err != nil {
			if errors.IsNoRecord(err) {
				JSONError(ctx, StatusError, fmt.Sprintf("用户'%d'不是该群成员", *req.OwnerID))
				return
			}
			log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Int64("member_id", *req.OwnerID).Msg("获取群成员信息失败")
			JSONError(ctx, StatusError, MessageInternalServerError)
			return
		}

		updateData.OwnerID = req.OwnerID
	}

	if updateData.OwnerID != nil {
		err = database.UpdateGroupTx(req.GroupID, updateData)
	} else {
		err = database.UpdateGroup(req.GroupID, updateData)
	}

	if err != nil {
		if errors.Is(err, errors.NotChange) {
			JSONError(ctx, StatusError, "未做任何改变")
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Bool("has_owner", updateData.OwnerID != nil).Msg("更新群信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	JSON(ctx)
}

// AddGroupMemberRequest 增加群成员请求参数
// @Description 增加群成员请求参数
type AddGroupMemberRequest struct {
	// GroupID 群ID
	GroupID int64 `json:"group_id" binding:"required" example:"1098"`

	// UserIDs 被邀请加入的用户ID列表
	UserIDs []int64 `json:"user_ids" binding:"required"  example:"1,2,3,4"`
}

// AddGroupMemberResponse 增加群成员返回参数
// @Description 增加群返回请求参数
type AddGroupMemberResponse struct {
	// Count 成功加入群的成员数量
	Count int64 `json:"count" binding:"required" example:"22"`
}

// AddGroupMemberHandler
// @Summary      增加群成员
// @Tags         群组
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      AddGroupMemberRequest  true  "请求JSON数据体"
// @Security 	 APIKeyHeader
// @Success      200  {object}  Response{data=AddGroupMemberResponse}
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/group/member/add [post]
func AddGroupMemberHandler(ctx *gin.Context) {
	req := new(AddGroupMemberRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	if req.GroupID <= 0 {
		JSONError(ctx, StatusError, "无效的群ID号")
		return
	}

	if len(req.UserIDs) == 0 {
		JSONError(ctx, StatusError, "加入成员列表不能为空")
		return
	}
	userIDs := req.UserIDs
	currentUser := LoginUserFromContext(ctx)

	// 判断群成员是否已经超过限制大小
	group, err := database.GetGroup(req.GroupID)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, "该群不存在")
			return
		}

		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Ints64("member_ids", userIDs).Msg("获取群成员信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	// 判断当前操作人员是否是该群成员
	_, err = database.GetGroupMember(req.GroupID, currentUser.ID)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, "您不是该群成员")
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Int64("member_id", currentUser.ID).Msg("获取群成员信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	// 判断群成员数量是否会达到上线
	// 1. 获取该群已经存在的成员数量
	groupMemberCnt, err := database.GetGroupMemberCount(req.GroupID)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, "该群已解散")
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Ints64("member_ids", userIDs).Msg("获取群成员数量失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	if groupMemberCnt+int64(len(userIDs)) > int64(group.MaxMember) {
		JSONError(ctx, StatusError, fmt.Sprintf("群成员超过上限."), map[string]any{"max": group.MaxMember, "total": groupMemberCnt})
		return
	}

	// 2. 获取已经加入该群成员的信息,防止多次插入,影响数据库自增ID
	members, err := database.GetGroupMembers(req.GroupID, userIDs)
	if err != nil && !errors.IsNoRecord(err) {
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Ints64("member_ids", userIDs).Msg("获取群成员信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	// 过滤出未添加的群成员
	var filterUserIDs []int64
	for i := 0; i < len(userIDs); i++ {
		var has bool
		for j := 0; j < len(members); j++ {
			if members[j].UserID == userIDs[i] {
				has = true
				break
			}
		}
		if !has {
			filterUserIDs = append(filterUserIDs, userIDs[i])
		}
	}

	// 查出加入的成员是否合法
	users, err := database.GetUsers(filterUserIDs)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, "找不到这些用户")
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Ints64("user_ids", filterUserIDs).Msg("获取用户的信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	var finalUserIDs []int64
	for i := 0; i < len(users); i++ {
		u := users[i]
		// 未被限制的用户才可以进群
		if u.Status == 1 {
			finalUserIDs = append(finalUserIDs, u.ID)
		}
	}

	addData := database.AddGroupMembersData{UserIDs: finalUserIDs, CreatorID: currentUser.ID, CreatedAt: time.Now()}
	cnt, err := database.AddGroupMembers(req.GroupID, &addData)
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Ints64("member_ids", userIDs).Ints64("final_member_ids", finalUserIDs).Msg("添加群成员失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	rsp := &AddGroupMemberResponse{
		Count: cnt,
	}
	JSON(ctx, rsp)
}

// UpdateGroupMemberRequest
// @Description 更新群成员信息请求参数
type UpdateGroupMemberRequest struct {
	// GroupID 群ID
	GroupID int64 `json:"group_id" binding:"required" example:"1098"`

	// UserID 成员用户ID
	UserID int64 `json:"user_id" binding:"required" example:"45"`

	// Role 角色, 只能群主才可操作
	Role *int `json:"role,omitempty" enums:"0,2" example:"0"`

	// SpeakStatus 发言权限, 管理者以上都可以操作, 管理员不能禁言管理员及以上权限的成员
	SpeakStatus *int `json:"speak_status" enums:"0,1" example:"1"`
}

// UpdateGroupMemberHandler
// @Summary      更新群成员
// @Tags         群组
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      UpdateGroupMemberRequest  true  "请求JSON数据体"
// @Security 	 APIKeyHeader
// @Success      200  {object}  Response{}
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/group/member/update [post]
func UpdateGroupMemberHandler(ctx *gin.Context) {
	req := new(UpdateGroupMemberRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	if req.GroupID <= 0 {
		JSONError(ctx, StatusError, "'group_id'无效")
		return
	}

	if req.UserID <= 0 {
		JSONError(ctx, StatusError, "'user_id'无效")
		return
	}

	currentUser := LoginUserFromContext(ctx)

	if req.UserID == currentUser.ID {
		JSONError(ctx, StatusError, "无法修改自身信息")
		return
	}

	if goutils.EqualAll(nil, req.Role, req.SpeakStatus) {
		JSONError(ctx, StatusError, "更改的项未填写")
		return
	}

	memberIDs := []int64{currentUser.ID, req.UserID}

	members, err := database.GetGroupMembers(req.GroupID, memberIDs)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, "不是该群成员")
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Ints64("member_ids", memberIDs).Msg("获取群成员信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	var editor *database.GroupMember
	var member *database.GroupMember
	for i := 0; i < len(members); i++ {
		u := members[i]
		if u.UserID == currentUser.ID {
			editor = u
		}

		if u.UserID == req.UserID {
			member = u
		}
	}

	if editor == nil {
		JSONError(ctx, StatusError, "您不是该群成员")
		return
	}

	if member == nil {
		JSONError(ctx, StatusError, "该用户不是该群成员")
		return
	}

	// 普通用户无权限
	if !goutils.In(editor.Role, 1, 2) {
		JSONError(ctx, StatusError, "无权限")
		return
	}

	// 禁止修改群主信息
	if member.Role == 1 {
		JSONError(ctx, StatusError, "无权限")
		return
	}

	// 管理员无法修改同级成员信息
	if editor.Role == 2 && member.Role == 2 {
		JSONError(ctx, StatusError, "无法修改同是管理员的信息")
		return
	}

	updateData := database.UpdateGroupMemberData{}
	// 改变角色:将目标角色设置成管理员(2-manager)/取消管理员(2-manager)
	//		条件: 编辑人角色必须是群主(1-owner)
	if req.Role != nil {
		if !goutils.In(*req.Role, 0, 2) {
			JSONError(ctx, StatusError, "无效的角色类型")
			return
		}

		// 操作人不是群主或者将对象修改成群主都不可行
		if editor.Role != 1 || *req.Role == 1 {
			JSONError(ctx, StatusError, "您不是群主,无法操作")
			return
		}

		// 如果是升级成管理员,则需要将其他参数设置成管理员权限状态
		if *req.Role == 2 {
			status := 1
			req.SpeakStatus = &status
		}

		updateData.Role = req.Role
	}

	// 改变目标发言状态(2-speak_status)
	//		条件:	1) 编辑人角色必须是管理员以上级别(1-owner||2-manager)
	//		     	2) 编辑人角色是管理员级别(2-manager),目标必须是管理员级别以下(0-normal)
	//
	if req.SpeakStatus != nil {
		// 管理员或者即将被设置成管理员的对象无法被禁言
		if *req.SpeakStatus == 0 && (member.Role >= 1 || (req.Role != nil && *req.Role >= 1)) {
			JSONError(ctx, StatusError, "管理员无法被禁言")
			return
		}

		updateData.SpeakStatus = req.SpeakStatus
	}

	err = database.UpdateGroupMember(req.GroupID, req.UserID, &updateData)
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Int64("member_id", req.UserID).Msg("更新群成员信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}
	JSON(ctx)

}

// RemoveGroupMemberRequest 移除群成员请求参数
// @Description 移除群成员请求参数
type RemoveGroupMemberRequest struct {
	// GroupID 群ID
	GroupID int64 `json:"group_id" binding:"required" min:"1" example:"1"`

	// UserIDs 当UserIDs为空时,表示主动当前用户退出群组,当不为空是表示将他人提出群组
	UserIDs []int64 `json:"user_ids" binding:"required" example:"1,2,3,4"`
}

// RemoveGroupMemberResponse 移除群成员返回参数
// @Description 移除群返回请求参数
type RemoveGroupMemberResponse struct {
	// Count 成功移除成员数量
	Count int64 `json:"count" binding:"required" example:"22"`
}

// RemoveGroupMemberHandler
// @Summary      移除群成员
// @Tags         群组
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      RemoveGroupMemberRequest  true  "请求JSON数据体"
// @Security 	 APIKeyHeader
// @Success      200  {object}  Response{data=RemoveGroupMemberResponse}
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/group/member/remove [post]
func RemoveGroupMemberHandler(ctx *gin.Context) {
	req := new(RemoveGroupMemberRequest)
	err := ctx.BindJSON(req)
	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	if req.GroupID <= 0 {
		JSONError(ctx, StatusError, "'group_id'无效")
		return
	}

	if len(req.UserIDs) == 0 {
		JSONError(ctx, StatusError, "'user_ids'不能为空")
		return
	}

	currentUser := LoginUserFromContext(ctx)

	// 判断当前操作人员是否是该群成员
	member, err := database.GetGroupMember(req.GroupID, currentUser.ID)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, "您不是该群成员")
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Int64("member_id", currentUser.ID).Msg("获取群成员信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	// 判断当前操作人员是否是管理员
	if member.Role == 0 {
		JSONError(ctx, StatusError, "您不是管理员")
		return
	}

	var cnt int64
	var rgmf = database.RemoveGroupMembersFilter{GroupID: req.GroupID, UserIDs: req.UserIDs}
	// 群主可以移除非群主的成员
	if member.Role == 1 {
		rgmf.Roles = []int{2} // 增加可删除管理员
	}
	cnt, err = database.RemoveGroupMembers(&rgmf)
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).Str("err_format", fmt.Sprintf("%+v", err)).Int64("group_id", req.GroupID).Ints64("member_id", req.UserIDs).Msg("移除群成员失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	rsp := &RemoveGroupMemberResponse{Count: cnt}
	JSON(ctx, rsp)

}
