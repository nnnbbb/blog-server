package services

import (
	"blog-server/db"
	"blog-server/models"
	"blog-server/utils"
	"fmt"
	"log"
	"strings"

	"github.com/huichen/sego"
	"gorm.io/gorm"
)

var Segmenter sego.Segmenter

// 初始化分词器
func init() {
	Segmenter.LoadDictionary("data/dictionary.txt") // 放置你的 sego 词典
	utils.Log("Sego segmenter initialized.")
}

// UpdateAllExistingPostsTokens 对数据库里所有文章生成 tokens
func UpdateAllExistingPostsTokens(batchSize int) error {
	var lastID uint = 0

	for {
		var posts []models.Post
		err := db.GetDB().
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(batchSize).
			Find(&posts).Error
		if err != nil {
			return err
		}

		if len(posts) == 0 {
			break // 已处理完所有文章
		}

		for _, post := range posts {
			if err := UpdatePostTokens(&post); err != nil {
				log.Printf("failed to update tokens for post ID %d: %v", post.ID, err)
			}
			lastID = post.ID
		}
	}

	utils.Log("All existing posts tokens updated successfully.")
	return nil
}

// 分词函数
func SegmentText(text string) []string {
	segments := Segmenter.Segment([]byte(text))
	words := sego.SegmentsToSlice(segments, true)
	var result []string
	for _, w := range words {
		if len(w) <= 1 {
			continue
		}
		result = append(result, w)
	}
	return result
}

// 转换为 PostgreSQL tsvector 可用字符串
func ToTSVector(words []string) string {
	return strings.Join(words, " ")
}

// 更新单篇文章的 tokens
func UpdatePostTokens(post *models.Post) error {
	titleWords := SegmentText(post.Title)
	contentWords := SegmentText(post.Content)

	titleVector := ToTSVector(titleWords)
	contentVector := ToTSVector(contentWords)

	// setweight 给 title 权重 A，content 权重 B
	tsvectorSQL := fmt.Sprintf(
		"setweight(to_tsvector('simple', '%s'), 'A') || setweight(to_tsvector('simple', '%s'), 'B')",
		titleVector,
		contentVector,
	)

	return db.GetDB().Model(post).Update("tokens", gorm.Expr(tsvectorSQL)).Error
}

// 批量更新所有文章 tokens
func UpdateAllPostTokens() error {
	var posts []models.Post
	if err := db.GetDB().Find(&posts).Error; err != nil {
		return err
	}

	for _, post := range posts {
		if err := UpdatePostTokens(&post); err != nil {
			return err
		}
	}
	return nil
}

// GenerateSummary 生成内容摘要
// content: 原文
// keyword: 搜索关键词，可以为空
// maxLen: 摘要最大长度
// contextLen: 关键词上下文长度
func GenerateSummary(content, keyword string, maxLen, contextLen int) string {
	if content == "" {
		return ""
	}

	// 把 content 转成 rune 切片，避免中文截断乱码
	runes := []rune(content)

	// 如果有关键词
	if keyword != "" {
		keywordRunes := []rune(keyword)
		idx := strings.Index(content, keyword) // 注意按 byte 索引
		if idx != -1 {
			// 把 byte 索引转换成 rune 索引
			runeIdx := len([]rune(content[:idx]))
			start := runeIdx - contextLen
			if start < 0 {
				start = 0
			}
			end := runeIdx + len(keywordRunes) + contextLen
			if end > len(runes) {
				end = len(runes)
			}
			summary := string(runes[start:end])
			if start > 0 {
				summary = "..." + summary
			}
			if end < len(runes) {
				summary = summary + "..."
			}
			return summary
		}
	}

	// 没找到关键词或者没有关键词，截取前 maxLen 个字符
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}

	return content
}
