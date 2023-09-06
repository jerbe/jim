package handler

import (
	"fmt"
	"github.com/jerbe/jim/log"
	"github.com/jerbe/jim/websocket"

	"github.com/gin-gonic/gin"
	gWebsocket "github.com/gorilla/websocket"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/19 23:55
  @describe :
*/

// WebsocketMessageRequest websocket请求参数
type WebsocketMessageRequest struct {
	Action   string `json:"action"`
	ActionID string `json:"action_id"`
	Data     any    `json:"data"`
}

// WebsocketMessageResponse websocket返回数据
type WebsocketMessageResponse struct {
	Action   string `json:"action"`
	ActionID string `json:"action_id"`
	Data     any    `json:"data"`
}

var upgrader = &gWebsocket.Upgrader{}

var websocketManager = websocket.DefaultManager

// WebsocketHandler websocket连接处理方法 `/api/ws`
func WebsocketHandler(ctx *gin.Context) {
	user := LoginUserFromContext(ctx)

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		JSONError(ctx, StatusError, "无法连接")
		return
	}

	defer func(conn *gWebsocket.Conn) {
		websocketManager.RemoveConnect(fmt.Sprintf("%d", user.ID), conn)
		err := conn.Close()
		if err != nil {
			log.ErrorFromGinContext(ctx).Err(err).
				Str("err_format", fmt.Sprintf("%+v", err)).
				Int64("user_id", user.ID).
				Str("remote", conn.RemoteAddr().String()).Msg("关闭websocket失败")
		}
	}(conn)

	websocketManager.AddConnect(fmt.Sprintf("%d", user.ID), conn)

	for {

		_, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		//go parseWSMessage(ctx, message)
	}
}
