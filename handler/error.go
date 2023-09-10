package handler

import (
	"fmt"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/9/2 22:14
  @describe :
*/

var (
	MessageInternalServerError = "内部服务错误"

	MessageInvalidParams = "参数无效"

	MessageNotFound = "找不到记录"

	MessageInvalidGroupID = "'group_id'无效"

	MessageInvalidUserID = "'user_id'无效"

	MessageInvalidUserIDs = "'user_ids'无效"

	MessageInvalidUsername = "'username'无效"

	MessageInvalidNickname = "'nickname'无效"

	MessageInvalidPassword = "'password'无效"

	MessageInvalidSessionType = "'session_type'无效"

	MessageInvalidReceiverID = "'receiver_id'无效"

	MessageInvalidType = "'type'无效"

	MessageInvalidTargetID = "'target_id'无效"

	MessageInvalidLastMessageID = "'last_message_id'无效"

	MessageInvalidLimit = "'limit'无效"

	MessageChatYourself = "不可与自己聊天"

	MessageNotFriends = "您与对方不是好友关系"

	MessageAlreadyFriends = "你们已经是好友关系"

	MessageBlockYou = "对方已将您拉黑"

	MessageInvalidTimeFormat = "时间格式无效"

	MessageAccountExists = "账户已存在"

	MessageAccountNotExists = "账户不存在"

	MessageAccountDisable = "账户已停用"

	MessageAccountDeleted = "账户已删除"

	MessageTargetNotExists = "对方不存在"

	MessageTargetDisable = "对方已停用"

	MessageTargetDeleted = "对方已删除"

	MessageIncorrectUsernameOrPassword = "用户名或密码错误"

	MessageConfirmPasswordWrong = "确认密码错误"

	MessageIncorrectUsernameOrPasswordMoreTimes = "用户或密码错误超过限制"

	MessageEmptyCaptchaID = "验证码ID未填写"

	MessageEmptyCaptcha = "验证码未填写"

	MessageInvalidCaptcha = "验证码无效"
)

// MessageInvalidFormat 格式化参数无效错误
func MessageInvalidFormat(val string) string {
	return fmt.Sprintf("'%s'无效", val)
}
