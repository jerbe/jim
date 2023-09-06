package main

import (
	"github.com/jerbe/jcache"
	"github.com/jerbe/jim/config"
	"github.com/jerbe/jim/database"
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
	// 初始化Http服务
	httpRouter := handler.Init()
	if err := httpRouter.Run(":8080"); err != nil {
		log.Fatal().Err(err).Msg("初始http服务失败")
	}

}

func init() {
	log.Info().Msg("服务初始化中...")
	defer log.Warn().Msg("服务已关闭")

	// 加载配置
	cfg, err := config.Init()
	if err != nil {
		log.Fatal().Err(err).Msg("加载配置文件失败")
	}

	err = pubsub.Init(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("初始化推收模块('pubsub')失败")
	}

	// 配置数据库
	if _, err = database.Init(cfg); err != nil {
		log.Fatal().Err(err).Msg("初始化数据模块('database')失败")
	}
	jcacheCfg := &jcache.Config{
		Redis: &jcache.RedisConfig{
			Mode:       cfg.Redis.Mode,
			MasterName: cfg.Redis.MasterName,
			Addrs:      cfg.Redis.Addrs,
			Database:   cfg.Redis.Database,
			Username:   cfg.Redis.Username,
			Password:   cfg.Redis.Password,
		},
	}
	err = jcache.Init(jcacheCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("初始化缓存模块('cache')失败")
	}

}
