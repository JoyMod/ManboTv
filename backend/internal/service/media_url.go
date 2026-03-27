package service

import (
	"net/url"
	"strings"
)

func normalizeMediaURL(raw, apiBase string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}

	value = strings.ReplaceAll(value, "&amp;", "&")

	if strings.HasPrefix(value, "data:") {
		return value
	}

	if strings.HasPrefix(value, "//") {
		return "https:" + value
	}

	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return value
	}

	base, err := url.Parse(strings.TrimSpace(apiBase))
	if err != nil || base == nil || base.Scheme == "" || base.Host == "" {
		return value
	}

	if strings.HasPrefix(value, "/") {
		return base.Scheme + "://" + base.Host + value
	}

	rel, err := url.Parse(value)
	if err != nil {
		return value
	}

	return base.ResolveReference(rel).String()
}

func appendProviderQuery(base string, params url.Values) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return "", err
	}

	query := parsed.Query()
	proxyTarget := strings.TrimSpace(query.Get("url"))
	if proxyTarget != "" {
		nested, nestedErr := url.Parse(proxyTarget)
		if nestedErr == nil && nested != nil {
			nestedQuery := nested.Query()
			for key, values := range params {
				for _, value := range values {
					nestedQuery.Set(key, value)
				}
			}
			nested.RawQuery = nestedQuery.Encode()
			query.Set("url", nested.String())
			parsed.RawQuery = query.Encode()
			return parsed.String(), nil
		}
	}

	for key, values := range params {
		for _, value := range values {
			query.Set(key, value)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}
