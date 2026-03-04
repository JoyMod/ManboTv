// internal/service/suggestion_service.go
package service

import (
	"context"
	"strings"
	"sync"
	"unicode"

	"go.uber.org/zap"
)

// SuggestionService 搜索建议服务
type SuggestionService interface {
	GetSuggestions(ctx context.Context, query string) ([]string, error)
	AddKeyword(ctx context.Context, keyword string)
}

// suggestionService 实现
type suggestionService struct {
	keywords []string
	logger   *zap.Logger
	mu       sync.RWMutex
}

// NewSuggestionService 创建搜索建议服务
func NewSuggestionService(logger *zap.Logger) SuggestionService {
	// 预置热门关键词
	defaultKeywords := []string{
		"流浪地球",
		"三体",
		"狂飙",
		"漫长的季节",
		"隐秘的角落",
		"沉默的真相",
		"庆余年",
		"赘婿",
		"人世间",
		"开端",
	}

	return &suggestionService{
		keywords: defaultKeywords,
		logger:   logger,
	}
}

// GetSuggestions 获取搜索建议
func (s *suggestionService) GetSuggestions(ctx context.Context, query string) ([]string, error) {
	if query == "" {
		return nil, nil
	}

	query = strings.ToLower(query)
	var suggestions []string

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, keyword := range s.keywords {
		if strings.Contains(strings.ToLower(keyword), query) {
			suggestions = append(suggestions, keyword)
			if len(suggestions) >= 10 {
				break
			}
		}
	}

	return suggestions, nil
}

// AddKeyword 添加关键词
func (s *suggestionService) AddKeyword(ctx context.Context, keyword string) {
	if keyword == "" {
		return
	}

	// 清理关键词
	keyword = s.cleanKeyword(keyword)
	if keyword == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否已存在
	for _, k := range s.keywords {
		if k == keyword {
			return
		}
	}

	s.keywords = append(s.keywords, keyword)
	s.logger.Debug("添加搜索关键词", zap.String("keyword", keyword))
}

// cleanKeyword 清理关键词
func (s *suggestionService) cleanKeyword(keyword string) string {
	// 移除首尾空白
	keyword = strings.TrimSpace(keyword)

	// 限制长度
	if len(keyword) > 50 {
		keyword = keyword[:50]
	}

	// 检查是否全是空白或特殊字符
	valid := false
	for _, r := range keyword {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.Is(unicode.Han, r) {
			valid = true
			break
		}
	}

	if !valid {
		return ""
	}

	return keyword
}
