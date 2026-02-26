package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"
)

func (s *Server) handleListClusters(c *gin.Context) {
	clusters, err := s.kubernetes.ListClusters(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "K8S_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"clusters": clusters})
}

func (s *Server) handleListPods(c *gin.Context) {
	cluster := c.Param("cluster")
	namespace := c.DefaultQuery("namespace", "default")
	label := c.Query("label")

	pods, err := s.kubernetes.ListPods(c.Request.Context(), cluster, namespace, label)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "K8S_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pods": pods})
}

func (s *Server) handleListDeployments(c *gin.Context) {
	cluster := c.Param("cluster")
	namespace := c.DefaultQuery("namespace", "default")

	deployments, err := s.kubernetes.ListDeployments(c.Request.Context(), cluster, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "K8S_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deployments": deployments})
}

func (s *Server) handleListServices(c *gin.Context) {
	cluster := c.Param("cluster")
	namespace := c.DefaultQuery("namespace", "default")

	services, err := s.kubernetes.ListServices(c.Request.Context(), cluster, namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "K8S_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"services": services})
}

func (s *Server) handleScaleDeployment(c *gin.Context) {
	cluster := c.Param("cluster")
	namespace := c.Param("ns")
	name := c.Param("name")

	var req struct {
		Replicas int `json:"replicas" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	if err := s.kubernetes.ScaleDeployment(c.Request.Context(), cluster, namespace, name, req.Replicas); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "K8S_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Deployment scaled to " + strconv.Itoa(req.Replicas) + " replicas",
		"deployment": gin.H{
			"name":      name,
			"namespace": namespace,
			"replicas":  req.Replicas,
		},
	})
}

func (s *Server) handleRestartPod(c *gin.Context) {
	cluster := c.Param("cluster")
	namespace := c.Param("ns")
	name := c.Param("name")

	if err := s.kubernetes.RestartPod(c.Request.Context(), cluster, namespace, name); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "K8S_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Pod deleted for restart",
	})
}
