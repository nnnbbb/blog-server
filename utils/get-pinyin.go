package utils

import (
	"strings"

	"blog-server/config"

	"github.com/mozillazg/go-pinyin"
)

func GetPinYin(hans string) string {
	// 优先匹配自定义词典
	if val, ok := config.CustomCityPinyin[hans]; ok {
		return val
	}

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
