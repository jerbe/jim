package handler

import (
	"github.com/jerbe/jim/config"
	"github.com/jerbe/jim/database"
	"github.com/jerbe/jim/pubsub"
	"log"
	"testing"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/16 19:13
  @describe :
*/

func TestMain(t *testing.M) {
	log.Println("服务初始化中...")
	defer log.Println("服务已关闭...")

	// 加载配置
	log.Println("加载配置中...")
	cfg, err := config.Init()
	if err != nil {
		log.Println("加载配置文件失败")
	}

	log.Println("配置推收模块('pubsub')...")
	err = pubsub.Init(cfg)
	if err != nil {
		log.Fatalln("初始化推收模块('pubsub')失败")
	}

	// 配置数据库
	if _, err = database.Init(cfg); err != nil {
		log.Fatalln("初始化数据模块('database')失败")
	}

	if err != nil {
		log.Fatalln("初始化缓存模块('cache')失败")
	}
	t.Run()
}
