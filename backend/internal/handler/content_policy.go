package handler

import (
	"context"
	"strings"

	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/service"
	"github.com/gin-gonic/gin"
)

const contentAccessModeCookieName = "content_access_mode"

func resolveContentPolicy(ctx context.Context, adminStorage model.AdminStorageService) service.ContentPolicy {
	if adminStorage == nil {
		return service.ContentPolicy{}
	}

	config, err := adminStorage.GetAdminConfig(ctx)
	if err != nil || config == nil {
		return service.ContentPolicy{}
	}

	return service.ContentPolicy{
		AccessMode:          config.SiteConfig.ContentAccessMode,
		DisableYellowFilter: config.SiteConfig.DisableYellowFilter,
		BlockedTags:         config.SiteConfig.BlockedContentTags,
	}
}

func resolveContentPolicyFromRequest(c *gin.Context, adminStorage model.AdminStorageService) service.ContentPolicy {
	policy := resolveContentPolicy(c.Request.Context(), adminStorage)

	overrideMode, err := c.Cookie(contentAccessModeCookieName)
	if err != nil {
		return policy
	}

	switch strings.TrimSpace(overrideMode) {
	case model.ContentAccessModeSafe, model.ContentAccessModeMixed, model.ContentAccessModeAdultOnly:
		policy.AccessMode = overrideMode
		policy.DisableYellowFilter = overrideMode != model.ContentAccessModeSafe
	}

	return policy
}

func filterResult(result model.SearchResult, policy service.ContentPolicy) (model.SearchResult, bool) {
	enriched := service.EnrichSearchResult(result)
	if service.IsBlockedContent(enriched, policy) {
		return model.SearchResult{}, false
	}
	return enriched, true
}

func filterResults(results []model.SearchResult, policy service.ContentPolicy) []model.SearchResult {
	filtered := make([]model.SearchResult, 0, len(results))
	for _, result := range results {
		if visible, ok := filterResult(result, policy); ok {
			filtered = append(filtered, visible)
		}
	}
	return filtered
}

func filterDetails(results []*model.SearchResult, policy service.ContentPolicy) []*model.SearchResult {
	filtered := make([]*model.SearchResult, 0, len(results))
	for _, result := range results {
		if result == nil {
			continue
		}
		visible, ok := filterResult(*result, policy)
		if !ok {
			continue
		}
		item := visible
		filtered = append(filtered, &item)
	}
	return filtered
}
