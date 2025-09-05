package server

import (
	"blog-server/controllers"
	"blog-server/forms"
	"blog-server/middlewares"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	health := new(controllers.HealthController)

	// router.Use(middlewares.ResponseWrapper())
	router.Use(middlewares.Recovery())
	router.GET("/health", health.Status)
	// router.Use(middlewares.AuthMiddleware())

	v1 := router.Group("v1")
	{
		userGroup := v1.Group("user")
		{
			user := new(controllers.UserController)
			userGroup.POST("/login", middlewares.ValidateBody[forms.LoginBody](), user.Login)

			router.Use(middlewares.JWTMiddleware())

			userGroup.GET("/:id", user.Retrieve)
		}
	}
	return router

}
