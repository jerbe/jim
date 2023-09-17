package database

import (
	"context"
	"strings"
	"time"

	"github.com/jerbe/jim/config"
	"github.com/jerbe/jim/errors"

	"github.com/jerbe/jcache"
	"github.com/jerbe/jcache/driver"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/11 15:55
  @describe :
*/

type Database struct {
	MySQL *sqlx.DB
	Redis redis.UniversalClient
	Mongo *mongo.Client
}

var (
	// MYSQL 库跟表
	DatabaseMySQLIM = "`jim`"

	// TableGroups 群组数据表
	TableGroups = DatabaseMySQLIM + ".`groups`"

	// TableGroupMembers 群组成员表
	TableGroupMembers = DatabaseMySQLIM + ".`group_member`"

	// TableUsers 用户数据表
	TableUsers = DatabaseMySQLIM + ".`users`"

	// TableUserRelation 用户关系表
	TableUserRelation = DatabaseMySQLIM + ".`user_relation`"

	// TableUserRelationInvite 用户关系邀请表
	TableUserRelationInvite = DatabaseMySQLIM + ".`user_relation_invite`"

	// MongoDB 库跟集合
	DatabaseMongodbIM = "jim"
	CollectionRoom    = "room"
	CollectionMessage = "message"
)

func constDatabase() {
	// TableGroups 群组数据表
	TableGroups = DatabaseMySQLIM + ".`groups`"

	// TableGroupMembers 群组成员表
	TableGroupMembers = DatabaseMySQLIM + ".`group_member`"

	// TableUsers 用户数据表
	TableUsers = DatabaseMySQLIM + ".`users`"

	// TableUserRelation 用户关系表
	TableUserRelation = DatabaseMySQLIM + ".`user_relation`"

	// TableUserRelationInvite 用户关系邀请表
	TableUserRelationInvite = DatabaseMySQLIM + ".`user_relation_invite`"
}

var (
	CacheKeyPrefix = config.GlobConfig().Main.ServerName

	GlobCtx = context.Background()

	GlobDB *Database

	GlobCache *jcache.Client

	initialized bool
)

func Init(cfg config.Config) (db *Database, err error) {
	if initialized {
		return nil, errors.New("数据库已经配置过")
	}

	CacheKeyPrefix = cfg.Main.ServerName

	// 初始化缓存
	c := &driver.RedisConfig{
		Mode:       cfg.Redis.Mode,
		MasterName: cfg.Redis.MasterName,
		Addrs:      cfg.Redis.Addrs,
		Database:   cfg.Redis.Database,
		Username:   cfg.Redis.Username,
		Password:   cfg.Redis.Password,
	}
	cache := jcache.NewClient(
		driver.NewMemory(),
		driver.NewRedis(driver.NewRedisOptionsWithConfig(c)))

	GlobCache = cache

	// 初始化MYSQL
	mysqlCfg := cfg.MySQL
	mysqlConnStr := mysqlCfg.URI
	if mysqlConnStr == "" {
		return nil, errors.New("mysql.uri为空")
	}
	timeout := time.Second * 10
	timoutCtx, _ := context.WithTimeout(context.Background(), timeout)
	mysqlDB, err := sqlx.ConnectContext(timoutCtx, "mysql", mysqlConnStr)
	if err != nil {
		return nil, err
	}
	mysqlDB.SetMaxOpenConns(10) // 设置数据库连接池的最大连接数
	mysqlDB.SetMaxIdleConns(5)

	if mysqlCfg.MainDB != "" {
		DatabaseMySQLIM = mysqlCfg.MainDB
	}
	db = new(Database)
	db.MySQL = mysqlDB

	// 初始化redis
	db.Redis, err = initRedis(cfg.Redis)
	if err != nil {
		return nil, err
	}

	// 初始化mongodb
	mongodbCfg := cfg.MongoDB
	mongodbConnStr := mongodbCfg.URI
	if mongodbConnStr == "" {
		return nil, errors.New("mongodb.uri为空")
	}

	mongoOpts := options.Client().ApplyURI(mongodbConnStr).SetTimeout(timeout)
	mongodbCli, err := mongo.Connect(GlobCtx, mongoOpts)
	if err != nil {
		return nil, err
	}

	err = mongodbCli.Ping(GlobCtx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	if mongodbCfg.MainDB != "" {
		DatabaseMongodbIM = mongodbCfg.MainDB
	}
	db.Mongo = mongodbCli

	// 变量重新整理一遍
	constDatabase()

	GlobDB = db
	initialized = true
	return db, nil
}

func initRedis(cfg config.Redis) (redis.UniversalClient, error) {
	var redisCli redis.UniversalClient
	if len(cfg.Addrs) == 0 {
		return nil, errors.New("未设置Addrs")
	}

	var dialTO = time.Second * 5

	switch strings.ToLower(cfg.Mode) {
	case "sentinel": // 哨兵模式
		if cfg.MasterName == "" {
			return nil, errors.New("redis.mode 哨兵(sentinel)模式下,未指定master_name")
		}

		// 返回 *redis.FailoverClient
		redisCli = redis.NewUniversalClient(&redis.UniversalOptions{
			MasterName:  cfg.MasterName,
			Addrs:       cfg.Addrs,
			Username:    cfg.Username,
			Password:    cfg.Password,
			DialTimeout: dialTO,
		})
	case "cluster": //集群模式
		// 返回 *redis.ClusterClient
		redisCli = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:       cfg.Addrs,
			Username:    cfg.Username,
			Password:    cfg.Password,
			DialTimeout: dialTO,
		})
	default: // 单例模式
		if len(cfg.Addrs) > 1 {
			return nil, errors.New("redis.mode 单例(single)模式下,addrs只允许一个元素")
		}
		// 返回 *redis.Client
		redisCli = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:       cfg.Addrs[0:1],
			Username:    cfg.Username,
			Password:    cfg.Password,
			DialTimeout: dialTO,
		})
	}
	return redisCli, nil
}
