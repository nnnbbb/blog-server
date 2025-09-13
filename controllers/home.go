package controllers

import (
	"blog-server/db"
	"blog-server/models"
	"blog-server/services"
	"blog-server/utils/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetNews 获取首页文章列表（返回简要信息）
func GetNews(c *gin.Context) {
	var posts []models.Post

	if err := db.DB.Order("created_at DESC").Limit(5).Find(&posts).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "获取失败")
		return
	}

	var newsItems []NewsItem
	for _, post := range posts {
		description := post.Content
		runes := []rune(post.Content)
		if len(runes) > 100 {
			description = string(runes[:100]) + "..."
		}

		tagNames, err := services.GetTagNamesByIDs(post.TagIDs)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "获取标签失败")
			return
		}
		newsItem := NewsItem{
			ID:          post.ID,
			Title:       post.Title,
			Description: description,
			Tags:        tagNames,
			AdjustTime:  post.AdjustTime.Format("2006-01-02 15:04"),
			ImgUrl:      post.ImgUrl,
		}
		newsItems = append(newsItems, newsItem)
	}

	response.Ok(c, newsItems)
}
