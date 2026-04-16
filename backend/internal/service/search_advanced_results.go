package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

func rankSearchResults(
	results []model.SearchResult,
	profile searchQueryProfile,
	options searchRuntimeOptions,
) []model.SearchResult {
	ranked := make([]model.SearchResult, 0, len(results))
	for _, result := range results {
		contentType := detectResultType(result)
		if len(options.Types) > 0 && !containsString(options.Types, contentType) {
			continue
		}

		yearValue := parseResultYear(result.Year)
		if options.YearFrom > 0 && yearValue > 0 && yearValue < options.YearFrom {
			continue
		}
		if options.YearTo > 0 && yearValue > 0 && yearValue > options.YearTo {
			continue
		}

		scored := result
		scored.MatchScore, scored.MatchReasons = calculateSearchScore(result, contentType, yearValue, profile, options)
		ranked = append(ranked, scored)
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		return compareRankedSearchResults(ranked[i], ranked[j], options.Sort)
	})

	return ranked
}

func calculateSearchScore(
	result model.SearchResult,
	contentType string,
	yearValue int,
	profile searchQueryProfile,
	options searchRuntimeOptions,
) (float64, []string) {
	score := float64(0)
	reasons := make([]string, 0, 8)
	title := normalizeComparableText(result.Title)
	query := profile.NormalizedQuery

	switch {
	case title != "" && query != "" && title == query:
		score += searchExactMatchScore
		reasons = append(reasons, "标题完全匹配")
	case title != "" && query != "" && strings.Contains(title, query):
		score += searchContainsMatchScore
		reasons = append(reasons, "标题包含关键词")
	case title != "" && query != "" && strings.Contains(query, title):
		score += searchContainsMatchScore / 2
		reasons = append(reasons, "标题与关键词高度接近")
	}

	for _, token := range profile.TokenSet {
		if token == "" || !strings.Contains(title, token) {
			continue
		}
		score += searchTokenMatchScore
		reasons = append(reasons, fmt.Sprintf("命中词 %s", token))
	}

	if options.PreferExact && title == query {
		score += searchTokenMatchScore
	}

	if profile.DetectedYear > 0 && yearValue == profile.DetectedYear {
		score += searchExactYearScore
		reasons = append(reasons, "年份匹配")
	}
	if options.YearFrom > 0 && options.YearTo > 0 && yearValue >= options.YearFrom && yearValue <= options.YearTo {
		score += searchRangeYearScore
	}

	if profile.DetectedType != "" && profile.DetectedType == contentType {
		score += searchQueryTypeScore
		reasons = append(reasons, "类型匹配")
	}
	if len(options.Types) > 0 && containsString(options.Types, contentType) {
		score += searchExplicitTypeScore
	}

	episodeBonus := len(result.Episodes)
	if episodeBonus > searchEpisodeBonusCap {
		episodeBonus = searchEpisodeBonusCap
	}
	score += float64(episodeBonus)
	if episodeBonus > 0 {
		reasons = append(reasons, fmt.Sprintf("可播放线路 %d", len(result.Episodes)))
	}

	if len(result.Tags) > 0 {
		score += float64(len(result.Tags))
	}

	return score, searchUniqueStrings(reasons)
}

func compareRankedSearchResults(left, right model.SearchResult, sortMode string) bool {
	switch sortMode {
	case "year_desc":
		leftYear := parseResultYear(left.Year)
		rightYear := parseResultYear(right.Year)
		if leftYear != rightYear {
			return leftYear > rightYear
		}
	case "year_asc":
		leftYear := parseResultYear(left.Year)
		rightYear := parseResultYear(right.Year)
		if leftYear != rightYear {
			if leftYear == 0 {
				return false
			}
			if rightYear == 0 {
				return true
			}
			return leftYear < rightYear
		}
	case "title":
		leftTitle := normalizeComparableText(left.Title)
		rightTitle := normalizeComparableText(right.Title)
		if leftTitle != rightTitle {
			return leftTitle < rightTitle
		}
	case "playable":
		if len(left.Episodes) != len(right.Episodes) {
			return len(left.Episodes) > len(right.Episodes)
		}
	}

	if left.MatchScore != right.MatchScore {
		return left.MatchScore > right.MatchScore
	}
	if len(left.Episodes) != len(right.Episodes) {
		return len(left.Episodes) > len(right.Episodes)
	}
	return parseResultYear(left.Year) > parseResultYear(right.Year)
}

func buildSearchAggregates(results []model.SearchResult) []model.SearchAggregateResult {
	states := make(map[string]*searchAggregateState)
	for _, result := range results {
		contentType := detectResultType(result)
		key := buildAggregateKey(result.Title, result.Year, contentType)
		state, exists := states[key]
		if !exists {
			state = &searchAggregateState{
				Key:     key,
				Title:   result.Title,
				Year:    result.Year,
				Type:    contentType,
				Cover:   result.Poster,
				Tags:    append([]string(nil), result.Tags...),
				Items:   []model.SearchResult{},
				Sources: map[string]string{},
			}
			states[key] = state
		}
		if state.Cover == "" && result.Poster != "" {
			state.Cover = result.Poster
		}
		state.Items = append(state.Items, result)
		if sourceKey := strings.TrimSpace(result.Source); sourceKey != "" {
			state.Sources[sourceKey] = strings.TrimSpace(result.SourceName)
		}
	}

	aggregates := make([]model.SearchAggregateResult, 0, len(states))
	for _, state := range states {
		sort.SliceStable(state.Items, func(i, j int) bool {
			if state.Items[i].MatchScore != state.Items[j].MatchScore {
				return state.Items[i].MatchScore > state.Items[j].MatchScore
			}
			return len(state.Items[i].Episodes) > len(state.Items[j].Episodes)
		})

		bestSource := ""
		bestSourceName := ""
		if len(state.Items) > 0 {
			bestSource = state.Items[0].Source
			bestSourceName = state.Items[0].SourceName
		}

		aggregates = append(aggregates, model.SearchAggregateResult{
			Key:            state.Key,
			Title:          state.Title,
			Year:           state.Year,
			Type:           state.Type,
			Cover:          state.Cover,
			SourceCount:    len(state.Sources),
			ResultCount:    len(state.Items),
			BestSource:     bestSource,
			BestSourceName: bestSourceName,
			Tags:           searchUniqueStrings(state.Tags),
			Items:          cloneSearchResults(state.Items),
		})
	}

	sort.SliceStable(aggregates, func(i, j int) bool {
		left := aggregateBestScore(aggregates[i])
		right := aggregateBestScore(aggregates[j])
		if left != right {
			return left > right
		}
		if aggregates[i].SourceCount != aggregates[j].SourceCount {
			return aggregates[i].SourceCount > aggregates[j].SourceCount
		}
		if aggregates[i].ResultCount != aggregates[j].ResultCount {
			return aggregates[i].ResultCount > aggregates[j].ResultCount
		}
		return parseResultYear(aggregates[i].Year) > parseResultYear(aggregates[j].Year)
	})

	return aggregates
}

func applySourceModeFilter(
	results []model.SearchResult,
	aggregates []model.SearchAggregateResult,
	sourceMode string,
) ([]model.SearchResult, []model.SearchAggregateResult) {
	if strings.TrimSpace(sourceMode) == "" || strings.TrimSpace(sourceMode) == searchDefaultSourceMode {
		return results, aggregates
	}

	allowedGroups := make(map[string]struct{}, len(aggregates))
	filteredAggregates := make([]model.SearchAggregateResult, 0, len(aggregates))
	for _, aggregate := range aggregates {
		if !matchesSourceMode(aggregate.SourceCount, sourceMode) {
			continue
		}
		allowedGroups[aggregate.Key] = struct{}{}
		filteredAggregates = append(filteredAggregates, aggregate)
	}

	filteredResults := make([]model.SearchResult, 0, len(results))
	for _, result := range results {
		key := buildAggregateKey(result.Title, result.Year, detectResultType(result))
		if _, ok := allowedGroups[key]; ok {
			filteredResults = append(filteredResults, result)
		}
	}

	return filteredResults, filteredAggregates
}

func matchesSourceMode(sourceCount int, sourceMode string) bool {
	switch sourceMode {
	case "multi", "multi_only":
		return sourceCount >= 2
	case "single", "single_only":
		return sourceCount < 2
	default:
		return true
	}
}

func buildSearchFacets(
	results []model.SearchResult,
	aggregates []model.SearchAggregateResult,
) model.SearchFacets {
	typeCounts := make(map[string]int)
	sourceCounts := make(map[string]int)
	sourceLabels := make(map[string]string)
	yearCounts := make(map[string]int)

	for _, aggregate := range aggregates {
		if aggregate.Type != "" {
			typeCounts[aggregate.Type]++
		}
		if strings.TrimSpace(aggregate.Year) != "" {
			yearCounts[aggregate.Year]++
		}
	}

	for _, result := range results {
		sourceKey := strings.TrimSpace(result.Source)
		if sourceKey == "" {
			continue
		}
		sourceCounts[sourceKey]++
		sourceLabels[sourceKey] = strings.TrimSpace(result.SourceName)
	}

	return model.SearchFacets{
		Types:   buildFacetBuckets(typeCounts, nil),
		Sources: buildFacetBuckets(sourceCounts, sourceLabels),
		Years:   buildYearFacetBuckets(yearCounts),
	}
}

func buildFacetBuckets(counts map[string]int, labels map[string]string) []model.SearchFacetBucket {
	buckets := make([]model.SearchFacetBucket, 0, len(counts))
	for value, count := range counts {
		label := value
		if labels != nil && strings.TrimSpace(labels[value]) != "" {
			label = strings.TrimSpace(labels[value])
		}
		buckets = append(buckets, model.SearchFacetBucket{
			Value: value,
			Label: label,
			Count: count,
		})
	}
	sort.SliceStable(buckets, func(i, j int) bool {
		if buckets[i].Count != buckets[j].Count {
			return buckets[i].Count > buckets[j].Count
		}
		return buckets[i].Label < buckets[j].Label
	})
	return buckets
}

func buildYearFacetBuckets(counts map[string]int) []model.SearchFacetBucket {
	buckets := make([]model.SearchFacetBucket, 0, len(counts))
	for value, count := range counts {
		buckets = append(buckets, model.SearchFacetBucket{
			Value: value,
			Label: value,
			Count: count,
		})
	}
	sort.SliceStable(buckets, func(i, j int) bool {
		leftYear := parseResultYear(buckets[i].Value)
		rightYear := parseResultYear(buckets[j].Value)
		if leftYear != rightYear {
			return leftYear > rightYear
		}
		return buckets[i].Label > buckets[j].Label
	})
	if len(buckets) > searchMaximumFacetYearSize {
		return buckets[:searchMaximumFacetYearSize]
	}
	return buckets
}

func filterVisibleSearchResults(results []model.SearchResult, policy ContentPolicy) []model.SearchResult {
	visible := make([]model.SearchResult, 0, len(results))
	for _, result := range results {
		enriched := EnrichSearchResult(result)
		if IsBlockedContent(enriched, policy) {
			continue
		}
		visible = append(visible, enriched)
	}
	return visible
}

func paginateSearchResults(
	results []model.SearchResult,
	page,
	pageSize int,
	returnAll bool,
) []model.SearchResult {
	if returnAll {
		return cloneSearchResults(results)
	}
	if len(results) == 0 {
		return []model.SearchResult{}
	}
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	if start >= len(results) {
		return []model.SearchResult{}
	}
	end := start + pageSize
	if end > len(results) {
		end = len(results)
	}
	return cloneSearchResults(results[start:end])
}

func buildSearchPageInfo(page, pageSize, total int, returnAll bool) model.PageInfo {
	if returnAll {
		return model.PageInfo{
			Page:       1,
			PageSize:   total,
			Total:      int64(total),
			TotalPages: 1,
		}
	}

	totalPages := 1
	if pageSize > 0 && total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	if totalPages <= 0 {
		totalPages = 1
	}
	return model.PageInfo{
		Page:       page,
		PageSize:   pageSize,
		Total:      int64(total),
		TotalPages: totalPages,
	}
}

func toSearchSourceStatus(result searchSourceResult) model.SearchSourceStatus {
	status := model.SearchSourceStatus{
		Source:      strings.TrimSpace(result.Site.Key),
		SourceName:  strings.TrimSpace(result.Site.Name),
		Status:      result.Status,
		ResultCount: len(result.Results),
		PageCount:   result.PageCount,
		ElapsedMs:   result.ElapsedMs,
	}
	if result.Err != nil {
		status.Error = result.Err.Error()
	}
	return status
}

func buildLegacySourceMap(statuses []model.SearchSourceStatus) map[string]string {
	legacy := make(map[string]string, len(statuses))
	for _, status := range statuses {
		if status.Source == "" {
			continue
		}
		legacy[status.Source] = status.Status
	}
	return legacy
}

func sortSearchStatuses(statuses []model.SearchSourceStatus, sites []model.ApiSite) []model.SearchSourceStatus {
	order := make(map[string]int, len(sites))
	for index, site := range sites {
		order[strings.TrimSpace(site.Key)] = index
	}
	sort.SliceStable(statuses, func(i, j int) bool {
		leftOrder, leftExists := order[statuses[i].Source]
		rightOrder, rightExists := order[statuses[j].Source]
		switch {
		case leftExists && rightExists && leftOrder != rightOrder:
			return leftOrder < rightOrder
		case leftExists != rightExists:
			return leftExists
		default:
			return statuses[i].Source < statuses[j].Source
		}
	})
	return statuses
}

func fillMissingSearchStatuses(
	statuses []model.SearchSourceStatus,
	sites []model.ApiSite,
	degraded bool,
) []model.SearchSourceStatus {
	if len(statuses) >= len(sites) {
		return statuses
	}

	seen := make(map[string]struct{}, len(statuses))
	for _, status := range statuses {
		seen[strings.TrimSpace(status.Source)] = struct{}{}
	}

	fallbackStatus := searchStatusError
	if degraded {
		fallbackStatus = searchStatusTimeout
	}

	completed := append([]model.SearchSourceStatus(nil), statuses...)
	for _, site := range sites {
		key := strings.TrimSpace(site.Key)
		if _, ok := seen[key]; ok {
			continue
		}
		completed = append(completed, model.SearchSourceStatus{
			Source:     key,
			SourceName: strings.TrimSpace(site.Name),
			Status:     fallbackStatus,
		})
	}

	return completed
}

func dedupeSearchResults(results []model.SearchResult) []model.SearchResult {
	seen := make(map[string]struct{}, len(results))
	deduped := make([]model.SearchResult, 0, len(results))
	for _, result := range results {
		key := strings.Join([]string{
			strings.TrimSpace(result.Source),
			strings.TrimSpace(result.ID),
			normalizeComparableText(result.Title),
			strings.TrimSpace(result.Year),
		}, "|")
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, result)
	}
	return deduped
}

func normalizeComparableText(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		" ", "",
		"-", "",
		"_", "",
		".", "",
		",", "",
		"，", "",
		"。", "",
		"·", "",
		"：", "",
		":", "",
		"!", "",
		"！", "",
		"?", "",
		"？", "",
		"（", "",
		"）", "",
		"(", "",
		")", "",
		"[", "",
		"]", "",
		"【", "",
		"】", "",
		"《", "",
		"》", "",
		"/", "",
		"／", "",
	)
	return replacer.Replace(normalized)
}

func extractYearHint(query string) int {
	trimmed := strings.TrimSpace(query)
	for index := 0; index <= len(trimmed)-4; index++ {
		segment := trimmed[index : index+4]
		year, err := strconv.Atoi(segment)
		if err != nil {
			continue
		}
		if year >= 1900 && year <= time.Now().Year()+1 {
			return year
		}
	}
	return 0
}

func parseResultYear(value string) int {
	year, err := strconv.Atoi(strings.TrimSpace(value))
	if err == nil {
		return year
	}
	return extractYearHint(value)
}

func detectSearchType(value string) string {
	text := strings.ToLower(strings.TrimSpace(value))
	switch {
	case strings.Contains(text, "动漫"), strings.Contains(text, "番"), strings.Contains(text, "anime"):
		return "anime"
	case strings.Contains(text, "综艺"), strings.Contains(text, "真人秀"), strings.Contains(text, "show"):
		return "variety"
	case strings.Contains(text, "电视剧"), strings.Contains(text, "美剧"), strings.Contains(text, "韩剧"), strings.Contains(text, "日剧"), strings.Contains(text, "tv"):
		return "tv"
	case strings.Contains(text, "电影"), strings.Contains(text, "movie"), strings.Contains(text, "影片"):
		return "movie"
	default:
		return ""
	}
}

func detectResultType(result model.SearchResult) string {
	text := strings.ToLower(strings.Join([]string{result.TypeName, result.Class, strings.Join(result.Tags, " ")}, " "))
	switch {
	case strings.Contains(text, "动漫"), strings.Contains(text, "番"), strings.Contains(text, "动画"), strings.Contains(text, "anime"):
		return "anime"
	case strings.Contains(text, "综艺"), strings.Contains(text, "真人秀"), strings.Contains(text, "脱口秀"), strings.Contains(text, "show"):
		return "variety"
	case strings.Contains(text, "剧"), strings.Contains(text, "tv"), strings.Contains(text, "连续"), strings.Contains(text, "美剧"), strings.Contains(text, "韩剧"):
		return "tv"
	default:
		return "movie"
	}
}

func buildAggregateKey(title, year, contentType string) string {
	return strings.Join([]string{
		normalizeComparableText(title),
		strings.TrimSpace(year),
		strings.TrimSpace(contentType),
	}, "|")
}

func aggregateBestScore(result model.SearchAggregateResult) float64 {
	if len(result.Items) == 0 {
		return 0
	}
	return result.Items[0].MatchScore + float64(result.SourceCount*searchMultiSourceScore)
}

func filterSitesBySelection(sites []model.ApiSite, selectedSources []string) []model.ApiSite {
	if len(selectedSources) == 0 {
		return cloneApiSites(sites)
	}
	allowed := make(map[string]struct{}, len(selectedSources))
	for _, source := range selectedSources {
		allowed[source] = struct{}{}
	}
	filtered := make([]model.ApiSite, 0, len(sites))
	for _, site := range sites {
		if _, ok := allowed[strings.TrimSpace(site.Key)]; ok {
			filtered = append(filtered, site)
		}
	}
	return filtered
}

func normalizeStringList(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			continue
		}
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		normalized = append(normalized, item)
	}
	return normalized
}

func uniqueStrings(values []string) []string {
	return normalizeStringList(values)
}

func searchUniqueStrings(values []string) []string {
	return normalizeStringList(values)
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == strings.TrimSpace(target) {
			return true
		}
	}
	return false
}

func isSearchTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "deadline exceeded") || strings.Contains(message, "context canceled") || strings.Contains(message, "timeout")
}

func cloneApiSites(sites []model.ApiSite) []model.ApiSite {
	if len(sites) == 0 {
		return []model.ApiSite{}
	}
	cloned := make([]model.ApiSite, len(sites))
	copy(cloned, sites)
	return cloned
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
