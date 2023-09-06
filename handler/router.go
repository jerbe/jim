package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jerbe/jim/pubsub"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023-08-10 12:39
  @describe :
*/

// @title           JIM Jerbe's Instant Messaging Service
// @version         1.0
// @description     Jerbe的即时通讯服务
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/

// @securityDefinitions.apikey  APIKeyHeader
// @in header
// @name Authorization

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

var rootRouter *gin.Engine

func Init() *gin.Engine {
	rootRouter = gin.New()
	rootRouter.Use(RecoverMiddleware(), CORSMiddleware())
	{
		authGroup := rootRouter.Group("/api/v1/auth", RequestLogMiddleware())
		authGroup.POST("/login", AuthLoginHandler)
		authGroup.POST("/register", AuthRegisterHandler)
		authGroup.POST("/logout", AuthLoginHandler)
	}

	// WebSocket连接
	rootRouter.GET("/api/v1/ws", WebsocketMiddleware(), WebsocketHandler)

	apiGroup := rootRouter.Group("/api/v1", RequestLogMiddleware(), CheckAuthMiddleware())
	{
		// 个人画像
		profile := apiGroup.Group("/profile")
		profile.GET("/info", GetProfileInfoHandler)
	}

	{
		// 聊天
		chat := apiGroup.Group("/chat")
		chat.POST("/message/send", SendChatMessageHandler)
		chat.POST("/message/rollback_", RollbackChatMessageHandler)
		chat.POST("/message/delete", DeleteChatMessageHandler)
	}

	{
		// 聊天
		friend := apiGroup.Group("/friend")
		friend.POST("/find", FindFriendHandler)
		friend.POST("/update", UpdateFriendHandle)

		friend.POST("/invite/add", AddFriendInviteHandler)
		friend.POST("/invite/update", UpdateFriendInviteHandler)
	}

	{
		// 聊天
		group := apiGroup.Group("/group")
		group.POST("/create", CreateGroupHandler)
		group.POST("/join", JoinGroupHandler)
		group.POST("/leave", LeaveGroupHandler)
		group.POST("/update", UpdateGroupHandler)

		group.POST("/member/add", AddGroupMemberHandler)
		group.POST("/member/update", UpdateGroupMemberHandler)
		group.POST("/member/remove", RemoveGroupMemberHandler)
	}

	var subscriber = pubsub.NewSubscriber()
	subscriber.Subscribe(pubsub.ChannelChatMessage, pubsub.PayloadTypeChatMessage, SubscribeChatMessageHandler)
	subscriber.Subscribe(pubsub.ChannelNotify, pubsub.PayloadTypeFriendInvite, SubscribeFriendInviteHandler)

	return rootRouter
}
