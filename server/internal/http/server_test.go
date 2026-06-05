package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mmandrille_meli/arpg-dev/server/internal/config"
	"github.com/mmandrille_meli/arpg-dev/server/internal/logging"
	"github.com/mmandrille_meli/arpg-dev/server/internal/metrics"
)

func newTestServer(t *testing.T, ready ReadyFunc) http.Handler {
	t.Helper()
	return New(Deps{
		Config:  config.Config{Addr: ":0", Env: "local", MetricsEnabled: true},
		Logger:  logging.NewTo(io.Discard, "local"),
		Metrics: metrics.New(),
		Ready:   ready,
	}).Handler()
}

func TestHealthz(t *testing.T) {
	h := newTestServer(t, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status field = %q, want ok", body["status"])
	}
	if rec.Header().Get("X-Correlation-Id") == "" {
		t.Fatal("missing X-Correlation-Id header")
	}
}

func TestReadyzReady(t *testing.T) {
	h := newTestServer(t, func(context.Context) error { return nil })
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestReadyzUnready(t *testing.T) {
	h := newTestServer(t, func(context.Context) error { return errors.New("db down") })
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["status"] != "unready" {
		t.Fatalf("status field = %q, want unready", body["status"])
	}
}

func TestMetricsExposed(t *testing.T) {
	h := newTestServer(t, nil)
	// Generate one request so the http counter has a sample.
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "arpg_http_requests_total") {
		t.Fatal("metrics output missing arpg_http_requests_total")
	}
}

func TestCorrelationIdPropagated(t *testing.T) {
	h := newTestServer(t, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Correlation-Id", "corr_provided")
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("X-Correlation-Id"); got != "corr_provided" {
		t.Fatalf("correlation id = %q, want corr_provided", got)
	}
}
