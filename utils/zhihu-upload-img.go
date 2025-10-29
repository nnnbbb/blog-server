package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

var cookie string
var xZst81 string

func init() {
	cookie = os.Getenv("ZHIHU_COOKIE")
	xZst81 = os.Getenv("ZHIHU_X_ZST81")
}

// 公共请求头
func getHeaders() http.Header {

	headers := http.Header{}
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) "+
		"Chrome/139.0.0.0 Safari/537.36")
	headers.Set("Cookie", cookie)
	headers.Set("x-requested-with", "fetch")
	headers.Set("x-zst-81", xZst81)
	headers.Set("Origin", "https://www.zhihu.com")
	headers.Set("Referer", "https://www.zhihu.com/")
	return headers
}

// 上传本地图片
type ZhihuUploadResponse struct {
	Src            string `json:"src"`
	AnimationCover string `json:"animation_cover_src"`
	Hash           string `json:"hash"`
	Watermark      string `json:"watermark"`
	WatermarkSrc   string `json:"watermark_src"`
	OriginalHash   string `json:"original_hash"`
	OriginalSrc    string `json:"original_src"`
	RawHeight      int    `json:"data-rawheight"`
	RawWidth       int    `json:"data-rawwidth"`
	WatermarkHash  string `json:"watermark_hash"`
	Class          string `json:"class"`
}

// 核心上传函数，传入 Reader 和文件名
func uploadImageFromReader(reader io.Reader, filename string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("picture", filename)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(part, reader)
	if err != nil {
		return "", err
	}

	_ = writer.WriteField("source", "answer")
	writer.Close()

	req, err := http.NewRequest("POST", "https://www.zhihu.com/api/v4/uploaded_images", body)
	if err != nil {
		return "", err
	}

	req.Header = getHeaders()
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Println("上传图片:", resp.StatusCode, string(respData))

	var result ZhihuUploadResponse
	if err := json.Unmarshal(respData, &result); err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://pic1.zhimg.com/80/%s_1440w.png", result.OriginalHash)
	return url, nil
}

// 上传本地图片
func UploadImageLocal(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return uploadImageFromReader(file, filepath.Base(filePath))
}

// 上传网络图片
func UploadImageFromURL(imgURL string) (string, error) {
	resp, err := http.Get(imgURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("下载图片失败: %d", resp.StatusCode)
	}

	return uploadImageFromReader(resp.Body, "remote.jpg")
}
