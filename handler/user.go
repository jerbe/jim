package handler

import (
	"github.com/gin-gonic/gin"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/15 16:35
  @describe :
*/

// ====================================
// ============ 账户信息 ================
// ====================================

// ProfileInfoResponse 个人画像的数据结构
type ProfileInfoResponse struct {
	Username     string `json:"username"`
	BirthDate    string `json:"birth_date" `
	Avatar       string `json:"avatar"`
	OnlineStatus int    `json:"online_status"`
}

// GetProfileInfoHandler 获取个人画像信息
func GetProfileInfoHandler(c *gin.Context) {
	currentUser := LoginUserFromContext(c)

	rspUser := &ProfileInfoResponse{
		Username:     currentUser.Username,
		Avatar:       currentUser.Avatar, // 如果为空,设置成默认地址?
		OnlineStatus: currentUser.OnlineStatus,
		BirthDate:    "0000-00-00",
	}

	if currentUser.BirthDate != nil {
		rspUser.BirthDate = currentUser.BirthDate.Format("2006-01-02")
	}

	JSON(c, rspUser)
}
