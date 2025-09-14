package server

import (
	"fmt"

	"blog-server/config"
	_ "blog-server/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Init() {
	config := config.GetConfig()
	r := NewRouter()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	s := config.GetString("server.address")
	fmt.Printf("Starting server on %s\n", s)
	r.Run(s)
}
