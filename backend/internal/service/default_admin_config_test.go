package service

import (
	"strings"
	"testing"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

func TestApplyDefaultAdminConfigBootstrapsEmptyConfig(t *testing.T) {
	config := applyDefaultAdminConfig(&model.AdminConfig{})

	if strings.TrimSpace(config.ConfigFile) == "" {
		t.Fatal("expected default config file to be populated")
	}
	if config.ConfigSubscription.URL != defaultConfigSubscriptionURL {
		t.Fatalf("expected default subscription url, got %q", config.ConfigSubscription.URL)
	}
	if config.SiteConfig.ContentAccessMode != model.ContentAccessModeSafe {
		t.Fatalf("expected safe mode, got %q", config.SiteConfig.ContentAccessMode)
	}
}

func TestApplyDefaultAdminConfigKeepsExistingSources(t *testing.T) {
	config := applyDefaultAdminConfig(&model.AdminConfig{
		SiteConfig: model.SiteConfig{
			ContentAccessMode: model.ContentAccessModeMixed,
		},
		VideoSources: []model.VideoSource{
			{Key: "custom", API: "https://example.com/api.php/provide/vod"},
		},
	})

	if strings.TrimSpace(config.ConfigFile) != "" {
		t.Fatalf("expected config file to remain empty, got %q", config.ConfigFile)
	}
	if config.SiteConfig.ContentAccessMode != model.ContentAccessModeMixed {
		t.Fatalf("expected existing access mode to be preserved, got %q", config.SiteConfig.ContentAccessMode)
	}
}
