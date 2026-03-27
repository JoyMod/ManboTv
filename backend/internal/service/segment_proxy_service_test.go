package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/config"
)

func newTestSegmentService() SegmentProxyService {
	return NewSegmentProxyService(&config.HTTPClientConfig{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
		Timeout:             5 * time.Second,
	}, zap.NewNop())
}

func TestProxySegmentFullBodyReadable(t *testing.T) {
	payload := strings.Repeat("segment-data-", 128)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp2t")
		_, _ = io.WriteString(w, payload)
	}))
	defer server.Close()

	svc := newTestSegmentService()
	resp, err := svc.ProxySegment(context.Background(), server.URL, "")
	if err != nil {
		t.Fatalf("ProxySegment returned error: %v", err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		t.Fatalf("failed to read proxied body: %v", readErr)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
	if string(body) != payload {
		t.Fatalf("unexpected body length: got=%d want=%d", len(body), len(payload))
	}
}

func TestProxySegmentForwardRangeHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Range"); got != "bytes=0-9" {
			t.Fatalf("range header not forwarded: %q", got)
		}

		w.Header().Set("Content-Type", "video/mp2t")
		w.Header().Set("Content-Range", "bytes 0-9/100")
		w.Header().Set("Accept-Ranges", "bytes")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = io.WriteString(w, "0123456789")
	}))
	defer server.Close()

	svc := newTestSegmentService()
	resp, err := svc.ProxySegment(context.Background(), server.URL, "bytes=0-9")
	if err != nil {
		t.Fatalf("ProxySegment returned error: %v", err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		t.Fatalf("failed to read proxied range body: %v", readErr)
	}

	if resp.StatusCode != http.StatusPartialContent {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
	if resp.ContentRange != "bytes 0-9/100" {
		t.Fatalf("unexpected content-range: %q", resp.ContentRange)
	}
	if string(body) != "0123456789" {
		t.Fatalf("unexpected body: %q", string(body))
	}
}
