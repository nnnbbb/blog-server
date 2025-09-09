package models

import "time"

type Post struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Title      string    `gorm:"size:255;not null" json:"title"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	ImgUrl     string    `gorm:"size:255" json:"img_url"`
	Tags       string    `gorm:"size:500" json:"tags"`
	AdjustTime time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"adjust_time"`

	Timestamps
}
