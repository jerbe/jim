package handler

import (
	"github.com/jerbe/jcache"
	"github.com/jerbe/jim/config"
	"github.com/jerbe/jim/database"
	"github.com/jerbe/jim/log"
	"github.com/jerbe/jim/pubsub"
	"testing"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/16 19:13
  @describe :
*/

func TestMain(t *testing.M) {
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

	jcacheCfg := jcache.Config{
		Redis: &jcache.RedisConfig{
			Mode:       cfg.Redis.Mode,
			MasterName: cfg.Redis.MasterName,
			Addrs:      cfg.Redis.Addrs,
			Database:   cfg.Redis.Database,
			Username:   cfg.Redis.Username,
			Password:   cfg.Redis.Password,
		},
	}
	err = jcache.Init(&jcacheCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("初始化缓存模块('cache')失败")
	}
	t.Run()
}
