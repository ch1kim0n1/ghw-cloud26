package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/api"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/config"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/db"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/services"
	"github.com/ch1kim0n1/ghw-cloud26/backend/internal/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger.Info("starting server", "addr", cfg.ServerAddr, "repo_root", cfg.RepoRoot)

	pathService := services.NewPathService(cfg)
	if err := pathService.EnsureRuntimeDirectories(); err != nil {
		logger.Error("ensure runtime directories", "error", err)
		os.Exit(1)
	}

	sqliteDB, err := db.Open(ctx, cfg.DatabasePath)
	if err != nil {
		logger.Error("open sqlite database", "error", err)
		os.Exit(1)
	}
	defer sqliteDB.Close()

	if err := db.ApplyMigrations(ctx, sqliteDB, cfg.MigrationsDir); err != nil {
		logger.Error("apply migrations", "error", err)
		os.Exit(1)
	}

	processor := worker.NewProcessor(logger, cfg.WorkerInterval)
	go processor.Run(ctx)

	handler := api.NewRouter(api.Dependencies{
		Config:         cfg,
		Logger:         logger,
		DB:             sqliteDB,
		AnalysisClient: services.NewNoopAnalysisClient(logger),
		OpenAIClient:   services.NewNoopOpenAIClient(logger),
		MLClient:       services.NewNoopMLClient(logger),
		SpeechClient:   services.NewNoopSpeechClient(logger),
		BlobClient:     services.NewNoopBlobStorageClient(logger),
		RenderClient:   services.NewNoopRenderClient(logger),
		CafaiGenerator: services.NewNoopCafaiGenerator(logger),
	})

	server := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown", "error", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("listen and serve", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
