package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/11 10:58
  @describe :
*/

type Response struct {
	Status int    `json:"status" binding:"required" example:"1"`
	Error  string `json:"error,omitempty" example:"系统内部错误"`
	Data   any    `json:"data,omitempty" `
}

// HTML 返回HTML模板数据
func HTML(c *gin.Context, name string, obj ...any) {
	if len(obj) > 0 {
		c.HTML(http.StatusOK, name, obj[0])
	} else {
		c.HTML(http.StatusOK, name, nil)
	}
}

// JSON 返回JSON格式的数据
func JSON(c *gin.Context, obj ...any) {
	var data any
	if len(obj) > 0 {
		data = obj[0]
	} else {
		data = struct{}{}
	}

	c.JSON(http.StatusOK, Response{Status: StatusOK, Data: data})
}

// JSONError 返回JSON格式的错误数据
func JSONError(c *gin.Context, errorCode int, errorMsg string, obj ...any) {
	var data any
	if len(obj) > 0 {
		data = obj[0]
	} else {
		data = struct{}{}
	}
	c.JSON(http.StatusOK, Response{Status: errorCode, Error: errorMsg, Data: data})
}

// JSONP 返回JSONP格式的数据
func JSONP(c *gin.Context, obj ...any) {
	var data any
	if len(obj) > 0 {
		data = obj[0]
	} else {
		data = struct{}{}
	}
	c.JSONP(http.StatusOK, Response{Status: StatusOK, Data: data})
}

// JSONPError 返回JSONP格式的错误数据
func JSONPError(c *gin.Context, errorCode int, errorMsg string, obj ...any) {
	var data any
	if len(obj) > 0 {
		data = obj[0]
	} else {
		data = struct{}{}
	}
	c.JSONP(http.StatusOK, Response{Status: errorCode, Error: errorMsg, Data: data})
}
