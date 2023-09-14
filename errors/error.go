package errors

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/20 11:46
  @describe :
*/

import (
	"database/sql"

	"github.com/jerbe/go-errors"

	"github.com/jerbe/jcache"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// InternalServerError 服务器异常
	InternalServerError = errors.New("internal server error")

	// NoRecords 没有记录
	NoRecords = errors.New("no records found")

	// NotChange 未改变
	NotChange = errors.New("not change")

	// ParamsInvalid 参数无效
	ParamsInvalid = errors.New("params invalid")
)

var (
	New = errors.New

	NewWithCaller = errors.NewWithCaller

	Wrap = errors.Wrap

	Is = errors.Is

	IsIn = errors.IsIn

	As = errors.As

	Unwrap = errors.Unwrap

	noRowsErrs = []error{
		redis.Nil,
		mongo.ErrNoDocuments,
		sql.ErrNoRows,
		jcache.ErrEmpty,
		NoRecords,
	}
)

// IsNoRecord 传入的错误信息是一个没有找到记录的错误
func IsNoRecord(err error) bool {
	for i := 0; i < len(noRowsErrs); i++ {
		if errors.Is(err, noRowsErrs[i]) {
			return true
		}
	}
	return false
}

// IsEmptyRecord 掺入的错误信息是有找到,但是值为空
func IsEmptyRecord(err error) bool {
	return errors.Is(err, jcache.ErrEmpty)
}
