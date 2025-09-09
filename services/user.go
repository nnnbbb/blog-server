package services

import (
	"errors"
	"time"

	"blog-server/db"
	"blog-server/forms"
	"blog-server/models"

	"github.com/google/uuid"
)

// 注册用户
func Signup(userPayload forms.UserSignup) (*models.User, error) {
	id := uuid.New().String()
	user := models.User{
		ID:       id,
		Name:     userPayload.Name,
		BirthDay: userPayload.BirthDay,
		Gender:   userPayload.Gender,
		PhotoURL: userPayload.PhotoURL,
		Time:     time.Now().UnixNano(),
		Active:   true,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		return nil, errors.New("failed to save user to database")
	}

	return &user, nil
}

// 根据 ID 查询用户
func GetUserByID(id string) (*models.User, error) {
	var user models.User
	if err := db.DB.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
