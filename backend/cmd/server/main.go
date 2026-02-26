package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seyunpark/hybrid_cloud_dashboard/internal/ai"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/api"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/data"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/docker"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/kubernetes"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/metrics"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/registry"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	slog.Info("configuration loaded", "port", cfg.Server.Port)

	// Initialize services
	dockerSvc, err := docker.NewService(cfg.Docker)
	if err != nil {
		slog.Error("failed to initialize docker service", "error", err)
		os.Exit(1)
	}

	k8sSvc, err := kubernetes.NewService(cfg.Clusters)
	if err != nil {
		slog.Error("failed to initialize kubernetes service", "error", err)
		os.Exit(1)
	}

	aiSvc, err := ai.NewService(cfg.AI)
	if err != nil {
		slog.Error("failed to initialize ai service", "error", err)
		os.Exit(1)
	}

	dataStore, err := data.NewStore(cfg.Database)
	if err != nil {
		slog.Error("failed to initialize data store", "error", err)
		os.Exit(1)
	}
	if err := dataStore.Init(); err != nil {
		slog.Error("failed to init data store", "error", err)
		os.Exit(1)
	}
	defer dataStore.Close()

	registrySvc, err := registry.NewService(cfg.Registry)
	if err != nil {
		slog.Error("failed to initialize registry service", "error", err)
		os.Exit(1)
	}

	// Initialize metrics collector
	metricsColl := metrics.NewCollector(cfg.Metrics)
	if cfg.Features.MetricsCollection {
		metricsColl.Start()
		defer metricsColl.Stop()
	}

	// Create and start HTTP server
	server := api.NewServer(cfg, dockerSvc, k8sSvc, aiSvc, dataStore, registrySvc, metricsColl)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server.Router(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server exited gracefully")
}
