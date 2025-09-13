package models

type Tag struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Timestamps
}
