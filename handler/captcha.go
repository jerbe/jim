package handler

import (
	"context"
	"fmt"
	"github.com/jerbe/jim/config"
	"github.com/jerbe/jim/database"
	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"image/color"
	"time"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/9/10 21:42
  @describe :
*/

const (
	// MaxLoginFailTimes 最多登录失败次数
	MaxLoginFailTimes = 5

	// NeedCaptchaLoginFailTimes 需要验证码的失败登录次数
	NeedCaptchaLoginFailTimes = 3
)

// RedisCaptchaStore 用于存储验证码的Redis结构
type RedisCaptchaStore struct {
	cli redis.UniversalClient
}

func (s *RedisCaptchaStore) genKey(id string) string {
	return fmt.Sprintf("%s:captcha:id:%s", config.GlobConfig().Main.ServerName, id)
}

func (s *RedisCaptchaStore) Set(id string, value string) error {
	return s.cli.Set(context.Background(), s.genKey(id), value, time.Minute).Err()
}

func (s *RedisCaptchaStore) Get(id string, clear bool) string {
	key := s.genKey(id)
	val := s.cli.Get(context.Background(), key).Val()
	if clear {
		s.cli.Del(context.Background(), key)
	}
	return val
}

func (s *RedisCaptchaStore) Verify(id string, answer string, clear bool) bool {
	key := s.genKey(id)
	val, err := s.cli.Get(context.Background(), key).Result()
	// 有设置清除标签,先删除key,
	if clear {
		s.cli.Del(context.Background(), key)
	}

	if err != nil {
		return false
	}

	return answer != "" && val == answer
}

// Captcha 验证码组件
type Captcha struct {
	Store base64Captcha.Store

	DriverAudio base64Captcha.Driver

	DriverString base64Captcha.Driver

	DriverChinese base64Captcha.Driver

	DriverLanguage base64Captcha.Driver

	DriverMath base64Captcha.Driver

	DriverDigit base64Captcha.Driver
}

// getCaptcha 获取验证码组件
func getCaptcha() *Captcha {
	return _captcha
}

var _captcha *Captcha

// InitCaptcha 初始化验证码相关配件
func InitCaptcha() {
	_captcha = &Captcha{
		Store: &RedisCaptchaStore{cli: database.GlobDB.Redis},

		DriverAudio:    base64Captcha.DefaultDriverAudio,
		DriverMath:     base64Captcha.NewDriverMath(20, 100, 2, 2, &color.RGBA{}, base64Captcha.DefaultEmbeddedFonts, nil),
		DriverLanguage: nil,
		DriverChinese:  base64Captcha.NewDriverChinese(20, 100, 2, 2, 6, base64Captcha.TxtChineseCharaters, &color.RGBA{}, base64Captcha.DefaultEmbeddedFonts, nil),
		DriverDigit:    base64Captcha.DefaultDriverDigit,
		DriverString:   base64Captcha.NewDriverString(20, 100, 2, 2, 6, base64Captcha.TxtAlphabet, &color.RGBA{}, base64Captcha.DefaultEmbeddedFonts, nil),
	}
}
