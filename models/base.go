package models

import (
	"time"

	"gorm.io/gorm"
)

// 通用时间字段
type Timestamps struct {
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
