package metrics

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
)

// Collector periodically gathers metrics from Docker and Kubernetes
// and broadcasts them to connected WebSocket clients.
type Collector struct {
	interval          time.Duration
	broadcastInterval time.Duration
	cancel            context.CancelFunc
	wg                sync.WaitGroup
}

// NewCollector creates a new metrics collector with the given configuration.
func NewCollector(cfg config.MetricsConfig) *Collector {
	return &Collector{
		interval:          time.Duration(cfg.Interval) * time.Second,
		broadcastInterval: time.Duration(cfg.BroadcastInterval) * time.Second,
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

func (c *Collector) collectLoop(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// TODO: Collect metrics from Docker and Kubernetes services
			// 1. Query Docker stats for all containers
			// 2. Query K8s metrics for all clusters
			// 3. Aggregate and store in memory
			// 4. Broadcast to WebSocket subscribers
		}
	}
}
