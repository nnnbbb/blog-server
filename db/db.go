package db

import (
	"blog-server/models"
	"blog-server/utils"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	host := "localhost"

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		host, user, password, dbname, port,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	// 自动迁移表
	if err := DB.AutoMigrate(
		&models.Post{},
		&models.Tag{},
	); err != nil {
		log.Fatal("failed to migrate database:", err)
	}

	// 调用 EnsureGinIndex 创建扩展和索引
	EnsureGinIndex()

	utils.Log("Database initialized.")
}

// EnsureGinIndex 确保 pg_trgm 扩展和 GIN 索引存在
func EnsureGinIndex() {
	sqls := []string{
		"CREATE EXTENSION IF NOT EXISTS pg_trgm;",
		//  posts 表 tokens 字段
		"CREATE INDEX IF NOT EXISTS idx_posts_tokens ON posts USING GIN(tokens);",
	}

	for _, sql := range sqls {
		if err := DB.Exec(sql).Error; err != nil {
			log.Fatalf("failed to execute %q: %v", sql, err)
		}
	}

	utils.Log("PG GIN index and extension ensured.")
}

func GetDB() *gorm.DB {
	return DB
}
