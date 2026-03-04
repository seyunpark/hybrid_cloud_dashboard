package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
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

func (s *Server) handleListKubeContexts(c *gin.Context) {
	kubeconfigPath := c.DefaultQuery("kubeconfig", "")

	contexts, err := s.kubernetes.ListKubeContexts(kubeconfigPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "KUBECONFIG_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"contexts": contexts})
}

func (s *Server) handleRegisterCluster(c *gin.Context) {
	var req models.RegisterClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	if req.Kubeconfig == "" {
		req.Kubeconfig = "~/.kube/config"
	}
	if req.Type == "" {
		req.Type = "kubernetes"
	}

	clusterCfg := config.ClusterConfig{
		Name:       req.Name,
		Type:       req.Type,
		Kubeconfig: req.Kubeconfig,
		Context:    req.Context,
		Registry:   req.Registry,
	}

	if err := s.kubernetes.AddCluster(c.Request.Context(), clusterCfg); err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "CLUSTER_EXISTS", Message: err.Error()},
		})
		return
	}

	// Persist to DB
	if s.data != nil {
		_ = s.data.SaveRegisteredCluster(c.Request.Context(), &models.RegisteredCluster{
			Name:       req.Name,
			Type:       req.Type,
			Kubeconfig: req.Kubeconfig,
			Context:    req.Context,
			Registry:   req.Registry,
		})
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Cluster %q registered successfully", req.Name),
	})
}

func (s *Server) handleUnregisterCluster(c *gin.Context) {
	name := c.Param("name")

	if err := s.kubernetes.RemoveCluster(name); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "CLUSTER_NOT_FOUND", Message: err.Error()},
		})
		return
	}

	// Remove from DB
	if s.data != nil {
		_ = s.data.DeleteRegisteredCluster(c.Request.Context(), name)
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Cluster %q unregistered successfully", name),
	})
}

func (s *Server) handleGetAIConfig(c *gin.Context) {
	c.JSON(http.StatusOK, s.ai.GetConfig())
}

func (s *Server) handleUpdateAIConfig(c *gin.Context) {
	var req models.UpdateAIConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	if req.Provider == "" && req.APIKey == "" && req.Model == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: "at least one field (provider, api_key, model) is required"},
		})
		return
	}

	s.ai.UpdateConfig(req.Provider, req.APIKey, req.Model)

	// Persist to DB
	if s.data != nil {
		ctx := c.Request.Context()
		if req.Provider != "" {
			_ = s.data.SaveSetting(ctx, "ai.provider", req.Provider)
		}
		if req.APIKey != "" {
			_ = s.data.SaveSetting(ctx, "ai.api_key", req.APIKey)
		}
		if req.Model != "" {
			_ = s.data.SaveSetting(ctx, "ai.model", req.Model)
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "AI configuration updated successfully",
	})
}

func (s *Server) handleListAIModels(c *gin.Context) {
	provider := c.Query("provider")
	apiKey := c.Query("api_key")

	modelList, err := s.ai.ListModels(c.Request.Context(), provider, apiKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "LIST_MODELS_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": modelList})
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
