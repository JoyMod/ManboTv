package handler

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/service"
)

const (
	searchDefaultPage     = 1
	searchDefaultPageSize = 20
	searchMaximumPageSize = 120
)

func parseSearchRequest(c *gin.Context) model.SearchRequest {
	page := searchParsePositiveInt(c.Query("page"), searchDefaultPage)
	pageSize := searchParsePositiveInt(c.Query("page_size"), searchDefaultPageSize)
	if pageSize > searchMaximumPageSize {
		pageSize = searchMaximumPageSize
	}

	return model.SearchRequest{
		Query:       strings.TrimSpace(c.Query("q")),
		Page:        page,
		PageSize:    pageSize,
		View:        strings.TrimSpace(c.Query("view")),
		Sort:        strings.TrimSpace(c.Query("sort")),
		Types:       strings.TrimSpace(c.Query("types")),
		Sources:     strings.TrimSpace(c.Query("sources")),
		YearFrom:    searchParsePositiveInt(c.Query("year_from"), 0),
		YearTo:      searchParsePositiveInt(c.Query("year_to"), 0),
		SourceMode:  strings.TrimSpace(c.Query("source_mode")),
		PreferExact: parseBoolQuery(c.Query("prefer_exact")),
		Stream:      parseBoolQuery(c.Query("stream")),
	}
}

func buildSearchParams(
	ctx context.Context,
	request model.SearchRequest,
	adminStorage model.AdminStorageService,
) service.SearchParams {
	params := service.SearchParams{
		Query:        request.Query,
		Page:         request.Page,
		PageSize:     request.PageSize,
		View:         request.View,
		Sort:         request.Sort,
		Types:        splitSearchList(request.Types),
		Sources:      splitSearchList(request.Sources),
		YearFrom:     request.YearFrom,
		YearTo:       request.YearTo,
		SourceMode:   request.SourceMode,
		PreferExact:  request.PreferExact,
		EnableStream: request.Stream,
	}

	if adminStorage == nil {
		return params
	}

	config, err := adminStorage.GetAdminConfig(ctx)
	if err != nil || config == nil {
		return params
	}

	params.MaxPages = config.SiteConfig.SearchDownstreamMaxPage
	params.MaxConcurrent = config.SiteConfig.SearchMaxConcurrent
	if strings.TrimSpace(params.Sort) == "" {
		params.Sort = strings.TrimSpace(config.SiteConfig.SearchDefaultSort)
	}
	if config.SiteConfig.SearchSourceTimeoutMs > 0 {
		params.SourceTimeout = time.Duration(config.SiteConfig.SearchSourceTimeoutMs) * time.Millisecond
	}
	params.EnableFastReturn = config.SiteConfig.FluidSearch
	params.EnableStream = params.EnableStream || config.SiteConfig.SearchEnableStream

	return params
}

func buildLegacySearchParams(
	ctx context.Context,
	request model.SearchRequest,
	adminStorage model.AdminStorageService,
) service.SearchParams {
	params := buildSearchParams(ctx, request, adminStorage)
	params.Page = searchDefaultPage
	params.PageSize = searchMaximumPageSize
	params.ReturnAll = true
	return params
}

func splitSearchList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	segments := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '，' || r == '|' || r == ' '
	})
	values := make([]string, 0, len(segments))
	seen := make(map[string]struct{}, len(segments))
	for _, segment := range segments {
		value := strings.TrimSpace(segment)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values
}

func searchParsePositiveInt(value string, fallback int) int {
	number, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || number < 0 {
		return fallback
	}
	return number
}

func parseBoolQuery(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
