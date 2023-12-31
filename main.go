package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/jerbe/jim/config"
	"github.com/jerbe/jim/database"
	"github.com/jerbe/jim/errors"
	"github.com/jerbe/jim/handler"
	"github.com/jerbe/jim/log"
	"github.com/jerbe/jim/pubsub"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023-08-10 00:54
  @describe :
*/

func main() {
	defer log.Warn().Msg("服务已关闭")

	// 初始化验证码服务
	handler.InitCaptcha()

	// 初始化订阅服务
	handler.InitSubscribe()

	// 初始化Http路由器
	mainHttpRouter := handler.InitRouter()
	mainHttpListenPort := fmt.Sprintf(":%d", config.GlobConfig().Http.MainListenPort)

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	mainHttpSvr := &http.Server{Addr: mainHttpListenPort, Handler: mainHttpRouter.Handler()}
	go func() {
		log.Info().Str("listen", mainHttpListenPort).Msg("main http服务运行中...")
		if err := mainHttpSvr.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error().Err(err).Str("listen", mainHttpListenPort).Msg("main http服务启动异常")
				sigCh <- syscall.SIGQUIT
			} else {
				log.Warn().Err(err).Str("listen", mainHttpListenPort).Msg("main http服务已关闭")
			}
		}
	}()

	// 开启对阻塞操作的跟踪
	runtime.SetBlockProfileRate(1)

	// 开启对锁调用的跟踪
	runtime.SetMutexProfileFraction(1)

	pprofHttpListenPort := fmt.Sprintf(":%d", config.GlobConfig().Http.PprofListenPort)
	pprofHttpSvr := &http.Server{Addr: pprofHttpListenPort, Handler: http.DefaultServeMux}
	go func() {
		log.Info().Str("listen", pprofHttpListenPort).Msg("pprof http服务运行中...")
		if err := pprofHttpSvr.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error().Err(err).Str("listen", pprofHttpListenPort).Msg("pprof http服务启动异常")
				sigCh <- syscall.SIGQUIT
			} else {
				log.Warn().Err(err).Str("listen", pprofHttpListenPort).Msg("pprof http服务已关闭")
			}
		}
	}()

	sig := <-sigCh
	log.Warn().Str("cause", sig.String()).Msg("系统即将退出")

	err := pprofHttpSvr.Shutdown(context.Background())
	if err != nil {
		log.Error().Err(err).Str("listen", pprofHttpListenPort).Msg("pprof http服务关闭异常")
	}

	err = mainHttpSvr.Shutdown(context.Background())
	if err != nil {
		log.Error().Err(err).Str("listen", mainHttpListenPort).Msg("main http服务关闭异常")
	}
}

func init() {
	log.Info().Msg("服务初始化中...")

	// 加载配置
	cfg, err := config.Init()
	if err != nil {
		log.Fatal().Err(err).Msg("配置文件初始化失败")
	}

	// 配置日志
	log.Init(cfg.Main.ServerName)

	// 配置推送模块
	err = pubsub.Init(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("推收模块('pubsub')初始化失败")
	}
	log.Info().Msg("推收模块('pubsub')初始完成")

	// 配置数据库
	if _, err = database.Init(cfg); err != nil {
		log.Fatal().Err(err).Msg("数据模块('database')初始化失败")
	}
}
