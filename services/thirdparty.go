package services

import (
	"blog-server/config"

	"encoding/json"

	"fmt"
	"net/http"
)

// 天气
// https://restapi.amap.com/v3/weather/weatherInfo?city=350100&key=apikey

type WeatherResponse struct {
	Status   string `json:"status"`
	Count    string `json:"count"`
	Info     string `json:"info"`
	InfoCode string `json:"infocode"`
	Lives    []Live `json:"lives"`
}

type Live struct {
	Province      string `json:"province"`
	City          string `json:"city"`
	Adcode        string `json:"adcode"`
	Weather       string `json:"weather"`
	Temperature   string `json:"temperature"`
	WindDirection string `json:"winddirection"`
	WindPower     string `json:"windpower"`
	Humidity      string `json:"humidity"`
	ReportTime    string `json:"reporttime"`
}

func GetWeather(city string, cityCode string) (*Live, error) {
	cfg := config.GetConfig()
	amapApiKey := cfg.GetString("server.amapApiKey")

	url := fmt.Sprintf("https://restapi.amap.com/v3/weather/weatherInfo?city=%s&key=%s", cityCode, amapApiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求天气数据失败: %w", err)
	}
	defer resp.Body.Close()

	var weather WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return nil, fmt.Errorf("解析天气数据失败: %w", err)
	}

	if len(weather.Lives) == 0 {
		return nil, fmt.Errorf("未获取到天气数据")
	}

	return &weather.Lives[0], nil
}
