package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"
)

func (s *Server) handleDeployDockerToK8s(c *gin.Context) {
	var req models.DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	// TODO: Implement actual AI-based manifest generation flow
	// 1. Get container info from Docker service
	// 2. Search similar deployments from data store
	// 3. Generate manifest via AI service
	// 4. Return deploy response with manifests

	c.JSON(http.StatusOK, models.DeployResponse{
		DeployID: "deploy-stub",
		Status:   "analyzing",
		AIAnalysis: &models.AIAnalysis{
			ServiceType:        "unknown",
			DetectedLanguage:   "unknown",
			SimilarDeployments: 0,
		},
	})
}

func (s *Server) handleExecuteDeploy(c *gin.Context) {
	deployID := c.Param("deploy_id")

	var req models.ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	// TODO: Implement actual deployment execution
	// 1. Push image to registry
	// 2. Apply Kubernetes manifests
	// 3. Monitor deployment status

	c.JSON(http.StatusOK, gin.H{
		"deploy_id": deployID,
		"status":    "deploying",
		"steps": []models.DeployStep{
			{Step: "push_image", Status: "pending"},
			{Step: "create_deployment", Status: "pending"},
			{Step: "create_service", Status: "pending"},
		},
	})
}

func (s *Server) handleGetDeployStatus(c *gin.Context) {
	deployID := c.Param("deploy_id")

	// TODO: Implement actual status retrieval from data store

	c.JSON(http.StatusOK, models.DeployStatus{
		DeployID: deployID,
		Status:   "pending",
		Steps:    []models.DeployStep{},
	})
}

func (s *Server) handleGetDeployHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 50
	}

	history, err := s.data.GetDeployHistory(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DATA_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deployments": history,
		"total":       len(history),
	})
}
