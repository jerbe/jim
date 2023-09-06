package utils

import (
	"crypto/md5"
	"encoding/hex"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/12 00:32
  @describe :
*/

// MD5 封装的一个MD5加密字节组
func MD5(src []byte) []byte {
	data := md5.Sum(src)
	res := make([]byte, 32, 32)
	hex.Encode(res, data[:])
	return res
}

// MD5String 封装的一个MD5加密字符串的方法
func MD5String(src string) string {
	return string(MD5([]byte(src)))
}

// passwordSecretKey 密码再次加密的密钥
// @TODO 在有用户注册后此加密密钥禁止再修改,否则将导致前期注册的账户将无法登录系统
const passwordSecretKey = "jim@jerbe.me"

// PasswordHash 封装的一个哈希密码的加密方法
// 使用双重MD5进行加密
func PasswordHash(pwd string) string {
	pwd = pwd + passwordSecretKey
	return string(MD5(MD5([]byte(pwd))))
}
