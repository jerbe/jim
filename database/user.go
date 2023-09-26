package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jerbe/jim/errors"
	"github.com/jerbe/jim/log"
	"github.com/jerbe/jim/utils"

	"github.com/jerbe/jcache/v2"

	"github.com/jmoiron/sqlx"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/11 22:58
  @describe :
*/

type User struct {
	ID           int64      `db:"id" json:"id"`
	Username     string     `db:"username" json:"username"`
	Password     string     `db:"password_hash" json:"password_hash"`
	Nickname     string     `db:"nickname" json:"nickname"`
	Avatar       string     `db:"avatar" json:"avatar"`
	BirthDate    *time.Time `db:"birth_date" json:"birth_date"`
	OnlineStatus int        `db:"online_status" json:"online_status"`
	Status       int        `db:"status" json:"status"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

func (u *User) MarshalBinary() (data []byte, err error) {
	return json.Marshal(u)
}

func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u *User) String() string {
	return fmt.Sprintf("%d, %s, %s, %d, %d, %s, %s, %s, %s",
		u.ID, u.Username, u.Nickname, u.OnlineStatus, u.Status, u.Avatar, u.BirthDate, u.CreatedAt, u.UpdatedAt)
}

// GetUser 根据ID获取用户信息
func GetUser(id int64, opts ...*GetOptions) (*User, error) {
	opt := MergeGetOptions(opts)
	// 判断是否使用了缓存
	if opt.UseCache() {
		cacheKey := cacheKeyFormatUserID(id)
		var user = new(User)
		value := GlobCache.Get(GlobCtx, cacheKey)
		if value.Err() == nil && value.Val() != "" {
			err := value.Scan(user)
			return user, err
		}
		if value.Err() == nil && value.Val() == "" {
			return nil, errors.NoRecords
		}
	}

	sqlStr := fmt.Sprintf("SELECT `id`,`username`,`password_hash`, `nickname`, `avatar`, `birth_date`, `online_status`, `status`, `created_at`,`updated_at` FROM %s WHERE `id` = ?", TableUsers)

	user := &User{}
	err := sqlx.Get(opt.SQLExt(), user, sqlStr, id)
	if err != nil {
		if err == sql.ErrNoRows {
			// 写入缓存,如果key不存在的话
			var cacheKey = cacheKeyFormatUserID(id)
			if err1 := GlobCache.SetNX(GlobCtx, cacheKey, nil, jcache.DefaultEmptySetNXDuration).Err(); err1 != nil {
				log.Error().Err(err1).Str("cache_key", cacheKey).Msg("缓存写入失败")
			}

			return nil, errors.Wrap(err)
		}
		return nil, errors.Wrap(err)
	}

	if opt.UpdateCache() {
		// 将数据写入到缓存
		GlobCache.Set(GlobCtx, cacheKeyFormatUserID(user.ID), user, jcache.DefaultExpirationDuration)
		GlobCache.Set(GlobCtx, cacheKeyFormatUsername(user.Username), user, jcache.DefaultExpirationDuration)
	}

	return user, nil
}

// GetUserByUsername  根据用户名获取用户信息
func GetUserByUsername(username string, opts ...*GetOptions) (*User, error) {
	opt := MergeGetOptions(opts)
	if opt.UseCache() {
		cacheKey := cacheKeyFormatUsername(username)
		exits := GlobCache.Exists(GlobCtx, cacheKey).Val()
		if exits > 0 {
			user := new(User)
			value := GlobCache.Get(GlobCtx, cacheKey)
			if value.Err() == nil && value.Val() != "" {
				err := value.Scan(user)
				return user, err
			}
			if value.Err() == nil && value.Val() == "" {
				return nil, errors.NoRecords
			}
		}
	}

	sqlQuery := fmt.Sprintf("SELECT `id`,`username`,`password_hash`,`nickname`,`avatar`,`birth_date`,`online_status`, `status`,`created_at`,`updated_at` FROM %s WHERE `username` = ?", TableUsers)

	user := &User{}
	err := sqlx.Get(opt.SQLExt(), user, sqlQuery, username)
	if err != nil {
		if err == sql.ErrNoRows {
			// 写入缓存,如果key不存在的话
			cacheKey := cacheKeyFormatUsername(username)
			if err1 := GlobCache.SetNX(GlobCtx, cacheKey, nil, jcache.DefaultEmptySetNXDuration).Err(); err1 != nil {
				log.Error().Err(err1).Str("cache_key", cacheKey).Msg("缓存写入失败")
			}
			return nil, errors.Wrap(err)
		}
		return nil, errors.Wrap(err)
	}

	// 将数据写入到缓存
	if opt.UpdateCache() {
		// 将数据写入到缓存
		GlobCache.Set(GlobCtx, cacheKeyFormatUserID(user.ID), user, jcache.DefaultExpirationDuration)
		GlobCache.Set(GlobCtx, cacheKeyFormatUsername(user.Username), user, jcache.DefaultExpirationDuration)
	}

	return user, nil
}

func GetUsers(ids []int64, opts ...*GetOptions) ([]*User, error) {
	opt := MergeGetOptions(opts)
	// 去重后的用户ID

	var uqIds []int64
	err := utils.SliceUnique(ids, &uqIds)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// 没匹配到缓存的用户IDs
	var users []*User

	// 判断是否使用了缓存
	if opt.UseCache() {
		var cacheKeys = make([]string, len(uqIds))

		for i := 0; i < len(uqIds); i++ {
			cacheKeys[i] = cacheKeyFormatUserID(uqIds[i])
		}

		var cacheUsers []*User
		err = GlobCache.MGetAndScan(GlobCtx, &cacheUsers, cacheKeys...)
		if err == nil && len(cacheUsers) == len(uqIds) {
			return cacheUsers, nil
		}

		var notMatchIDs []int64
		for j := 0; j < len(uqIds); j++ {
			var has bool
			for i := 0; i < len(cacheUsers); i++ {
				if uqIds[j] == cacheUsers[i].ID {
					has = true
					break
				}
			}
			if !has {
				notMatchIDs = append(notMatchIDs, uqIds[j])
			}
		}
		ids = notMatchIDs

		users = append(users, cacheUsers...)
	}

	sqlQuery := fmt.Sprintf("SELECT `id`,`username`,`password_hash`, `nickname`, `avatar`, `birth_date`, `online_status`, `status`, `updated_at`,`created_at` FROM %s WHERE `id` IN (?)", TableUsers)
	sqlQuery, sqlArgs, err := sqlx.In(sqlQuery, ids)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	var dbUsers []*User
	rs, err := opt.SQLExt().Query(sqlQuery, sqlArgs...)
	if err != nil && !errors.IsNoRecord(err) {
		return nil, errors.Wrap(err)
	}
	err = sqlx.StructScan(rs, &dbUsers)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	if opt.UpdateCache() && len(dbUsers) > 0 {
		for i := 0; i < len(dbUsers); i++ {
			// 将数据写入到缓存
			u := dbUsers[i]
			GlobCache.Set(GlobCtx, cacheKeyFormatUserID(u.ID), u, jcache.RandomExpirationDuration())
			GlobCache.Set(GlobCtx, cacheKeyFormatUsername(u.Username), u, jcache.RandomExpirationDuration())
		}
	}
	users = append(users, dbUsers...)

	if len(users) == 0 {
		return nil, errors.NoRecords
	}

	return users, nil
}

// SearchUsersFilter 获取用户过滤条件
type SearchUsersFilter struct {
	// UserID 用户ID
	UserID *int64 `json:"user_id,omitempty" db:"user_id"`

	// Nickname 昵称
	Nickname *string `json:"nickname,omitempty" db:"nickname"`

	// StartID 开始ID
	StartID int64 `json:"start_id" db:"start_id"`

	// Limit 限制集合大小
	Limit int64 `json:"limit" db:"limit"`
}

// SearchUsers 获取用户列表
func SearchUsers(filter *SearchUsersFilter, opts ...*GetOptions) ([]*User, error) {
	if filter == nil {
		return nil, errors.ParamsInvalid
	}

	if filter.UserID == nil && filter.Nickname == nil {
		return nil, errors.ParamsInvalid
	}

	whereSQL := make([]string, 0)
	if filter.UserID != nil {
		whereSQL = append(whereSQL, " `id` = :user_id ")
	} else {
		whereSQL = append(whereSQL, " `id` > :start_id ")
	}

	if filter.Nickname != nil {
		*filter.Nickname = strings.ReplaceAll(*filter.Nickname, "%", "\\%")
		whereSQL = append(whereSQL, " `nickname` LIKE CONCAT(:nickname,'%')")
	}

	opt := MergeGetOptions(opts)
	sqlQuery := fmt.Sprintf("SELECT `id`, `nickname`, `avatar`,`online_status` FROM %s WHERE %s ORDER BY `id` ASC LIMIT 0,:limit ", TableUsers, strings.Join(whereSQL, " AND "))
	var users []*User
	rows, err := sqlx.NamedQuery(opt.SQLExt(), sqlQuery, filter)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer rows.Close()

	for rows.Next() {
		user := new(User)
		err = rows.StructScan(user)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		users = append(users, user)
	}

	return users, nil
}

// UserExist 根据用户ID检测用户是否存在
func UserExist(id int64) (bool, error) {
	_, err := GetUser(id)
	if err == nil {
		return true, nil
	}
	if errors.IsNoRecord(err) {
		return false, nil
	}
	return false, errors.Wrap(err)
}

// UserExistByUsername 根据用户名判断用户是否存在
func UserExistByUsername(username string) (bool, error) {
	_, err := GetUserByUsername(username)
	if err == nil {
		return true, nil
	}
	if errors.IsNoRecord(err) {
		return false, nil
	}
	return false, err
}

// AddUser 添加用户
func AddUser(user *User, opts ...*SetOptions) error {
	opt := MergeSetOptions(opts)

	sqlStr := fmt.Sprintf("INSERT INTO %s "+
		"(`username`,`password_hash`,`nickname`,`birth_date`,`online_status`, `status`,`created_at`,`updated_at`) "+
		"VALUES "+
		"(:username, :password_hash, :nickname, :birth_date, :online_status, :status, :created_at, :updated_at)",
		TableUsers)

	var now = time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	rs, err := sqlx.NamedExec(opt.SQLExt(), sqlStr, user)
	if err != nil {
		return errors.Wrap(err)
	}
	if id, err := rs.LastInsertId(); err == nil {
		user.ID = id
	} else {
		return errors.Wrap(err)
	}

	if opt.UpdateCache() {
		// 将数据写入到缓存
		GlobCache.Set(GlobCtx, cacheKeyFormatUserID(user.ID), user, jcache.DefaultExpirationDuration)
		GlobCache.Set(GlobCtx, cacheKeyFormatUsername(user.Username), user, jcache.DefaultExpirationDuration)
	}

	return nil
}

// ==============================================================
// ================== CACHE CONTROL =============================
// ==============================================================
// cacheKeyFormatUserID 格式化用户的ID 缓存 key
func cacheKeyFormatUserID(id int64) string {
	return fmt.Sprintf("%s:user:id:%d", CacheKeyPrefix, id)
}

// cacheKeyFormatUsername 格式化用户的用户名 缓存 key
func cacheKeyFormatUsername(username string) string {
	return fmt.Sprintf("%s:user:username:%s", CacheKeyPrefix, username)
}
