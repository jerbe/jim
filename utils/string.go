package utils

import (
	"fmt"
	"github.com/google/uuid"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/16 12:52
  @describe :
*/

// FormatPrivateRoomID 格式化私人聊天室房间号
func FormatPrivateRoomID(userAID, userBID int64) string {
	var a, b = userAID, userBID
	if a > b {
		a, b = b, a
	}
	return fmt.Sprintf("%08x%08x", a, b)
}

// FormatGroupRoomID 格式化群组聊天室房间号
func FormatGroupRoomID(groupID int64) string {
	return fmt.Sprintf("%08x", groupID)
}

// FormatWorldRoomID 格式化世界聊天室房间号
func FormatWorldRoomID(worldID int64) string {
	return fmt.Sprintf("world_%04x", worldID)
}

// StringCut 裁剪字符串
func StringCut(data string, limit int) string {
	d := []rune(data)
	l := len(d)
	if l < limit {
		limit = l
	}
	return string(d[0:limit])
}

// StringLen 返回字符串的长度,因为中文需要3个字节(byte),所以,我们使用rune来代替
func StringLen(data string) int {
	return len([]rune(data))
}

// UUID 全球唯一标识
func UUID() string {
	return uuid.New().String()
}
