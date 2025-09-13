package forms

type GetWeatherQuery struct {
	City     string `form:"city" binding:"required"`
	CityCode string `form:"cityCode" binding:"required"`
}
