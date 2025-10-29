package controllers

import (
	"blog-server/db"
	"blog-server/forms"
	"blog-server/models"
	"blog-server/services"
	"blog-server/utils"
	"blog-server/utils/hefeng"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
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

// CrawlAndCreatePost 爬取知乎回答并创建文章
// @Summary 爬取知乎回答并创建文章
// @Description 通过 URL 爬取知乎回答，自动转换为文章保存
// @Tags thirdparty
// @Accept json
// @Produce json
// @Param data body forms.CrawlPostBody true "知乎回答 URL"
// @Success 200 {object} forms.PostResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /thirdparty/crawl [post]
func CrawlAndCreatePost(c *gin.Context, body forms.CrawlPostBody) (forms.PostResponse, error) {
	// 调用 Node.js 爬虫脚本
	cmd := exec.Command("node", "./scripts/crawler/zhihu-answer.js", body.Url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return forms.PostResponse{}, utils.NewAPIError(
			http.StatusInternalServerError,
			"爬虫执行失败: "+string(output),
			err,
		)
	}

	// 从输出中提取文件路径
	outputStr := string(output)
	var filepath string

	// 解析输出，查找 "Markdown 文件已保存:" 行
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Markdown 文件已保存:") {
			// 提取文件路径
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				filepath = strings.TrimSpace(strings.Join(parts[1:], ":"))
				break
			}
		}
	}

	if filepath == "" {
		return forms.PostResponse{}, utils.NewAPIError(
			http.StatusInternalServerError,
			"无法从爬虫输出中获取文件路径",
		)
	}

	// 读取 Markdown 文件
	content, err := os.ReadFile(filepath)
	if err != nil {
		return forms.PostResponse{}, utils.NewAPIError(
			http.StatusInternalServerError,
			"读取文件失败",
			err,
		)
	}

	// 从文件名中提取标题（去掉路径和 .md 扩展名）
	filename := strings.TrimSuffix(strings.Replace(filepath, "\\", "/", -1), ".md")
	parts := strings.Split(filename, "/")
	title := parts[len(parts)-1]
	imgUrl := "https://picx.zhimg.com/70/v2-14a2f0a03e53c29005aef575285dac0f_1440w.avis"

	// 处理标签：固定使用 "文章" 和 "转载"
	tagIDs, err := services.ResolveTagIDs([]string{"文章", "转载"})
	if err != nil {
		return forms.PostResponse{}, utils.NewAPIError(
			http.StatusInternalServerError,
			"标签处理失败",
			err,
		)
	}

	// 检查文章是否已存在
	var existingPost models.Post

	// 判断 force 参数，默认为 false
	forceCreate := false
	if body.Force != nil {
		forceCreate = *body.Force
	}

	// 如果文章已存在
	if err = db.DB.Where("title = ?", title).First(&existingPost).Error; err == nil {
		// 如果不强制更新，直接返回已存在的文章
		if !forceCreate {
			tagNames, err := services.GetTagNamesByIDs(existingPost.TagIDs)
			if err != nil {
				return forms.PostResponse{}, utils.NewAPIError(
					http.StatusInternalServerError,
					"获取标签名称失败",
					err,
				)
			}

			return forms.PostResponse{
				ID:         existingPost.ID,
				Title:      existingPost.Title,
				ImgUrl:     existingPost.ImgUrl,
				AdjustTime: existingPost.AdjustTime.Format("2006-01-02 15:04:05"),
				Tags:       tagNames,
			}, nil
		}

		// force=true，更新文章内容
		existingPost.Content = string(content)
		existingPost.TagIDs = tagIDs
		existingPost.ImgUrl = imgUrl

		if err := db.DB.Save(&existingPost).Error; err != nil {
			return forms.PostResponse{}, utils.NewAPIError(
				http.StatusInternalServerError,
				"文章更新失败",
				err,
			)
		}

		// 更新搜索 tokens
		if err := services.UpdatePostTokens(&existingPost); err != nil {
			return forms.PostResponse{}, utils.NewAPIError(
				http.StatusInternalServerError,
				"文章分词失败",
				err,
			)
		}

		// 获取标签名称
		tagNames, err := services.GetTagNamesByIDs(existingPost.TagIDs)
		if err != nil {
			return forms.PostResponse{}, utils.NewAPIError(
				http.StatusInternalServerError,
				"获取标签名称失败",
				err,
			)
		}

		// 返回更新后的文章信息
		return forms.PostResponse{
			ID:         existingPost.ID,
			Title:      existingPost.Title,
			ImgUrl:     existingPost.ImgUrl,
			AdjustTime: existingPost.AdjustTime.Format("2006-01-02 15:04:05"),
			Tags:       tagNames,
		}, nil
	}

	// 文章不存在，创建新文章
	post := models.Post{
		Title:   title,
		Content: string(content),
		ImgUrl:  imgUrl,
		TagIDs:  tagIDs,
	}

	if err := db.DB.Create(&post).Error; err != nil {
		return forms.PostResponse{}, utils.NewAPIError(
			http.StatusInternalServerError,
			"文章创建失败",
			err,
		)
	}

	// 生成搜索 tokens
	if err := services.UpdatePostTokens(&post); err != nil {
		return forms.PostResponse{}, utils.NewAPIError(
			http.StatusInternalServerError,
			"文章分词失败",
			err,
		)
	}

	// 获取标签名称
	tagNames, err := services.GetTagNamesByIDs(tagIDs)
	if err != nil {
		return forms.PostResponse{}, utils.NewAPIError(
			http.StatusInternalServerError,
			"获取标签名称失败",
			err,
		)
	}
	// 返回响应
	resp := forms.PostResponse{
		ID:         post.ID,
		Title:      post.Title,
		ImgUrl:     post.ImgUrl,
		AdjustTime: post.AdjustTime.Format("2006-01-02 15:04:05"),
		Tags:       tagNames,
	}

	return resp, nil
}
