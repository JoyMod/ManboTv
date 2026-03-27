package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	browseDefaultPageSize      = 25
	browseMaximumPageSize      = 50
	browseMaximumPageIndex     = 6
	browseDefaultSort          = "default"
	browseDefaultKind          = "movie"
	browseFallbackIDBase       = 1000
	browseHotScoreMultiplier   = 10
	browseHotScoreYearFallback = 0
)

type browseContentItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Cover string `json:"cover"`
	Rate  string `json:"rate,omitempty"`
	Year  string `json:"year,omitempty"`
	Type  string `json:"type,omitempty"`
}

type browseBootstrapPayload struct {
	Kind         string              `json:"kind"`
	Items        []browseContentItem `json:"items"`
	TotalResults int                 `json:"total_results"`
	Page         int                 `json:"page"`
	PageSize     int                 `json:"page_size"`
	HasMore      bool                `json:"has_more"`
}

type browseQuery struct {
	Kind       string
	APIType    string
	DefaultTag string
	Sort       string
	Page       int
	PageSize   int
	Filters    map[string][]string
}

type browseFetchPlan struct {
	FetchTags   []string
	FetchGroups map[string][]string
}

type doubanAggregateItem struct {
	ID    string
	Title string
	Cover string
	Rate  string
	Year  string
	Type  string
}

func fetchDoubanAggregatePage(
	ctx context.Context,
	client *http.Client,
	contentType string,
	tag string,
	sortKey string,
	pageSize int,
	pageStart int,
) ([]DoubanItem, error) {
	resolvedTag := normalizeDoubanTag(contentType, tag)
	target := fmt.Sprintf(
		"https://movie.douban.com/j/search_subjects?type=%s&tag=%s&sort=%s&page_limit=%d&page_start=%d",
		url.QueryEscape(contentType),
		url.QueryEscape(resolvedTag),
		url.QueryEscape(resolveDoubanSort(sortKey)),
		pageSize,
		pageStart,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("create douban request failed: %w", err)
	}

	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	)
	req.Header.Set("Referer", "https://movie.douban.com/")
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request douban failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("douban status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read douban response failed: %w", err)
	}

	var payload DoubanApiResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode douban response failed: %w", err)
	}

	items := make([]DoubanItem, 0, len(payload.Subjects))
	for _, item := range payload.Subjects {
		items = append(items, DoubanItem{
			ID:     item.ID,
			Title:  item.Title,
			Poster: item.Cover,
			Rate:   item.Rate,
		})
	}

	return items, nil
}

func parseBrowseQueryValues(raw string) []string {
	segments := strings.Split(strings.TrimSpace(raw), ",")
	values := make([]string, 0, len(segments))
	for _, segment := range segments {
		value := strings.TrimSpace(segment)
		if value == "" || value == "全部" {
			continue
		}
		values = append(values, value)
	}
	return uniqueStrings(values)
}

func parseBrowseQuery(kind string, values url.Values) browseQuery {
	resolvedKind := normalizeBrowseKind(kind)
	query := browseQuery{
		Kind:       resolvedKind,
		APIType:    resolveBrowseAPIType(resolvedKind),
		DefaultTag: resolveBrowseDefaultTag(resolvedKind),
		Sort:       firstNonEmptyString(strings.TrimSpace(values.Get("sort")), browseDefaultSort),
		Page:       clampInt(parsePositiveInt(values.Get("page"), 0), 0, browseMaximumPageIndex),
		PageSize:   clampInt(parsePositiveInt(values.Get("pageSize"), browseDefaultPageSize), 1, browseMaximumPageSize),
		Filters:    map[string][]string{},
	}

	filterKeys := []string{"type", "region", "feature", "status", "year"}
	for _, key := range filterKeys {
		query.Filters[key] = parseBrowseQueryValues(values.Get(key))
	}

	return query
}

func buildBrowseFetchPlan(filters map[string][]string, defaultTag string) browseFetchPlan {
	fetchGroups := make(map[string][]string)
	for groupID, values := range filters {
		if len(values) == 0 {
			continue
		}
		if groupID == "year" {
			yearTags := make([]string, 0, len(values))
			for _, value := range values {
				if !isFetchableYearTag(value) {
					continue
				}
				yearTags = append(yearTags, normalizeYearFetchValue(value))
			}
			if len(yearTags) > 0 {
				fetchGroups[groupID] = uniqueStrings(yearTags)
			}
			continue
		}
		fetchGroups[groupID] = uniqueStrings(values)
	}

	fetchTags := make([]string, 0, len(fetchGroups))
	for _, values := range fetchGroups {
		fetchTags = append(fetchTags, values...)
	}
	fetchTags = uniqueStrings(fetchTags)
	if len(fetchTags) == 0 {
		fetchTags = []string{defaultTag}
	}

	return browseFetchPlan{
		FetchTags:   fetchTags,
		FetchGroups: fetchGroups,
	}
}

func buildBrowsePayload(
	ctx context.Context,
	client *http.Client,
	logger *zap.Logger,
	query browseQuery,
) browseBootstrapPayload {
	fetchPlan := buildBrowseFetchPlan(query.Filters, query.DefaultTag)
	tagBuckets := make(map[string][]doubanAggregateItem, len(fetchPlan.FetchTags))
	tagHasMore := make(map[string]bool, len(fetchPlan.FetchTags))
	var lock sync.Mutex

	group, groupCtx := errgroup.WithContext(ctx)
	for _, tag := range fetchPlan.FetchTags {
		tag := tag
		group.Go(func() error {
			items := make([]doubanAggregateItem, 0, (query.Page+1)*query.PageSize)
			hasMore := false

			for pageIndex := 0; pageIndex <= query.Page; pageIndex++ {
				pageStart := pageIndex * query.PageSize
				list, err := fetchDoubanAggregatePage(
					groupCtx,
					client,
					query.APIType,
					tag,
					query.Sort,
					query.PageSize,
					pageStart,
				)
				if err != nil {
					logger.Warn(
						"browse bootstrap tag degraded",
						zap.String("kind", query.Kind),
						zap.String("tag", tag),
						zap.Int("page", pageIndex),
						zap.Error(err),
					)
					break
				}

				items = append(items, mapBrowseAggregateItems(list, query.Kind)...)
				hasMore = len(list) == query.PageSize
				if !hasMore {
					break
				}
			}

			lock.Lock()
			tagBuckets[tag] = dedupeAggregateItems(items)
			tagHasMore[tag] = hasMore
			lock.Unlock()
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		logger.Warn("browse bootstrap interrupted", zap.String("kind", query.Kind), zap.Error(err))
	}

	items := mergeBrowseAggregateItems(tagBuckets, fetchPlan, query.Filters, query.DefaultTag, query.Sort)
	hasMore := false
	for _, tag := range fetchPlan.FetchTags {
		if tagHasMore[tag] {
			hasMore = true
			break
		}
	}

	return browseBootstrapPayload{
		Kind:         query.Kind,
		Items:        items,
		TotalResults: len(items),
		Page:         query.Page,
		PageSize:     query.PageSize,
		HasMore:      hasMore,
	}
}

func mapBrowseAggregateItems(items []DoubanItem, kind string) []doubanAggregateItem {
	mapped := make([]doubanAggregateItem, 0, len(items))
	for index, item := range items {
		mapped = append(mapped, doubanAggregateItem{
			ID:    resolveAggregateID(item.ID, kind, index),
			Title: firstNonEmptyString(item.Title, "未知标题"),
			Cover: firstNonEmptyString(item.Poster, "/placeholder-poster.svg"),
			Rate:  item.Rate,
			Year:  item.Year,
			Type:  kind,
		})
	}
	return mapped
}

func dedupeAggregateItems(items []doubanAggregateItem) []doubanAggregateItem {
	seen := make(map[string]struct{}, len(items))
	unique := make([]doubanAggregateItem, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item.ID]; exists {
			continue
		}
		seen[item.ID] = struct{}{}
		unique = append(unique, item)
	}
	return unique
}

func mergeBrowseAggregateItems(
	tagBuckets map[string][]doubanAggregateItem,
	fetchPlan browseFetchPlan,
	filters map[string][]string,
	defaultTag string,
	sortKey string,
) []browseContentItem {
	activeGroups := make([][]string, 0, len(fetchPlan.FetchGroups))
	for _, tags := range fetchPlan.FetchGroups {
		if len(tags) > 0 {
			activeGroups = append(activeGroups, tags)
		}
	}

	groupSets := make([]map[string]struct{}, 0, len(activeGroups))
	for _, tags := range activeGroups {
		itemSet := make(map[string]struct{})
		for _, tag := range tags {
			for _, item := range tagBuckets[tag] {
				itemSet[item.ID] = struct{}{}
			}
		}
		groupSets = append(groupSets, itemSet)
	}

	sourceTags := fetchPlan.FetchTags
	if len(activeGroups) == 0 {
		sourceTags = []string{defaultTag}
	}

	candidateMap := make(map[string]doubanAggregateItem)
	for _, tag := range sourceTags {
		for _, item := range tagBuckets[tag] {
			candidateMap[item.ID] = item
		}
	}

	filtered := make([]doubanAggregateItem, 0, len(candidateMap))
	for _, item := range candidateMap {
		if !matchesBrowseGroupSets(item.ID, groupSets) {
			continue
		}
		if !matchesBrowseYearFilters(item.Year, filters["year"]) {
			continue
		}
		filtered = append(filtered, item)
	}

	sortedItems := sortBrowseAggregateItems(filtered, sortKey)
	result := make([]browseContentItem, 0, len(sortedItems))
	for _, item := range sortedItems {
		result = append(result, browseContentItem{
			ID:    item.ID,
			Title: item.Title,
			Cover: item.Cover,
			Rate:  item.Rate,
			Year:  item.Year,
			Type:  item.Type,
		})
	}
	return result
}

func sortBrowseAggregateItems(items []doubanAggregateItem, sortKey string) []doubanAggregateItem {
	sortedItems := append([]doubanAggregateItem(nil), items...)
	switch strings.TrimSpace(sortKey) {
	case "rating":
		sort.SliceStable(sortedItems, func(i, j int) bool {
			return parseFloatScore(sortedItems[i].Rate) > parseFloatScore(sortedItems[j].Rate)
		})
	case "latest":
		sort.SliceStable(sortedItems, func(i, j int) bool {
			return parseYearScore(sortedItems[i].Year) > parseYearScore(sortedItems[j].Year)
		})
	case "hot", "comments", "follow":
		sort.SliceStable(sortedItems, func(i, j int) bool {
			return buildHotScore(sortedItems[i]) > buildHotScore(sortedItems[j])
		})
	}
	return sortedItems
}

func matchesBrowseGroupSets(id string, groupSets []map[string]struct{}) bool {
	for _, groupSet := range groupSets {
		if _, ok := groupSet[id]; !ok {
			return false
		}
	}
	return true
}

func matchesBrowseYearFilters(itemYear string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}

	if strings.TrimSpace(itemYear) == "" {
		for _, filter := range filters {
			if !isFetchableYearTag(filter) {
				return false
			}
		}
		return true
	}

	for _, filter := range filters {
		if matchesYearBucket(itemYear, filter) {
			return true
		}
	}
	return false
}

func matchesYearBucket(itemYear string, filter string) bool {
	year := parseYearScore(itemYear)
	if year <= 0 {
		return false
	}
	switch {
	case isFourDigitValue(filter):
		return year == parsePositiveInt(filter, 0)
	case strings.HasSuffix(filter, "年代") || filter == "90s":
		return year >= 1990 && year <= 1999
	case filter == "2010s":
		return year >= 2010 && year <= 2019
	case filter == "2000s":
		return year >= 2000 && year <= 2009
	case filter == "更早" || filter == "earlier":
		return year < 1990
	case filter == "经典":
		return year <= 2019
	default:
		return false
	}
}

func resolveDoubanSort(sortKey string) string {
	switch strings.TrimSpace(strings.ToLower(sortKey)) {
	case "latest", "time", "recent":
		return "time"
	case "rating", "rank":
		return "rank"
	default:
		return "recommend"
	}
}

func normalizeBrowseKind(kind string) string {
	switch strings.TrimSpace(kind) {
	case "tv", "anime", "variety":
		return kind
	default:
		return browseDefaultKind
	}
}

func resolveBrowseAPIType(kind string) string {
	if kind == "movie" {
		return "movie"
	}
	return "tv"
}

func resolveBrowseDefaultTag(kind string) string {
	switch kind {
	case "anime":
		return "日本动画"
	case "variety":
		return "综艺"
	case "tv":
		return "国产剧"
	default:
		return "热门"
	}
}

func resolveAggregateID(rawID string, prefix string, index int) string {
	id := strings.TrimSpace(rawID)
	if id != "" {
		return id
	}
	return fmt.Sprintf("%s-%d", prefix, browseFallbackIDBase+index)
}

func parsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func clampInt(value int, minValue int, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func parseFloatScore(value string) float64 {
	score, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}
	return score
}

func parseYearScore(value string) int {
	return parsePositiveInt(strings.TrimSpace(value), 0)
}

func buildHotScore(item doubanAggregateItem) float64 {
	return parseFloatScore(item.Rate)*browseHotScoreMultiplier +
		float64(maxInt(parseYearScore(item.Year), browseHotScoreYearFallback))
}

func maxInt(value int, fallback int) int {
	if value > fallback {
		return value
	}
	return fallback
}

func isFetchableYearTag(value string) bool {
	trimmed := strings.TrimSpace(value)
	if isFourDigitValue(trimmed) {
		return true
	}
	return len(trimmed) >= 4 && isFourDigitValue(trimmed[:4])
}

func isFourDigitValue(value string) bool {
	return len(value) == 4 && parsePositiveInt(value, -1) > 0
}

func normalizeYearFetchValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) >= 4 && isFourDigitValue(trimmed[:4]) {
		return trimmed[:4]
	}
	return trimmed
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		unique = append(unique, trimmed)
	}
	return unique
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
