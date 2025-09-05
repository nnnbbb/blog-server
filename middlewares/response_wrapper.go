package middlewares

import (
	"bytes"
	"encoding/json"

	"github.com/gin-gonic/gin"
)

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	// 拦截但不输出（只缓存）
	return w.body.Write(b)
}

func ResponseWrapper() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 替换 response writer
		bw := &bodyWriter{body: bytes.NewBuffer([]byte{}), ResponseWriter: c.Writer}
		c.Writer = bw

		c.Next()

		// 拿到原始响应内容
		respBytes := bw.body.Bytes()
		if len(respBytes) == 0 {
			return
		}

		// 判断是否是 JSON
		var original interface{}
		if err := json.Unmarshal(respBytes, &original); err != nil {
			// 不是 JSON，原样输出
			c.Writer = bw.ResponseWriter
			c.Writer.WriteHeaderNow()
			c.Writer.Write(respBytes)
			return
		}

		// 设置 Content-Type
		c.Writer = bw.ResponseWriter
		c.Header("Content-Type", "application/json")
		c.AbortWithStatusJSON(c.Writer.Status(), gin.H{
			"code": c.Writer.Status(),
			"data": original,
		})
	}
}
