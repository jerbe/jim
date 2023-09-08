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
	RequestID string `json:"request_id" binding:"required" example:"11234"`
	Status    int    `json:"status" binding:"required" example:"1"`
	Error     string `json:"error,omitempty" example:"系统内部错误"`
	Data      any    `json:"data,omitempty" `
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
func JSON(ctx *gin.Context, obj ...any) {
	var data any
	if len(obj) > 0 {
		data = obj[0]
	} else {
		data = struct{}{}
	}

	ctx.JSON(http.StatusOK, Response{RequestID: getAndStoreRequestID(ctx), Status: StatusOK, Data: data})
}

// JSONError 返回JSON格式的错误数据
func JSONError(ctx *gin.Context, errorCode int, errorMsg string, obj ...any) {
	var data any
	if len(obj) > 0 {
		data = obj[0]
	} else {
		data = struct{}{}
	}
	ctx.JSON(http.StatusOK, Response{RequestID: getAndStoreRequestID(ctx), Status: errorCode, Error: errorMsg, Data: data})
}

// JSONP 返回JSONP格式的数据
func JSONP(ctx *gin.Context, obj ...any) {
	var data any
	if len(obj) > 0 {
		data = obj[0]
	} else {
		data = struct{}{}
	}
	ctx.JSONP(http.StatusOK, Response{RequestID: getAndStoreRequestID(ctx), Status: StatusOK, Data: data})
}

// JSONPError 返回JSONP格式的错误数据
func JSONPError(ctx *gin.Context, errorCode int, errorMsg string, obj ...any) {
	var data any
	if len(obj) > 0 {
		data = obj[0]
	} else {
		data = struct{}{}
	}
	ctx.JSONP(http.StatusOK, Response{RequestID: getAndStoreRequestID(ctx), Status: errorCode, Error: errorMsg, Data: data})
}
