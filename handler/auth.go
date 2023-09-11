package handler

import (
	"context"
	"fmt"
	"github.com/mojocn/base64Captcha"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jerbe/jim/config"
	"github.com/jerbe/jim/database"
	"github.com/jerbe/jim/errors"
	"github.com/jerbe/jim/log"
	"github.com/jerbe/jim/utils"

	"github.com/golang-jwt/jwt/v5"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023-08-10 12:40
  @describe :
*/

// ====================================
// ============ USER REGISTER =========
// ====================================

// AuthRegisterRequest
// @Description 用户注册请求参数
type AuthRegisterRequest struct {
	// Username 账户名
	Username string `json:"username" binding:"required" example:"admin"`

	// Password 密码
	Password string `json:"password" binding:"required" minLength:"8" example:"password"`

	// ConfirmPassword 确认密码
	ConfirmPassword string `json:"confirm_password" binding:"required" minLength:"8" example:"password"`

	// BirthDate 生辰八字
	BirthDate string `json:"birth_date,omitempty" minLength:"8" example:"2016-01-02"`

	// Nickname 昵称
	Nickname string `json:"nickname,omitempty"  minLength:"2" example:"昵称"`

	// Captcha 验证码
	Captcha string `json:"captcha" binding:"required" `

	// CaptchaID 验证码ID, 通过调用 /api/v1/auth/captcha 获得
	CaptchaID string `json:"captcha_id" binding:"required" `
}

// AuthRegisterResponse
// @Description 用户注册返回数据
type AuthRegisterResponse struct {
	// UserID 用户ID
	UserID *int64 `json:"user_id" binding:"required" example:"10086"`

	// Username 账户名
	Username string `json:"username" binding:"required" example:"admin"`

	// Nickname 昵称
	Nickname string `json:"nickname" binding:"required" example:"the king"`

	// BirthDate 生辰八字
	BirthDate string `form:"birth_date" binding:"required" example:"2016-01-02"`

	// Avatar 头像地址
	Avatar string `json:"avatar" binding:"required" example:"https://www.baidu.com/logo.png"`
}

// AuthRegisterHandler
// @Summary      注册
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      AuthRegisterRequest  true  "请求JSON数据体"
// @Success      200  {object}  Response{data=AuthRegisterResponse}
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/auth/register [post]
func AuthRegisterHandler(ctx *gin.Context) {
	// 获取表单数据
	var req AuthRegisterRequest
	if err := ctx.BindJSON(&req); err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	if req.Password != req.ConfirmPassword {
		JSONError(ctx, StatusError, MessageConfirmPasswordWrong)
		return
	}

	if req.CaptchaID == "" {
		JSONError(ctx, StatusError, MessageEmptyCaptchaID)
		return
	}

	if req.Captcha == "" {
		JSONError(ctx, StatusError, MessageEmptyCaptcha)
		return
	}

	// 验证码检测
	if !getCaptcha().Store.Verify(req.CaptchaID, req.Captcha, true) {
		JSONError(ctx, StatusError, MessageInvalidCaptcha)
		return
	}

	// 查找用户是否已经存在
	exist, err := database.UserExistByUsername(req.Username)
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).Str("username", req.Username).Msg("判断用户是否存在失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	if exist {
		JSONError(ctx, StatusError, MessageAccountExists)
		return
	}

	// 进行数据插入
	user := &database.User{
		Username:     req.Username,
		Nickname:     req.Nickname,
		Password:     utils.PasswordHash(req.Password),
		Status:       1,
		OnlineStatus: 1,
	}

	if req.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			JSONError(ctx, StatusError, MessageInvalidTimeFormat)
			return
		}
		user.BirthDate = &birthDate
	}

	// 进行缓存更新
	if err := database.AddUser(user); err != nil {
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).Str("username", user.Username).Msg("添加用户失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	} else {
		JSON(ctx)
	}
}

// ====================================
// ============ USER LOGIN ============
// ====================================

// AuthLoginRequest
// @Description 用户登陆请求参数
type AuthLoginRequest struct {
	// Username 账户名
	Username string `json:"username" binding:"required" example:"admin"`

	// Password 密码
	Password string `json:"password" binding:"required" example:"password"`

	// Captcha 验证码;当登录错误超过次数时,必须填
	Captcha string `json:"captcha"`

	// CaptchaID 验证码ID; 当登录错误超过次数时,必须填; 通过调用 /api/v1/auth/captcha 获得
	CaptchaID string `json:"captcha_id"`
}

// AuthLoginResponse 用户登陆返回数据
// @Description 用户登陆返回参数
type AuthLoginResponse struct {
	// Token 认证Token
	Token *string `json:"token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjE3MjUyNDkxMDZ9.S8bgf9lopUflrEcoiBGFToqWh4a9T-lCy1WqTiB9vTI"`

	// ExpiresAt 到期时间
	ExpiresAt *int64 `json:"expires_at,omitempty" example:"1725249106"`

	// FailTimes 累计失败次数
	FailTimes *int64 `json:"fail_times,omitempty"`

	// NeedCaptcha 是否需要验证码
	NeedCaptcha *bool `json:"need_captcha,omitempty"`
}

// AuthLoginHandler
// @Summary      登录
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        jsonRaw    body      AuthLoginRequest  true  "请求JSON数据体"
// @Success      200  {object}  Response{data=AuthLoginResponse}
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/auth/login [post]
func AuthLoginHandler(ctx *gin.Context) {
	req := &AuthLoginRequest{}
	if err := ctx.BindJSON(req); err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	// 1. 验证登录失败次数是否超过限制
	loginFailRedisKey := fmt.Sprintf("%s:user:login_fail:%s", config.GlobConfig().Main.ServerName, req.Username)
	times, err := database.GlobDB.Redis.Get(context.Background(), loginFailRedisKey).Int64()

	if err != nil && !errors.IsNoRecord(err) {
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).
			Str("username", req.Username).
			Str("redis_key", loginFailRedisKey).
			Msg("获取用户登录失败次数发生错误")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	resp := &AuthLoginResponse{}
	if times >= NeedCaptchaLoginFailTimes {
		resp.FailTimes = &times
		needCaptcha := true
		resp.NeedCaptcha = &needCaptcha
	}

	// 已经达到了失败次数
	if times >= MaxLoginFailTimes {
		resp.FailTimes = &times
		JSONError(ctx, StatusError, MessageIncorrectUsernameOrPasswordMoreTimes, resp)
		return
	}

	// 2. 验证是否启用了验证码,并且验证码正确
	if times >= NeedCaptchaLoginFailTimes {
		if req.CaptchaID == "" {
			JSONError(ctx, StatusError, MessageEmptyCaptcha, resp)
			return
		}

		if req.Captcha == "" {
			JSONError(ctx, StatusError, MessageEmptyCaptcha, resp)
			return
		}

		if !getCaptcha().Store.Verify(req.CaptchaID, req.Captcha, true) {
			JSONError(ctx, StatusError, MessageInvalidCaptcha, resp)
			return
		}
	}

	// 3. 验证用户是否存在
	user, err := database.GetUserByUsername(req.Username)
	if err != nil {
		if errors.IsNoRecord(err) {
			JSONError(ctx, StatusError, MessageAccountNotExists)
			return
		}
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).
			Str("username", req.Username).
			Msg("获取用户信息失败")
		JSONError(ctx, StatusError, MessageInternalServerError, resp)
		return
	}

	// 4. 验证密码是否正确
	var password = utils.PasswordHash(req.Password)
	if password != user.Password {
		// 递增失败次数
		incr, err := database.GlobDB.Redis.Incr(context.Background(), loginFailRedisKey).Result()
		if err != nil {
			log.ErrorFromGinContext(ctx).Err(err).
				Str("err_format", fmt.Sprintf("%+v", err)).
				Str("username", req.Username).
				Str("redis_key", loginFailRedisKey).
				Msg("递增用户登录失败业务发生错误")
			JSONError(ctx, StatusError, MessageInternalServerError, resp)
			return
		}

		// 设置登录失败次数的redis的超时时间
		database.GlobDB.Redis.Expire(context.Background(), loginFailRedisKey, time.Minute*5)

		// 判断是否已经达到了登录失败次数极限
		if incr >= MaxLoginFailTimes {
			incr = MaxLoginFailTimes
			resp.FailTimes = &incr
			JSONError(ctx, StatusError, MessageIncorrectUsernameOrPasswordMoreTimes, resp)
			return
		}

		// 判断是否已经达到了登录失败需要验证码的的极限
		if incr >= NeedCaptchaLoginFailTimes {
			resp.FailTimes = &incr
			needCaptcha := true
			resp.NeedCaptcha = &needCaptcha
		}

		JSONError(ctx, StatusError, MessageIncorrectUsernameOrPassword, resp)
		return
	}

	// 5. 删除登录失败的rediskey
	database.GlobDB.Redis.Del(context.Background(), loginFailRedisKey)

	// 6. 验证账户已经被禁用
	if user.Status == 0 {
		JSONError(ctx, StatusError, MessageAccountDisable, resp)
		return
	}

	// 7. 验证账户已经被删除
	if user.Status == 2 {
		JSONError(ctx, StatusError, MessageAccountDeleted, resp)
		return
	}

	// 8. 生成token
	// 这里用于调试,所以设置成一年,上线时按需要使用time.Add指定过期时长
	//expiresAt := time.Now().Add(time.Hour)
	expiresAt := time.Now().AddDate(1, 0, 0)
	var claims = UserClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt), // 1小时后失效
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signKey := []byte(config.GlobConfig().Main.JwtSigningKey)
	tokenStr, err := token.SignedString(signKey)
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).Msg("签名失败")
		JSONError(ctx, StatusError, MessageInternalServerError, resp)
		return
	}

	// 9. 设置token
	expiresAtUnix := expiresAt.Unix()
	resp.Token = &tokenStr
	resp.ExpiresAt = &expiresAtUnix
	resp.FailTimes = nil
	resp.NeedCaptcha = nil
	JSON(ctx, resp)
}

// GetCaptchaRequest
// @Description 获取验证码请求参数
type GetCaptchaRequest struct {
	Type string `form:"type" example:"string"`
}

// GetCaptchaResponse
// @Description 获取验证码返回数据
type GetCaptchaResponse struct {
	// ID 验证码验证ID
	ID string `json:"id" binding:"required" example:"7uh37xVCN0oGarKZ79nx"`

	// Type 验证码类型
	Type string `json:"type" binding:"required" example:"audio"`

	// Data 验证码数据;有图片数据,也有音频数据,前端需要根据type生成对应媒体数据
	Data string `json:"data" binding:"required"`
}

// GetCaptchaHandler
// @Summary      登录
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        type    query      string  true  "capthca类型;audio,string,math,chinese,digit"
// @Success      200  {object}  Response{data=GetCaptchaResponse}
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /v1/auth/captcha [get]
func GetCaptchaHandler(ctx *gin.Context) {
	req := new(GetCaptchaRequest)
	err := ctx.BindQuery(req)

	if err != nil {
		JSONError(ctx, StatusError, err.Error())
		return
	}

	var driver base64Captcha.Driver
	switch req.Type {
	case "audio":
		driver = getCaptcha().DriverAudio
	case "string":
		driver = getCaptcha().DriverString
	case "math":
		driver = getCaptcha().DriverMath
	case "chinese":
		driver = getCaptcha().DriverChinese
	case "language":
		driver = getCaptcha().DriverLanguage
	default:
		req.Type = "digit"
		driver = getCaptcha().DriverDigit
	}

	c := base64Captcha.NewCaptcha(driver, getCaptcha().Store)
	id, b64s, err := c.Generate()
	if err != nil {
		log.ErrorFromGinContext(ctx).Err(err).
			Str("err_format", fmt.Sprintf("%+v", err)).Msg("生成'captch'失败")
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	rsp := &GetCaptchaResponse{ID: id, Data: b64s, Type: req.Type}
	JSON(ctx, rsp)
}

// ====================================
// ============ USER LOGOUT ===========
// ====================================

// AuthLogoutHandler 用户登陆请求方法
func AuthLogoutHandler(c *gin.Context) {
	JSON(c)
}
