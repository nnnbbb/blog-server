package controllers

import (
	"blog-server/services"
	"blog-server/utils"
	"blog-server/utils/hefeng"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	weatherCache *utils.Cache[gin.H]
	ipCityCache  *utils.Cache[*IpCityInfo]
)

func init() {
	// 缓存 100 个城市，缓存 60 分钟
	weatherCache, _ = utils.NewCache[gin.H](100, 60*time.Minute)
	// 缓存 1000 个 IP 城市映射，缓存 7*24 小时
	ipCityCache, _ = utils.NewCache[*IpCityInfo](1000, 7*24*time.Hour)
}

// IpCityInfo 表示 IP 对应的城市信息
type IpCityInfo struct {
	Addr     string `json:"addr"`
	City     string `json:"city"`
	CityCode string `json:"cityCode"` // 根据实际返回字段修改
}

func GBKToUTF8(s string) ([]byte, error) {
	reader := transform.NewReader(strings.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// GetCityByIP 根据客户端 IP 获取城市信息
func GetCityByIP(r *http.Request) (*IpCityInfo, error) {
	// 获取客户端 IP
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP, _, _ = net.SplitHostPort(r.RemoteAddr)
	} else {
		// X-Forwarded-For 可能有多个 IP，取第一个
		clientIP = strings.Split(clientIP, ",")[0]
	}

	if val, ok := ipCityCache.Get(clientIP); ok {
		return val, nil
	}

	// 调用 pconline 接口
	url := fmt.Sprintf("https://whois.pconline.com.cn/ipJson.jsp?ip=%s&json=true", clientIP)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0") // 模拟浏览器

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	utf8Body, err := GBKToUTF8(string(body))
	if err != nil {
		return nil, err
	}

	var ipInfo IpCityInfo
	if err := json.Unmarshal(utf8Body, &ipInfo); err != nil {
		return nil, err
	}

	// 写入缓存
	ipCityCache.Set(clientIP, &ipInfo)

	return &ipInfo, nil
}

// 获取天气
func GetWeather(c *gin.Context) (gin.H, error) {

	cityInfo, err := GetCityByIP(c.Request)
	if err != nil {
		return nil, utils.NewAPIError(http.StatusInternalServerError, "获取城市信息错误", err)
	}

	city := cityInfo.City
	cityCode := cityInfo.CityCode

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
