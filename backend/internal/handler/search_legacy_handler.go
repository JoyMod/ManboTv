package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// SearchLegacy handles GET /api/search.
func (h *SearchHandler) SearchLegacy(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusOK, gin.H{"results": []interface{}{}})
		return
	}

	sites := h.resolveSites(c)
	results, err := h.service.Search(c.Request.Context(), query, sites)
	if err != nil {
		h.logger.Error("legacy search failed", zap.Error(err), zap.String("query", query))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败", "results": []interface{}{}})
		return
	}
	results = filterResults(results, resolveContentPolicyFromRequest(c, h.adminStorage))

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// SearchSingleLegacy handles GET /api/search/one.
func (h *SearchHandler) SearchSingleLegacy(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	siteKey := strings.TrimSpace(c.Query("resourceId"))
	if siteKey == "" {
		siteKey = strings.TrimSpace(c.Query("site"))
	}

	if query == "" || siteKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数: q 或 resourceId", "result": nil})
		return
	}

	targetSite, ok := h.findSite(h.resolveSites(c), siteKey)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("未找到指定的视频源: %s", siteKey), "result": nil})
		return
	}

	results, err := h.service.SearchSingle(c.Request.Context(), targetSite, query)
	if err != nil {
		h.logger.Error("legacy single search failed", zap.Error(err), zap.String("site", siteKey), zap.String("query", query))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败", "result": nil})
		return
	}
	results = filterResults(results, resolveContentPolicyFromRequest(c, h.adminStorage))

	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到结果", "result": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// SearchResourcesLegacy handles GET /api/search/resources.
func (h *SearchHandler) SearchResourcesLegacy(c *gin.Context) {
	c.JSON(http.StatusOK, h.resolveSites(c))
}

// SearchStreamLegacy handles GET /api/search/ws as SSE stream.
func (h *SearchHandler) SearchStreamLegacy(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET")
	c.Header("Access-Control-Allow-Headers", "Content-Type")

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

	if !send(gin.H{
		"type":         "start",
		"query":        query,
		"totalSources": len(h.resolveSites(c)),
		"timestamp":    time.Now().UnixMilli(),
	}) {
		return
	}

	type streamMessage struct {
		Payload interface{}
	}

	ctx := c.Request.Context()
	sites := h.resolveSites(c)
	policy := resolveContentPolicyFromRequest(c, h.adminStorage)
	ch := make(chan streamMessage, len(sites))

	var wg sync.WaitGroup
	for _, site := range sites {
		site := site
		wg.Add(1)
		go func() {
			defer wg.Done()

			searchCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
			defer cancel()

			results, err := h.service.SearchSingle(searchCtx, site, query)
			if err != nil {
				ch <- streamMessage{Payload: gin.H{
					"type":       "source_error",
					"source":     site.Key,
					"sourceName": site.Name,
					"error":      err.Error(),
					"timestamp":  time.Now().UnixMilli(),
				}}
				return
			}
			results = filterResults(results, policy)

			ch <- streamMessage{Payload: gin.H{
				"type":       "source_result",
				"source":     site.Key,
				"sourceName": site.Name,
				"results":    results,
				"timestamp":  time.Now().UnixMilli(),
			}}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	completed := 0
	totalResults := 0
	for msg := range ch {
		completed++
		if payload, ok := msg.Payload.(gin.H); ok {
			if results, ok := payload["results"].([]model.SearchResult); ok {
				totalResults += len(results)
			}
		}
		if !send(msg.Payload) {
			return
		}
	}

	_ = send(gin.H{
		"type":             "complete",
		"totalResults":     totalResults,
		"completedSources": completed,
		"timestamp":        time.Now().UnixMilli(),
	})
}

func (h *SearchHandler) findSite(sites []model.ApiSite, key string) (model.ApiSite, bool) {
	for _, site := range sites {
		if site.Key == key {
			return site, true
		}
	}
	return model.ApiSite{}, false
}
