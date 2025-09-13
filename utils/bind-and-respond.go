package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError struct {
	Code    int
	Message string
	Err     error
}

func (e *APIError) Error() string {
	return e.Err.Error()
}

// 可选第三个参数
func NewAPIError(code int, message string, opts ...error) *APIError {
	var err error
	if len(opts) > 0 && opts[0] != nil {
		err = opts[0]
	} else {
		err = errors.New(message)
	}

	return &APIError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// 统一错误处理
func handleError(c *gin.Context, err error) {
	var apiErr *APIError
	status := http.StatusInternalServerError
	msg := "内部错误"

	if errors.As(err, &apiErr) {
		status = apiErr.Code
		msg = apiErr.Message
	} else {
		msg = err.Error()
	}

	c.AbortWithStatusJSON(status, gin.H{
		"code":  status,
		"error": msg,
	})
}

// 不带请求参数
func BindAndRespond[TRes any](handler func(c *gin.Context) (TRes, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := handler(c)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "success",
			"data":    res,
		})
	}
}

// 带请求参数
func BindAndRespondR[TReq any, TRes any](handler func(c *gin.Context, req TReq) (TRes, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TReq
		var err error

		switch c.Request.Method {
		case http.MethodGet, http.MethodDelete:
			err = c.ShouldBindQuery(&req)
		default:
			err = c.ShouldBind(&req)
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":  http.StatusBadRequest,
				"error": err.Error(),
			})
			return
		}

		res, err := handler(c, req)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "success",
			"data":    res,
		})
	}
}
