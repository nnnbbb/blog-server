package hefeng

import (
	"blog-server/utils"
	"compress/gzip"

	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CityLookupResponse 对应接口返回的 JSON 结构
type CityLookupResponse struct {
	Code     string     `json:"code"`
	Location []Location `json:"location"`
	Refer    Refer      `json:"refer"`
}

type Location struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	Lon       string `json:"lon"` // 经度
	Lat       string `json:"lat"` // 纬度
	Adm2      string `json:"adm2"`
	Adm1      string `json:"adm1"`
	Country   string `json:"country"`
	Tz        string `json:"tz"`
	UTCOffset string `json:"utcOffset"`
	IsDst     string `json:"isDst"`
	Type      string `json:"type"`
	Rank      string `json:"rank"`
	FxLink    string `json:"fxLink"`
}

type Refer struct {
	Sources []string `json:"sources"`
	License []string `json:"license"`
}

// DoRequest 封装 GET 请求，支持 gzip 解压和 JSON 解析
func DoRequest(url, token string, result interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// Bearer 前缀处理
	if token != "" && !strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = "Bearer " + token
	}
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "go-qweather-client/1.0")

	// 使用 context 带超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 非 2xx 返回错误
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API 返回 HTTP %d: %s", resp.StatusCode, string(body))
	}

	var reader io.ReadCloser = resp.Body
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Encoding")), "gzip") {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("gzip 解压失败: %w", err)
		}
		defer gz.Close()
		reader = gz
	}

	if err := json.NewDecoder(reader).Decode(result); err != nil {
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	return nil
}

func LookupCity(apiHost, token, location string) (*CityLookupResponse, error) {
	if apiHost == "" || location == "" {
		return nil, fmt.Errorf("apiHost 和 location 不能为空")
	}

	url := fmt.Sprintf("%s/geo/v2/city/lookup?location=%s", apiHost, url.QueryEscape(location))
	var resp CityLookupResponse
	if err := DoRequest(url, token, &resp); err != nil {
		return nil, err
	}

	if resp.Code != "200" {
		return &resp, fmt.Errorf("api 返回 code=%s", resp.Code)
	}

	return &resp, nil
}

// AirIndex 表示空气质量指标
type AirIndex struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	AQI      any    `json:"aqi"`
	Category string `json:"category"`
}
type AirQualityResponse struct {
	Indexes []AirIndex `json:"indexes"` // 复用 AirIndex

	Pollutants []struct {
		Code          string `json:"code"`
		Name          string `json:"name"`
		Concentration struct {
			Value float64 `json:"value"`
			Unit  string  `json:"unit"`
		} `json:"concentration"`
	} `json:"pollutants"`
	Stations []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"stations"`
}

func GetAirQuality(apiHost, token, lat, lon string) (*AirQualityResponse, error) {
	url := fmt.Sprintf("%s/airquality/v1/current/%s/%s", apiHost, lat, lon)
	var res AirQualityResponse
	if err := DoRequest(url, token, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type CityAirQuality struct {
	City       string
	Lat        string
	Lon        string
	AirQuality *AirQualityResponse
}

func GetCityAirQuality(cityName, apiHost string) (*CityAirQuality, error) {
	// 生成 JWT token
	token, err := GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("生成 token 失败: %w", err)
	}

	city := strings.TrimSuffix(cityName, "市")
	location := utils.GetPinYin(city)

	cityResp, err := LookupCity(apiHost, token, location)
	if err != nil {
		return nil, fmt.Errorf("城市查询失败: %w", err)
	}
	if len(cityResp.Location) == 0 {
		return nil, fmt.Errorf("未找到城市: %s", cityName)
	}

	loc := cityResp.Location[0]
	aqiResp, err := GetAirQuality(apiHost, token, loc.Lat, loc.Lon)
	if err != nil {
		return nil, fmt.Errorf("空气质量查询失败: %w", err)
	}

	return &CityAirQuality{
		City:       cityName,
		Lat:        loc.Lat,
		Lon:        loc.Lon,
		AirQuality: aqiResp,
	}, nil
}

func GetAQI(city string) (string, error) {
	apiHost := "https://m263yw33ef.re.qweatherapi.com"

	result, err := GetCityAirQuality(city, apiHost)
	if err != nil {
		fmt.Println("获取城市空气质量失败:", err)
		return "", err
	}

	// fmt.Printf("城市: %s, 经纬度: %s,%s\n", result.City, result.Lat, result.Lon)
	// for _, idx := range result.AirQuality.Indexes {
	// 	fmt.Printf("%s AQI: %v, 分类: %s\n", idx.Name, idx.AQI, idx.Category)
	// }
	AirQuality := &result.AirQuality.Indexes[0]
	aqi := fmt.Sprintf("%v %s", AirQuality.AQI, AirQuality.Category)

	return aqi, nil
}
