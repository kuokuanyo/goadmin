package modules

import (
	"github.com/satori/go.uuid"
	"html/template"
	"strconv"
)

// 判斷第二個參數(string)是否存在[]string(第一個參數中)
func InArray(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

// 判斷參數del後回傳del+s(參數)+del或[s(參數)]
func Delimiter(del, s string) string {
	if del == "[" {
		return "[" + s + "]"
	}
	return del + s + del
}

func FilterField(filed, delimiter string) string {
	if delimiter == "[" {
		return "[" + filed + "]"
	}
	return delimiter + filed + delimiter
}

func InArrayWithoutEmpty(arr []string, str string) bool {
	if len(arr) == 0 {
		return true
	}
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

func RemoveBlankFromArray(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func Uuid() string {
	uid, _ := uuid.NewV4()
	rst := uid.String()
	return rst
}

func SetDefault(source, def string) string {
	if source == "" {
		return def
	}
	return source
}

func GetPage(page string) (pageInt int) {
	if page == "" {
		pageInt = 1
	} else {
		pageInt, _ = strconv.Atoi(page)
	}
	return
}

func AorB(condition bool, a, b string) string {
	if condition {
		return a
	}
	return b
}

func AorEmpty(condition bool, a string) string {
	if condition {
		return a
	}
	return ""
}

func AorBHTML(condition bool, a, b template.HTML) template.HTML {
	if condition {
		return a
	}
	return b
}
