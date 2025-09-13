package server

import (
	"blog-server/controllers"
	"blog-server/middlewares"
	"blog-server/utils"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // 前端地址
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-App-Version"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	health := new(controllers.HealthController)

	// router.Use(middlewares.ResponseWrapper())
	router.Use(middlewares.Recovery())
	router.GET("/health", health.Status)
	// router.Use(middlewares.AuthMiddleware())

	api := router.Group("api")
	{
		userGroup := api.Group("user")
		{
			user := new(controllers.UserController)
			userGroup.POST("/login", utils.BindAndRespondR(user.Login))

			// router.Use(middlewares.JWTMiddleware())
			// userGroup.Use(middlewares.JWTMiddleware())

			userGroup.GET("/:id", user.Retrieve)
		}

		// 首页接口
		homeGroup := api.Group("home")
		{
			homeGroup.GET("/get-news", controllers.GetNews) // 获取首页文章列表
		}

		// 文章相关路由（暂时不需要认证，方便测试）
		postGroup := api.Group("blog")
		{
			// 搜索文章
			postGroup.GET("/search", controllers.SearchPosts)
			// 根据标签获取文章
			postGroup.GET("/tag", controllers.GetPostsByTag)
			postGroup.GET("/get-tags", controllers.GetTags)
			// 获取单篇文章
			postGroup.GET("/fetch-blog-by-seq",
				utils.BindAndRespondR(controllers.GetPost),
			)

			// 创建文章
			postGroup.POST("", utils.BindAndRespondR(controllers.CreatePost))

			// 更新文章
			postGroup.PUT("/:id", controllers.UpdatePost)
			// 删除文章
			postGroup.DELETE("/:id", controllers.DeletePost)
		}

		thirdpartyGroup := api.Group("thirdparty")
		{
			thirdpartyGroup.GET(
				"/get-weather-by-city",
				utils.BindAndRespondR(controllers.GetWeather),
			)
			thirdpartyGroup.GET(
				"/random-image-url",
				utils.BindAndRespond(controllers.GetRomdomImage),
			)
		}

	}
	return router

}
