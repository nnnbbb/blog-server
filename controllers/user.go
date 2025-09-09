package controllers

import (
	"net/http"
	"time"

	"blog-server/config"
	"blog-server/forms"
	"blog-server/services"
	"blog-server/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	day = 24 * time.Hour
)

type UserController struct{}

func (u UserController) Retrieve(c *gin.Context) {
	if c.Param("id") != "" {
		user, err := services.GetUserByID(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve user", "error": err})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "User founded!", "user": user})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}

func GenerateJWT(username string) (string, error) {
	config := config.GetConfig()
	jwtKey := config.GetString("server.jwtKey")
	stringKey := []byte(jwtKey)

	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(2 * day).Unix(), // 2 day 有效期
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(stringKey)
}

func (u UserController) Login(c *gin.Context) {
	// 假设从请求中获取用户名密码进行校验
	form := c.MustGet("payload").(forms.LoginBody)
	username := form.Username
	password := form.Password

	if username == "admin" && password == "123456" {
		token, err := GenerateJWT(username)

		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		response.Ok(c, gin.H{"token": token}, "登录成功")
	} else {
		response.Fail(c, http.StatusUnauthorized, "用户名或密码错误")
	}
}
