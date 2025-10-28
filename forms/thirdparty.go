package forms

type GetWeatherQuery struct {
	City     string `form:"city" binding:"required"`
	CityCode string `form:"cityCode" binding:"required"`
}

type CrawlPostBody struct {
	Url   string `json:"url" binding:"required,url"`
	Force *bool  `json:"force"` // 可选参数，默认 false
}
