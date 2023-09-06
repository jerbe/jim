package errors

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/20 11:46
  @describe :
*/

import (
	"database/sql"
	"fmt"
	"github.com/jerbe/jcache"
	"io"
	"runtime"
	"strconv"
	"sync/atomic"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

var errID = uint64(0)

// newErrorID 获取错误ID
func newErrorID() uint64 {
	return atomic.AddUint64(&errID, 1)
}

// Error 栈错误
type Error struct {
	cause   error
	message string
	caller  string
	id      uint64
}

func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return t.id == e.id && t.message == e.message
	}
	return false
}

func (e *Error) Error() string {
	if e.cause != nil {
		msg := e.message
		if msg == "" {
			return e.cause.Error()
		}
		msg += " <= " + e.cause.Error()
		return msg
	}
	return e.message
}

func (e *Error) Cause() error {
	return e.cause
}

func (e *Error) Wrap(target error) error {
	ne := *e
	ne.cause = target
	ne.caller = caller()
	return &ne
}

func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if e.cause != nil {
				fmt.Fprintf(s, "%+v\n", e.cause)
			}
			msg := ""
			if e.message != "" {
				if e.caller != "" {
					msg = " => "
				}
				msg += e.message
			}
			fmt.Fprintf(s, "%s%s", e.caller, msg)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, e.Error())

	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

// New 生成错误
func New(message string) *Error {
	return &Error{message: message, id: newErrorID()}
}

func Wrap(err error) error {
	if err == nil {
		return nil
	}
	return &Error{
		cause:  err,
		caller: caller(),
		id:     newErrorID(),
	}
}

var (
	// InternalServerError 服务器异常
	InternalServerError = New("internal server error")

	// NoRecords 没有记录
	NoRecords = New("no records found")

	// NotChange 未改变
	NotChange = New("not change")

	// ParamsInvalid 参数无效
	ParamsInvalid = New("params invalid")
)

var noRowsErrs = []error{
	redis.Nil,
	mongo.ErrNoDocuments,
	sql.ErrNoRows,
	jcache.ErrEmpty,
	NoRecords,
}

// IsNoRecord 传入的错误信息是一个没有找到记录的错误
func IsNoRecord(err error) bool {
	for i := 0; i < len(noRowsErrs); i++ {
		if Is(err, noRowsErrs[i]) {
			return true
		}
	}
	return false
}

// IsEmptyRecord 掺入的错误信息是有找到,但是值为空
func IsEmptyRecord(err error) bool {
	return Is(err, jcache.ErrEmpty)
}

func caller() string {
	pc, file, line, _ := runtime.Caller(2)
	return "[" + runtime.FuncForPC(pc).Name() + "]:" + file + ":" + strconv.Itoa(line)
}
