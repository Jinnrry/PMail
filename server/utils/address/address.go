package address

import "strings"

// IsValidEmailAddress 检查是否是有效的邮箱地址
func IsValidEmailAddress(str string) bool {
	ars := strings.Split(str, "@")
	if len(ars) != 2 {
		return false
	}
	return strings.Contains(ars[1], ".")
}
