package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jerbe/jcache"
	"github.com/jerbe/jim/errors"
	"github.com/jerbe/jim/log"
	"github.com/jerbe/jim/utils"
	"github.com/jmoiron/sqlx"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/15 17:40
  @describe :
*/

// UserRelation 用户关系信息
type UserRelation struct {
	ID int64 `db:"id" json:"id"`

	// UserAID 用户A ID
	UserAID int64 `db:"user_a_id" json:"user_a_id"`

	// UserBID 用户B ID
	UserBID int64 `db:"user_b_id" json:"user_b_id"`

	// Status 好友状态
	Status int `db:"status" json:"status"`

	// BlockStatus 拉黑状态
	BlockStatus int `db:"block_status" json:"block_status"`

	// RemarkOnA B对A的备注,也就是B显示A的备注名称
	RemarkOnA string `db:"remark_on_a" json:"remark_on_a"`

	// RemarkOnB A对B的备注,也就是A显示B的备注名称
	RemarkOnB string `db:"remark_on_b" json:"remark_on_b"`

	// UpdatedAt 更新时间
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// CreatedAt 创建时间
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func (r *UserRelation) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

func (r *UserRelation) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}

// AddUserRelation 添加一条用户关系
func AddUserRelation(relation *UserRelation, opts ...*SetOptions) (id int64, err error) {
	opt := MergeSetOptions(opts)

	// 基本参数验证,防止入脏数据
	if relation.UserAID == 0 {
		return 0, errors.New("user_a_id can not zero")
	}
	if relation.UserBID == 0 {
		return 0, errors.New("user_b_id can not zero")
	}

	if relation.UpdatedAt.IsZero() && relation.CreatedAt.IsZero() {
		now := time.Now()
		relation.CreatedAt, relation.UpdatedAt = now, now
	}
	if !relation.UpdatedAt.IsZero() && relation.CreatedAt.IsZero() {
		relation.CreatedAt = relation.UpdatedAt
	}
	if !relation.CreatedAt.IsZero() && relation.UpdatedAt.IsZero() {
		relation.UpdatedAt = relation.CreatedAt
	}

	defer func() {
		if err == nil && opt.UpdateCache() {
			jcache.Set(GlobCtx, cacheKeyFormatUserRelationID(relation.ID), relation, jcache.DefaultExpirationDuration)
			jcache.Set(GlobCtx, cacheKeyFormatUserRelationUserIDs(relation.UserAID, relation.UserBID), relation, jcache.DefaultExpirationDuration)
		}
	}()

	{
		// 尝试先更新
		// 先找到是否有记录存在
		existRelation, err := GetUserRelationByUsersID(relation.UserAID, relation.UserBID)
		// 如果有错误并且不是没记录的错误,则报错
		if err != nil && !errors.IsNoRecord(err) {
			return 0, errors.Wrap(err)
		}

		updateFilter := &UpdateUserRelationFilter{UserAID: relation.UserAID, UserBID: relation.UserBID}
		updateData := &UpdateUserRelationData{
			Status:      &relation.Status,
			BlockStatus: &relation.BlockStatus,
			UpdatedAt:   relation.UpdatedAt,
		}
		if relation.RemarkOnA != "" {
			updateData.RemarkOnA = &relation.RemarkOnA
		}
		if relation.RemarkOnB != "" {
			updateData.RemarkOnB = &relation.RemarkOnB
		}

		// 数据库里已经存在记录
		if err == nil && existRelation != nil {
			_, err := UpdateUserRelation(updateFilter, updateData, opt)
			if err != nil {
				return 0, errors.Wrap(err)
			}

			relation.ID = existRelation.ID
			relation.CreatedAt = existRelation.CreatedAt
			return relation.ID, nil
		}
	}

	// 插入关系数据
	relationClone := *relation
	relationClone.UserAID, relationClone.UserBID = utils.SortInt(relation.UserAID, relation.UserBID)
	sqlQuery := fmt.Sprintf(
		"INSERT INTO %s (`user_a_id`, `user_b_id`, `status`, `block_status`, `updated_at`, `created_at`) VALUES (:user_a_id, :user_b_id, :status, :block_status, :updated_at, :created_at) ON DUPLICATE KEY UPDATE `status` = :status, `block_status` = :block_status, `updated_at` =:updated_at",
		TableUserRelation)
	rs, err := sqlx.NamedExec(opt.SQLExt(), sqlQuery, relationClone)
	if err != nil {
		return 0, errors.Wrap(err)
	}

	id, err = rs.LastInsertId()
	if err != nil {
		return id, errors.Wrap(err)
	}
	relation.ID = id

	return id, nil
}

// CreateUserRelationTx 事务方式创建用户关系
// 1. 更新用户好友邀请表
// 2. 添加用户关系表
func CreateUserRelationTx(relation *UserRelation) (id int64, err error) {
	var tx *sqlx.Tx
	tx, err = GlobDB.MySQL.BeginTxx(GlobCtx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, errors.Wrap(err)
	}
	// 更新建立关系邀请表
	var invite *UserRelationInvite

	defer func() {
		if err == nil {
			err = tx.Commit()
			if err == nil {
				log.Error().Err(err).Msg("事务提交失败")
			}
		}

		if err != nil {
			err = tx.Rollback()
			if err != nil {
				log.Error().Err(err).Msg("事务回滚失败")
			}
			return
		}

		// 删除缓存
		jcache.Del(GlobCtx, cacheKeyFormatUserRelationID(relation.ID))
		jcache.Del(GlobCtx, cacheKeyFormatUserRelationUserIDs(relation.UserAID, relation.UserAID))

		if invite != nil {
			jcache.Del(GlobCtx, cacheKeyFormatUserRelationInviteID(invite.ID))
			jcache.Del(GlobCtx, cacheKeyFormatUserRelationInviteUserIDs(invite.UserID, invite.TargetID))
		}
	}()

	setOptions := NewSetOptions().SetSQLExt(tx)
	getOptions := NewGetOptionsFromSetOptions(setOptions)

	invite, err = GetLastUserRelationInvite(relation.UserAID, relation.UserBID, getOptions)
	if err != nil && !errors.IsNoRecord(err) {
		return 0, errors.Wrap(err)
	}

	if err == nil && invite != nil {
		updateFilter := &UpdateUserRelationInviteFilter{
			ID: invite.ID,
		}

		updateData := &UpdateUserRelationInviteData{
			Status:    UserRelationInviteStatusAgree,
			UpdatedAt: relation.UpdatedAt,
		}

		_, err = UpdateUserRelationInvite(updateFilter, updateData, setOptions)
		if err != nil {
			return 0, errors.Wrap(err)
		}
	}

	id, err = AddUserRelation(relation, setOptions)
	return id, errors.Wrap(err)
}

// GetUserRelation 获取用户关系信息
func GetUserRelation(id int64, opts ...*GetOptions) (*UserRelation, error) {
	opt := MergeGetOptions(opts)
	if opt.UseCache() {
		cacheKey := cacheKeyFormatUserRelationID(id)
		exists, _ := jcache.Exists(GlobCtx, cacheKey)
		if exists {
			relation := &UserRelation{}
			err := jcache.CheckAndScan(GlobCtx, relation, cacheKey)
			if err == nil {
				return relation, nil
			}

			// 如果有记录,并且记录内容为空,则表示被标记成查询空
			if errors.IsEmptyRecord(err) {
				return nil, errors.Wrap(err)
			}
		}
	}

	sqlQuery := fmt.Sprintf("SELECT `id`, `user_a_id`, `user_b_id`, `status`, `block_status`, `updated_at`, `created_at` FROM %s WHERE `id` = ? ", TableUserRelation)

	relation := &UserRelation{}
	err := sqlx.Get(opt.SQLExt(), relation, sqlQuery, id)
	if err != nil {
		if err == sql.ErrNoRows {
			// 写入缓存,如果key不存在的话
			var cacheKey = cacheKeyFormatUserRelationID(id)
			if err := jcache.SetNX(GlobCtx, cacheKey, nil, jcache.DefaultEmptySetNXDuration); err != nil {
				log.Error().Err(err).Str("cache_key", cacheKey).Msg("缓存写入失败")
			}
		}
		return nil, errors.Wrap(err)
	}

	if opt.UpdateCache() {
		jcache.Set(GlobCtx, cacheKeyFormatUserRelationID(relation.ID), relation, jcache.DefaultExpirationDuration)
		jcache.Set(GlobCtx, cacheKeyFormatUserRelationUserIDs(relation.UserAID, relation.UserBID), relation, jcache.DefaultExpirationDuration)
	}

	return relation, nil
}

// GetUserRelationByUsersID 根据用户id获取用户关系
func GetUserRelationByUsersID(userAID int64, userBID int64, opts ...*GetOptions) (*UserRelation, error) {
	opt := MergeGetOptions(opts)
	if opt.UseCache() {
		cacheKey := cacheKeyFormatUserRelationUserIDs(userAID, userBID)
		exists, _ := jcache.Exists(GlobCtx, cacheKey)
		if exists {
			relation := &UserRelation{}
			err := jcache.CheckAndScan(GlobCtx, relation, cacheKey)
			if err == nil {
				return relation, nil
			}

			// 如果有记录,并且记录内容为空,则表示被标记成查询空
			if errors.IsEmptyRecord(err) {
				return nil, errors.Wrap(err)
			}
		}

	}

	// 入库已经a比b小,省去OR条件
	a, b := utils.SortInt(userAID, userBID)
	sqlQuery := fmt.Sprintf("SELECT `id`, `user_a_id`, `user_b_id`, `status`,`block_status`, `updated_at`, `created_at` FROM %s WHERE `user_a_id` = ? AND `user_b_id` = ?", TableUserRelation)

	relation := &UserRelation{}
	err := sqlx.Get(opt.SQLExt(), relation, sqlQuery, a, b)
	if err != nil {
		if err == sql.ErrNoRows {
			// 写入缓存,如果key不存在的话
			var cacheKey = cacheKeyFormatUserRelationUserIDs(a, b)
			if err := jcache.SetNX(GlobCtx, cacheKey, nil, jcache.DefaultEmptySetNXDuration); err != nil {
				log.Error().Err(err).Str("cache_key", cacheKey).Msg("缓存写入失败")
			}
		}
		return nil, errors.Wrap(err)
	}

	if opt.UpdateCache() {
		// 将关系数据写入到缓存中去
		jcache.Set(GlobCtx, cacheKeyFormatUserRelationID(relation.ID), relation, jcache.DefaultExpirationDuration)
		jcache.Set(GlobCtx, cacheKeyFormatUserRelationUserIDs(relation.UserAID, relation.UserBID), relation, jcache.DefaultExpirationDuration)

	}
	return relation, nil
}

// UpdateUserRelationFilter 更新用户关系过滤器
type UpdateUserRelationFilter struct {
	ID int64 `db:"id" json:"id"`

	// UserAID 用户A ID
	UserAID int64 `db:"user_a_id" json:"user_a_id"`

	// UserBID 用户B ID
	UserBID int64 `db:"user_b_id" json:"user_b_id"`
}

// UpdateUserRelationData 更新用户关系配置
type UpdateUserRelationData struct {
	// RemarkOnA B给A的备注
	RemarkOnA *string `db:"remark_on_a" json:"remark_on_a"`

	// RemarkOnB A给B的备注
	RemarkOnB *string `db:"remark_on_b" json:"remark_on_b"`

	// Status 好友状态
	Status *int `db:"status" json:"status"`

	// BlockStatus 拉黑状态
	BlockStatus *int `db:"block_status" json:"block_status"`

	// UpdatedAt 更新时间
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// SetStatus 设置好友状态
func (opt *UpdateUserRelationData) SetStatus(val int) *UpdateUserRelationData {
	opt.Status = &val
	return opt
}

// SetBlockStatus 设置拉黑状态
func (opt *UpdateUserRelationData) SetBlockStatus(val int) *UpdateUserRelationData {
	opt.BlockStatus = &val
	return opt
}

// SetUpdatedAt 设置更新时间
func (opt *UpdateUserRelationData) SetUpdatedAt(val time.Time) *UpdateUserRelationData {
	opt.UpdatedAt = val
	return opt
}

// UpdateUserRelation 更新用户关系资料
func UpdateUserRelation(filter *UpdateUserRelationFilter, data *UpdateUserRelationData, opts ...*SetOptions) (int64, error) {
	opt := MergeSetOptions(opts)

	// 与源数据一样,不用更新
	if utils.Equal(nil, data.Status, data.BlockStatus, data.RemarkOnA, data.RemarkOnB) {
		return 0, errors.ParamsInvalid
	}

	var whereSQLs []string
	var setSQLs []string
	if filter.ID > 0 {
		whereSQLs = append(whereSQLs, "`id` = :id")
	}

	if (filter.UserAID > 0 && filter.UserBID <= 0) || (filter.UserBID > 0 && filter.UserAID <= 0) {
		return 0, errors.ParamsInvalid
	}

	if filter.UserAID > 0 && filter.UserBID > 0 {
		filter.UserAID, filter.UserBID = utils.SortInt(filter.UserAID, filter.UserBID)
		whereSQLs = append(whereSQLs, "`user_a_id` = :user_a_id", "`user_b_id` = :user_b_id")
	}

	if len(whereSQLs) == 0 {
		return 0, errors.ParamsInvalid
	}

	if data.Status != nil {
		setSQLs = append(setSQLs, "`status` = :status ")
		whereSQLs = append(whereSQLs, " `status` != :status ")
	}

	if data.BlockStatus != nil {
		setSQLs = append(setSQLs, "`block_status` = :block_status ")
		whereSQLs = append(whereSQLs, " `block_status` != :block_status ")
	}

	if data.RemarkOnA != nil {
		setSQLs = append(setSQLs, "`remark_on_a` = :remark_on_a ")
		whereSQLs = append(whereSQLs, " `remark_on_a` != :remark_on_a ")
	}

	if data.RemarkOnB != nil {
		setSQLs = append(setSQLs, "`remark_on_b` = :remark_on_b ")
		whereSQLs = append(whereSQLs, " `remark_on_b` != :remark_on_b ")
	}

	if len(setSQLs) == 0 {
		return 0, errors.ParamsInvalid
	}

	if data.UpdatedAt.IsZero() {
		t := time.Now()
		data.UpdatedAt = t
	}

	setSQLs = append(setSQLs, "`updated_at` = :updated_at ")

	// 实例化一个struct
	type namedStruct struct {
		*UpdateUserRelationFilter
		*UpdateUserRelationData
	}

	sqlQuery := fmt.Sprintf("UPDATE %s SET %s WHERE %s", TableUserRelation,
		strings.Join(setSQLs, ","),
		strings.Join(whereSQLs, " AND "))
	rs, err := sqlx.NamedExec(opt.SQLExt(), sqlQuery, namedStruct{filter, data})
	if err != nil {
		return 0, errors.Wrap(err)
	}
	cnt, err := rs.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err)
	}

	if opt.UpdateCache() {
		relation := new(UserRelation)
		var got bool
		if filter.ID > 0 {
			got = true
			err = jcache.CheckAndScan(GlobCtx, relation, cacheKeyFormatUserRelationID(filter.ID))
		}

		if filter.UserAID > 0 && (!got || err != nil) {
			got = true
			err = jcache.CheckAndScan(GlobCtx, relation, cacheKeyFormatUserRelationUserIDs(filter.UserAID, filter.UserBID))
		}

		if got && err == nil {
			jcache.Del(GlobCtx, cacheKeyFormatUserRelationID(relation.ID))
			jcache.Del(GlobCtx, cacheKeyFormatUserRelationUserIDs(relation.UserAID, relation.UserBID))
		}
	}
	return cnt, nil
}

// UpdateUserRelationTx 事务模式更新用户关系资料
// 1. 更新用户关系表
// 2. 更新用户好友邀请表
func UpdateUserRelationTx(filter *UpdateUserRelationFilter, data *UpdateUserRelationData) (rowsAffected int64, err error) {
	var tx *sqlx.Tx
	var relation *UserRelation
	tx, err = GlobDB.MySQL.BeginTxx(GlobCtx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, errors.Wrap(err)
	}
	var invite *UserRelationInvite

	defer func() {
		if err != nil {
			err1 := tx.Rollback()
			if err1 != nil {
				log.Error().Err(err1).Msg("事务回滚失败")
			}
			return
		}

		err1 := tx.Commit()
		if err1 != nil {
			log.Error().Err(err1).Msg("事务提交失败")
			return
		}

		// 删除相关缓存
		if relation != nil {
			jcache.Del(GlobCtx, cacheKeyFormatUserRelationID(relation.ID))
			jcache.Del(GlobCtx, cacheKeyFormatUserRelationUserIDs(relation.UserAID, relation.UserBID))
		}
		if invite != nil {
			jcache.Del(GlobCtx, cacheKeyFormatUserRelationInviteID(invite.ID))
			jcache.Del(GlobCtx, cacheKeyFormatUserRelationInviteUserIDs(invite.UserID, invite.TargetID))
		}
	}()

	setOptions := NewSetOptions().SetSQLExt(tx)
	getOptions := NewGetOptionsFromSetOptions(setOptions)

	// 先获取数据
	if filter.ID > 0 {
		relation, err = GetUserRelation(filter.ID)
	} else if filter.UserAID > 0 {
		relation, err = GetUserRelationByUsersID(filter.UserAID, filter.UserBID)
	}

	if err != nil {
		return 0, errors.Wrap(err)
	}
	if err == nil && relation == nil {
		return 0, errors.NoRecords
	}

	// 进行用户关系的修改
	rowsAffected, err = UpdateUserRelation(filter, data)
	if err != nil {
		return 0, errors.Wrap(err)
	}

	// 如果状态是成为好友,才会更新邀请记录
	if data.Status != nil && *data.Status == 0b11 {
		// 获取最后一条用户关系的邀请记录
		invite, err = GetLastUserRelationInvite(relation.UserAID, relation.UserBID, getOptions)
		if err != nil && !errors.IsNoRecord(err) {
			return 0, errors.Wrap(err)
		}

		// 如果存在邀请记录并且没有错误
		if err == nil && invite != nil {
			updateFilter := &UpdateUserRelationInviteFilter{
				ID: invite.ID,
			}
			updateData := &UpdateUserRelationInviteData{
				Status:    UserRelationInviteStatusAgree,
				UpdatedAt: relation.UpdatedAt,
			}

			_, err = UpdateUserRelationInvite(updateFilter, updateData, setOptions)
			if err != nil && !errors.InIs(err, errors.NoRecords, errors.NotChange) {
				return 0, errors.Wrap(err)
			}
		}
	}

	return rowsAffected, nil
}

// ==============================================================
// ================== CACHE MANAGE ==============================
// ==============================================================

// cacheKeyFormatUserRelationID 格式化用户关系ID的缓存 key
func cacheKeyFormatUserRelationID(id int64) string {
	return fmt.Sprintf("%s:user_relation:id:%d", CacheKeyPrefix, id)
}

// cacheKeyFormatUserRelationUserIDs 格式化用户关系的用户ID 缓存 key
func cacheKeyFormatUserRelationUserIDs(userAID, userBID int64) string {
	a, b := utils.SortInt(userAID, userBID)
	return fmt.Sprintf("%s:user_relation:user_id:%d_%d", CacheKeyPrefix, a, b)
}
