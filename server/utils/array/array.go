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

// Unique 数组去重
func Unique[T comparable](slice []T) []T {
	mp := map[T]bool{}
	for _, v := range slice {
		mp[v] = true
	}
	ret := []T{}
	for t, _ := range mp {
		ret = append(ret, t)
	}
	return ret
}

// Merge 求并集
func Merge[T any](slice1, slice2 []T) []T {
	s1Len := len(slice1)

	slice3 := make([]T, s1Len+len(slice2))
	for i, t := range slice1 {
		slice3[i] = t
	}

	for i, t := range slice2 {
		slice3[s1Len+i] = t
	}

	return slice3
}

// Intersect 求交集
func Intersect[T comparable](slice1, slice2 []T) []T {
	m := make(map[T]bool)
	nn := make([]T, 0)
	for _, v := range slice1 {
		m[v] = true
	}

	for _, v := range slice2 {
		exist, _ := m[v]
		if exist {
			nn = append(nn, v)
		}
	}
	return nn
}

// Difference 求差集 slice1-并集
func Difference[T comparable](slice1, slice2 []T) []T {
	m := make(map[T]bool)
	nn := make([]T, 0)
	inter := Intersect(slice1, slice2)
	for _, v := range inter {
		m[v] = true
	}

	for _, value := range slice1 {
		exist, _ := m[value]
		if !exist {
			nn = append(nn, value)
		}
	}
	return nn
}

// InArray 判断元素是否在数组中
func InArray[T comparable](needle T, haystack []T) bool {
	for _, t := range haystack {
		if needle == t {
			return true
		}
	}

	return false
}
