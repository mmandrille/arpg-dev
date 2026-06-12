// Package httpapi wires the server's HTTP surface: health/readiness/metrics
// today, plus auth, sessions, inspection, and the WebSocket upgrade added by
// later tasks. It lives in directory internal/http but is named httpapi so the
// standard net/http import is not shadowed.
package httpapi

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/auth"
	"github.com/mmandrille_meli/arpg-dev/server/internal/config"
	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	"github.com/mmandrille_meli/arpg-dev/server/internal/ids"
	"github.com/mmandrille_meli/arpg-dev/server/internal/logging"
	"github.com/mmandrille_meli/arpg-dev/server/internal/metrics"
	"github.com/mmandrille_meli/arpg-dev/server/internal/realtime"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

// ReadyFunc reports whether the server's dependencies are ready to serve. A nil
// ReadyFunc is treated as always-ready.
type ReadyFunc func(ctx context.Context) error

// Deps are the collaborators a Server needs.
type Deps struct {
	Config   config.Config
	Logger   *slog.Logger
	Metrics  *metrics.Metrics
	Store    store.Repository
	Auth     *auth.Service
	Realtime *realtime.Hub
	Rules    *game.Rules
	Ready    ReadyFunc
}

// Server holds HTTP dependencies and builds the routed handler.
type Server struct {
	cfg      config.Config
	log      *slog.Logger
	metrics  *metrics.Metrics
	store    store.Repository
	auth     *auth.Service
	realtime *realtime.Hub
	rules    *game.Rules
	ready    ReadyFunc
}

// New constructs a Server.
func New(d Deps) *Server {
	return &Server{
		cfg:      d.Config,
		log:      logging.Component(d.Logger, "http"),
		metrics:  d.Metrics,
		store:    d.Store,
		auth:     d.Auth,
		realtime: d.Realtime,
		rules:    d.Rules,
		ready:    d.Ready,
	}
}

// Handler builds the fully routed, middleware-wrapped HTTP handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /readyz", s.handleReadyz)
	if s.cfg.MetricsEnabled {
		mux.Handle("GET /metrics", s.metrics.Handler())
	}
	s.registerAuthRoutes(mux)
	s.registerCharacterRoutes(mux)
	s.registerSessionRoutes(mux)
	s.registerAccountStashRoutes(mux)
	s.registerMarketRoutes(mux)
	s.registerInspectRoutes(mux)
	s.registerRealtimeRoutes(mux)
	return s.withMiddleware(mux)
}

// --- handlers ---------------------------------------------------------------

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	if s.ready != nil {
		if err := s.ready(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "unready",
				"error":  err.Error(),
			})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// --- middleware -------------------------------------------------------------

func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return s.recoverMW(s.observeMW(next))
}

// observeMW assigns/propagates a correlation id and records access logs and
// HTTP metrics. The correlation id is echoed in the X-Correlation-Id header.
func (s *Server) observeMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corr := r.Header.Get("X-Correlation-Id")
		if corr == "" {
			corr = ids.New("corr")
		}
		ctx := logging.ContextWithCorrelation(r.Context(), corr)
		r = r.WithContext(ctx)
		w.Header().Set("X-Correlation-Id", corr)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r)
		dur := time.Since(start)

		route := r.Pattern
		if route == "" {
			route = "unmatched"
		}
		s.metrics.HTTPRequests.WithLabelValues(r.Method, route, fmt.Sprintf("%d", rec.status)).Inc()
		s.log.LogAttrs(r.Context(), slog.LevelInfo, "http_request",
			slog.String("correlation_id", corr),
			slog.String("method", r.Method),
			slog.String("route", route),
			slog.Int("status", rec.status),
			slog.Duration("duration", dur),
		)
	})
}

// recoverMW converts a panic into a 500 + structured error instead of crashing.
func (s *Server) recoverMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				corr, _ := logging.CorrelationFromContext(r.Context())
				s.log.LogAttrs(r.Context(), slog.LevelError, "panic_recovered",
					slog.String("correlation_id", corr),
					slog.Any("panic", rec),
				)
				writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// --- helpers ----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// errorBody is the canonical structured error shape for HTTP responses.
type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	var b errorBody
	b.Error.Code = code
	b.Error.Message = message
	writeJSON(w, status, b)
}

// statusRecorder captures the response status and transparently supports the
// optional interfaces a WebSocket upgrade and streaming need.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.wroteHeader = true
	return r.ResponseWriter.Write(b)
}

func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("underlying ResponseWriter does not support Hijack")
	}
	return h.Hijack()
}

func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
