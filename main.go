package main

import (
	"flag"
	"fmt"
	"os"

	"blog-server/config"
	"blog-server/db"
	"blog-server/server"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	environment := flag.String("e", "development", "")
	flag.Usage = func() {
		fmt.Println("Usage: server -e {mode}")
		os.Exit(1)
	}
	flag.Parse()
	config.Init(*environment)

	// 读取 .env 文件
	err := godotenv.Load()
	if err != nil {
		log.Fatal("加载 .env 文件失败: ", err)
	}

	db.InitDB()
	server.Init()
}
