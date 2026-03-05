package handler

import (
	"context"
	"strings"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

func resolveVideoSites(
	ctx context.Context,
	adminStorage model.AdminStorageService,
	fallback []model.ApiSite,
	username string,
	ownerUser string,
) []model.ApiSite {
	if adminStorage == nil {
		return cloneSites(fallback)
	}

	videoSources, err := adminStorage.GetVideoSources(ctx)
	if err != nil {
		return cloneSites(fallback)
	}

	sites := make([]model.ApiSite, 0, len(videoSources))
	for _, source := range videoSources {
		if source.Disabled {
			continue
		}

		key := strings.TrimSpace(source.Key)
		api := strings.TrimSpace(source.API)
		if key == "" || api == "" {
			continue
		}

		sites = append(sites, model.ApiSite{
			Key:    key,
			Name:   strings.TrimSpace(source.Name),
			API:    api,
			Detail: strings.TrimSpace(source.Detail),
		})
	}

	if len(sites) == 0 {
		return []model.ApiSite{}
	}

	username = strings.TrimSpace(username)
	if username == "" || username == strings.TrimSpace(ownerUser) {
		return sites
	}

	cfg, err := adminStorage.GetAdminConfig(ctx)
	if err != nil || cfg == nil {
		return sites
	}

	var user *model.UserConfig
	for i := range cfg.UserConfig {
		if strings.TrimSpace(cfg.UserConfig[i].Username) == username {
			user = &cfg.UserConfig[i]
			break
		}
	}

	if user == nil {
		storedUser, lookupErr := adminStorage.GetUser(ctx, username)
		if lookupErr == nil && storedUser != nil {
			user = storedUser
		}
	}

	if user == nil || user.Banned {
		return sites
	}

	if len(user.EnabledAPIs) > 0 {
		enabled := make(map[string]struct{}, len(user.EnabledAPIs))
		for _, key := range user.EnabledAPIs {
			enabled[strings.TrimSpace(key)] = struct{}{}
		}
		return filterSitesBySet(sites, enabled)
	}

	if len(user.Tags) > 0 && len(cfg.UserGroups) > 0 {
		enabled := make(map[string]struct{})
		for _, tag := range user.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			for _, group := range cfg.UserGroups {
				if strings.TrimSpace(group.Name) != tag {
					continue
				}
				for _, apiKey := range group.EnabledAPIs {
					apiKey = strings.TrimSpace(apiKey)
					if apiKey != "" {
						enabled[apiKey] = struct{}{}
					}
				}
			}
		}
		if len(enabled) > 0 {
			return filterSitesBySet(sites, enabled)
		}
	}

	return sites
}

func filterSitesBySet(sites []model.ApiSite, enabled map[string]struct{}) []model.ApiSite {
	filtered := make([]model.ApiSite, 0, len(sites))
	for _, site := range sites {
		if _, ok := enabled[strings.TrimSpace(site.Key)]; ok {
			filtered = append(filtered, site)
		}
	}
	return filtered
}

func cloneSites(sites []model.ApiSite) []model.ApiSite {
	if len(sites) == 0 {
		return []model.ApiSite{}
	}
	out := make([]model.ApiSite, len(sites))
	copy(out, sites)
	return out
}
