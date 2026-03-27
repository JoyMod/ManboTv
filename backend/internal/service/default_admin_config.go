package service

import (
	_ "embed"
	"strings"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

const (
	defaultSearchDownstreamMaxPage = 5
	defaultSiteInterfaceCacheTime  = 3600
	defaultConfigSubscriptionURL   = "https://raw.githubusercontent.com/hafrey1/LunaTV-config/main/LunaTV-config.txt"
)

//go:embed defaults/admin_config.json
var defaultAdminConfigFile string

func newDefaultAdminConfig() *model.AdminConfig {
	return &model.AdminConfig{
		SiteConfig: model.SiteConfig{
			SiteName:                "ManboTV",
			SearchDownstreamMaxPage: defaultSearchDownstreamMaxPage,
			SiteInterfaceCacheTime:  defaultSiteInterfaceCacheTime,
			FluidSearch:             true,
			ContentAccessMode:       model.ContentAccessModeSafe,
			BlockedContentTags:      []string{},
		},
		VideoSources:     []model.VideoSource{},
		LiveSources:      []model.LiveSource{},
		CustomCategories: []model.CustomCategory{},
		UserConfig:       []model.UserConfig{},
		UserGroups:       []model.UserGroup{},
		ConfigSubscription: model.ConfigSubscription{
			URL:        defaultConfigSubscriptionURL,
			AutoUpdate: true,
		},
		ConfigFile: strings.TrimSpace(defaultAdminConfigFile),
	}
}

func applyDefaultAdminConfig(config *model.AdminConfig) *model.AdminConfig {
	if config == nil {
		return newDefaultAdminConfig()
	}

	defaultConfig := newDefaultAdminConfig()

	if strings.TrimSpace(config.SiteConfig.SiteName) == "" {
		config.SiteConfig.SiteName = defaultConfig.SiteConfig.SiteName
	}
	if config.SiteConfig.SearchDownstreamMaxPage <= 0 {
		config.SiteConfig.SearchDownstreamMaxPage = defaultSearchDownstreamMaxPage
	}
	if config.SiteConfig.SiteInterfaceCacheTime <= 0 {
		config.SiteConfig.SiteInterfaceCacheTime = defaultSiteInterfaceCacheTime
	}
	if strings.TrimSpace(config.SiteConfig.ContentAccessMode) == "" {
		config.SiteConfig.ContentAccessMode = model.ContentAccessModeSafe
	}
	if config.SiteConfig.BlockedContentTags == nil {
		config.SiteConfig.BlockedContentTags = []string{}
	}
	if config.VideoSources == nil {
		config.VideoSources = []model.VideoSource{}
	}
	if config.LiveSources == nil {
		config.LiveSources = []model.LiveSource{}
	}
	if config.CustomCategories == nil {
		config.CustomCategories = []model.CustomCategory{}
	}
	if config.UserConfig == nil {
		config.UserConfig = []model.UserConfig{}
	}
	if config.UserGroups == nil {
		config.UserGroups = []model.UserGroup{}
	}

	if shouldBootstrapDefaultSources(config) {
		config.ConfigFile = defaultConfig.ConfigFile
		config.ConfigSubscription = defaultConfig.ConfigSubscription
	}

	return config
}

func shouldBootstrapDefaultSources(config *model.AdminConfig) bool {
	if config == nil {
		return true
	}

	hasExplicitSources := len(config.VideoSources) > 0 ||
		len(config.LiveSources) > 0 ||
		len(config.CustomCategories) > 0
	if hasExplicitSources {
		return false
	}

	hasUserSpecificConfig := len(config.UserConfig) > 0 ||
		len(config.UserGroups) > 0
	if hasUserSpecificConfig {
		return false
	}

	return strings.TrimSpace(config.ConfigFile) == ""
}
