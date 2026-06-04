package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chatbot-go-admin/server/internal/config"
	"chatbot-go-admin/server/internal/db"
	"chatbot-go-admin/server/internal/httpserver"
	"chatbot-go-admin/server/internal/logging"
	"chatbot-go-admin/server/internal/schema"
	"chatbot-go-admin/server/internal/repository"
	"chatbot-go-admin/server/internal/seed"
	"chatbot-go-admin/server/internal/service"
)

func main() {
	cfg := config.Load()
	logger := logging.New(cfg.LogLevel)
	defer func() { _ = logger.Sync() }()

	logger.Info("chatbot-go-admin server starting",
		logging.String("http_addr", cfg.HTTPAddr),
		logging.String("log_level", cfg.LogLevel),
	)

	gormDB, err := db.OpenPostgres(cfg.DB)
	if err != nil {
		logger.Fatal("open db failed", logging.Err(err))
	}
	logger.Info("database connected",
		logging.String("host", cfg.DB.Host),
		logging.String("port", fmt.Sprintf("%d", cfg.DB.Port)),
		logging.String("database", cfg.DB.DBName),
	)

	if err := schema.CheckAndInit(context.Background(), gormDB, logger, cfg.SchemaDir); err != nil {
		logger.Fatal("schema check failed", logging.Err(err))
	}
	logger.Info("schema initialized", logging.String("schema_dir", cfg.SchemaDir))

	if err := seed.EnsureDefaultAdmin(context.Background(), gormDB, logger); err != nil {
		logger.Fatal("seed admin failed", logging.Err(err))
	}
	logger.Info("default admin seed ready")

	if err := seed.MigratePrefixedIDsToUUID(context.Background(), gormDB, logger); err != nil {
		logger.Fatal("migrate legacy ids failed", logging.Err(err))
	}
	logger.Info("legacy id migration finished")

	cleanupCtx, stopCleanup := context.WithCancel(context.Background())
	defer stopCleanup()
	deviceAudioSvc := service.NewDeviceAudioService(
		repository.NewBizDeviceAudioRepository(gormDB),
		repository.NewBizUserSpeakerRepository(gormDB),
		service.NewVoiceprintClientFromConfig(cfg, logger),
		service.NewProfileImportClientFromConfig(cfg, logger),
		logger,
	)
	go deviceAudioSvc.RunCleanupLoop(cleanupCtx)
	logger.Info("device audio cleanup worker started")

	engine := httpserver.NewEngine(logger, gormDB, cfg)
	httpserver.LogStartup(logger, engine, cfg)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("http server listening", logging.String("addr", cfg.HTTPAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("http server failed", logging.Err(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	stopCleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	logger.Info("http server stopped")
}

