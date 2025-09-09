package models

type User struct {
	ID       string `gorm:"type:uuid;primaryKey" json:"user_id"`
	Name     string `gorm:"size:100;not null" json:"name"`
	BirthDay string `gorm:"size:20" json:"birthday"`
	Gender   string `gorm:"size:10" json:"gender"`
	PhotoURL string `gorm:"size:255" json:"photo_url"`
	Time     int64  `json:"current_time"`
	Active   bool   `json:"active"`

	Timestamps
}

func (User) TableName() string {
	return "TableUsers" // 如果你要固定表名
}
