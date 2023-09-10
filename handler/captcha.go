package handler

import (
	"github.com/jerbe/jim/database"
	"github.com/mojocn/base64Captcha"
	"image/color"
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
