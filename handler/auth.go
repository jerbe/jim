package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jerbe/jim/config"
	"github.com/jerbe/jim/database"
	"github.com/jerbe/jim/errors"
	"github.com/jerbe/jim/log"
	"github.com/jerbe/jim/utils"
	"time"
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
}

// AuthLoginResponse 用户登陆返回数据
// @Description 用户登陆返回参数
type AuthLoginResponse struct {
	// Token 认证Token
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjE3MjUyNDkxMDZ9.S8bgf9lopUflrEcoiBGFToqWh4a9T-lCy1WqTiB9vTI"`

	// ExpiresAt 到期时间
	ExpiresAt int64 `json:"expires_at" binding:"required" example:"1725249106"`
}

// AuthLoginHandler
// @Summary      登陆
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
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	// 账户已经被禁用
	if user.Status == 0 {
		JSONError(ctx, StatusError, MessageAccountDisable)
		return
	}

	// 账户已经被删除
	if user.Status == 2 {
		JSONError(ctx, StatusError, MessageAccountDeleted)
		return
	}

	var password = utils.PasswordHash(req.Password)

	if password != user.Password {
		JSONError(ctx, StatusError, MessageIncorrectUsernameOrPassword)
		return
	}

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
		JSONError(ctx, StatusError, MessageInternalServerError)
		return
	}

	resp := AuthLoginResponse{
		Token:     tokenStr,
		ExpiresAt: expiresAt.Unix(),
	}

	JSON(ctx, resp)
}

// ====================================
// ============ USER LOGOUT ===========
// ====================================

// AuthLogoutHandler 用户登陆请求方法
func AuthLogoutHandler(c *gin.Context) {
	JSON(c)
}
