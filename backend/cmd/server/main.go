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
	logger.Info("starting server", "addr", cfg.ServerAddr, "repo_root", cfg.RepoRoot, "provider_profile", cfg.ProviderProfile)

	pathService := services.NewPathService(cfg)
	if err := pathService.EnsureRuntimeDirectories(); err != nil {
		logger.Error("ensure runtime directories", "error", err)
		os.Exit(1)
	}
	if err := services.EnsureRuntimeDependencies(); err != nil {
		logger.Error("check runtime dependencies", "error", err)
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

	analysisClient, openAIClient, err := services.NewPhaseTwoClients(cfg, logger)
	if err != nil {
		logger.Error("configure phase 2 provider clients", "error", err, "provider_profile", cfg.ProviderProfile)
		os.Exit(1)
	}
	mlClient, err := services.NewPhaseThreeClient(cfg, logger)
	if err != nil {
		logger.Error("configure phase 3 provider clients", "error", err, "provider_profile", cfg.ProviderProfile)
		os.Exit(1)
	}
	blobClient, renderClient, err := services.NewPhaseFourClients(cfg, logger)
	if err != nil {
		logger.Error("configure phase 4 provider clients", "error", err, "provider_profile", cfg.ProviderProfile)
		os.Exit(1)
	}
	websiteAdsImageClient := services.NewWebsiteAdsImageClient(cfg, logger)
	auditLogger := services.NewAsyncJobAuditLogger(services.NewNotionAuditLogger(cfg, logger), logger)
	defer auditLogger.Close()
	defer auditLogger.Wait()
	auditHealthCtx, auditHealthCancel := context.WithTimeout(ctx, cfg.NotionRequestTimeout)
	auditHealth := auditLogger.Health(auditHealthCtx)
	auditHealthCancel()
	if auditHealth.Enabled && auditHealth.Status != "healthy" {
		logger.Error("notion audit connectivity check failed", "status", auditHealth.Status, "details", auditHealth.Details)
		os.Exit(1)
	}
	logger.Info("audit sink status", "enabled", auditHealth.Enabled, "status", auditHealth.Status, "details", auditHealth.Details)

	jobService := services.NewJobService(
		sqliteDB,
		db.NewJobsRepository(sqliteDB),
		db.NewCampaignsRepository(sqliteDB),
		db.NewProductsRepository(sqliteDB),
		db.NewJobLogsRepository(sqliteDB),
		db.NewScenesRepository(sqliteDB),
		db.NewSlotsRepository(sqliteDB),
		db.NewPreviewsRepository(sqliteDB),
		analysisClient,
		openAIClient,
		mlClient,
		services.NewFFmpegAnchorFrameExtractor(cfg.ArtifactsDir),
		services.NewLocalStorageService(),
		blobClient,
		renderClient,
		cfg.PreviewsDir,
		cfg.CacheDir,
	)
	jobService.SetAuditLogger(auditLogger)

	processor := worker.NewProcessor(logger, cfg.WorkerInterval)
	processor.SetOnTick(func(tickCtx context.Context) {
		if err := jobService.ProcessPendingAnalysis(tickCtx); err != nil {
			logger.Error("process pending analysis", "error", err)
		}
	})
	go processor.Run(ctx)

	handler := api.NewRouter(api.Dependencies{
		Config:                cfg,
		Logger:                logger,
		DB:                    sqliteDB,
		AnalysisClient:        analysisClient,
		OpenAIClient:          openAIClient,
		MLClient:              mlClient,
		WebsiteAdsImageClient: websiteAdsImageClient,
		AnchorFrameExtractor:  services.NewFFmpegAnchorFrameExtractor(cfg.ArtifactsDir),
		SpeechClient:          services.NewNoopSpeechClient(logger),
		BlobClient:            blobClient,
		RenderClient:          renderClient,
		CafaiGenerator:        services.NewNoopCafaiGenerator(logger),
		AuditLogger:           auditLogger,
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
