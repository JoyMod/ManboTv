// internal/handler/douban_handler.go
package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
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
	Tag     string       `json:"tag,omitempty"`
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
	sort := normalizeDoubanSort(c.DefaultQuery("sort", "recommend"))
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

	resolvedTag := normalizeDoubanTag(contentType, tag)

	h.logger.Debug("豆瓣搜索",
		zap.String("type", contentType),
		zap.String("tag", tag),
		zap.String("resolvedTag", resolvedTag),
		zap.String("sort", sort),
		zap.Int("pageSize", pageSize),
		zap.Int("pageStart", pageStart),
	)

	// Top250 特殊处理
	if resolvedTag == "top250" {
		h.handleTop250(c, pageStart)
		return
	}

	// 构建请求 URL
	target := fmt.Sprintf(
		"https://movie.douban.com/j/search_subjects?type=%s&tag=%s&sort=%s&page_limit=%d&page_start=%d",
		url.QueryEscape(contentType),
		url.QueryEscape(resolvedTag),
		url.QueryEscape(sort),
		pageSize,
		pageStart,
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
		Tag:     resolvedTag,
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
	catalog := buildDoubanFilterCatalog()

	if c.Query("view") == "full" {
		c.JSON(http.StatusOK, model.Success(catalog))
		return
	}

	c.JSON(http.StatusOK, model.Success(flattenDoubanFilterCatalog(catalog)))
}

func normalizeDoubanSort(sort string) string {
	switch strings.TrimSpace(strings.ToLower(sort)) {
	case "time", "latest", "recent":
		return "time"
	case "rank", "rating":
		return "rank"
	default:
		return "recommend"
	}
}

func normalizeDoubanTag(contentType, raw string) string {
	tag := strings.TrimSpace(raw)
	if tag == "" {
		if contentType == "movie" {
			return "热门"
		}
		return "热门"
	}

	movieAlias := map[string]string{
		"美国":      "欧美",
		"国产":      "华语",
		"剧情":      "热门",
		"豆瓣高分":    "豆瓣高分",
		"获奖佳作":    "经典",
		"新片热映":    "最新",
		"经典重温":    "经典",
		"90年代":    "经典",
		"IMAX":    "热门",
		"4K":      "热门",
		"90s":     "经典",
		"更早":      "经典",
		"earlier": "经典",
	}

	tvAlias := map[string]string{
		"国产":      "国产剧",
		"国内":      "综艺",
		"港台":      "综艺",
		"连载中":     "最近热门",
		"已完结":     "高分经典",
		"即将开播":    "最新上线",
		"新番":      "日本动画",
		"剧场版":     "日本动画",
		"OVA":     "日本动画",
		"2026-冬":  "日本动画",
		"2025-秋":  "日本动画",
		"2025-夏":  "日本动画",
		"2025-春":  "日本动画",
		"90年代":    "高分经典",
		"90s":     "高分经典",
		"更早":      "高分经典",
		"earlier": "高分经典",
	}

	if contentType == "movie" {
		if v, ok := movieAlias[tag]; ok {
			return v
		}
		return tag
	}

	if v, ok := tvAlias[tag]; ok {
		return v
	}

	animeTags := map[string]struct{}{
		"热血": {}, "恋爱": {}, "搞笑": {}, "悬疑": {}, "科幻": {}, "机战": {},
		"运动": {}, "校园": {}, "魔法": {}, "冒险": {}, "战斗": {}, "日常": {},
		"治愈": {}, "奇幻": {}, "后宫": {}, "百合": {}, "耽美": {}, "神魔": {},
		"推理": {}, "动漫": {},
	}
	if _, ok := animeTags[tag]; ok {
		return "日本动画"
	}

	varietyTags := map[string]struct{}{
		"真人秀": {}, "脱口秀": {}, "音乐": {}, "情感": {}, "竞技": {},
		"美食": {}, "旅行": {}, "游戏": {}, "访谈": {}, "选秀": {},
		"晚会": {}, "文化": {}, "亲子": {}, "舞蹈": {}, "时尚": {},
		"明星": {}, "汽车": {},
	}
	if _, ok := varietyTags[tag]; ok {
		return "综艺"
	}

	return tag
}

func buildDoubanFilterCatalog() map[string][]gin.H {
	return map[string][]gin.H{
		"movie": {
			{"id": "type", "label": "类型", "options": []string{"全部", "动作", "喜剧", "爱情", "科幻", "恐怖", "悬疑", "犯罪", "动画", "纪录片", "战争", "古装", "奇幻", "冒险", "剧情", "惊悚", "武侠", "家庭", "传记", "历史", "音乐", "运动", "西部", "灾难", "青春", "儿童"}},
			{"id": "region", "label": "地区", "options": []string{"全部", "华语", "美国", "韩国", "日本", "印度", "泰国", "法国", "英国", "德国", "俄罗斯", "西班牙", "意大利", "其他"}},
			{"id": "feature", "label": "特色", "options": []string{"全部", "豆瓣高分", "获奖佳作", "新片热映", "经典重温", "IMAX", "4K"}},
			{"id": "year", "label": "年代", "options": []string{"全部", "2026", "2025", "2024", "2023", "2022", "2021", "2020", "2019", "2010s", "2000s", "90年代", "更早"}},
		},
		"tv": {
			{"id": "type", "label": "类型", "options": []string{"全部", "古装", "都市", "悬疑", "爱情", "武侠", "奇幻", "谍战", "军旅", "喜剧", "家庭", "科幻", "青春", "传奇", "农村", "历史", "宫廷", "仙侠", "甜宠", "职场", "校园", "穿越", "民国"}},
			{"id": "region", "label": "地区", "options": []string{"全部", "国产", "美剧", "韩剧", "日剧", "港剧", "台剧", "泰剧", "英剧"}},
			{"id": "status", "label": "状态", "options": []string{"全部", "连载中", "已完结", "即将开播"}},
			{"id": "year", "label": "年代", "options": []string{"全部", "2026", "2025", "2024", "2023", "2022", "2021", "2020", "2019", "2010s", "2000s", "90年代"}},
		},
		"anime": {
			{"id": "type", "label": "类型", "options": []string{"全部", "热血", "恋爱", "搞笑", "悬疑", "科幻", "机战", "运动", "校园", "魔法", "冒险", "战斗", "日常", "治愈", "奇幻", "后宫", "百合", "耽美", "神魔", "推理", "音乐"}},
			{"id": "region", "label": "地区", "options": []string{"全部", "日本", "国产", "欧美"}},
			{"id": "status", "label": "状态", "options": []string{"全部", "连载中", "已完结", "新番", "剧场版", "OVA"}},
			{"id": "year", "label": "年份", "options": []string{"全部", "2026冬", "2025秋", "2025夏", "2025春", "2024", "2023", "2022", "2021", "2020", "经典"}},
		},
		"variety": {
			{"id": "type", "label": "类型", "options": []string{"全部", "真人秀", "脱口秀", "音乐", "情感", "竞技", "美食", "旅行", "游戏", "访谈", "选秀", "晚会", "喜剧", "文化", "亲子", "舞蹈", "时尚", "明星", "汽车"}},
			{"id": "region", "label": "地区", "options": []string{"全部", "国内", "韩国", "日本", "欧美", "港台"}},
			{"id": "status", "label": "状态", "options": []string{"全部", "连载中", "已完结", "即将开播"}},
			{"id": "year", "label": "年代", "options": []string{"全部", "2026", "2025", "2024", "2023", "2022", "2021", "2020"}},
		},
	}
}

func flattenDoubanFilterCatalog(catalog map[string][]gin.H) []gin.H {
	result := make([]gin.H, 0, 256)
	for contentType, groups := range catalog {
		for _, group := range groups {
			groupID, _ := group["id"].(string)
			options, _ := group["options"].([]string)
			for _, option := range options {
				if option == "全部" {
					continue
				}
				result = append(result, gin.H{
					"type":      contentType,
					"dimension": groupID,
					"name":      option,
					"key":       option,
				})
			}
		}
	}
	return result
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
