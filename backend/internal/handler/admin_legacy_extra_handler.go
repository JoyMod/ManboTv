package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/util"
)

// HandleSourceValidate validates sources with SSE stream.
// GET /api/admin/source/validate?q=xxx
func (h *AdminLegacyHandler) HandleSourceValidate(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("q"))
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	sources, err := h.adminStorage.GetVideoSources(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取视频源失败"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	writer := c.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	send := func(payload interface{}) bool {
		data, err := json.Marshal(payload)
		if err != nil {
			return false
		}
		if _, err := fmt.Fprintf(writer, "data: %s\n\n", data); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	if !send(gin.H{"type": "start", "totalSources": len(sources)}) {
		return
	}

	ctx := c.Request.Context()
	msgCh := make(chan gin.H, len(sources))
	var wg sync.WaitGroup

	for _, source := range sources {
		source := source
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, err := h.validateVideoSource(ctx, source, keyword)
			if err != nil {
				msgCh <- gin.H{"type": "source_error", "source": source.Key, "status": "invalid"}
				return
			}
			msgCh <- gin.H{"type": "source_result", "source": source.Key, "status": status}
		}()
	}

	go func() {
		wg.Wait()
		close(msgCh)
	}()

	completed := 0
	for msg := range msgCh {
		completed++
		if !send(msg) {
			return
		}
	}

	_ = send(gin.H{"type": "complete", "completedSources": completed})
}

func (h *AdminLegacyHandler) validateVideoSource(ctx context.Context, source model.VideoSource, keyword string) (string, error) {
	target := fmt.Sprintf("%s?ac=videolist&wd=%s", source.API, url.QueryEscape(keyword))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "invalid", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "invalid", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "invalid", fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "invalid", err
	}

	var payload struct {
		List []struct {
			VodName string `json:"vod_name"`
		} `json:"list"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "invalid", err
	}

	if len(payload.List) == 0 {
		return "no_results", nil
	}

	for _, item := range payload.List {
		if strings.Contains(strings.ToLower(item.VodName), strings.ToLower(keyword)) {
			return "valid", nil
		}
	}

	return "no_results", nil
}

func (h *AdminLegacyHandler) refreshLiveSourceChannels(ctx context.Context, key string) (int, error) {
	sources, err := h.adminStorage.GetLiveSources(ctx)
	if err != nil {
		return 0, err
	}

	updated := false
	for i := range sources {
		if key != "" && sources[i].Key != key {
			continue
		}
		if sources[i].Disabled {
			sources[i].ChannelNumber = 0
			updated = true
			continue
		}
		count, err := countM3UChannels(ctx, sources[i].URL, sources[i].UA)
		if err != nil {
			h.logger.Warn("刷新直播源失败", zap.String("key", sources[i].Key), zap.Error(err))
			sources[i].ChannelNumber = 0
		} else {
			sources[i].ChannelNumber = count
		}
		updated = true
	}

	if updated {
		if err := h.adminStorage.SaveLiveSources(ctx, sources); err != nil {
			return 0, err
		}
	}

	if key == "" {
		total := 0
		for _, source := range sources {
			total += source.ChannelNumber
		}
		return total, nil
	}

	for _, source := range sources {
		if source.Key == key {
			return source.ChannelNumber, nil
		}
	}

	return 0, nil
}

func countM3UChannels(ctx context.Context, targetURL string, ua string) (int, error) {
	if strings.TrimSpace(targetURL) == "" {
		return 0, fmt.Errorf("empty live source url")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(ua) == "" {
		ua = "AptvPlayer/1.4.10"
	}
	request.Header.Set("User-Agent", ua)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#EXTINF:") {
			count++
		}
	}

	return count, nil
}

func (h *AdminLegacyHandler) fetchAndDecodeSubscription(ctx context.Context, targetURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	decoded, err := util.DecodeBase58(strings.TrimSpace(string(body)))
	if err != nil {
		return "", err
	}

	if !json.Valid(decoded) {
		return "", fmt.Errorf("decoded content is not valid json")
	}

	return string(decoded), nil
}

func (h *AdminLegacyHandler) resetAdminConfig(ctx context.Context) error {
	config := &model.AdminConfig{
		SiteConfig: model.SiteConfig{
			SiteName:                "ManboTV",
			SearchDownstreamMaxPage: 5,
			SiteInterfaceCacheTime:  3600,
			FluidSearch:             true,
		},
		UserConfig:         []model.UserConfig{},
		VideoSources:       []model.VideoSource{},
		LiveSources:        []model.LiveSource{},
		CustomCategories:   []model.CustomCategory{},
		UserGroups:         []model.UserGroup{},
		ConfigFile:         "",
		ConfigSubscription: model.ConfigSubscription{},
	}

	return h.adminStorage.SaveAdminConfig(ctx, config)
}

// HandleReset resets admin config to defaults.
// GET /api/admin/reset
func (h *AdminLegacyHandler) HandleReset(c *gin.Context) {
	if !h.adminHandler.requireOwner(c) {
		return
	}

	if err := h.resetAdminConfig(c.Request.Context()); err != nil {
		h.logger.Error("重置管理员配置失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "重置管理员配置失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
