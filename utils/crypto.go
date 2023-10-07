package utils

import (
	goutils "github.com/jerbe/go-utils"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/12 00:32
  @describe :
*/

// passwordSecretKey 密码再次加密的密钥
// @TODO 在有用户注册后此加密密钥禁止再修改,否则将导致前期注册的账户将无法登录系统
const passwordSecretKey = "jim@jerbe.me"

// PasswordHash 封装的一个哈希密码的加密方法
// 使用双重MD5进行加密
func PasswordHash(pwd string) string {
	pwd = pwd + passwordSecretKey
	return string(goutils.MD5(goutils.MD5([]byte(pwd))))
}
