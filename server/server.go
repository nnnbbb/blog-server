package server

import (
	"fmt"

	"blog-server/config"
)

func Init() {
	config := config.GetConfig()
	r := NewRouter()
	s := config.GetString("server.address")
	fmt.Printf("Starting server on %s\n", s)
	r.Run(s)
}
