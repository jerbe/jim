package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jerbe/jcache"
	"github.com/jerbe/jim/errors"
	"github.com/jerbe/jim/log"
	"github.com/jerbe/jim/utils"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/21 12:03
  @describe :
*/

const (
	// UserRelationInviteStatusPending 等待确认用户关系邀请
	UserRelationInviteStatusPending = iota

	// UserRelationInviteStatusAgree 接收用户关系邀请
	UserRelationInviteStatusAgree

	// UserRelationInviteStatusReject 拒绝用户关系邀请
	UserRelationInviteStatusReject
)

// UserRelationInvite 用户关系邀请
type UserRelationInvite struct {
	// ID 请求ID
	ID int64 `db:"id" json:"id"`

	// UserID 用户ID
	UserID int64 `db:"user_id" json:"user_id"`

	// TargetID 申请目标的用户ID
	TargetID int64 `db:"target_id" json:"target_id"`

	// Note 申请备注
	Note string `db:"note" json:"note,omitempty"`

	// Reply 申请回复
	Reply string `db:"reply" json:"reply,omitempty"`

	// Status 状态 0:待确认,1:已通过,2:已拒绝
	Status int `db:"status" json:"status"`

	// UpdatedAt 更新时间
	UpdatedAt time.Time `db:"updated_at" json:"updated_at,omitempty"`

	// CreatedAt 创建时间
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	// UQFlag 控制记录是唯一的申请记录标记, 使用userID跟targetID联合
	UQFlag string `db:"uq_flag" json:"uq_flag,omitempty"`
}

func (i *UserRelationInvite) MarshalBinary() (data []byte, err error) {
	return json.Marshal(i)
}

func (i *UserRelationInvite) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, i)
}

// AddUserRelationInvite 添加一条用户关系建立的请求记录
func AddUserRelationInvite(invite *UserRelationInvite, opts ...*SetOptions) error {
	opt := MergeSetOptions(opts)

	if invite.UpdatedAt.IsZero() && !invite.CreatedAt.IsZero() {
		invite.UpdatedAt = invite.CreatedAt
	}

	if !invite.UpdatedAt.IsZero() && invite.CreatedAt.IsZero() {
		invite.CreatedAt = invite.UpdatedAt
	}

	if invite.UpdatedAt.IsZero() && invite.CreatedAt.IsZero() {
		now := time.Now()
		invite.UpdatedAt, invite.CreatedAt = now, now
	}
	invite.Status = 0

	// 升序排序
	a, b := utils.SortInt(invite.UserID, invite.TargetID)
	invite.UQFlag = fmt.Sprintf("%d_%d", a, b)

	sqlQuery := fmt.Sprintf("INSERT INTO %s (`user_id`,`target_id`,`note`,`reply`,`status`,`updated_at`,`created_at`,`uq_flag`) VALUES (:user_id, :target_id, :note, :reply, :status, :updated_at, :created_at, :uq_flag)", TableUserRelationInvite)
	result, err := sqlx.NamedExec(opt.SQLExt(), sqlQuery, invite)
	if err != nil {
		return errors.Wrap(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err)
	}
	invite.ID = id

	if opt.UpdateCache() {
		GlobCache.Set(GlobCtx, cacheKeyFormatUserRelationInviteID(invite.ID), invite, jcache.DefaultExpirationDuration)
		GlobCache.Set(GlobCtx, cacheKeyFormatUserRelationInviteUserIDs(invite.UserID, invite.TargetID), invite, jcache.DefaultExpirationDuration)
	}
	return nil
}

// GetUserRelationInvite 获取一条建立关系邀请记录
func GetUserRelationInvite(id int64, opts ...*GetOptions) (*UserRelationInvite, error) {
	opt := MergeGetOptions(opts)
	if opt.UseCache() {
		cacheKey := cacheKeyFormatUserRelationInviteID(id)
		exits, err := GlobCache.Exists(GlobCtx, cacheKey)
		if exits > 0 {
			invite := &UserRelationInvite{}
			err = GlobCache.CheckAndScan(GlobCtx, invite, cacheKey)
			if err == nil {
				return invite, nil
			}

			if errors.IsEmptyRecord(err) {
				return nil, errors.Wrap(err)
			}
		}

	}

	// 通过数据库获取一条记录
	sqlQuery := fmt.Sprintf("SELECT `id`, `user_id`,`target_id`,`note`,`reply`,`status`,`updated_at`,`created_at`,`uq_flag` FROM %s WHERE `id` = ?", TableUserRelationInvite)

	invite := new(UserRelationInvite)
	err := sqlx.Get(opt.SQLExt(), invite, sqlQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 写入缓存,如果key不存在的话
			var cacheKey = cacheKeyFormatUserRelationInviteID(id)
			if err := GlobCache.SetNX(GlobCtx, cacheKey, nil, jcache.DefaultEmptySetNXDuration); err != nil {
				log.Error().Err(err).Str("key", cacheKey).Msg("写入缓存失败")
			}
		}
		return nil, errors.Wrap(err)
	}

	if opt.UpdateCache() {
		GlobCache.Set(GlobCtx, cacheKeyFormatUserRelationInviteID(id), invite, jcache.DefaultExpirationDuration)
		GlobCache.Set(GlobCtx, cacheKeyFormatUserRelationInviteUserIDs(invite.UserID, invite.TargetID), invite, jcache.DefaultExpirationDuration)
	}

	return invite, nil
}

// GetLastUserRelationInvite 获取一条建立用户关系请求记录
func GetLastUserRelationInvite(userID, targetID int64, opts ...*GetOptions) (*UserRelationInvite, error) {
	opt := MergeGetOptions(opts)
	if opt.UseCache() {
		cacheKey := cacheKeyFormatUserRelationInviteUserIDs(userID, targetID)
		exists, _ := GlobCache.Exists(GlobCtx, cacheKey)
		if exists > 0 {
			invite := &UserRelationInvite{}
			err := GlobCache.CheckAndScan(GlobCtx, invite, cacheKey)
			if err == nil {
				return invite, nil
			}
			if errors.IsEmptyRecord(err) {
				return nil, errors.Wrap(err)
			}
		}

	}

	sqlQuery := fmt.Sprintf("SELECT `id`, `user_id`,`target_id`,`note`,`reply`,`status`,`updated_at`,`created_at`,`uq_flag` FROM %s WHERE ((`user_id` = ? AND `target_id` = ?) OR (`user_id` = ? AND `target_id` = ?)) AND `status` = 0", TableUserRelationInvite)

	invite := new(UserRelationInvite)
	err := sqlx.Get(opt.SQLExt(), invite, sqlQuery, userID, targetID, targetID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 写入缓存,如果key不存在的话
			var cacheKey = cacheKeyFormatUserRelationInviteUserIDs(userID, targetID)
			if err := GlobCache.SetNX(GlobCtx, cacheKey, nil, jcache.DefaultEmptySetNXDuration); err != nil {
				log.Warn().Err(err).Str("key", cacheKey).Str("err_format", fmt.Sprintf("%+v", err)).Msg("写入缓存失败")
			}
		}
		return nil, errors.Wrap(err)
	}

	if opt.UpdateCache() {
		GlobCache.Set(GlobCtx, cacheKeyFormatUserRelationInviteID(invite.ID), invite, jcache.DefaultExpirationDuration)
		GlobCache.Set(GlobCtx, cacheKeyFormatUserRelationInviteUserIDs(invite.UserID, invite.TargetID), invite, jcache.DefaultExpirationDuration)
	}

	return invite, nil
}

// UpdateUserRelationInviteFilter 更新用户关系邀请过滤器
type UpdateUserRelationInviteFilter struct {
	// ID 邀请记录ID. (查询条件)
	ID int64 `db:"id"`

	// UserID 用户ID
	UserID int64 `db:"user_id"`

	// TargetID 目标ID
	TargetID int64 `db:"target_id"`
}

// UpdateUserRelationInviteData 更新用户关系邀请数据
type UpdateUserRelationInviteData struct {
	// Status 邀请状态,更新是只能大于0. (设置结果)
	Status int `db:"status"`

	// Reply 状态回复. (设置结果)
	Reply string `db:"reply"`

	// UpdatedAt 更新时间,如果未设置,将自动生成当前时间. (设置结果)
	UpdatedAt time.Time `db:"updated_at"`
}

// UpdateUserRelationInvite 更新用户关系邀请
func UpdateUserRelationInvite(filter *UpdateUserRelationInviteFilter, data *UpdateUserRelationInviteData, opts ...*SetOptions) (int64, error) {
	opt := MergeSetOptions(opts)

	// 基本参数判断
	// 如果 ID 未设置,则返回
	if filter.ID < 0 || filter.UserID < 0 || filter.TargetID < 0 {
		return 0, errors.ParamsInvalid
	}

	// 查询参数
	whereSQLs := []string{}
	if filter.ID > 0 {
		whereSQLs = append(whereSQLs, " `id` = :id ")
	}

	if filter.UserID > 0 {
		whereSQLs = append(whereSQLs, " ((`user_id` = :user_id AND `target_id` = :target_id) OR (`user_id` = :target_id AND `target_id` = :user_id))")
	}

	whereSQLs = append(whereSQLs, "`status` = 0 ")

	// 设置参数
	setSQLs := []string{
		"`status` = :status ",
	}

	if data.Reply != "" {
		setSQLs = append(setSQLs, "`reply` = :reply ")
	}

	if data.UpdatedAt.IsZero() {
		now := time.Now()
		data.UpdatedAt = now
	}

	setSQLs = append(setSQLs, fmt.Sprintf("`uq_flag` = CONCAT(`uq_flag`, '_%d')", data.UpdatedAt.Unix()), "`updated_at` = :updated_at")

	// 邀请记录,提前查出.才可判断是否有必要进行更新
	var err error
	type named struct {
		*UpdateUserRelationInviteFilter
		*UpdateUserRelationInviteData
	}

	// 执行更新
	sqlQuery := fmt.Sprintf("UPDATE %s SET %s WHERE %s", TableUserRelationInvite, strings.Join(setSQLs, ","), strings.Join(whereSQLs, " AND "))
	rs, err := sqlx.NamedExec(opt.SQLExt(), sqlQuery, named{filter, data})
	if err != nil {
		return 0, errors.Wrap(err)
	}
	cnt, err := rs.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err)
	}

	if opt.UpdateCache() {
		invite := new(UserRelationInvite)
		var got bool

		idKey := cacheKeyFormatUserRelationInviteID(filter.ID)
		usersKey := cacheKeyFormatUserRelationInviteUserIDs(filter.UserID, filter.TargetID)
		if filter.ID > 0 {
			got = true
			err = GlobCache.CheckAndScan(GlobCtx, invite, idKey)
		}

		if filter.UserID > 0 && (!got || err != nil) {
			got = true
			err = GlobCache.CheckAndScan(GlobCtx, invite, usersKey)
		}

		if got && err == nil {
			GlobCache.Del(GlobCtx, idKey)
			GlobCache.Del(GlobCtx, usersKey)
		}
	}

	return cnt, nil
}

// ========================================================================
// ======================= CACHE CONTROL ==================================
// ========================================================================
// cacheKeyFormatUserRelationInviteID 格式化根据ID获取建立用户关系请求的键
func cacheKeyFormatUserRelationInviteID(id int64) string {
	return fmt.Sprintf("%s:user_relation_invite:id:%d", CacheKeyPrefix, id)
}

// cacheKeyFormatUserRelationInviteUserIDs 格式化根据用户ID跟目标用户ID获取建立用户关系请求的键
func cacheKeyFormatUserRelationInviteUserIDs(userID, targetID int64) string {
	a, b := utils.SortInt(userID, targetID)
	return fmt.Sprintf("%s:user_relation_invite:user_id:%d_%d", CacheKeyPrefix, a, b)
}
