package controllers

import (
	"blog-server/services"
	"blog-server/utils"
	"blog-server/utils/hefeng"
	"blog-server/utils/response"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetWeather(c *gin.Context) {
	city := c.Query("city")
	cityCode := c.Query("cityCode")
	if cityCode == "" {
		response.Error(c, http.StatusBadRequest, "参数 cityCode 不能为空")
		return
	}

	live, err := services.GetWeather(city, cityCode)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	aqi, err := hefeng.GetAQI(city)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	cityNameCn := strings.ReplaceAll(city, "市", "")
	cityNameEn := utils.GetPinYin(cityNameCn)

	response.Ok(c, gin.H{
		"cityNameEn": cityNameEn,
		"cityNameCn": cityNameCn,

		"weather":     live.Weather,
		"temperature": live.Temperature,
		"humidity":    fmt.Sprintf("%s%%", live.Humidity),
		"wind":        fmt.Sprintf("%s风 %s级", live.WindDirection, live.WindPower),
		"time":        live.ReportTime,
		"aqi":         aqi,
	})
}
