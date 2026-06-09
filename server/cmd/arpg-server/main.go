// Command arpg-server runs the realtime game server and platform services.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mmandrille_meli/arpg-dev/server/internal/auth"
	"github.com/mmandrille_meli/arpg-dev/server/internal/config"
	"github.com/mmandrille_meli/arpg-dev/server/internal/game"
	httpapi "github.com/mmandrille_meli/arpg-dev/server/internal/http"
	"github.com/mmandrille_meli/arpg-dev/server/internal/logging"
	"github.com/mmandrille_meli/arpg-dev/server/internal/metrics"
	"github.com/mmandrille_meli/arpg-dev/server/internal/realtime"
	"github.com/mmandrille_meli/arpg-dev/server/internal/store"
)

func main() {
	migrateOnly := flag.Bool("migrate-only", false, "apply database migrations and exit")
	flag.Parse()

	cfg := config.Load()
	log := logging.New(cfg.Env)

	if err := cfg.Validate(); err != nil {
		log.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	if err := run(cfg, log, *migrateOnly); err != nil {
		log.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger, migrateOnly bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	m := metrics.New()

	// Connect to Postgres and apply migrations before serving traffic.
	connectCtx, connectCancel := context.WithTimeout(ctx, 15*time.Second)
	db, err := store.Connect(connectCtx, cfg.DatabaseURL)
	connectCancel()
	if err != nil {
		return err
	}
	defer db.Close()

	migrateCtx, migrateCancel := context.WithTimeout(ctx, 30*time.Second)
	err = db.Migrate(migrateCtx)
	migrateCancel()
	if err != nil {
		return err
	}
	log.Info("migrations applied")

	if migrateOnly {
		return nil
	}

	cleanupCtx, cleanupCancel := context.WithTimeout(ctx, 15*time.Second)
	resetConnected, err := db.ResetConnectedSessionMembers(cleanupCtx)
	if err != nil {
		cleanupCancel()
		return err
	}
	if resetConnected > 0 {
		log.Info("reset stale session connection flags", "count", resetConnected)
	}
	staleCutoff := time.Now().Add(-12 * time.Hour)
	deleted, deleteErr := db.DeleteStaleEmptySessions(cleanupCtx, staleCutoff)
	cleanupCancel()
	if deleteErr != nil {
		return deleteErr
	}
	if deleted > 0 {
		log.Info("deleted stale empty sessions", "count", deleted, "older_than", "12h")
	}

	authSvc := auth.NewService(cfg.DevToken, db, db)

	rulesDir, err := game.FindSharedRulesDir()
	if err != nil {
		return err
	}
	rules, err := game.LoadRules(rulesDir)
	if err != nil {
		return err
	}
	log.Info("rules loaded", "dir", rulesDir)

	hub := realtime.NewHub(db, rules, log, m)

	srv := httpapi.New(httpapi.Deps{
		Config:   cfg,
		Logger:   log,
		Metrics:  m,
		Store:    db,
		Auth:     authSvc,
		Realtime: hub,
		Rules:    rules,
		Ready:    db.Ping,
	})

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("server listening", "addr", cfg.Addr, "env", cfg.Env)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return httpServer.Shutdown(shutdownCtx)
}
