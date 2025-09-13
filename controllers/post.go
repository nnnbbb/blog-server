package controllers

import (
	"net/http"

	"blog-server/db"
	"blog-server/forms"
	"blog-server/models"
	"blog-server/services"
	"blog-server/utils"
	"blog-server/utils/response"

	"github.com/gin-gonic/gin"
)

type CreatePostBody struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
	ImgUrl  string   `json:"img_url"`
}

// NewsItem 首页最新文章结构
type NewsItem struct {
	ID          uint     `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	ImgUrl      string   `json:"img_url"`
	AdjustTime  string   `json:"adjustTime"`
}

// CreatePost  创建文章
func CreatePost(c *gin.Context, body forms.CreatePostBody) (models.Post, error) {
	tagIDs, err := services.ResolveTagIDs(body.Tags)
	if err != nil {
		return models.Post{}, utils.NewAPIError(http.StatusInternalServerError, "标签处理失败", err)
	}

	post := models.Post{
		Title:   body.Title,
		Content: body.Content,
		ImgUrl:  body.ImgUrl,
		TagIDs:  tagIDs,
	}

	if err := db.DB.Create(&post).Error; err != nil {
		return post, utils.NewAPIError(http.StatusInternalServerError, "文章创建失败", err)
	}

	return post, nil
}

// GetPost 获取单篇文章
func GetPost(c *gin.Context, q forms.FetchPostQuery) (models.Post, error) {
	id := q.Seq

	var post models.Post
	if err := db.DB.First(&post, id).Error; err != nil {
		return post, utils.NewAPIError(http.StatusBadRequest, "文章获取失败", err)
	}

	// 压缩 Markdown 字段
	compressed, err := utils.CompressAndEncode([]byte(post.Content))
	if err != nil {
		return post, utils.NewAPIError(http.StatusInternalServerError, "压缩文章失败", err)
	}

	// 返回时替换原字段
	post.Content = compressed

	return post, nil
}

// GetTags 获取所有标签
func GetTags(c *gin.Context) {
	var tags []models.Tag
	if err := db.DB.Find(&tags).Error; err != nil {
		response.Error(c, http.StatusNotFound, "获取失败")
		return
	}
	var ts []string
	for _, t := range tags {
		ts = append(ts, t.Name)
	}
	response.Ok(c, ts, "获取成功")
}

// UpdatePost 更新文章
func UpdatePost(c *gin.Context) {
	id := c.Param("id")
	var post models.Post
	if err := db.DB.First(&post, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "文章不存在")
		return
	}

	var postReq CreatePostBody
	if err := c.ShouldBindJSON(&postReq); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 调用通用方法获取 tagIDs
	tagIDs, err := services.ResolveTagIDs(postReq.Tags)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "标签处理失败")
		return
	}

	post.Title = postReq.Title
	post.Content = postReq.Content
	post.ImgUrl = postReq.ImgUrl
	post.TagIDs = tagIDs

	if err := db.DB.Save(&post).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "文章更新失败")
		return
	}

	response.Ok(c, post, "文章更新成功")
}

// DeletePost 删除文章
func DeletePost(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Delete(&models.Post{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "文章删除失败")
		return
	}
	response.Ok(c, nil, "文章删除成功")
}

func GetPostsByTag(c *gin.Context) {
	tag := c.Query("tag")
	if tag == "" {
		response.Error(c, http.StatusBadRequest, "标签参数不能为空")
		return
	}

	var posts []models.Post
	if err := db.DB.Where("tags LIKE ?", "%"+tag+"%").Find(&posts).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "获取文章失败")
		return
	}
	response.Ok(c, posts, "获取文章成功")
}

// SearchPosts 搜索文章（支持标题、内容和标签搜索）
func SearchPosts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "查询参数不能为空")
		return
	}

	var posts []models.Post
	searchPattern := "%" + query + "%"

	// 在标题、内容和标签中搜索
	if err := db.DB.Where("title LIKE ? OR content LIKE ? OR tags LIKE ?",
		searchPattern, searchPattern, searchPattern).Find(&posts).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "搜索文章失败")
		return
	}
	response.Ok(c, posts, "搜索文章成功")
}
