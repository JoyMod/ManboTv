package service

import (
	"strings"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

type ContentPolicy struct {
	AccessMode          string
	DisableYellowFilter bool
	BlockedTags         []string
}

type contentTagRule struct {
	Tag      string
	Keywords []string
}

var contentTextReplacer = strings.NewReplacer(
	"\n", " ",
	"\r", " ",
	"\t", " ",
	"｜", " ",
	"|", " ",
	"/", " ",
	"／", " ",
	"·", " ",
	"・", " ",
	"，", " ",
	",", " ",
	"、", " ",
	"：", " ",
	":", " ",
	"（", " ",
	"）", " ",
	"(", " ",
	")", " ",
)

var channelRules = []contentTagRule{
	{Tag: "电影", Keywords: []string{"电影", "影片", "院线", "movie"}},
	{Tag: "剧集", Keywords: []string{"电视剧", "连续剧", "短剧", "网剧", "剧集", "tv", "剧场"}},
	{Tag: "动漫", Keywords: []string{"动漫", "动画", "番剧", "国创", "anime", "cartoon"}},
	{Tag: "综艺", Keywords: []string{"综艺", "真人秀", "脱口秀", "variety", "show"}},
}

var contentTagRules = []contentTagRule{
	{Tag: "剧情", Keywords: []string{"剧情", "情感"}},
	{Tag: "喜剧", Keywords: []string{"喜剧", "搞笑", "轻喜", "爆笑"}},
	{Tag: "爱情", Keywords: []string{"爱情", "恋爱", "甜宠", "romance"}},
	{Tag: "动作", Keywords: []string{"动作", "战斗", "热血"}},
	{Tag: "科幻", Keywords: []string{"科幻", "机战", "赛博"}},
	{Tag: "悬疑", Keywords: []string{"悬疑", "推理", "烧脑", "谜案"}},
	{Tag: "惊悚", Keywords: []string{"惊悚", "惊栗", "惊魂"}},
	{Tag: "恐怖", Keywords: []string{"恐怖", "恐影", "灵异", "鬼怪"}},
	{Tag: "犯罪", Keywords: []string{"犯罪", "警匪", "刑侦", "罪案", "黑帮"}},
	{Tag: "冒险", Keywords: []string{"冒险", "探险", "闯关"}},
	{Tag: "奇幻", Keywords: []string{"奇幻", "魔幻", "玄幻", "魔法", "神话"}},
	{Tag: "武侠", Keywords: []string{"武侠", "江湖"}},
	{Tag: "古装", Keywords: []string{"古装", "古风", "宫廷", "宫斗", "权谋"}},
	{Tag: "家庭", Keywords: []string{"家庭", "亲情", "家长里短", "亲子"}},
	{Tag: "历史", Keywords: []string{"历史", "史诗", "年代"}},
	{Tag: "战争", Keywords: []string{"战争", "抗战", "军旅", "谍战"}},
	{Tag: "传记", Keywords: []string{"传记", "人物传", "纪实人物"}},
	{Tag: "纪录片", Keywords: []string{"纪录", "纪录片", "documentary"}},
	{Tag: "动画", Keywords: []string{"动画", "动画电影"}},
	{Tag: "音乐", Keywords: []string{"音乐", "歌舞", "演唱会", "乐队"}},
	{Tag: "运动", Keywords: []string{"运动", "体育", "篮球", "足球", "电竞"}},
	{Tag: "灾难", Keywords: []string{"灾难", "末日"}},
	{Tag: "青春", Keywords: []string{"青春", "校园", "成长", "少年"}},
	{Tag: "儿童", Keywords: []string{"儿童", "少儿"}},
	{Tag: "同性", Keywords: []string{"同性", "耽美", "百合", "lgbt", "bl", "gl"}},
	{Tag: "短片", Keywords: []string{"短片", "短剧", "微电影"}},
	{Tag: "西部", Keywords: []string{"西部", "牛仔"}},
	{Tag: "真人秀", Keywords: []string{"真人秀", "观察类"}},
	{Tag: "脱口秀", Keywords: []string{"脱口秀", "talk show"}},
	{Tag: "美食", Keywords: []string{"美食", "下饭"}},
	{Tag: "旅行", Keywords: []string{"旅行", "旅拍", "游记"}},
	{Tag: "访谈", Keywords: []string{"访谈", "对话"}},
	{Tag: "选秀", Keywords: []string{"选秀", "竞演"}},
	{Tag: "电影解说", Keywords: []string{"电影解说", "影视解说", "解说"}},
	{Tag: "花絮", Keywords: []string{"花絮", "幕后", "预告"}},
}

var adultKeywords = []string{
	"18禁", "十八禁", "成人", "情色", "色情", "福利", "伦理",
	"写真", "sm", "av", "jav", "h漫", "做爱", "无码", "有码",
	"porn", "erotic", "sex", "情欲", "淫", "欲望", "🔞", "色猫",
	"迷奸", "内射", "乱伦", "口交", "约炮", "裸聊", "调教",
	"成人视频", "无码视频", "网红主播",
}

var adultSpecificRules = []contentTagRule{
	{Tag: "十八禁", Keywords: []string{"18禁", "十八禁", "r18", "成人向"}},
	{Tag: "情色", Keywords: []string{"情色", "色情", "情欲", " erotic "}},
	{Tag: "福利", Keywords: []string{"福利", "写真", "无码", "有码"}},
	{Tag: "伦理", Keywords: []string{"伦理", "伦理片"}},
}

func EnrichSearchResult(result model.SearchResult) model.SearchResult {
	tags, isAdult := ExtractContentTags(result.Title, result.Class, result.TypeName, result.Desc)
	if len(tags) > 0 {
		result.Tags = tags
	}
	result.IsAdult = isAdult
	return result
}

func ExtractContentTags(title, className, typeName, desc string) ([]string, bool) {
	primaryText := normalizeContentText(strings.Join([]string{title, className, typeName}, " "))
	if primaryText == "" {
		return nil, false
	}
	fullText := normalizeContentText(strings.Join([]string{title, className, typeName, desc}, " "))

	tags := make([]string, 0, len(contentTagRules))
	seen := make(map[string]struct{}, len(contentTagRules))

	addTag := func(tag string) {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return
		}
		if _, ok := seen[tag]; ok {
			return
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}

	for _, rule := range channelRules {
		if containsAnyKeyword(primaryText, rule.Keywords) {
			addTag(rule.Tag)
		}
	}

	for _, rule := range contentTagRules {
		if containsAnyKeyword(primaryText, rule.Keywords) {
			addTag(rule.Tag)
		}
	}

	isAdult := containsAnyKeyword(fullText, adultKeywords)
	if isAdult {
		addTag("成人")
		for _, rule := range adultSpecificRules {
			if containsAnyKeyword(fullText, rule.Keywords) {
				addTag(rule.Tag)
			}
		}
	}

	return tags, isAdult
}

func IsBlockedContent(result model.SearchResult, policy ContentPolicy) bool {
	result = EnrichSearchResult(result)
	searchable := normalizeContentText(strings.Join([]string{
		result.Title,
		result.Class,
		result.TypeName,
		result.Desc,
		result.Source,
		result.SourceName,
		strings.Join(result.Tags, " "),
	}, " "))
	isAdultContent := result.IsAdult || containsAnyKeyword(searchable, adultKeywords)

	switch resolveContentAccessMode(policy.AccessMode, policy.DisableYellowFilter) {
	case model.ContentAccessModeAdultOnly:
		if !isAdultContent {
			return true
		}
	case model.ContentAccessModeSafe:
		if isAdultContent {
			return true
		}
	default:
		// mixed 模式不过滤成人内容
	}

	if policy.AccessMode == "" && !policy.DisableYellowFilter && isAdultContent {
		return true
	}

	normalizedTags := normalizeTagSet(result.Tags)
	for _, blockedTag := range policy.BlockedTags {
		normalizedBlocked := normalizeContentText(blockedTag)
		if normalizedBlocked == "" {
			continue
		}
		if _, ok := normalizedTags[normalizedBlocked]; ok {
			return true
		}
		if searchable != "" && strings.Contains(searchable, normalizedBlocked) {
			return true
		}
	}

	return false
}

func normalizeTagSet(tags []string) map[string]struct{} {
	normalized := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		normalizedTag := normalizeContentText(tag)
		if normalizedTag == "" {
			continue
		}
		normalized[normalizedTag] = struct{}{}
	}
	return normalized
}

func containsAnyKeyword(contentText string, keywords []string) bool {
	for _, keyword := range keywords {
		normalizedKeyword := normalizeContentText(keyword)
		if normalizedKeyword == "" {
			continue
		}
		if strings.Contains(contentText, normalizedKeyword) {
			return true
		}
	}
	return false
}

func normalizeContentText(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return ""
	}
	normalized = contentTextReplacer.Replace(normalized)
	return strings.Join(strings.Fields(normalized), " ")
}

func resolveContentAccessMode(mode string, disableYellowFilter bool) string {
	switch strings.TrimSpace(mode) {
	case model.ContentAccessModeSafe, model.ContentAccessModeMixed, model.ContentAccessModeAdultOnly:
		return mode
	default:
		if disableYellowFilter {
			return model.ContentAccessModeMixed
		}
		return model.ContentAccessModeSafe
	}
}
