package log

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jerbe/jim/utils"

	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	serviceName = "[jim_web_server]"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/18 21:40
  @describe :
*/

var requestLogger zerolog.Logger
var infoLogger zerolog.Logger
var errorLogger zerolog.Logger
var consoleWriter zerolog.ConsoleWriter

func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return file + ":[" + runtime.FuncForPC(pc).Name() + "]:" + strconv.Itoa(line)
	}
	consoleWriter = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = os.Stdout
	})
}

// Init 初始化
func Init(svrName string) {
	if svrName == serviceName {
		return
	}
	Info().Str("old", serviceName).Str("new", svrName).Msg("改变改变服务名称")

	serviceName = svrName
	requestLogger = newZeroLogger(fmt.Sprintf("%s-request.log", serviceName))
	infoLogger = newZeroLogger(fmt.Sprintf("%s-info.log", serviceName))
	errorLogger = newZeroLogger(fmt.Sprintf("%s-error.log", serviceName))
}

func newZeroLogger(filename string) zerolog.Logger {
	mode := os.Getenv("LOG_MODE")
	write := newRollingFile(filename)
	if mode == strings.ToLower("dev") {
		write = zerolog.MultiLevelWriter(write, consoleWriter)
	}
	return zerolog.New(write).With().Timestamp().Str("service", serviceName).Logger()
}

func newRollingFile(filename string) io.Writer {
	pwd, _ := os.Getwd()
	l, err := lumberjack.NewRoller(path.Join(pwd, filename), 10000000*256, &lumberjack.Options{
		MaxAge:    time.Minute,
		LocalTime: true,
	})
	if err != nil {
		panic(err)
	}

	return l
}

func Request() *zerolog.Event {
	l := requestLogger.With().Str("hostname", utils.GetHostname()).Logger()
	return (&l).Info()
}

func Debug() *zerolog.Event {
	l := infoLogger.With().Str("hostname", utils.GetHostname()).Caller().Logger()
	return (&l).Debug()
}

func Info() *zerolog.Event {
	l := infoLogger.With().Str("hostname", utils.GetHostname()).Caller().Logger()
	return (&l).Info()
}

func Warn() *zerolog.Event {
	l := errorLogger.With().Str("hostname", utils.GetHostname()).Caller().Logger()
	return (&l).Warn()
}

func Error() *zerolog.Event {
	l := errorLogger.With().Str("hostname", utils.GetHostname()).Caller().Stack().Logger()
	return (&l).Error()
}

func Fatal() *zerolog.Event {
	l := errorLogger.With().Str("hostname", utils.GetHostname()).Caller().Logger()
	return (&l).Fatal()
}

func Panic() *zerolog.Event {
	l := errorLogger.With().Str("hostname", utils.GetHostname()).Caller().Logger()
	return (&l).Panic()
}

func parseGinContextToLog(evt *zerolog.Event, ctx *gin.Context) *zerolog.Event {
	clientIP, _ := utils.GetClientIP(ctx.Request)
	evt.Str("request_id", ctx.GetString("REQUEST_ID")).
		Str("user_id", ctx.GetString("LOGIN_USER_ID")).
		Str("method", ctx.Request.Method).
		Str("client_ip", clientIP).
		Str("host", ctx.Request.Host).
		Str("uri", ctx.Request.URL.Path)

	if ctx.Request.URL.RawQuery != "" {
		evt.Str("query", ctx.Request.URL.RawQuery)
	}
	return evt
}

func InfoFromGinContext(ctx *gin.Context) *zerolog.Event {
	return parseGinContextToLog(Info(), ctx)
}

func WarnFromGinContext(ctx *gin.Context) *zerolog.Event {
	return parseGinContextToLog(Warn(), ctx)
}

func ErrorFromGinContext(ctx *gin.Context) *zerolog.Event {
	return parseGinContextToLog(Error(), ctx)
}

func FatalFromGinContext(ctx *gin.Context) *zerolog.Event {
	return parseGinContextToLog(Fatal(), ctx)
}

func PanicFromGinContext(ctx *gin.Context) *zerolog.Event {
	return parseGinContextToLog(Panic(), ctx)
}
