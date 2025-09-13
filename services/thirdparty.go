package services

import (
	"io"
	"log"
	"os"
	"time"

	"encoding/json"

	"fmt"
	"net/http"
	"net/url"
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
	amapApiKey := os.Getenv("AMAP_API_KEY")

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

type DmoeResponse struct {
	Code   string `json:"code"`
	ImgURL string `json:"imgurl"`
	Width  string `json:"width"`
	Height string `json:"height"`
}

func getRealImageURL(fullURL string) string {
	u, err := url.Parse(fullURL)
	if err != nil {
		log.Println("解析 URL 失败:", err)
		return fullURL
	}

	// 从查询参数获取 url 值
	values := u.Query()
	realURL := values.Get("url")
	if realURL == "" {
		return fullURL
	}

	return realURL
}

func FetchRandomImage() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	url := "https://www.dmoe.cc/random.php?return=json"
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// 解析 JSON
	var data DmoeResponse
	if err := json.Unmarshal(body, &data); err != nil {
		panic(err)
	}

	return getRealImageURL(data.ImgURL), nil
}
