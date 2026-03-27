package service

import (
	"crypto/tls"
	"net/http"
	"strings"
	"time"

	"github.com/JoyMod/ManboTV/backend/internal/config"
)

const MinimumProxyTimeout = 30 * time.Second

func buildHTTPClients(cfg *config.HTTPClientConfig) (*http.Client, *http.Client) {
	clientTimeout := cfg.Timeout
	if clientTimeout < MinimumProxyTimeout {
		clientTimeout = MinimumProxyTimeout
	}

	baseTransport := &http.Transport{
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
	}
	insecureTransport := baseTransport.Clone()
	insecureTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec

	return &http.Client{
			Timeout:   clientTimeout,
			Transport: baseTransport,
		}, &http.Client{
			Timeout:   clientTimeout,
			Transport: insecureTransport,
		}
}

func shouldRetryWithoutTLSVerify(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "certificate signed by unknown authority") ||
		strings.Contains(message, "tls: failed to verify certificate") ||
		strings.Contains(message, "x509:")
}
