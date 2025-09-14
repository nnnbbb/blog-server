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

type PostResponse struct {
	ID         uint     `json:"id"`
	Title      string   `json:"title"`
	ImgUrl     string   `json:"imgUrl"`
	Content    string   `json:"content"`
	AdjustTime string   `json:"adjustTime"`
	Tags       []string `json:"tags"`
}

type FetchPostsQuery struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

type PostItem struct {
	ID         uint     `json:"id"`
	Title      string   `json:"title"`
	ImgUrl     string   `json:"imgUrl"`
	Tags       []string `json:"tags"`
	AdjustTime string   `json:"adjustTime"` // 格式化后的时间
}

type PostsPage struct {
	Total int64      `json:"total"`
	List  []PostItem `json:"list"`
}
