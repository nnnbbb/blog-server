package controllers

import (
	"net/http"
	"strings"
	"time"

	"blog-server/db"
	"blog-server/forms"
	"blog-server/models"
	"blog-server/services"
	"blog-server/utils"
	"blog-server/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
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

// CreatePost 创建文章
// @Summary 创建文章
// @Description 根据传入的参数创建一篇新文章
// @Tags blog
// @Accept json
// @Produce json
// @Param data body forms.CreatePostBody true "文章信息"
// @Success 200 {object} forms.PostResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /blog [post]
func CreatePost(c *gin.Context, body forms.CreatePostBody) (forms.PostResponse, error) {
	tagIDs, err := services.ResolveTagIDs(body.Tags)
	if err != nil {
		return forms.PostResponse{}, utils.NewAPIError(http.StatusInternalServerError, "标签处理失败", err)
	}

	post := models.Post{
		Title:   body.Title,
		Content: body.Content,
		ImgUrl:  body.ImgUrl,
		TagIDs:  tagIDs,
	}

	if err := db.DB.Create(&post).Error; err != nil {
		return forms.PostResponse{}, utils.NewAPIError(http.StatusInternalServerError, "文章创建失败", err)
	}
	// --- 自动生成 tokens ---
	if err := services.UpdatePostTokens(&post); err != nil {
		// 如果分词失败，可以选择回滚或记录错误
		return forms.PostResponse{}, utils.NewAPIError(http.StatusInternalServerError, "文章分词失败", err)
	}
	// 转换成响应对象返回前端
	resp := forms.PostResponse{
		ID:         post.ID,
		Title:      post.Title,
		ImgUrl:     post.ImgUrl,
		AdjustTime: post.AdjustTime.Format("2006-01-02 15:04:05"),
	}

	return resp, nil
}

// GetPost 获取单篇文章
func GetPost(c *gin.Context, q forms.FetchPostQuery) (forms.PostResponse, error) {
	id := q.Seq

	var post models.Post
	if err := db.DB.First(&post, id).Error; err != nil {
		return forms.PostResponse{}, utils.NewAPIError(http.StatusBadRequest, "文章获取失败", err)
	}

	// 压缩 Markdown 字段
	compressed, err := utils.CompressAndEncode([]byte(post.Content))
	if err != nil {
		return forms.PostResponse{}, utils.NewAPIError(http.StatusInternalServerError, "压缩文章失败", err)
	}
	tagNames, err := services.GetTagNamesByIDs(post.TagIDs)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取标签失败")
		return forms.PostResponse{}, utils.NewAPIError(http.StatusInternalServerError, "获取标签失败", err)
	}
	// 返回时替换原字段
	post.Content = compressed
	resp := forms.PostResponse{
		ID:         post.ID,
		Title:      post.Title,
		ImgUrl:     post.ImgUrl,
		Tags:       tagNames,
		Content:    compressed,
		AdjustTime: post.AdjustTime.Format("2006-01-02 15:04:05"),
	}

	return resp, nil
}

// GetPosts 获取文章分页列表（不返回 content，调整时间格式化）
func GetPosts(c *gin.Context, q forms.FetchPostsQuery) (forms.PostsPage, error) {
	var posts []models.Post
	var total int64

	// 计算总数
	if err := db.DB.Model(&models.Post{}).Count(&total).Error; err != nil {
		return forms.PostsPage{}, utils.NewAPIError(http.StatusInternalServerError, "查询总数失败", err)
	}

	offset := (q.Page - 1) * q.PageSize

	// 查询分页数据
	if err := db.DB.Order("created_at DESC").Limit(q.PageSize).Offset(offset).Find(&posts).Error; err != nil {
		return forms.PostsPage{}, utils.NewAPIError(http.StatusInternalServerError, "查询文章失败", err)
	}

	// 转换为 DTO
	list := make([]forms.PostItem, len(posts))
	for i, p := range posts {
		tagNames, err := services.GetTagNamesByIDs(p.TagIDs)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "获取标签失败")
			return forms.PostsPage{}, utils.NewAPIError(http.StatusInternalServerError, "获取标签失败", err)
		}

		list[i] = forms.PostItem{
			ID:         p.ID,
			Title:      p.Title,
			ImgUrl:     p.ImgUrl,
			Tags:       tagNames,
			AdjustTime: p.AdjustTime.Format("2006年01月02日 15:04"),
		}
	}

	return forms.PostsPage{
		Total: total,
		List:  list,
	}, nil
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

// 搜索文章
func SearchPosts(c *gin.Context, q forms.SearchPosts) ([]forms.PostItem, error) {
	tsQuery := q.Q

	// 结构体用于扫描 SQL 查询结果
	var posts []struct {
		ID         uint
		Title      string
		TagIDs     pq.Int64Array `gorm:"type:integer[]"`
		ImgUrl     string
		Content    string
		AdjustTime time.Time
		Score      float64
	}

	sql := `
		SELECT id, title, content, tag_ids, img_url, adjust_time,
		       ts_rank(tokens, to_tsquery('simple', ?)) AS score
		FROM posts
		WHERE tokens @@ to_tsquery('simple', ?)
		ORDER BY score DESC
		LIMIT 20
	`

	if err := db.GetDB().Raw(sql, tsQuery, tsQuery, tsQuery).Scan(&posts).Error; err != nil {
		return nil, utils.NewAPIError(http.StatusInternalServerError, "查询失败", err)
	}

	// 构造返回数据
	list := make([]forms.PostItem, len(posts))
	for i, p := range posts {
		// 获取标签名
		tagNames, err := services.GetTagNamesByIDs(p.TagIDs)
		if err != nil {
			return nil, utils.NewAPIError(http.StatusInternalServerError, "获取标签失败", err)
		}

		// 生成摘要
		contentSummary := services.GenerateSummary(p.Content, tsQuery, 200, 50)

		list[i] = forms.PostItem{
			ID:         p.ID,
			Title:      p.Title,
			ImgUrl:     p.ImgUrl,
			Tags:       tagNames,
			AdjustTime: p.AdjustTime.Format("2006年01月02日 15:04"),
			Content:    strings.ReplaceAll(contentSummary, "\n", ""),
		}
	}

	return list, nil
}
