package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jerbe/jim/errors"
	"github.com/jerbe/jim/log"
	"github.com/jerbe/jim/utils"

	"github.com/jerbe/jcache"

	"github.com/jmoiron/sqlx"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/29 19:08
  @describe :
*/

// Group 群组信息
type Group struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	MaxMember   int       `db:"max_member" json:"max_member"`
	OwnerID     int64     `db:"owner_id" json:"owner_id"`
	SpeakStatus int       `db:"speak_status" json:"speak_status"`
	CreatorID   int64     `db:"creator_id" json:"creator_id"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdaterID   int64     `db:"updater_id" json:"updater_id"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// GroupMember 群成员信息
type GroupMember struct {
	ID          int64     `db:"id" json:"id"`
	GroupID     int64     `db:"group_id" json:"group_id"`
	UserID      int64     `db:"user_id" json:"user_id"`
	Role        int       `db:"role" json:"role"`
	SpeakStatus int       `db:"speak_status" json:"speak_status"`
	UpdaterID   int64     `db:"updater_id" json:"updater_id"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	CreatorID   int64     `db:"creator_id" json:"creator_id"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// AddGroup 添加群组
func AddGroup(group *Group, opts ...*SetOptions) (int64, error) {
	opt := MergeSetOptions(opts)

	if group.OwnerID == 0 {
		return 0, errors.New("owner id is zero")
	}

	if group.CreatorID == 0 {
		return 0, errors.New("creator id is zero")
	}

	if group.UpdaterID == 0 {
		return 0, errors.New("updater id is zero")
	}

	now := time.Now()
	if group.UpdatedAt.IsZero() {
		group.UpdatedAt = now
	}

	if group.CreatedAt.IsZero() {
		group.CreatedAt = now
	}

	sqlQuery := fmt.Sprintf("INSERT INTO %s (`name`,`max_member`,`owner_id`,`speak_status`,`updated_at`,`updater_id`,`creator_id`,`created_at`) VALUES(:name, :max_member, :owner_id,:speak_status,:updated_at,:updater_id,:creator_id,:created_at)", TableGroups)

	result, err := sqlx.NamedExec(opt.SQLExt(), sqlQuery, group)
	if err != nil {
		return 0, errors.Wrap(err)
	}

	insertId, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err)
	}

	group.ID = insertId

	if opt.UpdateCache() {
		GlobCache.Set(GlobCtx, cacheKeyFormatGroupID(insertId), group, jcache.DefaultExpirationDuration)
	}

	return insertId, nil
}

// GetGroup 获取一条群组消息
func GetGroup(id int64, opts ...*GetOptions) (*Group, error) {
	opt := MergeGetOptions(opts)

	if opt.UseCache() {
		cacheKey := cacheKeyFormatGroupID(id)
		group := new(Group)
		err := GlobCache.CheckAndScan(GlobCtx, group, cacheKey)
		if err == nil {
			return group, nil
		}

		// 如果是空记录,则直接返回找不到
		if errors.IsEmptyRecord(err) {
			return nil, errors.Wrap(err)
		}
	}

	sqlQuery := fmt.Sprintf("SELECT `id`,`name`,`max_member`,`owner_id`,`speak_status`,`updated_at`,`updater_id`,`creator_id`,`created_at` FROM %s WHERE `id` = ? ", TableGroups)

	group := new(Group)
	err := sqlx.Get(opt.SQLExt(), group, sqlQuery, id)
	if err != nil {
		if errors.IsNoRecord(err) {
			// 写入缓存,如果key不存在的话
			var cacheKey = cacheKeyFormatGroupID(id)
			if e := GlobCache.SetNX(GlobCtx, cacheKey, nil, jcache.DefaultEmptySetNXDuration); e != nil {
				log.Error().Err(e).Str("err_format", fmt.Sprintf("%+v", e)).Str("cache_key", cacheKey).Msg("缓存写入失败")
			}
		}
		return nil, errors.Wrap(err)
	}

	if opt.UpdateCache() {
		cacheKey := cacheKeyFormatGroupID(id)
		GlobCache.Set(GlobCtx, cacheKey, group, jcache.RandomExpirationDuration())
	}

	return group, nil
}

type UpdateGroupData struct {
	Name        *string   `db:"name" json:"name"`
	MaxMember   *int      `db:"max_member" json:"max_member"`
	OwnerID     *int64    `db:"owner_id" json:"owner_id"`
	SpeakStatus *int      `db:"speak_status" json:"speak_status"`
	UpdaterID   int64     `db:"updater_id" json:"updater_id"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// UpdateGroup 更新群信息
func UpdateGroup(groupID int64, data *UpdateGroupData, opts ...*SetOptions) error {
	opt := MergeSetOptions(opts)

	if utils.Equal(nil, data.Name, data.MaxMember, data.OwnerID, data.SpeakStatus) {
		return errors.Wrap(errors.NotChange)
	}

	var setSQLs []string
	if data.Name != nil {
		setSQLs = append(setSQLs, "`name` = :name")
	}

	if data.MaxMember != nil {
		setSQLs = append(setSQLs, "max_member = :max_member")
	}

	if data.OwnerID != nil {
		setSQLs = append(setSQLs, "owner_id = :owner_id")
	}

	if data.SpeakStatus != nil {
		setSQLs = append(setSQLs, "speak_status = :speak_status")
	}

	setSQLs = append(setSQLs, "`updated_at` = :updated_at, `updater_id` = :updater_id")

	setSQL, sqlArgs, err := sqlx.Named(strings.Join(setSQLs, ","), data)
	if err != nil {
		return errors.Wrap(err)
	}
	sqlArgs = append(sqlArgs, groupID)

	sqlQuery := fmt.Sprintf("UPDATE %s SET %s WHERE `id` = ? ", TableGroups, setSQL)

	rs, err := opt.SQLExt().Exec(sqlQuery, sqlArgs...)
	if err != nil {
		return errors.Wrap(err)
	}

	cnt, err := rs.RowsAffected()
	if err != nil {
		return errors.Wrap(err)
	}

	if cnt == 0 {
		return errors.Wrap(errors.NotChange)
	}

	defer func() {
		if opt.UpdateCache() {
			GlobCache.Del(GlobCtx, cacheKeyFormatGroupID(groupID))
		}
	}()

	return nil

}

// UpdateGroupTx 使用事务方式更新群信息
func UpdateGroupTx(groupID int64, data *UpdateGroupData) (err error) {
	if utils.Equal(nil, data.Name, data.MaxMember, data.OwnerID, data.SpeakStatus) {
		return errors.Wrap(errors.NotChange)
	}

	ctx, _ := context.WithTimeout(GlobCtx, time.Second*5)
	tx, err := GlobDB.MySQL.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return errors.Wrap(err)
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			if err != nil {
				log.Error().Err(err).Msg("事务提交失败")
			}
		}

		// 如果提交失败,可以做回滚
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			e := tx.Rollback()
			if e != nil {
				log.Error().Err(e).Msg("事务回滚失败")
			}
			return
		}
	}()

	err = UpdateGroup(groupID, data)
	if err != nil {
		return errors.Wrap(err)
	}

	if data.OwnerID != nil {
		sqlQuery := fmt.Sprintf("UPDATE %s SET `role` = 0, `updated_at`=? , `updater_id`=? WHERE `group_id`=? AND `role`=1", TableGroupMembers)
		_, err = tx.Exec(sqlQuery, data.UpdatedAt, data.UpdaterID, groupID)
		if err != nil {
			return errors.Wrap(err)
		}

		sqlQuery = fmt.Sprintf("UPDATE %s SET `role` = 1, `updated_at`=? , `updater_id`=? WHERE `group_id`=? AND `user_id`=?", TableGroupMembers)
		_, err = tx.Exec(sqlQuery, data.UpdatedAt, data.UpdaterID, groupID, data.OwnerID)
		if err != nil {
			return errors.Wrap(err)
		}
	}

	return nil
}

type AddGroupMembersData struct {
	UserIDs   []int64   `db:"user_i_ds"`
	CreatorID int64     `db:"creator_id"`
	CreatedAt time.Time `db:"created_at"`
}

// AddGroupMembers 增加群成员
func AddGroupMembers(groupID int64, data *AddGroupMembersData, opts ...*SetOptions) (int64, error) {
	opt := MergeSetOptions(opts)
	userIDs := data.UserIDs
	if groupID == 0 || len(userIDs) == 0 {
		return 0, errors.Wrap(errors.ParamsInvalid)
	}

	var l = len(userIDs)
	var fields = 6
	valueSQL := make([]string, l)
	valueArgs := make([]any, l*fields)

	for i := 0; i < l; i++ {
		valueSQL[i] = " (?,?,?,?,?,?) "
		valueArgs[i*fields] = groupID
		valueArgs[i*fields+1] = userIDs[i]
		valueArgs[i*fields+2] = data.CreatorID
		valueArgs[i*fields+3] = data.CreatedAt
		valueArgs[i*fields+4] = data.CreatorID
		valueArgs[i*fields+5] = data.CreatedAt
	}

	sqlQuery := fmt.Sprintf("INSERT IGNORE INTO %s (`group_id`,`user_id`,`updater_id`,`updated_at`,`creator_id`,`created_at`) VALUES %s", TableGroupMembers, strings.Join(valueSQL, ","))

	rs, err := opt.SQLExt().Exec(sqlQuery, valueArgs...)
	if err != nil {
		return 0, errors.Wrap(err)
	}
	if opt.UpdateCache() {
		GlobCache.Del(GlobCtx, cacheKeyFormatGroupMembers(groupID))
	}

	rowsAffected, err := rs.RowsAffected()
	return rowsAffected, errors.Wrap(err)
}

// GetGroupMemberCount 获取群成员数量
func GetGroupMemberCount(groupID int64, opts ...*GetOptions) (int64, error) {
	opt := MergeGetOptions(opts)
	if opt.UseCache() {
		cacheKey := cacheKeyFormatGroupMembers(groupID)
		exists, _ := GlobCache.Exists(GlobCtx, cacheKey)
		if exists > 0 {
			cnt, err := GlobCache.HLen(GlobCtx, cacheKey)
			if err == nil {
				return cnt, nil
			}
		}
	}

	sqlQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `group_id` = ?", TableGroupMembers)
	var cnt int64
	err := sqlx.Get(opt.SQLExt(), &cnt, sqlQuery, groupID)
	if err != nil {
		return 0, errors.Wrap(err)
	}
	return cnt, nil
}

// GetGroupMemberIDs 获取群成员ID数组
func GetGroupMemberIDs(groupID int64, opts ...*GetOptions) ([]int64, error) {
	opt := MergeGetOptions(opts)
	var userIDs []int64
	if opt.UseCache() {
		cacheKey := cacheKeyFormatGroupMembers(groupID)
		exists, _ := GlobCache.Exists(GlobCtx, cacheKey)
		if exists > 0 {
			err := GlobCache.HKeysAndScan(GlobCtx, userIDs, cacheKey)
			if err == nil {
				return userIDs, nil
			}
		}
	}

	sqlQuery := fmt.Sprintf("SELECT `user_id` FROM %s WHERE `group_id` = ?", TableGroupMembers)

	rs, err := opt.SQLExt().Query(sqlQuery, groupID)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	defer rs.Close()
	for rs.Next() {
		var id int64
		err = rs.Scan(&id)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		userIDs = append(userIDs, id)
	}

	return userIDs, nil
}

// GetGroupMemberIDsString 获取
func GetGroupMemberIDsString(groupID int64, opts ...*GetOptions) ([]string, error) {
	opt := MergeGetOptions(opts)
	var userIDs []string
	if opt.UseCache() {
		cacheKey := cacheKeyFormatGroupMembers(groupID)
		exists, _ := GlobCache.Exists(GlobCtx, cacheKey)
		if exists > 0 {
			keys, err := GlobCache.HKeys(GlobCtx, cacheKey)
			if err == nil {
				return keys, nil
			}
		}
	}

	sqlQuery := fmt.Sprintf("SELECT `user_id` FROM %s WHERE `group_id` = ?", TableGroupMembers)

	rs, err := opt.SQLExt().Query(sqlQuery, groupID)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	defer rs.Close()
	for rs.Next() {
		var id string
		err = rs.Scan(&id)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		userIDs = append(userIDs, id)
	}

	return userIDs, nil
}

type RemoveGroupMembersFilter struct {
	// 群ID
	GroupID int64 `db:"group_id"`

	// UserIDs 用户ID列表
	UserIDs []int64 `db:"user_ids"`

	// Roles 角色列表
	Roles []int `db:"roles"`
}

// RemoveGroupMembers 移除群成员,不区分管理员
func RemoveGroupMembers(filter *RemoveGroupMembersFilter, opts ...*SetOptions) (int64, error) {
	opt := MergeSetOptions(opts)
	groupID := filter.GroupID
	membersIDs := filter.UserIDs
	roles := filter.Roles
	// 不管怎么样,都会追加普通角色
	roles = append(filter.Roles, 0)

	if groupID <= 0 {
		return 0, errors.Wrap(errors.New("'group_id' <= '0'"))
	}

	if len(membersIDs) == 0 {
		return 0, errors.Wrap(errors.New("'user_ids' length is '0'"))
	}

	sqlQuery := fmt.Sprintf("DELETE FROM %s WHERE `group_id` = ? AND `user_id` IN (?) AND `role` IN (?) AND `role` != 1", TableGroupMembers)
	sqlQuery, sqlArgs, err := sqlx.In(sqlQuery, groupID, membersIDs, roles)
	if err != nil {
		return 0, errors.Wrap(err)
	}

	rs, err := opt.SQLExt().Exec(sqlQuery, sqlArgs...)
	if err != nil {
		return 0, errors.Wrap(err)
	}

	if opt.UpdateCache() {
		// 直接删除整个Key最干脆
		var keys []string
		for i := 0; i < len(membersIDs); i++ {
			keys = append(keys, strconv.FormatInt(membersIDs[i], 10))
		}
		GlobCache.HDel(GlobCtx, cacheKeyFormatGroupMembers(groupID), keys...)
	}

	rowsAffected, err := rs.RowsAffected()
	return rowsAffected, errors.Wrap(err)
}

// GetGroupAllMembers 获取群成员资料
func GetGroupAllMembers(groupID int64, opts ...*GetOptions) ([]*GroupMember, error) {
	opt := MergeGetOptions(opts)
	gms := make([]*GroupMember, 0)
	if opt.UseCache() {
		cacheKey := cacheKeyFormatGroupMembers(groupID)
		exists, _ := GlobCache.Exists(GlobCtx, cacheKey)
		if exists > 0 {
			err := GlobCache.HValsAndScan(GlobCtx, gms, cacheKeyFormatGroupMembers(groupID))
			if err == nil {
				return gms, nil
			}
		}
	}

	sqlQuery := fmt.Sprintf("SELECT `id`,`group_id`,`user_id`,`role`,`speak_status`,`updated_at`,`created_at` FROM %s WHERE `group_id` = ? ORDER BY `id` ASC", TableGroupMembers)

	rows, err := opt.SQLExt().Query(sqlQuery, groupID)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	defer rows.Close()
	err = sqlx.StructScan(rows, &gms)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	if opt.UpdateCache() {
		cacheKey := cacheKeyFormatGroupMembers(groupID)
		m := make(map[string]*GroupMember)
		for i := 0; i < len(gms); i++ {
			m[strconv.FormatInt(gms[i].UserID, 10)] = gms[i]
		}

		GlobCache.HSet(GlobCtx, cacheKey, m)
		GlobCache.Expire(GlobCtx, cacheKey, jcache.RandomExpirationDuration())
	}

	return gms, nil
}

// GetGroupMembers 获取多个群成员资料
func GetGroupMembers(groupID int64, memberIDs []int64, opts ...*GetOptions) ([]*GroupMember, error) {
	opt := MergeGetOptions(opts)
	gms := make([]*GroupMember, 0)
	if opt.UseCache() {
		cacheKey := cacheKeyFormatGroupMembers(groupID)
		exists, _ := GlobCache.Exists(GlobCtx, cacheKey)
		if exists > 0 {
			var memberKeys = make([]string, len(memberIDs))
			for i := 0; i < len(memberIDs); i++ {
				memberKeys[i] = fmt.Sprintf("%d", memberIDs[i])
			}

			err := GlobCache.HMGetAndScan(GlobCtx, gms, cacheKeyFormatGroupMembers(groupID), memberKeys...)
			if err == nil && len(gms) == len(memberIDs) {
				return gms, nil
			}
		}
	}

	sqlQuery := fmt.Sprintf("SELECT `id`,`group_id`,`user_id`,`role`,`speak_status`,`updated_at`,`created_at` FROM %s WHERE `group_id` = ? AND `user_id` IN (?) ORDER BY `id` ASC", TableGroupMembers)
	sqlQuery, sqlArgs, err := sqlx.In(sqlQuery, groupID, memberIDs)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	rows, err := opt.SQLExt().Query(sqlQuery, sqlArgs...)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	defer rows.Close()
	err = sqlx.StructScan(rows, &gms)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	if opt.UpdateCache() {
		cacheKey := cacheKeyFormatGroupMembers(groupID)
		m := make(map[string]*GroupMember)
		for i := 0; i < len(gms); i++ {
			m[strconv.FormatInt(gms[i].UserID, 10)] = gms[i]
		}

		GlobCache.HSet(GlobCtx, cacheKey, m)
		GlobCache.Expire(GlobCtx, cacheKey, jcache.RandomExpirationDuration())
	}

	return gms, nil
}

// GetGroupMember 获取某个群成员资料
func GetGroupMember(groupID, memberID int64, opts ...*GetOptions) (*GroupMember, error) {
	opt := MergeGetOptions(opts)
	member := new(GroupMember)
	if opt.UseCache() {
		cacheKey := cacheKeyFormatGroupMembers(groupID)
		exists, _ := GlobCache.Exists(GlobCtx, cacheKey)
		if exists > 0 {
			err := GlobCache.HGetAndScan(GlobCtx, member, cacheKeyFormatGroupMembers(groupID), fmt.Sprintf("%d", memberID))
			if err == nil {
				return member, nil
			}
		}
	}

	sqlQuery := fmt.Sprintf("SELECT `id`,`group_id`,`user_id`,`role`,`speak_status`,`updated_at`,`created_at` FROM %s WHERE `group_id` = ? AND `user_id` = ?", TableGroupMembers)

	err := sqlx.Get(opt.SQLExt(), member, sqlQuery, groupID, memberID)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	if opt.UpdateCache() {
		cacheKey := cacheKeyFormatGroupMembers(groupID)

		if exists, _ := GlobCache.Exists(GlobCtx, cacheKey); exists > 0 {
			GlobCache.HSet(GlobCtx, cacheKey, fmt.Sprintf("%d", memberID), member)
		}
	}

	return member, nil
}

// UpdateGroupMemberData 更新群组成员数据
type UpdateGroupMemberData struct {
	// Role 成员角色;     0:普通成员,1:拥有者,2:管理员
	Role *int `db:"role" json:"role"`

	// SpeakStatus 发言状态;   1:可发言, 0:禁止发言
	SpeakStatus *int `db:"speak_status" json:"speak_status"`

	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	UpdaterID int64 `db:"updater_id" json:"updater_id"`
}

// UpdateGroupMember 更新群组成员信息
func UpdateGroupMember(groupID, userID int64, data *UpdateGroupMemberData, opts ...*SetOptions) error {
	opt := MergeSetOptions(opts)

	var sqlArgs []any
	var setSQLs []string

	if groupID <= 0 || userID <= 0 {
		return errors.Wrap(errors.ParamsInvalid)
	}

	if data.Role == nil && data.SpeakStatus == nil {
		return errors.Wrap(errors.ParamsInvalid)
	}

	if data.Role != nil {
		setSQLs = append(setSQLs, " `role` = ? ")
		sqlArgs = append(sqlArgs, data.Role)
	}

	if data.SpeakStatus != nil {
		setSQLs = append(setSQLs, " `speak_status` = ? ")
		sqlArgs = append(sqlArgs, data.SpeakStatus)
	}

	setSQLs = append(setSQLs, " `updater_id`=?, `updated_at`=? ")
	sqlArgs = append(sqlArgs, data.UpdaterID, data.UpdatedAt, groupID, userID)

	sqlQuery := fmt.Sprintf("UPDATE %s SET %s WHERE `group_id` = ? AND `user_id` = ? ", TableGroupMembers, strings.Join(setSQLs, ","))

	_, err := opt.SQLExt().Exec(sqlQuery, sqlArgs...)
	if err != nil {
		return errors.Wrap(err)
	}

	if opt.UpdateCache() {
		cacheKey := cacheKeyFormatGroupMembers(groupID)
		GlobCache.Del(GlobCtx, cacheKey)
	}
	return nil
}

const (
	GroupMaxMember = 10
)

// CreateGroupTx 事务方式创建群组
func CreateGroupTx(creatorID int64, memberIDs []int64) (group *Group, err error) {
	var tx *sqlx.Tx
	timeOutCtx, _ := context.WithTimeout(GlobCtx, time.Second*5)
	tx, err = GlobDB.MySQL.BeginTxx(timeOutCtx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, errors.Wrap(err)
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			if err != nil {
				log.Error().Err(err).Msg("事务提交失败")
			}
		}

		// 如果提交失败,可以做回滚
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			e := tx.Rollback()
			if e != nil {
				log.Error().Err(e).Msg("事务回滚失败")
			}
			return
		}
	}()

	userIDs := []int64{creatorID}
	err = utils.SliceUnique(memberIDs, &memberIDs)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	userIDs = append(userIDs, memberIDs...)

	setOpt := NewSetOptionsWithSQLExt(tx)
	getOpt := NewGetOptionsFromSetOptions(setOpt)

	// 查出所有成员名称信息
	users, err := GetUsers(userIDs, getOpt)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	var groupName = ""
	for i := 0; i < len(users); i++ {
		if i != 0 {
			groupName += ","
		}
		groupName += users[i].Nickname
	}

	groupName = utils.StringCut(groupName, 50)

	owner := userIDs[0]

	now := time.Now()
	group = &Group{
		Name:        groupName,
		MaxMember:   GroupMaxMember,
		OwnerID:     owner,
		SpeakStatus: 1,
		CreatorID:   owner,
		CreatedAt:   now,
		UpdaterID:   owner,
		UpdatedAt:   now,
	}

	id, err := AddGroup(group, setOpt)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	addData := &AddGroupMembersData{
		UserIDs:   userIDs,
		CreatedAt: now,
		CreatorID: owner,
	}
	_, err = AddGroupMembers(id, addData, setOpt)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// 更新创建人的权限为群主
	roleOwner := 1
	err = UpdateGroupMember(id, owner, &UpdateGroupMemberData{Role: &roleOwner}, setOpt)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	group.ID = id
	return group, nil
	//}

}

// ==============================================================
// ================== CACHE CONTROL =============================
// ==============================================================
// cacheKeyFormatGroupID 格式化群组的ID 缓存 key
func cacheKeyFormatGroupID(id int64) string {
	return fmt.Sprintf("%s:group:id:%d", CacheKeyPrefix, id)
}

// cacheKeyFormatGroupMembers 格式化群组成员 缓存 key
func cacheKeyFormatGroupMembers(id int64) string {
	return fmt.Sprintf("%s:group:members:id:%d", CacheKeyPrefix, id)
}
