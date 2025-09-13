package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ValidateRequest 根据请求方法自动绑定参数到上下文
// GET/DELETE -> Query
// 其他方法 -> JSON/Form
// key 用于设置到 c.Set 的上下文 key
func ValidateRequest[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var data T
		var err error

		switch c.Request.Method {
		case http.MethodGet, http.MethodDelete:
			err = c.ShouldBindQuery(&data)
		default:
			err = c.ShouldBind(&data)
		}

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "parameter error",
				"data":    err.Error(),
			})
			return
		}

		// 设置到上下文
		c.Set("payload", data)
		c.Next()
	}
}
