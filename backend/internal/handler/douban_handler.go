// internal/handler/douban_handler.go
package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// DoubanHandler 豆瓣处理器
type DoubanHandler struct {
	logger     *zap.Logger
	httpClient *http.Client
}

// NewDoubanHandler 创建豆瓣处理器
func NewDoubanHandler(logger *zap.Logger) *DoubanHandler {
	return &DoubanHandler{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// DoubanItem 豆瓣条目
type DoubanItem struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Poster string `json:"poster"`
	Rate   string `json:"rate"`
	Year   string `json:"year"`
}

// DoubanResult 豆瓣结果
type DoubanResult struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	List    []DoubanItem `json:"list"`
}

// DoubanApiResponse 豆瓣 API 响应
type DoubanApiResponse struct {
	Subjects []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Cover string `json:"cover"`
		Rate  string `json:"rate"`
	} `json:"subjects"`
}

// Search 豆瓣搜索
// GET /api/v1/douban?type=movie&tag=热门&pageSize=16&pageStart=0
func (h *DoubanHandler) Search(c *gin.Context) {
	contentType := c.Query("type")
	tag := c.Query("tag")
	pageSize := 16
	pageStart := 0

	if ps := c.Query("pageSize"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 && n <= 100 {
			pageSize = n
		}
	}
	if ps := c.Query("pageStart"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n >= 0 {
			pageStart = n
		}
	}

	// 验证参数
	if contentType == "" || tag == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少必要参数: type 或 tag"))
		return
	}

	if contentType != "tv" && contentType != "movie" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "type 参数必须是 tv 或 movie"))
		return
	}

	h.logger.Debug("豆瓣搜索",
		zap.String("type", contentType),
		zap.String("tag", tag),
		zap.Int("pageSize", pageSize),
		zap.Int("pageStart", pageStart),
	)

	// Top250 特殊处理
	if tag == "top250" {
		h.handleTop250(c, pageStart)
		return
	}

	// 构建请求 URL
	target := fmt.Sprintf(
		"https://movie.douban.com/j/search_subjects?type=%s&tag=%s&sort=recommend&page_limit=%d&page_start=%d",
		contentType, tag, pageSize, pageStart,
	)

	// 调用豆瓣 API
	doubanData, err := h.fetchDoubanData(target)
	if err != nil {
		h.logger.Error("豆瓣搜索失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取豆瓣数据失败"))
		return
	}

	// 转换数据格式
	var list []DoubanItem
	for _, item := range doubanData.Subjects {
		list = append(list, DoubanItem{
			ID:     item.ID,
			Title:  item.Title,
			Poster: item.Cover,
			Rate:   item.Rate,
			Year:   "",
		})
	}

	result := DoubanResult{
		Code:    200,
		Message: "获取成功",
		List:    list,
	}

	c.JSON(http.StatusOK, result)
}

// GetRecommends 豆瓣推荐
// GET /api/v1/douban/recommends?type=movie&category=xxx
func (h *DoubanHandler) GetRecommends(c *gin.Context) {
	contentType := c.DefaultQuery("type", "movie")
	category := c.Query("category")

	if category == "" {
		category = "热门"
	}

	h.logger.Debug("豆瓣推荐",
		zap.String("type", contentType),
		zap.String("category", category),
	)

	// 复用搜索逻辑
	target := fmt.Sprintf(
		"https://movie.douban.com/j/search_subjects?type=%s&tag=%s&sort=recommend&page_limit=16&page_start=0",
		contentType, category,
	)

	doubanData, err := h.fetchDoubanData(target)
	if err != nil {
		h.logger.Error("获取豆瓣推荐失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取推荐失败"))
		return
	}

	var list []DoubanItem
	for _, item := range doubanData.Subjects {
		list = append(list, DoubanItem{
			ID:     item.ID,
			Title:  item.Title,
			Poster: item.Cover,
			Rate:   item.Rate,
			Year:   "",
		})
	}

	result := DoubanResult{
		Code:    200,
		Message: "获取成功",
		List:    list,
	}

	c.JSON(http.StatusOK, result)
}

// GetCategories 豆瓣分类
// GET /api/v1/douban/categories
func (h *DoubanHandler) GetCategories(c *gin.Context) {
	// 返回豆瓣分类列表
	categories := []gin.H{
		// 电影分类
		{"type": "movie", "name": "热门", "key": "热门"},
		{"type": "movie", "name": "最新", "key": "最新"},
		{"type": "movie", "name": "经典", "key": "经典"},
		{"type": "movie", "name": "豆瓣高分", "key": "豆瓣高分"},
		{"type": "movie", "name": "冷门佳片", "key": "冷门佳片"},
		{"type": "movie", "name": "华语", "key": "华语"},
		{"type": "movie", "name": "欧美", "key": "欧美"},
		{"type": "movie", "name": "韩国", "key": "韩国"},
		{"type": "movie", "name": "日本", "key": "日本"},
		{"type": "movie", "name": "动作", "key": "动作"},
		{"type": "movie", "name": "喜剧", "key": "喜剧"},
		{"type": "movie", "name": "爱情", "key": "爱情"},
		{"type": "movie", "name": "科幻", "key": "科幻"},
		{"type": "movie", "name": "悬疑", "key": "悬疑"},
		{"type": "movie", "name": "恐怖", "key": "恐怖"},
		{"type": "movie", "name": "治愈", "key": "治愈"},
		// 电视剧分类
		{"type": "tv", "name": "热门", "key": "热门"},
		{"type": "tv", "name": "美剧", "key": "美剧"},
		{"type": "tv", "name": "英剧", "key": "英剧"},
		{"type": "tv", "name": "韩剧", "key": "韩剧"},
		{"type": "tv", "name": "日剧", "key": "日剧"},
		{"type": "tv", "name": "国产剧", "key": "国产剧"},
		{"type": "tv", "name": "港剧", "key": "港剧"},
		{"type": "tv", "name": "日本动画", "key": "日本动画"},
		{"type": "tv", "name": "综艺", "key": "综艺"},
		{"type": "tv", "name": "纪录片", "key": "纪录片"},
	}

	c.JSON(http.StatusOK, model.Success(categories))
}

// fetchDoubanData 获取豆瓣数据
func (h *DoubanHandler) fetchDoubanData(url string) (*DoubanApiResponse, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://movie.douban.com/")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result DoubanApiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	return &result, nil
}

// handleTop250 处理豆瓣 Top250
func (h *DoubanHandler) handleTop250(c *gin.Context, pageStart int) {
	target := fmt.Sprintf("https://movie.douban.com/top250?start=%d&filter=", pageStart)

	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		h.logger.Error("创建请求失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取失败"))
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://movie.douban.com/")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Error("请求豆瓣 Top250 失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取失败"))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取失败"))
		return
	}

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("读取响应失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取失败"))
		return
	}

	// 正则匹配影片信息
	moviePattern := regexp.MustCompile(`<div class="item">[\s\S]*?<a[^>]+href="https?://movie\.douban\.com/subject/(\d+)/"[\s\S]*?<img[^>]+alt="([^"]+)"[^>]*src="([^"]+)"[\s\S]*?<span class="rating_num"[^>]*>([^<]*)</span>[\s\S]*?</div>`)
	matches := moviePattern.FindAllStringSubmatch(string(html), -1)

	var movies []DoubanItem
	for _, match := range matches {
		if len(match) >= 5 {
			id := match[1]
			title := match[2]
			cover := match[3]
			rate := match[4]

			// 处理图片 URL，使用 HTTPS
			cover = regexp.MustCompile(`^http:`).ReplaceAllString(cover, "https:")

			movies = append(movies, DoubanItem{
				ID:     id,
				Title:  title,
				Poster: cover,
				Rate:   rate,
				Year:   "",
			})
		}
	}

	result := DoubanResult{
		Code:    200,
		Message: "获取成功",
		List:    movies,
	}

	c.JSON(http.StatusOK, result)
}
