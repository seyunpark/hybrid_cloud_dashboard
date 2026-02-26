package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"
)

func (s *Server) handleGetClustersConfig(c *gin.Context) {
	type clusterInfo struct {
		Name           string `json:"name"`
		Type           string `json:"type"`
		KubeconfigPath string `json:"kubeconfig_path"`
		Context        string `json:"context"`
		Registry       string `json:"registry"`
	}

	clusters := make([]clusterInfo, len(s.cfg.Clusters))
	for i, cl := range s.cfg.Clusters {
		clusters[i] = clusterInfo{
			Name:           cl.Name,
			Type:           cl.Type,
			KubeconfigPath: cl.Kubeconfig,
			Context:        cl.Context,
			Registry:       cl.Registry,
		}
	}

	c.JSON(http.StatusOK, gin.H{"clusters": clusters})
}

func (s *Server) handleGetAIConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"provider":          s.cfg.AI.Provider,
		"model":             s.cfg.AI.Model,
		"temperature":       s.cfg.AI.Temperature,
		"few_shot_enabled":  s.cfg.AI.FewShot.Enabled,
		"few_shot_examples": s.cfg.AI.FewShot.MaxExamples,
	})
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleReady(c *gin.Context) {
	checks := map[string]string{
		"docker":   "ok",
		"database": "ok",
	}

	// TODO: Perform actual dependency health checks
	// - Docker API connectivity
	// - Kubernetes API connectivity
	// - AI API connectivity
	// - Database connectivity

	c.JSON(http.StatusOK, models.ReadyResponse{
		Status:    "ready",
		Checks:    checks,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
