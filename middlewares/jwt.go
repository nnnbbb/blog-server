package middlewares

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware(requireAuth bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			if requireAuth {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少或无效的 Authorization 头"})
				c.Abort()
				return
			}
			// 非强制模式下继续执行
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		jwtKey := os.Getenv("jwtKey")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 检查签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtKey), nil
		})

		if err != nil || !token.Valid {
			if requireAuth {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 token"})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// 将用户信息传入 context（可选）
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if username, ok := claims["username"].(string); ok {
				c.Set("username", username)
			}
		}

		c.Next()
	}
}
