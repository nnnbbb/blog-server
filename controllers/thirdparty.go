package controllers

import (
	"blog-server/forms"
	"blog-server/services"
	"blog-server/utils"
	"blog-server/utils/hefeng"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var weatherCache *utils.Cache[gin.H]

func init() {
	// 缓存 100 个城市，缓存 60 分钟
	weatherCache, _ = utils.NewCache[gin.H](100, 60*time.Minute)
}

// 获取天气
func GetWeather(c *gin.Context, q forms.GetWeatherQuery) (gin.H, error) {
	city := q.City
	cityCode := q.CityCode

	if cityCode == "" {
		return nil, utils.NewAPIError(http.StatusBadRequest, "参数 cityCode 不能为空")
	}

	cacheKey := city + ":" + cityCode
	if val, ok := weatherCache.Get(cacheKey); ok {
		return val, nil
	}

	// 调用服务
	live, err := services.GetWeather(city, cityCode)
	if err != nil {
		return nil, utils.NewAPIError(http.StatusInternalServerError, "获取天气失败", err)
	}
	aqi, err := hefeng.GetAQI(city)
	if err != nil {
		return nil, utils.NewAPIError(http.StatusInternalServerError, "获取空气质量失败", err)
	}

	cityNameCn := strings.ReplaceAll(city, "市", "")
	cityNameEn := utils.GetPinYin(cityNameCn)

	data := gin.H{
		"cityNameEn":  cityNameEn,
		"cityNameCn":  cityNameCn,
		"weather":     live.Weather,
		"temperature": live.Temperature + "°",
		"humidity":    fmt.Sprintf("%s%%", live.Humidity),
		"wind":        fmt.Sprintf("%s风 %s级", live.WindDirection, live.WindPower),
		"time":        live.ReportTime,
		"aqi":         aqi,
	}

	// 放入缓存
	weatherCache.Set(cacheKey, data)

	return data, nil
}

func GetRomdomImage(c *gin.Context) (string, error) {
	imgURL, err := services.FetchRandomImage()
	if err != nil {
		return "", utils.NewAPIError(http.StatusInternalServerError, "获取图片失败", err)
	}
	return imgURL, nil
}
