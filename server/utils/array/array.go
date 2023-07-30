package array

import (
	"github.com/spf13/cast"
	"strings"
)

func Join[T any](arg []T, str string) string {
	var ret strings.Builder
	for i, t := range arg {
		if i == 0 {
			ret.WriteString(cast.ToString(t))
		} else {
			ret.WriteString(str)
			ret.WriteString(cast.ToString(t))
		}
	}
	return ret.String()
}
