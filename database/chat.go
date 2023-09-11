package database

import (
	"encoding/json"
	"fmt"
	"github.com/jerbe/jcache"
	"github.com/jerbe/jim/errors"
	"github.com/jerbe/jim/log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/13 13:32
  @describe :
*/

const (
	// ChatMessageSessionTypePrivate 私聊会话类型
	ChatMessageSessionTypePrivate = 1

	// ChatMessageSessionTypeGroup 群聊会话类型
	ChatMessageSessionTypeGroup = 2

	// ChatMessageSessionTypeWorld 世界会话类型
	ChatMessageSessionTypeWorld = 99
)

const (
	// 消息类型: 1-纯文本,2-图片,3-语音,4-视频, 5-位置

	// ChatMessageTypePlainText 文本类型
	ChatMessageTypePlainText = 1

	// ChatMessageTypePicture 图像类型
	ChatMessageTypePicture = 2

	// ChatMessageTypeVoice 语音类型
	ChatMessageTypeVoice = 3

	// ChatMessageTypeVideo 视频类型
	ChatMessageTypeVideo = 4

	// ChatMessageTypeLocation 位置类型
	ChatMessageTypeLocation = 5
)

const (
	// ChatMessageBodyFormatGIF GIF类型
	ChatMessageBodyFormatGIF = "gif"

	// ChatMessageBodyFormatJPEG JPEG类型
	ChatMessageBodyFormatJPEG = "jpeg"

	// ChatMessageBodyFormatWEBP GIF类型
	ChatMessageBodyFormatWEBP = "webp"

	// ChatMessageBodyFormatMP3 mp3类型
	ChatMessageBodyFormatMP3 = "mp3"

	// ChatMessageBodyFormatVMA vma类型
	ChatMessageBodyFormatVMA = "vma"

	// ChatMessageBodyFormatMP4 mp4类型
	ChatMessageBodyFormatMP4 = "mp4"
)

/*
{
    "msg_id":1, // 递增
		"type":1, // 消息类型: 1-纯文本,2-图片,3-语音,4-视频, 5-位置
		"session_type":1,  // 会话类型, 1-私聊,2-群聊
		"room_id":"12345",  // 房间号
		"sender_id":1,  // 发送人ID
		"receiver_id":1, // 接收人ID, // 当,sessin_type=1时为私聊对象ID,当session_type=2时为群聊房间ID
		"send_status":1, // 消息发送状态,1-已发送,2-未抵达,3-已抵达
		"read_status":1, // 已读状态, 0-未读,1-已读
		"status":1, // 消息状态, 1-正常,2-已删除,3-已撤回
		"body":"{  // 消息主体
			// 文本信息。适用消息类型: 1
			\"text\": \"纯文本消息\",
			// 来源地址。通用字段，适用消息类型: 2,3,4
			\"src\": \"图片、语音、视频等地址\",

			// 文件格式。适用消息类型: 2,3,4
			\"format\": \"avi\",

			// 文件大小。适用消息类型: 2,3,4
			\"size\": \"123GB\",
			// 位置信息。适用消息类型: 5
			\"longitude\":\"100644.12323\",

			// 位置信息-纬度。 适用消息类型: 5
			\"latitude\":\"100644.12323\",

			// 位置信息-地图缩放级别。 适用消息类型: 5
			"\scale\": 123.3,

			// 位置信息。适用消息类型: 5
			\"location_label\": \"福建省厦门市某个地址\"
		}",
		"time":1, // 消息发送时间, 时间戳?
}
*/
// ChatMessage 聊天消息体
type ChatMessage struct {
	// ID
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// 消息ID
	MessageID int64 `bson:"message_id" json:"message_id"`

	// 房间号ID
	RoomID string `bson:"room_id" json:"room_id"`

	// 消息类型: 1-纯文本,2-图片,3-语音,4-视频, 5-位置
	Type int `bson:"type" json:"type"`

	// 会话类型, 1-私聊,2-群聊
	SessionType int `bson:"session_type" json:"session_type"`

	// 发送人ID
	SenderID int64 `bson:"sender_id" json:"sender_id"`

	// 接收人ID
	ReceiverID int64 `bson:"receiver_id" json:"receiver_id"`

	// 消息发送状态,1-已发送,2-未抵达,3-已抵达
	SendStatus int `bson:"send_status" json:"send_status"`

	// 已读状态, 0-未读,1-已读
	ReadStatus int `bson:"read_status" json:"read_status"`

	// 消息状态, 1-正常,2-已删除,3-已撤回
	Status int `bson:"status" json:"status"`

	// 消息主体
	Body ChatMessageBody `bson:"body" json:"body"`

	// 消息发送时间, 要用时间戳?
	CreatedAt int64 `bson:"created_at" json:"created_at"` // 消息时间

	// 消息最后更新时间,要用时间戳?
	UpdatedAt int64 `bson:"updated_at" json:"updated_at"`
}

// MarshalBinary 实现encoding.BinaryMarshaler接口
func (m *ChatMessage) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

// UnmarshalBinary 实现encoding.BinaryUnmarshaler接口
func (m *ChatMessage) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}

// ChatMessageBody 消息主体
type ChatMessageBody struct {
	// 文本信息。适用消息类型: 1
	Text string `bson:"text,omitempty" json:"text,omitempty"`

	// 来源地址。通用字段，适用消息类型: 2,3,4
	Src string `bson:"src,omitempty" json:"src,omitempty"`

	// 文件格式。适用消息类型: 2,3,4
	Format string `bson:"format,omitempty" json:"format,omitempty"`

	// 文件大小。适用消息类型: 2,3,4
	Size string `bson:"size,omitempty" json:"size,omitempty"`

	// 位置信息-经度。 适用消息类型: 5
	Longitude string `bson:"longitude,omitempty" json:"longitude,omitempty"`

	// 位置信息-纬度。 适用消息类型: 5
	Latitude string `bson:"latitude,omitempty" json:"latitude,omitempty"`

	// 位置信息-地图缩放级别。 适用消息类型: 5
	Scale float64 `bson:"scale,omitempty" json:"scale,omitempty"`

	// 位置信息标签。适用消息类型: 5
	LocationLabel string `bson:"location_label,omitempty" json:"location_label,omitempty"`
}

// ChatRoom 聊天室房间数据
type ChatRoom struct {
	// ID 房间号
	ID primitive.ObjectID `bson:"id" json:"id"`

	// RoomID 房间ID
	RoomID string `bson:"room_id" json:"room_id"`

	// SessionType 房间类型, 1-私聊,2-群聊
	SessionType int `bson:"session_type" json:"session_type"`

	// LastMessageID 最后一条消息ID,用于前端进行过滤排序
	LastMessageID int64 `bson:"last_message_id" json:"last_message_id"`

	// LastMessage 最后一条消息
	LastMessage ChatMessage `bson:"last_message" json:"last_message"`

	// CreatedAt 创建时间
	CreatedAt int64 `bson:"created_at" json:"created_at"`

	// UpdatedAt 最后更新时间
	UpdatedAt int64 `bson:"updated_at" json:"updated_at"`
}

// AddChatMessage 添加一条聊天消息
func AddChatMessage(msg *ChatMessage) error {
	now := msg.CreatedAt
	// mongo客户端

	db := GlobDB.Mongo.Database(DatabaseMongodbIM)
	srs := db.Collection(CollectionRoom).
		FindOneAndUpdate(GlobCtx, bson.M{
			"room_id": msg.RoomID,
		}, bson.M{
			"$inc": bson.M{"last_message_id": 1},
			"$set": bson.M{
				"updated_at": now,
			},
			"$setOnInsert": bson.M{
				"room_id":      msg.RoomID,
				"session_type": msg.SessionType,
				"created_at":   now,
			},
		}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	room := new(ChatRoom)
	err := srs.Decode(room)
	if err != nil {
		return errors.Wrap(err)
	}

	// 设置消息ID
	msg.MessageID = room.LastMessageID

	// 递增房间号的消息排列索引
	rs, err := db.Collection(CollectionMessage).InsertOne(GlobCtx, msg)
	if err != nil {
		// @TODO 推送到重试管道,进行重试插入
		return errors.Wrap(err)
	}
	msg.ID = rs.InsertedID.(primitive.ObjectID)

	// 更新聊天室的最后一条消息
	_, err = db.Collection(CollectionRoom).
		UpdateOne(GlobCtx, bson.M{
			"room_id": msg.RoomID,
			"$and": bson.A{
				bson.M{
					"$or": bson.A{
						bson.M{"last_message": bson.M{
							"$eq": nil,
						}},
						bson.M{"last_message.message_id": bson.M{"$lt": msg.MessageID}},
					},
				},
			},
		}, bson.M{
			"$set": bson.M{
				"last_message": msg,
			},
		})
	if err != nil {
		return errors.Wrap(err)
	}

	// 将聊天数据推送到缓存定长队列中去
	// @TODO 暂时这样push,但是这样处理不准确,需要做优化. 因为每个响应的是时长不一样,可能导致顺序不是正确的,甚至是断续的. 如 [1,3,2,4,7,6,5,8,9,11,10,19]
	cacheKey := cacheKeyFormatLastMessageList(msg.RoomID, msg.SessionType)
	err = jcache.Push(GlobCtx, cacheKey, msg)
	if err != nil && err.Error() == "WRONGTYPE Operation against a key holding the wrong kind of value" {
		jcache.Del(GlobCtx, cacheKey)
	}

	if err == nil {
		jcache.Expire(GlobCtx, cacheKey, jcache.RandomExpirationDuration())
	}

	return nil
}

// AddChatMessageTx 使用事务添加一条聊天消息
func AddChatMessageTx(msg *ChatMessage) error {
	now := msg.CreatedAt
	err := GlobDB.Mongo.UseSession(GlobCtx, func(sessionCtx mongo.SessionContext) error {
		//
		//t := time.Second
		_, err := sessionCtx.WithTransaction(sessionCtx, func(ctxb mongo.SessionContext) (interface{}, error) {
			// mongo客户端
			db := ctxb.Client().Database(DatabaseMongodbIM)
			srs := db.Collection(CollectionRoom).
				FindOneAndUpdate(ctxb, bson.M{
					"room_id": msg.RoomID,
				}, bson.M{
					"$inc": bson.M{"last_message_id": 1},
					"$set": bson.M{
						"updated_at": now,
					},
					"$setOnInsert": bson.M{
						"session_type": msg.SessionType,
						"created_at":   now,
					},
				}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
			room := new(ChatRoom)
			err := srs.Decode(room)
			if err != nil {
				return nil, errors.Wrap(err)
			}

			// 设置消息ID为递增的ID
			msg.MessageID = room.LastMessageID

			// 递增房间号的消息排列索引
			rs, err := db.Collection(CollectionMessage).InsertOne(ctxb, msg)
			if err != nil {
				return nil, errors.Wrap(err)
			}
			msg.ID = rs.InsertedID.(primitive.ObjectID)

			// 更新聊天室的最后一条消息
			_, err = db.Collection(CollectionRoom).
				UpdateOne(ctxb, bson.M{
					"room_id": msg.RoomID,
				}, bson.M{
					"$set": bson.M{
						"last_message": msg,
					},
				})
			if err != nil {
				return nil, errors.Wrap(err)
			}

			// 将聊天数据推送到缓存定长队列中去
			cacheKey := cacheKeyFormatLastMessageList(msg.RoomID, msg.SessionType)
			err = jcache.Push(GlobCtx, cacheKey, msg)
			if err != nil {
				jcache.Del(GlobCtx, cacheKey)
				log.Error().Err(err).Msgf("推送消息到缓存列表失败:\n %+v", err)
			}
			if err == nil {
				jcache.Expire(GlobCtx, cacheKey, jcache.RandomExpirationDuration())
			}

			return nil, nil
		}, options.Transaction().SetWriteConcern(writeconcern.Majority()).SetReadConcern(readconcern.Snapshot()))
		if err != nil {
			return errors.Wrap(err)
		}

		return nil
	})
	return errors.Wrap(err)
}

// RollbackChatMessage 撤回一条消息
func RollbackChatMessage(id primitive.ObjectID) (bool, error) {
	now := time.Now()
	rs, err := GlobDB.Mongo.Database(DatabaseMongodbIM).
		Collection(CollectionMessage).
		UpdateOne(GlobCtx, bson.M{
			"_id":        id,
			"status":     bson.M{"$ne": 3},
			"created_at": bson.M{"$gt": now.Add(-2 * time.Minute).UnixMilli()}, // 2分钟内禁止撤回
		}, bson.M{
			"$set": bson.M{
				"status":     3,
				"updated_at": now.UnixMilli(),
			},
		})
	if err != nil {
		return false, errors.Wrap(err)
	}

	return rs.ModifiedCount > 0, nil
}

type GetChatMessageListOptions struct {
	GetOptions

	Sort any

	LastMessageID int64

	Limit int64
}

const (
	defaultLimit     = 20
	defaultLastLimit = 20
	defaultMaxLimit  = 1000
)

func (opt *GetChatMessageListOptions) SetLimit(val int64) *GetChatMessageListOptions {
	opt.Limit = val
	return opt
}

func (opt *GetChatMessageListOptions) SetSort(val any) *GetChatMessageListOptions {
	opt.Sort = val
	return opt
}

func (opt *GetChatMessageListOptions) SetLastMessageID(val int64) *GetChatMessageListOptions {
	opt.LastMessageID = val
	return opt
}

func NewChatMessageListOptions() *GetChatMessageListOptions {
	return &GetChatMessageListOptions{
		Limit:         defaultLimit,
		Sort:          bson.M{"message_id": -1},
		LastMessageID: 0,
	}
}

type GetChatMessageListFilter struct {
	RoomID        string `bson:"room_id"`
	SessionType   int    `bson:"session_type"`
	Sort          any    `bson:"sort"`
	LastMessageID *int64 `bson:"last_message_id"`
	Limit         *int   `bson:"limit"`
}

func (f *GetChatMessageListFilter) SetLimit(val int) *GetChatMessageListFilter {
	f.Limit = &val
	return f
}

func (f *GetChatMessageListFilter) SetSort(val any) *GetChatMessageListFilter {
	f.Sort = val
	return f
}

func (f *GetChatMessageListFilter) SetLastMessageID(val int64) *GetChatMessageListFilter {
	f.LastMessageID = &val
	return f
}

// GetChatMessageList 获取消息列表
func GetChatMessageList(filter *GetChatMessageListFilter, opts ...*GetOptions) ([]*ChatMessage, error) {
	if filter.Sort == nil {
		filter.Sort = bson.M{"message_id": -1}
	}

	if filter.LastMessageID == nil {
		filter.LastMessageID = new(int64)
	}

	if filter.Limit == nil {
		filter.Limit = new(int)
	}

	rs, err := GlobDB.Mongo.Database(DatabaseMongodbIM).
		Collection(CollectionMessage).
		Find(GlobCtx, bson.M{
			"room_id":      filter.RoomID,
			"session_type": filter.SessionType,
			"message_id": bson.M{
				"$gte": *filter.LastMessageID,
				"$lt":  (*filter.LastMessageID) + int64(*filter.Limit),
			},
		}, options.Find().SetSort(filter.Sort))
	if err != nil {
		if errors.IsNoRecord(err) {
			return nil, errors.Wrap(err)
		}
		return nil, errors.Wrap(err)
	}
	defer rs.Close(GlobCtx)
	messages := make([]*ChatMessage, 0)
	for rs.Next(GlobCtx) {
		msg := new(ChatMessage)
		err = rs.Decode(msg)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// GetLastChatMessageList 获取最近的聊天消息列表
func GetLastChatMessageList(roomID string, sessionType int, opts ...*GetOptions) ([]*ChatMessage, error) {
	opt := MergeGetOptions(opts)
	cacheKey := cacheKeyFormatLastMessageList(roomID, sessionType)

	// 如果有使用缓存,则从缓存中获取
	if opt.UseCache() {
		exists, _ := jcache.Exists(GlobCtx, cacheKey)
		if exists {
			var messages []*ChatMessage
			err := jcache.RangAndScan(GlobCtx, &messages, cacheKey, defaultLastLimit)
			if err == nil {
				return messages, nil
			}

			// 如果有记录,并且记录内容为空,则表示被标记成查询空
			if errors.IsEmptyRecord(err) {
				return nil, errors.Wrap(err)
			}
		}
	}

	rs, err := GlobDB.Mongo.Database(DatabaseMongodbIM).
		Collection(CollectionMessage).
		Find(GlobCtx, bson.M{
			"room_id":      roomID,
			"session_type": sessionType,
		}, options.Find().SetSort(bson.M{"message_id": -1}).SetLimit(defaultLastLimit))

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// @todo 设置缓存为找不到记录

			cacheKey := cacheKeyFormatLastMessageList(roomID, sessionType)
			if err := jcache.SetNX(GlobCtx, cacheKey, nil, jcache.DefaultEmptySetNXDuration); err != nil {
				log.Error().Err(err).Str("cache_key", cacheKey).Msg("缓存写入失败")
			}

			return nil, errors.Wrap(err)
		}
		return nil, errors.Wrap(err)
	}

	defer rs.Close(GlobCtx)
	messages := make([]*ChatMessage, 0, defaultLastLimit)
	pushData := make([]any, 0, defaultLastLimit)
	for rs.Next(GlobCtx) {
		msg := new(ChatMessage)
		err = rs.Decode(msg)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		messages = append(messages, msg)
		pushData = append(pushData, msg)
	}

	err = jcache.Push(GlobCtx, cacheKey, pushData...)
	if err != nil && err.Error() == "WRONGTYPE Operation against a key holding the wrong kind of value" {
		jcache.Del(GlobCtx, cacheKey)
		err = jcache.Push(GlobCtx, cacheKey, pushData...)
		if err != nil {
			log.Warn().Err(err).Msg("缓存插入聊天消息失败")
		}
	}
	if err == nil {
		jcache.Expire(GlobCtx, cacheKey, jcache.RandomExpirationDuration())
	}

	// 设置缓存
	return messages, nil
}

// ==================================================================================
// ============================== 缓存操作 ============================================
// ==================================================================================
// cacheKeyFormatLastMessageList 格式化最后消息列表的缓存 key
func cacheKeyFormatLastMessageList(roomID string, sessionType int) string {
	return fmt.Sprintf("%s:chat_message:last_list:%d_%s", CacheKeyPrefix, sessionType, roomID)
}
