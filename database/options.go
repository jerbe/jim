package database

import "github.com/jmoiron/sqlx"

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/18 16:48
  @describe :
*/

// ============================================================
// ==================== SQLOptions ============================
// ============================================================

type SQLOptions struct {
	// ext 数据库执行器
	ext sqlx.Ext
}

// SetSQLExt 设置sql执行器
func (opt *SQLOptions) SetSQLExt(ext sqlx.Ext) *SQLOptions {
	if opt.ext != nil {
		return opt
	}

	opt.ext = ext
	return opt
}

// SQLExt 获取SQL执行器
func (opt *SQLOptions) SQLExt() sqlx.Ext {
	if opt.ext != nil {
		return opt.ext
	}
	return GlobDB.MySQL
}

// IsSQLTx 是否是一个事务执行器
func (opt *SQLOptions) IsSQLTx() bool {
	_, ok := opt.ext.(*sqlx.Tx)
	return ok
}

// ============================================================
// ==================== GetOptions ============================
// ============================================================

// GetOptions 获取数据的配置
type GetOptions struct {
	SQLOptions

	// useCache 获取数据时优先使用缓存
	useCache bool

	// updateCache 获取数据时同时更新缓存
	updateCache bool
}

// SetUseCache 设置是否优先使用缓存
func (opt *GetOptions) SetUseCache(val bool) *GetOptions {
	opt.useCache = val
	return opt
}

// UseCache 获取是否优先使用缓存
func (opt *GetOptions) UseCache() bool {
	// 如果是一个事务执行过程,不允许获取缓存数据
	if opt.IsSQLTx() {
		return false
	}
	return opt.useCache
}

// SetUpdateCache 设置获取数据时同时更新缓存
func (opt *GetOptions) SetUpdateCache(val bool) *GetOptions {
	opt.useCache = val
	return opt
}

// UpdateCache 获取数据时同时更新缓存
func (opt *GetOptions) UpdateCache() bool {
	// 如果是一个事务执行过程,不允许获取缓存数据
	if opt.IsSQLTx() {
		return false
	}
	return opt.updateCache
}

// SetSQLExt 设置是否使用其他sql执行器
func (opt *GetOptions) SetSQLExt(ext sqlx.Ext) *GetOptions {
	opt.SQLOptions.SetSQLExt(ext)

	// 如果是一个事务执行过程,不允许获取缓存数据
	opt.useCache = !opt.IsSQLTx()
	return opt
}

// NewGetOptions 新生成一个获取信息的配置
func NewGetOptions() *GetOptions {
	opt := &GetOptions{
		useCache:    true,
		updateCache: true,
	}
	return opt
}

// NewGetOptionsFromSetOptions 从新生成一个获取信息的配置
func NewGetOptionsFromSetOptions(setOpt *SetOptions) *GetOptions {
	opt := &GetOptions{
		SQLOptions:  setOpt.SQLOptions,
		useCache:    setOpt.updateCache,
		updateCache: setOpt.updateCache,
	}
	return opt
}

// MergeGetOptions 合并获取配置
func MergeGetOptions(opts []*GetOptions) *GetOptions {
	if len(opts) > 0 {
		return opts[0]
	} else {
		return NewGetOptions()
	}
}

// ============================================================
// ==================== SetOptions ============================
// ============================================================

// SetOptions 获取数据的配置
type SetOptions struct {
	SQLOptions

	// updateCache 获取数据时优先使用缓存
	updateCache bool
}

// SetUpdateCache 设置是否更新缓存
func (opt *SetOptions) SetUpdateCache(val bool) *SetOptions {
	// 如果是一个事务执行过程,不允许获取缓存数据
	if opt.IsSQLTx() {
		opt.updateCache = false
		return opt
	}

	opt.updateCache = val
	return opt
}

// UpdateCache 设置是否更新缓存
func (opt *SetOptions) UpdateCache() bool {
	// 如果是一个事务执行过程,不允许获取缓存数据
	if opt.IsSQLTx() {
		return false
	}
	return opt.updateCache
}

// SetSQLExt 设置是否使用其他sql执行器
func (opt *SetOptions) SetSQLExt(ext sqlx.Ext) *SetOptions {
	opt.SQLOptions.SetSQLExt(ext)

	// 如果是一个事务执行过程,不允许获取缓存数据
	opt.updateCache = !opt.IsSQLTx()
	return opt
}

// NewSetOptions 新生成一个设置获取信息的配置
func NewSetOptions() *SetOptions {
	opt := &SetOptions{
		updateCache: true,
	}
	return opt
}

// NewSetOptionsWithSQLExt 新生成一个设置获取信息的配置
func NewSetOptionsWithSQLExt(ext sqlx.Ext) *SetOptions {
	opt := &SetOptions{
		updateCache: true,
	}
	opt.SetSQLExt(ext)
	return opt
}

// NewSetOptionsFromGetOptions 从新生成一个获取信息的配置
func NewSetOptionsFromGetOptions(getOpt *GetOptions) *SetOptions {
	opt := &SetOptions{
		SQLOptions:  getOpt.SQLOptions,
		updateCache: getOpt.updateCache,
	}
	return opt
}

// MergeSetOptions 合并配置
func MergeSetOptions(opts []*SetOptions) *SetOptions {
	if len(opts) > 0 {
		return opts[0]
	} else {
		return NewSetOptions()
	}
}
