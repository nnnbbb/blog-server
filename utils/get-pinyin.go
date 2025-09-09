package utils

import (
	"strings"

	"github.com/mozillazg/go-pinyin"
)

func GetPinYin(hans string) string {
	args := pinyin.NewArgs()
	// 返回的是二维切片，例如 [["fu"], ["zhou"]]
	py := pinyin.Pinyin(hans, args)

	// 拼接
	var result []string
	for _, s := range py {
		if len(s) > 0 {
			result = append(result, s[0])
		}
	}

	return strings.Join(result, "")
}
