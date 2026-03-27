package handler

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/repository/redis"
	"github.com/JoyMod/ManboTV/backend/internal/service"
)

const (
	posterRecoverCachePrefix = "poster:recover:"
	posterRecoverCacheTTL    = 24 * time.Hour
	posterRecoverMissTTL     = 1 * time.Hour
	posterRecoverMissValue   = "-"
	posterRecoverQueryJoiner = " "
)

type posterRecoverRequest struct {
	Title  string `form:"title" binding:"required"`
	Year   string `form:"year"`
	Type   string `form:"type"`
	Cover  string `form:"cover"`
	Source string `form:"source"`
}

type PosterHandler struct {
	searchService service.SearchService
	logger        *zap.Logger
	sites         []model.ApiSite
	adminStorage  model.AdminStorageService
	ownerUser     string
	redisClient   *redis.Client
}

func NewPosterHandler(
	searchService service.SearchService,
	logger *zap.Logger,
	sites []model.ApiSite,
	adminStorage model.AdminStorageService,
	ownerUser string,
	redisClient *redis.Client,
) *PosterHandler {
	return &PosterHandler{
		searchService: searchService,
		logger:        logger,
		sites:         sites,
		adminStorage:  adminStorage,
		ownerUser:     ownerUser,
		redisClient:   redisClient,
	}
}

func (h *PosterHandler) Recover(c *gin.Context) {
	poster := h.recoverPoster(c)
	c.JSON(http.StatusOK, model.Success(gin.H{
		"poster": poster,
	}))
}

func (h *PosterHandler) RecoverLegacy(c *gin.Context) {
	poster := h.recoverPoster(c)
	c.JSON(http.StatusOK, gin.H{
		"poster": poster,
	})
}

func (h *PosterHandler) recoverPoster(c *gin.Context) string {
	var req posterRecoverRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("海报恢复参数无效", zap.Error(err))
		return ""
	}

	cacheKey := buildPosterRecoverCacheKey(req)
	if cacheKey != "" {
		if cachedPoster, ok := h.getCachedPoster(c.Request.Context(), cacheKey); ok {
			return cachedPoster
		}
	}

	sites := resolveVideoSites(
		c.Request.Context(),
		h.adminStorage,
		h.sites,
		c.GetString("username"),
		h.ownerUser,
	)

	queries := buildPosterRecoverQueries(req.Title)
	results := make([]model.SearchResult, 0)
	for _, query := range queries {
		found, err := h.searchService.Search(c.Request.Context(), query, sites)
		if err != nil {
			h.logger.Warn("海报恢复搜索失败",
				zap.String("title", req.Title),
				zap.String("query", query),
				zap.Error(err),
			)
			continue
		}
		results = append(results, found...)
	}

	poster := pickRecoveredPoster(results, req)
	if cacheKey != "" {
		h.cachePoster(c.Request.Context(), cacheKey, poster)
	}

	return poster
}

func (h *PosterHandler) getCachedPoster(ctx context.Context, cacheKey string) (string, bool) {
	if h.redisClient == nil {
		return "", false
	}

	value, err := h.redisClient.Get(ctx, cacheKey)
	if err != nil || value == "" {
		return "", false
	}
	if value == posterRecoverMissValue {
		return "", true
	}
	return value, true
}

func (h *PosterHandler) cachePoster(ctx context.Context, cacheKey, poster string) {
	if h.redisClient == nil || cacheKey == "" {
		return
	}

	value := strings.TrimSpace(poster)
	ttl := posterRecoverCacheTTL
	if value == "" {
		value = posterRecoverMissValue
		ttl = posterRecoverMissTTL
	}

	if err := h.redisClient.Set(ctx, cacheKey, value, ttl); err != nil {
		h.logger.Warn("写入海报恢复缓存失败",
			zap.String("key", cacheKey),
			zap.Error(err),
		)
	}
}

func buildPosterRecoverCacheKey(req posterRecoverRequest) string {
	normalizedTitle := normalizePosterText(req.Title)
	if normalizedTitle == "" {
		return ""
	}

	hash := sha1.Sum([]byte(strings.Join([]string{
		normalizedTitle,
		strings.TrimSpace(req.Year),
		strings.TrimSpace(req.Type),
	}, "|")))

	return posterRecoverCachePrefix + hex.EncodeToString(hash[:])
}

func pickRecoveredPoster(results []model.SearchResult, req posterRecoverRequest) string {
	normalizedTitle := normalizePosterText(req.Title)
	normalizedCover := strings.TrimSpace(req.Cover)
	bestScore := 0
	bestPoster := ""

	for _, result := range results {
		poster := strings.TrimSpace(result.Poster)
		if poster == "" || poster == normalizedCover {
			continue
		}

		score := posterTitleScore(result.Title, normalizedTitle)
		score += posterYearScore(result.Year, req.Year)
		score += posterTypeScore(result, req.Type)
		if req.Source != "" && strings.TrimSpace(result.Source) != "" && strings.TrimSpace(result.Source) != strings.TrimSpace(req.Source) {
			score++
		}

		if score > bestScore {
			bestScore = score
			bestPoster = poster
		}
	}

	return bestPoster
}

func normalizePosterText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}

	replacer := strings.NewReplacer(
		" ", "",
		"-", "",
		"_", "",
		".", "",
		":", "",
		"：", "",
		"!", "",
		"！", "",
		"?", "",
		"？", "",
		",", "",
		"，", "",
		"。", "",
		"·", "",
		"'", "",
		"\"", "",
		"“", "",
		"”", "",
		"‘", "",
		"’", "",
		"/", "",
	)
	return replacer.Replace(value)
}

func buildPosterRecoverQueries(title string) []string {
	original := strings.TrimSpace(title)
	if original == "" {
		return []string{}
	}

	normalized := normalizePosterText(original)
	collapsed := strings.Join(strings.Fields(original), posterRecoverQueryJoiner)

	queries := []string{original}
	if collapsed != "" && collapsed != original {
		queries = append(queries, collapsed)
	}
	if normalized != "" && normalized != original && normalized != collapsed {
		queries = append(queries, normalized)
	}

	return dedupePosterQueries(queries)
}

func dedupePosterQueries(queries []string) []string {
	seen := make(map[string]struct{}, len(queries))
	result := make([]string, 0, len(queries))
	for _, query := range queries {
		trimmed := strings.TrimSpace(query)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func posterTitleScore(candidateTitle, normalizedTitle string) int {
	currentTitle := normalizePosterText(candidateTitle)
	if currentTitle == "" || normalizedTitle == "" {
		return 0
	}
	if currentTitle == normalizedTitle {
		return 4
	}
	if strings.Contains(currentTitle, normalizedTitle) || strings.Contains(normalizedTitle, currentTitle) {
		return 2
	}
	return 0
}

func posterYearScore(candidateYear, targetYear string) int {
	candidateYear = strings.TrimSpace(candidateYear)
	targetYear = strings.TrimSpace(targetYear)
	if candidateYear == "" || targetYear == "" {
		return 0
	}
	if candidateYear == targetYear {
		return 2
	}
	return 0
}

func posterTypeScore(result model.SearchResult, targetType string) int {
	targetType = strings.ToLower(strings.TrimSpace(targetType))
	if targetType == "" {
		return 0
	}

	text := strings.ToLower(strings.TrimSpace(result.TypeName + " " + result.Class))
	switch targetType {
	case "movie":
		if strings.Contains(text, "电影") {
			return 2
		}
	case "tv":
		if strings.Contains(text, "剧") || strings.Contains(text, "电视剧") {
			return 2
		}
	case "variety":
		if strings.Contains(text, "综艺") {
			return 2
		}
	case "anime":
		if strings.Contains(text, "动漫") || strings.Contains(text, "动画") || strings.Contains(text, "番") {
			return 2
		}
	}
	return 0
}
