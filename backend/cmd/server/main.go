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

func loadSavedAIConfig(store data.Store, aiSvc ai.Service) {
	ctx := context.Background()
	settings, err := store.GetAllSettings(ctx, "ai.")
	if err != nil {
		slog.Warn("failed to load saved AI config", "error", err)
		return
	}
	provider := settings["ai.provider"]
	apiKey := settings["ai.api_key"]
	model := settings["ai.model"]
	if provider != "" || apiKey != "" || model != "" {
		aiSvc.UpdateConfig(provider, apiKey, model)
		slog.Info("restored AI configuration from database",
			"provider", provider, "model", model, "has_key", apiKey != "")
	}
}

func loadSavedStackDeploys(store data.Store, server *api.Server) {
	ctx := context.Background()
	records, err := store.ListStackDeploys(ctx, 100)
	if err != nil {
		slog.Warn("failed to load saved stack deploys", "error", err)
		return
	}
	restored := 0
	for i := range records {
		rec := &records[i]
		switch rec.Status {
		case "generating", "deploying":
			// Interrupted by restart — mark as failed
			now := time.Now()
			rec.Status = "failed"
			rec.CompletedAt = &now
			if err := store.UpdateStackDeploy(ctx, rec); err != nil {
				slog.Warn("failed to mark interrupted deploy as failed", "deploy_id", rec.DeployID, "error", err)
			}
			slog.Info("marked interrupted stack deploy as failed", "deploy_id", rec.DeployID)
		case "pending":
			// Was waiting for user approval — restore to in-memory
			server.RestoreStackDeploy(rec)
			restored++
			slog.Info("restored pending stack deploy", "deploy_id", rec.DeployID, "stack", rec.StackName)
		}
		// deployed, failed, cancelled, undeployed — remain in DB only, queryable via API
	}
	slog.Info("stack deploy recovery complete", "total", len(records), "restored_to_memory", restored)
}

func loadSavedClusters(store data.Store, k8sSvc kubernetes.Service) {
	ctx := context.Background()
	clusters, err := store.GetRegisteredClusters(ctx)
	if err != nil {
		slog.Warn("failed to load saved clusters", "error", err)
		return
	}
	for _, c := range clusters {
		clusterCfg := config.ClusterConfig{
			Name:       c.Name,
			Type:       c.Type,
			Kubeconfig: c.Kubeconfig,
			Context:    c.Context,
			Registry:   c.Registry,
		}
		if err := k8sSvc.AddCluster(ctx, clusterCfg); err != nil {
			slog.Warn("failed to restore cluster", "name", c.Name, "error", err)
			continue
		}
		slog.Info("restored cluster from database", "name", c.Name, "context", c.Context)
	}
}

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

	// Restore persisted AI configuration
	loadSavedAIConfig(dataStore, aiSvc)

	// Restore persisted registered clusters
	loadSavedClusters(dataStore, k8sSvc)

	registrySvc, err := registry.NewService(cfg.Registry)
	if err != nil {
		slog.Error("failed to initialize registry service", "error", err)
		os.Exit(1)
	}

	// Initialize metrics collector
	metricsColl := metrics.NewCollector(cfg.Metrics, dockerSvc, k8sSvc)
	if cfg.Features.MetricsCollection {
		metricsColl.Start()
		defer metricsColl.Stop()
	}

	// Create and start HTTP server
	server := api.NewServer(cfg, dockerSvc, k8sSvc, aiSvc, dataStore, registrySvc, metricsColl)

	// Restore persisted stack deployments
	loadSavedStackDeploys(dataStore, server)

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
