package utils

import (
	"fmt"
	"log"
)

// 定义彩色 log
var BlueLog = log.New(log.Writer(), "\033[34m", log.LstdFlags)

// 封装 Println，让输出后自动重置颜色
func Log(v ...any) {
	BlueLog.Output(2, fmt.Sprint(v...)+"\033[0m")
}
