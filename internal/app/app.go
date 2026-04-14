package app

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"sipi/internal/config"
	"sipi/internal/http/routes"
	"sipi/internal/platform/db"
	"sipi/internal/platform/jwt"
	"sipi/internal/platform/logger"

	"gorm.io/gorm"
)

const shutdownTimeout = 10 * time.Second

type App struct {
	config       *config.Config
	dependencies *Dependencies
	httpServer   *http.Server
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	deps, err := buildDependencies(cfg)
	if err != nil {
		return nil, err
	}

	router := routes.Setup(routes.Dependencies{
		Config:       deps.Config,
		DB:           deps.DB,
		Logger:       deps.Logger,
		TokenManager: deps.TokenManager,
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.App.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	return &App{
		config:       cfg,
		dependencies: deps,
		httpServer:   server,
	}, nil
}

func (a *App) Run() error {
	a.dependencies.Logger.Info("starting HTTP server", "port", a.config.App.Port)

	serverErrors := make(chan error, 1)

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
		close(serverErrors)
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		if err != nil {
			return err
		}
	case <-shutdownCtx.Done():
		a.dependencies.Logger.Info("shutdown signal received")
	}

	if err := a.shutdown(); err != nil {
		return err
	}

	a.dependencies.Logger.Info("application stopped gracefully")

	return nil
}

func (a *App) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	if err := closeDB(a.dependencies.DB); err != nil {
		return err
	}

	return nil
}

func buildDependencies(cfg *config.Config) (*Dependencies, error) {
	appLogger := logger.New()

	gormDB, err := db.NewPostgres(cfg)
	if err != nil {
		return nil, err
	}

	if err := migrate(gormDB); err != nil {
		return nil, err
	}

	return &Dependencies{
		Config:       cfg,
		DB:           gormDB,
		Logger:       appLogger,
		TokenManager: jwt.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL),
	}, nil
}

func closeDB(gormDB *gorm.DB) error {
	if gormDB == nil {
		return nil
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("get sql db for shutdown: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close sql db: %w", err)
	}

	return nil
}
