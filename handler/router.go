package handler

import (
	"time"

	"github.com/jerbe/jim/pubsub"

	goutils "github.com/jerbe/go-utils"

	"github.com/gin-gonic/gin"
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

// @securityDefinitions.apikey  APIKeyQuery
// @in query
// @name token

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

func InitRouter() *gin.Engine {
	var rootRouter *gin.Engine
	rootRouter = gin.New()
	rootRouter.Use(RecoverMiddleware(), CORSMiddleware())
	{
		authGroup := rootRouter.Group("/api/v1/auth", RequestLogMiddleware())
		authGroup.POST("/login", AuthLoginHandler)
		authGroup.POST("/register", AuthRegisterHandler)
		authGroup.POST("/logout", AuthLoginHandler)

		authGroup.POST("/captcha", GetCaptchaHandler)
	}

	// WebSocket连接
	rootRouter.GET("/api/v1/ws", WebsocketMiddleware(), WebsocketHandler)

	apiGroup := rootRouter.Group("/api/v1", RateLimitMiddleware(goutils.NewLimiter(1000000, time.Second)), RequestLogMiddleware(), CheckAuthMiddleware())
	{
		// 个人画像
		profile := apiGroup.Group("/profile")
		profile.GET("/info", GetProfileInfoHandler)
	}

	{
		// 聊天
		chat := apiGroup.Group("/chat")
		chat.POST("/message/send", SendChatMessageHandler)
		chat.POST("/message/rollback", RollbackChatMessageHandler)
		chat.POST("/message/delete", DeleteChatMessageHandler)
		chat.GET("/message/last", GetLastChatMessagesHandler)
	}

	{
		// 聊天
		friend := apiGroup.Group("/friend")
		friend.GET("/find", FindFriendHandler)
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

	return rootRouter
}

// InitSubscribe 初始化订阅
func InitSubscribe() {
	var subscriber = pubsub.NewSubscriber()
	subscriber.Subscribe(pubsub.ChannelChatMessage, pubsub.PayloadTypeChatMessage, SubscribeChatMessageHandler)
	subscriber.Subscribe(pubsub.ChannelNotify, pubsub.PayloadTypeFriendInvite, SubscribeFriendInviteHandler)
}
