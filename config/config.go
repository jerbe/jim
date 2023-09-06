package config

import (
	"os"

	"github.com/jerbe/jim/log"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Main    Main    `yaml:"main"`
	Http    Http    `yaml:"http"`
	Redis   Redis   `yaml:"redis"`
	MySQL   MySQL   `yaml:"mysql"`
	MongoDB MongoDB `yaml:"mongodb"`
}

type Main struct {
	JwtSigningKey string `yaml:"jwt_signing_key"`
}

type Redis struct {
	// Mode 模式
	// 支持:single,sentinel,cluster
	Mode       string   `yaml:"mode"`
	MasterName string   `yaml:"master_name"`
	Addrs      []string `yaml:"addrs"`
	Database   string   `yaml:"database"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
}

type Http struct {
	Port string `yaml:"port"`
}

type MySQL struct {
	URI string `yaml:"uri"`
}

type MongoDB struct {
	URI string `yaml:"uri"`
}

var _cfg Config

func Init() (cfg Config, err error) {
	// 加载配置文件
	pwd, _ := os.Getwd()

	filename := pwd + "/config/config.yml"
	_, err = os.Stat(filename)
	if os.IsNotExist(err) {
		log.Fatal().Err(err).Str("filename", filename).Msg("文件不存在")
		return
	} else if err != nil {
		log.Fatal().Err(err).Str("filename", filename).Msg("文件错误")
		return
	}

	var f *os.File
	f, err = os.Open(filename)

	if err != nil {
		log.Fatal().Err(err).Msg("打开配置文件失败")
	}

	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal().Err(err).Msg("关闭配置文件失败")
		}
	}()

	if err = yaml.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatal().Err(err).Msg("解析配置文件失败")
	}
	_cfg = cfg
	return
}

func GlobConfig() Config {
	return _cfg
}
