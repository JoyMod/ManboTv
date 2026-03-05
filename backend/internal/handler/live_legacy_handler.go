package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/JoyMod/ManboTV/backend/internal/service"
)

// GetSourcesLegacy handles GET /api/live/sources.
func (h *LiveHandler) GetSourcesLegacy(c *gin.Context) {
	sources, err := h.service.GetSources(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取直播源失败"})
		return
	}

	filtered := make([]service.LiveSource, 0, len(sources))
	for _, source := range sources {
		if !source.Disabled {
			filtered = append(filtered, source)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    filtered,
	})
}

// GetChannelsLegacy handles GET /api/live/channels.
func (h *LiveHandler) GetChannelsLegacy(c *gin.Context) {
	source := strings.TrimSpace(c.Query("source"))
	if source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少直播源参数"})
		return
	}

	channels, err := h.service.GetChannels(c.Request.Context(), source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取频道信息失败"})
		return
	}

	data := make([]gin.H, 0, len(channels.Channels))
	for _, channel := range channels.Channels {
		data = append(data, gin.H{
			"id":    channel.ID,
			"tvgId": channel.TvgID,
			"name":  channel.Name,
			"logo":  channel.Logo,
			"group": channel.Group,
			"url":   channel.URL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetEPGLegacy handles GET /api/live/epg.
func (h *LiveHandler) GetEPGLegacy(c *gin.Context) {
	source := strings.TrimSpace(c.Query("source"))
	tvgID := strings.TrimSpace(c.Query("tvgId"))
	if source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少直播源参数"})
		return
	}
	if tvgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少频道tvg-id参数"})
		return
	}

	channels, err := h.service.GetChannels(c.Request.Context(), source)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"tvgId":    tvgID,
				"source":   source,
				"epgUrl":   "",
				"programs": []service.EPGItem{},
			},
		})
		return
	}

	epgList, err := h.service.GetEPG(c.Request.Context(), source, tvgID)
	if err != nil {
		epgList = []service.EPGItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tvgId":    tvgID,
			"source":   source,
			"epgUrl":   channels.EPGUrl,
			"programs": epgList,
		},
	})
}

// PrecheckLegacy handles GET /api/live/precheck.
func (h *LiveHandler) PrecheckLegacy(c *gin.Context) {
	rawURL := strings.TrimSpace(c.Query("url"))
	source := strings.TrimSpace(c.Query("moontv-source"))
	if rawURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing url"})
		return
	}

	ua, found := h.lookupSourceUA(c.Request.Context(), source)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Source not found"})
		return
	}
	if ua == "" {
		ua = "AptvPlayer/1.4.10"
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, rawURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch", "message": err.Error()})
		return
	}
	req.Header.Set("User-Agent", ua)

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch", "message": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch", "message": resp.Status})
		return
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.Contains(contentType, "video/mp4") {
		c.JSON(http.StatusOK, gin.H{"success": true, "type": "mp4"})
		return
	}
	if strings.Contains(contentType, "video/x-flv") || strings.Contains(contentType, "flv") {
		c.JSON(http.StatusOK, gin.H{"success": true, "type": "flv"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "type": "m3u8"})
}

func (h *LiveHandler) lookupSourceUA(ctx context.Context, sourceKey string) (string, bool) {
	if sourceKey == "" {
		return "", false
	}
	sources, err := h.service.GetSources(ctx)
	if err != nil {
		return "", false
	}
	for _, source := range sources {
		if source.Key == sourceKey {
			return source.UA, true
		}
	}
	return "", false
}
