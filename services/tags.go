package services

import (
	"blog-server/db"
	"blog-server/models"
)

// ResolveTagIDs 根据标签名数组查询/创建标签，并返回对应的 ID 数组
func ResolveTagIDs(tagNames []string) ([]int64, error) {
	var tagIDs []int64

	for _, tagName := range tagNames {
		var tag models.Tag
		if err := db.DB.Where("name = ?", tagName).First(&tag).Error; err != nil {
			// 标签不存在就新建
			tag = models.Tag{Name: tagName}
			if err := db.DB.Create(&tag).Error; err != nil {
				return nil, err
			}
		}
		tagIDs = append(tagIDs, int64(tag.ID))
	}

	return tagIDs, nil
}

// GetTagNamesByIDs 根据 tagID 列表返回对应的 tagName
func GetTagNamesByIDs(tagIDs []int64) ([]string, error) {
	if len(tagIDs) == 0 {
		return []string{}, nil
	}

	var tags []models.Tag
	if err := db.DB.Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
		return nil, err
	}

	var names []string
	for _, tag := range tags {
		names = append(names, tag.Name)
	}

	return names, nil
}
