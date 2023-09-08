package handler

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/jerbe/jim/config"
	"github.com/jerbe/jim/database"
	"github.com/jerbe/jim/errors"
	"github.com/jerbe/jim/log"
	"github.com/jerbe/jim/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/15 11:29
  @describe :
*/

// UserClaims 用户认证使用的一些解码资料
type UserClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// CORSMiddleware
func CORSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//r := ctx.Request
		w := ctx.Writer

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, HEAD, PATCH, OPTIONS, GET, PUT")
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}
		ctx.Next()

	}
}

// RecoverMiddleware 回复中间件
func RecoverMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						seStr := strings.ToLower(se.Error())
						if strings.Contains(seStr, "broken pipe") ||
							strings.Contains(seStr, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				l := log.ErrorFromGinContext(ctx).Int("status_code", ctx.Writer.Status())

				stack := stack(3)
				httpRequest, _ := httputil.DumpRequest(ctx.Request, false)
				headers := strings.Split(string(httpRequest), "\r\n")
				for idx, header := range headers {
					current := strings.Split(header, ":")
					if current[0] == "Authorization" {
						headers[idx] = current[0] + ": *"
					}
				}
				headersToStr := strings.Join(headers, "\r\n")

				l.Str("headers", headersToStr).
					Bytes("stack", stack).
					Send()

				if brokenPipe {
					// If the connection is dead, we can't write a status to it.
					ctx.Error(err.(error)) //nolint: errcheck
					ctx.Abort()
				} else {
					ctx.AbortWithStatus(http.StatusInternalServerError)
				}
			}
		}()
		ctx.Next()
	}
}

// WebsocketMiddleware websocket用的中间件
// 因为websocket连接会导致协程阻塞,无法进行到下一步,导致日志也一并堵塞
// 需要在ctx.Next() 之前调用 日志的send或者msg
func WebsocketMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientIP, _ := utils.GetClientIP(ctx.Request)
		requestID := getAndStoreRequestID(ctx)
		l := log.Request().
			Str("request_id", requestID).
			Str("method", ctx.Request.Method).
			Str("client_ip", clientIP).
			Str("host", ctx.Request.Host).
			Str("uri", ctx.Request.URL.Path)

		if ctx.Request.URL.RawQuery != "" {
			l.Str("query", ctx.Request.URL.RawQuery)
		}

		l.Int("status_code", ctx.Writer.Status())
		ctx.Set(REQUEST_LOGGER_CONTEXT_KEY, l)
		ok := checkAuth(ctx)

		l.Send()
		if !ok {
			return
		}
		ctx.Next()
	}
}

// RequestLogMiddleware 请求日志中间件
func RequestLogMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientIP, _ := utils.GetClientIP(ctx.Request)
		requestID := getAndStoreRequestID(ctx)
		l := log.Request().
			Str("request_id", requestID).
			Str("method", ctx.Request.Method).
			Str("client_ip", clientIP).
			Str("host", ctx.Request.Host).
			Str("uri", ctx.Request.URL.Path)
		if ctx.Request.URL.RawQuery != "" {
			q := ctx.Request.URL.Query()
			if q.Has("token") {
				q.Set("token", "****")
			}
			query := strings.ReplaceAll(q.Encode(), "%2A", "*")
			l.Str("query", query)
		}
		l.Int("status_code", ctx.Writer.Status())
		startTime := time.Now()
		ctx.Set(REQUEST_LOGGER_CONTEXT_KEY, l)
		ctx.Next()
		l.Str("delay", time.Now().Sub(startTime).String())
		l.Send()

	}
}

// CheckAuthMiddleware 检验认证中间件
// 一般http请求使用的认证中间件,websocket使用 WebsocketMiddleware 方法进行独立鉴权及日志收集
func CheckAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !checkAuth(ctx) {
			return
		}
		ctx.Next()

	}
}

const (
	LOGIN_USER_CONTEXT_KEY     = "LOGIN_USER"
	LOGIN_USER_ID_CONTEXT_KEY  = "LOGIN_USER_ID"
	REQUEST_LOGGER_CONTEXT_KEY = "REQUEST_LOGGER"
	REQUEST_ID_CONTEXT_KEY     = "REQUEST_ID"
)

// getAndStoreRequestID 获取请求ID如果没有的情况下设置新的请求ID
func getAndStoreRequestID(ctx *gin.Context) string {
	if requestID := ctx.GetString(REQUEST_ID_CONTEXT_KEY); requestID != "" {
		return requestID
	}
	requestID := utils.UUID()
	ctx.Set(REQUEST_ID_CONTEXT_KEY, requestID)
	return requestID
}

// checkAuth 检验认证公共方法
func checkAuth(ctx *gin.Context) bool {
	l, _ := ctx.Get(REQUEST_LOGGER_CONTEXT_KEY)
	logEvent, _ := l.(*zerolog.Event)

	authToken := ctx.GetHeader("Authorization")
	if authToken == "" {
		authToken = ctx.Query("token")
	}

	if authToken == "" {
		ctx.Abort()
		JSONError(ctx, StatusError, "token不能为空")
		return false
	}

	claims := &UserClaims{}
	token, err := jwt.ParseWithClaims(authToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("不正确的签名方法")
		}
		return []byte(config.GlobConfig().Main.JwtSigningKey), nil
	})

	if err != nil {
		ctx.Abort()
		JSONError(ctx, StatusError, "token不正确")
		return false
	}

	// 验证账户状态
	if !token.Valid {
		ctx.Abort()
		JSONError(ctx, StatusError, "token验证不通过")
		return false
	}

	if logEvent != nil {
		logEvent.Int64("user_id", claims.UserID)
	}

	//@ TODO 有必要通过数据库再次查询用户是否存在?
	user, err := database.GetUser(claims.UserID)
	if err != nil {
		ctx.Abort()
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, MessageAccountNotExists)
			return false
		}
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).
			Int64("user_id", claims.UserID).
			Msg("获取用户信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return false
	}

	// 账户被禁用
	if user.Status == 0 {
		ctx.Abort()
		JSONError(ctx, StatusError, MessageAccountDisable)
		return false
	}

	// 账户被删除
	if user.Status == 2 {
		ctx.Abort()
		JSONError(ctx, StatusError, MessageAccountDeleted)
		return false
	}

	// 设置用户信息到上下文中去
	ctx.Set(LOGIN_USER_ID_CONTEXT_KEY, user.ID)
	ctx.Set(LOGIN_USER_CONTEXT_KEY, user)
	return true
}

// LoginUserFromContext 从上下文中获取当前登录的的用户信息
// 如果通过 checkAuth 方法鉴权过的,下文的 *gin.Context 必能找到用户信息
func LoginUserFromContext(ctx *gin.Context) *database.User {
	data, ok := ctx.Get(LOGIN_USER_CONTEXT_KEY)
	if !ok {
		panic(errors.New(fmt.Sprintf("%s key in context is nil ", LOGIN_USER_CONTEXT_KEY)))
	}

	user, ok := data.(*database.User)
	if !ok {
		panic(errors.New(fmt.Sprintf("%s data was not *database.Username", LOGIN_USER_CONTEXT_KEY)))
	}
	return user
}

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// stack returns a nicely formatted stack frame, skipping skip frames.
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contain dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.ReplaceAll(name, centerDot, dot)
	return name
}
