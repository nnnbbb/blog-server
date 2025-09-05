package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ValidateBody[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload T
		if err := c.ShouldBind(&payload); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "parameter error",
				"data":    err.Error(),
			})
			return
		}
		// 将解析结果放入上下文中，供控制器使用
		c.Set("payload", payload)
		c.Next()
	}
}
