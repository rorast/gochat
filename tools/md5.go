package tools

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

// MD5 加密
func MakePassword(plainpwd, salt string) string {
	return MD5Encode(plainpwd + salt)
}

// MD5 解密
func ValidPassword(plainpwd, salt string, password string) bool {
	md := MD5Encode(plainpwd + salt)
	fmt.Println(md + "				" + password)
	return md == password
}

// 小寫
func Md5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	tempStr := h.Sum(nil)
	return hex.EncodeToString(tempStr)
}

// 大寫
func MD5Encode(data string) string {
	return strings.ToUpper(Md5Encode(data))
}
