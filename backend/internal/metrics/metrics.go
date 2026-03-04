package metrics

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/docker"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/kubernetes"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"
)

// Snapshot holds the latest collected metrics.
type Snapshot struct {
	Containers []models.Container `json:"containers"`
	Clusters   []models.Cluster   `json:"clusters"`
	CollectedAt time.Time         `json:"collected_at"`
}

// Collector periodically gathers metrics from Docker and Kubernetes
// and stores the latest snapshot for WebSocket broadcast.
type Collector struct {
	interval          time.Duration
	broadcastInterval time.Duration
	cancel            context.CancelFunc
	wg                sync.WaitGroup

	docker     docker.Service
	kubernetes kubernetes.Service

	mu       sync.RWMutex
	snapshot *Snapshot
}

// NewCollector creates a new metrics collector with the given configuration.
func NewCollector(cfg config.MetricsConfig, dockerSvc docker.Service, k8sSvc kubernetes.Service) *Collector {
	return &Collector{
		interval:          time.Duration(cfg.Interval) * time.Second,
		broadcastInterval: time.Duration(cfg.BroadcastInterval) * time.Second,
		docker:            dockerSvc,
		kubernetes:        k8sSvc,
		snapshot:          &Snapshot{},
	}
}

// Start begins the metrics collection loop in a background goroutine.
func (c *Collector) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.collectLoop(ctx)
	}()

	slog.Info("metrics collector started",
		"interval", c.interval.String(),
		"broadcast_interval", c.broadcastInterval.String(),
	)
}

// Stop gracefully shuts down the metrics collector.
func (c *Collector) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
	slog.Info("metrics collector stopped")
}

// GetSnapshot returns the latest metrics snapshot.
func (c *Collector) GetSnapshot() *Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snapshot
}

func (c *Collector) collectLoop(ctx context.Context) {
	// Collect immediately on start
	c.collect(ctx)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collect(ctx)
		}
	}
}

func (c *Collector) collect(ctx context.Context) {
	snap := &Snapshot{
		CollectedAt: time.Now(),
	}

	if c.docker != nil {
		containers, err := c.docker.ListContainers(ctx, false)
		if err != nil {
			slog.Debug("metrics: failed to list containers", "error", err)
		} else {
			snap.Containers = containers
		}
	}

	if c.kubernetes != nil {
		clusters, err := c.kubernetes.ListClusters(ctx)
		if err != nil {
			slog.Debug("metrics: failed to list clusters", "error", err)
		} else {
			snap.Clusters = clusters
		}
	}

	c.mu.Lock()
	c.snapshot = snap
	c.mu.Unlock()
}
