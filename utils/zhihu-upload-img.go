package utils

import (
	"bytes"
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
func UploadImageLocal(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建 form 文件字段
	part, err := writer.CreateFormFile("picture", filepath.Base(filePath))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	// 添加额外字段
	_ = writer.WriteField("source", "answer")
	writer.Close()

	req, err := http.NewRequest("POST", "https://www.zhihu.com/api/v4/uploaded_images", body)
	if err != nil {
		return err
	}

	req.Header = getHeaders()
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respData, _ := io.ReadAll(resp.Body)
	fmt.Println("上传本地图片:", resp.StatusCode, string(respData))
	return nil
}

// 上传网络图片
func UploadImageFromURL(imgURL string) error {
	resp, err := http.Get(imgURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("下载图片失败: %d", resp.StatusCode)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 模拟文件上传
	part, err := writer.CreateFormFile("picture", "remote.jpg")
	if err != nil {
		return err
	}
	_, err = io.Copy(part, resp.Body)
	if err != nil {
		return err
	}

	_ = writer.WriteField("source", "answer")
	writer.Close()

	req, err := http.NewRequest("POST", "https://www.zhihu.com/api/v4/uploaded_images", body)
	if err != nil {
		return err
	}

	req.Header = getHeaders()
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp2, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp2.Body.Close()

	respData, _ := io.ReadAll(resp2.Body)
	fmt.Println("上传网络图片:", resp2.StatusCode, string(respData))
	return nil
}
