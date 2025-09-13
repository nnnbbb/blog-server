package forms

type FetchPostQuery struct {
	Seq int `form:"seq" binding:"required"`
}

type CreatePostBody struct {
	Title   string   `json:"title" binding:"required"`
	Content string   `json:"content" binding:"required"`
	Tags    []string `json:"tags" binding:"required"` // JSON 数组
	ImgUrl  string   `json:"imgUrl" binding:"required"`
}
