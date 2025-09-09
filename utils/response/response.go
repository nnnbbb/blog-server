package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"` // 可省略为空
}

// 成功响应
func Ok(c *gin.Context, data interface{}, messages ...string) {
	message := "success"
	if len(messages) > 0 && messages[0] != "" {
		message = messages[0]
	}
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	})
}

// 失败响应（自定义 HTTP 状态码）
func Fail(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// 简化错误响应（默认无 data）
func Error(c *gin.Context, code int, message string) {
	Fail(c, code, message)
}
