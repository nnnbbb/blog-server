package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Post struct {
	ID         uint          `gorm:"primaryKey" json:"id"`
	Title      string        `gorm:"size:255;not null" json:"title"`
	Content    string        `gorm:"type:text;not null" json:"content"`
	ImgUrl     string        `gorm:"size:255" json:"img_url"`
	TagIDs     pq.Int64Array `gorm:"type:integer[]" json:"tag_ids"`
	AdjustTime time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"adjust_time"`
	Tokens     string        `gorm:"type:tsvector" json:"-"`
	IsPrivate  bool          `gorm:"default:false" json:"is_private"`
	IsPinned   bool          `gorm:"default:false" json:"is_pinned"`

	Timestamps
}

func (p *Post) BeforeUpdate(tx *gorm.DB) (err error) {
	p.AdjustTime = time.Now().UTC()
	return nil
}
