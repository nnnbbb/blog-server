package controllers

import (
	"blog-server/db"
	"blog-server/forms"
	"blog-server/models"
	"blog-server/services"
	"blog-server/utils/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetNews 获取首页文章列表（返回简要信息）
func GetNews(c *gin.Context) {
	var posts []models.Post
	query := db.DB.Model(&models.Post{})

	// 判断登录状态
	_, loggedIn := c.Get("username")
	if !loggedIn {
		query = query.Where("is_private = ?", false)
	}

	if err := query.Order("is_pinned DESC").Order("created_at DESC").Limit(5).Find(&posts).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "获取失败")
		return
	}

	var newsItems []forms.NewsItem
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

		newsItems = append(newsItems, forms.NewsItem{
			ID:          post.ID,
			Title:       post.Title,
			Description: description,
			Tags:        tagNames,
			AdjustTime:  post.AdjustTime.Format("2006-01-02 15:04"),
			ImgUrl:      post.ImgUrl,
		})
	}

	response.Ok(c, newsItems)
}
