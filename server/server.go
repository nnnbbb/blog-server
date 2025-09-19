package server

import (
	"fmt"
	"os"

	_ "blog-server/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Init() {
	r := NewRouter()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	s := os.Getenv("address")
	fmt.Printf("Starting server on %s\n", s)
	r.Run(s)
}
