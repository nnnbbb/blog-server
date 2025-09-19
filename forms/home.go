package forms

// NewsItem 首页最新文章结构
type NewsItem struct {
	ID          uint     `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	ImgUrl      string   `json:"img_url"`
	AdjustTime  string   `json:"adjustTime"`
}
