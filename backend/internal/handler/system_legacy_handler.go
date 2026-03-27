package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/util"
)

const defaultServerVersion = "1.0.0"

// LegacySystemHandler handles legacy non-v1 endpoints.
type LegacySystemHandler struct {
	adminStorage model.AdminStorageService
	logger       *zap.Logger
	httpClient   *http.Client
}

// NewLegacySystemHandler creates a legacy system handler.
func NewLegacySystemHandler(adminStorage model.AdminStorageService, logger *zap.Logger) *LegacySystemHandler {
	return &LegacySystemHandler{
		adminStorage: adminStorage,
		logger:       logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetServerConfig handles GET /api/server-config.
func (h *LegacySystemHandler) GetServerConfig(c *gin.Context) {
	cfg, err := h.adminStorage.GetAdminConfig(c.Request.Context())
	if err != nil {
		h.logger.Warn("获取站点配置失败，使用默认配置", zap.Error(err))
	}

	siteName := "ManboTV"
	if cfg != nil && strings.TrimSpace(cfg.SiteConfig.SiteName) != "" {
		siteName = cfg.SiteConfig.SiteName
	}

	storageType := strings.TrimSpace(c.GetHeader("x-storage-type"))
	if storageType == "" {
		storageType = strings.TrimSpace(getEnv("NEXT_PUBLIC_STORAGE_TYPE", ""))
	}
	if storageType == "" {
		storageType = "localstorage"
	}

	version := strings.TrimSpace(getEnv("APP_VERSION", ""))
	if version == "" {
		version = defaultServerVersion
	}

	c.JSON(http.StatusOK, gin.H{
		"SiteName":    siteName,
		"StorageType": storageType,
		"Version":     version,
	})
}

// GetHome handles GET /api/home.
func (h *LegacySystemHandler) GetHome(c *gin.Context) {
	sections := buildHomeSections(c.Request.Context(), h.httpClient, h.logger)
	c.JSON(http.StatusOK, gin.H{
		"banner":   buildHomeBanner(sections),
		"sections": sections,
	})
}

// GetBangumiCalendar handles GET /api/bangumi/calendar.
func (h *LegacySystemHandler) GetBangumiCalendar(c *gin.Context) {
	request, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, "https://api.bgm.tv/calendar", nil)
	if err != nil {
		c.JSON(http.StatusOK, fallbackBangumiCalendar())
		return
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "ManboTV/1.0")

	resp, err := h.httpClient.Do(request)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			_ = resp.Body.Close()
		}
		c.JSON(http.StatusOK, fallbackBangumiCalendar())
		return
	}
	defer resp.Body.Close()

	var payload []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusOK, fallbackBangumiCalendar())
		return
	}

	result := make([]gin.H, 0, len(payload))
	for _, day := range payload {
		weekday, _ := day["weekday"].(map[string]interface{})
		items, _ := day["items"].([]interface{})

		parsedItems := make([]gin.H, 0, len(items))
		for _, item := range items {
			v, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			parsedItems = append(parsedItems, gin.H{
				"id":       v["id"],
				"name":     v["name"],
				"name_cn":  v["name_cn"],
				"images":   v["images"],
				"rating":   v["rating"],
				"air_date": v["air_date"],
			})
		}

		result = append(result, gin.H{
			"weekday": gin.H{
				"en": valueAsString(weekday, "en"),
				"cn": valueAsString(weekday, "cn"),
				"ja": valueAsString(weekday, "ja"),
			},
			"items": parsedItems,
		})
	}

	c.JSON(http.StatusOK, result)
}

// RunCron handles GET /api/cron.
func (h *LegacySystemHandler) RunCron(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 25*time.Second)
	defer cancel()

	if err := h.refreshConfigSubscription(ctx); err != nil {
		h.logger.Warn("cron 刷新配置失败", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Cron job executed successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (h *LegacySystemHandler) refreshConfigSubscription(ctx context.Context) error {
	cfg, err := h.adminStorage.GetAdminConfig(ctx)
	if err != nil {
		return fmt.Errorf("get admin config failed: %w", err)
	}

	if cfg == nil || cfg.ConfigSubscription.URL == "" || !cfg.ConfigSubscription.AutoUpdate {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.ConfigSubscription.URL, nil)
	if err != nil {
		return fmt.Errorf("build subscription request failed: %w", err)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch subscription failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("subscription status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read subscription body failed: %w", err)
	}

	decoded, err := util.DecodeBase58(strings.TrimSpace(string(body)))
	if err != nil {
		return fmt.Errorf("decode subscription failed: %w", err)
	}

	if !json.Valid(decoded) {
		return fmt.Errorf("decoded subscription is not valid json")
	}

	cfg.ConfigFile = string(decoded)
	cfg.ConfigSubscription.LastCheck = time.Now().Format(time.RFC3339)

	if err := h.adminStorage.SaveAdminConfig(ctx, cfg); err != nil {
		return fmt.Errorf("save admin config failed: %w", err)
	}

	return nil
}

func fallbackBangumiCalendar() []gin.H {
	return []gin.H{
		{"weekday": gin.H{"en": "Mon", "cn": "周一", "ja": "月曜日"}, "items": []gin.H{}},
		{"weekday": gin.H{"en": "Tue", "cn": "周二", "ja": "火曜日"}, "items": []gin.H{}},
		{"weekday": gin.H{"en": "Wed", "cn": "周三", "ja": "水曜日"}, "items": []gin.H{}},
		{"weekday": gin.H{"en": "Thu", "cn": "周四", "ja": "木曜日"}, "items": []gin.H{}},
		{"weekday": gin.H{"en": "Fri", "cn": "周五", "ja": "金曜日"}, "items": []gin.H{}},
		{"weekday": gin.H{"en": "Sat", "cn": "周六", "ja": "土曜日"}, "items": []gin.H{}},
		{"weekday": gin.H{"en": "Sun", "cn": "周日", "ja": "日曜日"}, "items": []gin.H{}},
	}
}

func valueAsString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, _ := m[key].(string)
	return v
}

func getEnv(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value != "" {
		return value
	}
	return fallback
}
