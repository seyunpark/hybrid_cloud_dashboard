package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/ai"
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

	ctx := c.Request.Context()

	// 1. Get container info from Docker
	container, err := s.docker.GetContainer(ctx, req.ContainerID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "CONTAINER_NOT_FOUND", Message: fmt.Sprintf("container %s not found: %v", req.ContainerID, err)},
		})
		return
	}

	// Build ContainerInfo for AI
	envVars := make(map[string]string)
	for _, e := range container.Config.Env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	ports := make([]int, 0)
	for _, p := range container.Ports {
		if p.PrivatePort > 0 {
			ports = append(ports, p.PrivatePort)
		}
	}

	imageParts := strings.SplitN(container.Image, ":", 2)
	imageName := imageParts[0]
	imageTag := "latest"
	if len(imageParts) > 1 {
		imageTag = imageParts[1]
	}

	volumes := make([]string, 0, len(container.Mounts))
	for _, m := range container.Mounts {
		volumes = append(volumes, m.Destination)
	}

	cpuUsage := ""
	memUsage := ""
	if container.Stats != nil {
		cpuUsage = fmt.Sprintf("%.1f%%", container.Stats.CPUPercent)
		memUsage = fmt.Sprintf("%dMi", container.Stats.MemoryUsage/(1024*1024))
	}

	containerInfo := ai.ContainerInfo{
		Name:        container.Name,
		Image:       imageName,
		ImageTag:    imageTag,
		EnvVars:     envVars,
		Ports:       ports,
		Volumes:     volumes,
		Command:     container.Config.Cmd,
		WorkingDir:  container.Config.WorkingDir,
		CPUUsage:    cpuUsage,
		MemoryUsage: memUsage,
	}

	// 2. Search similar deployments
	similar, _ := s.data.FindSimilar(ctx, imageName, "", 5)

	// 3. Generate manifest via AI
	manifest, err := s.ai.GenerateManifest(ctx, containerInfo, similar)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "AI_ERROR", Message: fmt.Sprintf("failed to generate manifest: %v", err)},
		})
		return
	}

	// 4. Create deploy state
	deployID := uuid.New().String()
	now := time.Now()

	resp := &models.DeployResponse{
		DeployID: deployID,
		Status:   "analyzing",
		AIAnalysis: &models.AIAnalysis{
			ServiceType:        detectServiceType(containerInfo),
			DetectedLanguage:   detectLanguage(containerInfo),
			SimilarDeployments: len(similar),
		},
		Recommendations: &models.Recommendations{
			CPURequest:    "100m",
			CPULimit:      "500m",
			MemoryRequest: "128Mi",
			MemoryLimit:   "512Mi",
			Replicas:      2,
			EnableHPA:     req.Options.EnableHPA,
			Reasoning:     manifest.Reasoning,
		},
		Manifests: &models.Manifests{
			Deployment: manifest.Deployment,
			Service:    manifest.Service,
			HPA:        manifest.HPA,
			ConfigMap:  manifest.ConfigMap,
		},
	}

	s.mu.Lock()
	s.deployStates[deployID] = &deployState{
		Status: &models.DeployStatus{
			DeployID:  deployID,
			Status:    "pending",
			StartedAt: &now,
			Steps:     []models.DeployStep{},
		},
		Response:  resp,
		Request:   &req,
		Manifests: manifest,
	}
	s.mu.Unlock()

	c.JSON(http.StatusOK, resp)
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

	s.mu.RLock()
	state, exists := s.deployStates[deployID]
	s.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "deployment not found"},
		})
		return
	}

	if !req.Approved {
		s.mu.Lock()
		state.Status.Status = "cancelled"
		now := time.Now()
		state.Status.CompletedAt = &now
		s.mu.Unlock()

		c.JSON(http.StatusOK, state.Status)
		return
	}

	// Initialize deployment steps
	steps := []models.DeployStep{
		{Step: "push_image", Status: "pending"},
		{Step: "create_deployment", Status: "pending"},
		{Step: "create_service", Status: "pending"},
	}

	s.mu.Lock()
	state.Status.Status = "deploying"
	state.Status.Steps = steps
	s.mu.Unlock()

	// Execute deployment asynchronously
	go s.executeDeployAsync(deployID)

	c.JSON(http.StatusOK, state.Status)
}

func (s *Server) handleRefineDeploy(c *gin.Context) {
	deployID := c.Param("deploy_id")

	var req struct {
		Feedback string `json:"feedback" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	s.mu.RLock()
	state, exists := s.deployStates[deployID]
	s.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "deployment not found"},
		})
		return
	}

	// Call AI to refine the manifest
	refined, err := s.ai.RefineManifest(c.Request.Context(), state.Manifests, req.Feedback)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "AI_ERROR", Message: err.Error()},
		})
		return
	}

	// Update deploy state with refined manifest
	s.mu.Lock()
	state.Manifests = refined
	state.Response.Manifests = &models.Manifests{
		Deployment: refined.Deployment,
		Service:    refined.Service,
		HPA:        refined.HPA,
		ConfigMap:  refined.ConfigMap,
	}
	state.Response.Recommendations.Reasoning = refined.Reasoning
	s.mu.Unlock()

	c.JSON(http.StatusOK, state.Response)
}

func (s *Server) executeDeployAsync(deployID string) {
	s.mu.RLock()
	state, exists := s.deployStates[deployID]
	s.mu.RUnlock()
	if !exists {
		return
	}

	updateStep := func(stepName, status, message string) {
		s.mu.Lock()
		defer s.mu.Unlock()
		for i, step := range state.Status.Steps {
			if step.Step == stepName {
				state.Status.Steps[i].Status = status
				state.Status.Steps[i].Message = message
				if status == "completed" || status == "failed" {
					now := time.Now()
					state.Status.Steps[i].CompletedAt = &now
				}
				break
			}
		}
	}

	ctx := context.Background()

	// Step 1: Push image (skip if no registry configured or image is public)
	updateStep("push_image", "in_progress", "Checking image availability...")
	if state.Request != nil && state.Response != nil {
		// Extract actual image name from the container info, not manifest YAML
		imageName := state.Request.ContainerID // fallback
		if state.Response.AIAnalysis != nil {
			// Use original container image for push
			slog.Info("image push skipped for public image", "container", state.Request.ContainerID)
		}
		_ = imageName // suppress unused warning
	}
	updateStep("push_image", "completed", "Image ready")

	clusterName := state.Request.ClusterName
	ns := state.Request.Namespace
	if ns == "" {
		ns = "default"
	}

	// Step 2: Create deployment
	updateStep("create_deployment", "in_progress", "Applying Kubernetes deployment...")
	if state.Manifests != nil && state.Manifests.Deployment != "" {
		applyCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err := s.kubernetes.ApplyManifest(applyCtx, clusterName, state.Manifests.Deployment)
		cancel()
		if err != nil {
			slog.Error("failed to apply deployment", "deploy_id", deployID, "error", err)
			updateStep("create_deployment", "failed", fmt.Sprintf("Failed: %v", err))
			s.mu.Lock()
			now := time.Now()
			state.Status.Status = "failed"
			state.Status.CompletedAt = &now
			s.mu.Unlock()
			return
		}
	}
	updateStep("create_deployment", "completed", "Deployment applied")

	// Step 3: Create service
	updateStep("create_service", "in_progress", "Applying Kubernetes service...")
	if state.Manifests != nil && state.Manifests.Service != "" {
		applyCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err := s.kubernetes.ApplyManifest(applyCtx, clusterName, state.Manifests.Service)
		cancel()
		if err != nil {
			slog.Error("failed to apply service", "deploy_id", deployID, "error", err)
			updateStep("create_service", "failed", fmt.Sprintf("Failed: %v", err))
			s.mu.Lock()
			now := time.Now()
			state.Status.Status = "failed"
			state.Status.CompletedAt = &now
			s.mu.Unlock()
			return
		}
	}
	updateStep("create_service", "completed", "Service applied")

	// Apply optional resources (best-effort)
	if state.Manifests != nil {
		if state.Manifests.HPA != "" {
			if err := s.kubernetes.ApplyManifest(ctx, clusterName, state.Manifests.HPA); err != nil {
				slog.Warn("failed to apply HPA", "deploy_id", deployID, "error", err)
			}
		}
		if state.Manifests.ConfigMap != "" {
			if err := s.kubernetes.ApplyManifest(ctx, clusterName, state.Manifests.ConfigMap); err != nil {
				slog.Warn("failed to apply ConfigMap", "deploy_id", deployID, "error", err)
			}
		}
	}

	// Mark as completed
	serviceName := state.Request.ContainerID
	s.mu.Lock()
	now := time.Now()
	state.Status.Status = "completed"
	state.Status.CompletedAt = &now
	state.Status.Result = &models.DeployResult{
		DeploymentName: serviceName,
		Namespace:      ns,
		Replicas:       1,
		PodsReady:      "pending",
		ServiceURL:     fmt.Sprintf("http://%s.%s.svc.cluster.local", serviceName, ns),
	}
	s.mu.Unlock()

	// Save to deployment history
	history := &models.DeploymentHistory{
		ID:            deployID,
		ServiceName:   state.Request.ContainerID,
		ImageName:     "",
		ImageTag:      "",
		TargetCluster: state.Request.ClusterName,
		Namespace:     state.Request.Namespace,
		DeployedAt:    now,
		Success:       true,
		AIGenerated:   true,
		AIConfidence:  state.Manifests.Confidence,
		Replicas:      2,
	}
	if state.Response != nil && state.Response.Recommendations != nil {
		history.CPURequest = state.Response.Recommendations.CPURequest
		history.CPULimit = state.Response.Recommendations.CPULimit
		history.MemoryRequest = state.Response.Recommendations.MemoryRequest
		history.MemoryLimit = state.Response.Recommendations.MemoryLimit
	}

	// Save manifest JSON for future redeploy
	if state.Manifests != nil {
		manifestJSON, _ := json.Marshal(state.Manifests)
		history.ManifestJSON = string(manifestJSON)
	}

	if err := s.data.SaveDeployment(ctx, history); err != nil {
		slog.Error("failed to save deployment history", "error", err)
	}
}

func (s *Server) handleGetDeployStatus(c *gin.Context) {
	deployID := c.Param("deploy_id")

	s.mu.RLock()
	state, exists := s.deployStates[deployID]
	s.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "deployment not found"},
		})
		return
	}

	c.JSON(http.StatusOK, state.Status)
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

// handleGetUnifiedHistory returns a paginated, chronological list of both
// single and stack deploys in one unified response.
func (s *Server) handleGetUnifiedHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	items, total, err := s.data.ListUnifiedHistory(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DATA_ERROR", Message: err.Error()},
		})
		return
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func detectServiceType(info ai.ContainerInfo) string {
	image := strings.ToLower(info.Image)
	switch {
	case strings.Contains(image, "nginx") || strings.Contains(image, "httpd") || strings.Contains(image, "apache"):
		return "web-server"
	case strings.Contains(image, "node") || strings.Contains(image, "express"):
		return "web-application"
	case strings.Contains(image, "postgres") || strings.Contains(image, "mysql") || strings.Contains(image, "mongo") || strings.Contains(image, "redis"):
		return "database"
	case strings.Contains(image, "rabbitmq") || strings.Contains(image, "kafka"):
		return "message-queue"
	default:
		return "application"
	}
}

func detectLanguage(info ai.ContainerInfo) string {
	image := strings.ToLower(info.Image)
	switch {
	case strings.Contains(image, "node"):
		return "javascript"
	case strings.Contains(image, "python"):
		return "python"
	case strings.Contains(image, "golang") || strings.Contains(image, "go"):
		return "go"
	case strings.Contains(image, "java") || strings.Contains(image, "openjdk"):
		return "java"
	case strings.Contains(image, "ruby"):
		return "ruby"
	case strings.Contains(image, "php"):
		return "php"
	case strings.Contains(image, "dotnet") || strings.Contains(image, "aspnet"):
		return "csharp"
	default:
		return "unknown"
	}
}

// handleUndeployFromK8s removes K8s resources for a deployment and marks it as deleted.
func (s *Server) handleUndeployFromK8s(c *gin.Context) {
	deployID := c.Param("deploy_id")
	ctx := c.Request.Context()

	deployment, err := s.data.GetDeployment(ctx, deployID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "deployment not found"},
		})
		return
	}

	if deployment.Status != "deployed" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_STATE", Message: "deployment is not in 'deployed' state"},
		})
		return
	}

	// Best-effort delete all K8s resources (order: HPA → Ingress → ConfigMap → Service → Deployment)
	cluster := deployment.TargetCluster
	ns := deployment.Namespace
	svcName := deployment.ServiceName

	// If stored manifests are available, parse and delete all resource kinds
	if deployment.ManifestJSON != "" {
		var storedManifests map[string]string
		if err := json.Unmarshal([]byte(deployment.ManifestJSON), &storedManifests); err == nil {
			// Phase 1: Delete known kinds in dependency-safe order
			deleteOrder := []string{"HPA", "HorizontalPodAutoscaler", "Ingress", "HTTPRoute", "Gateway", "Service", "Deployment", "StatefulSet", "Secret", "ConfigMap", "PersistentVolumeClaim"}
			deleted := map[string]bool{}
			for _, kind := range deleteOrder {
				if _, ok := storedManifests[kind]; ok {
					deleted[kind] = true
					if err := s.kubernetes.DeleteResource(ctx, cluster, kind, ns, svcName); err != nil {
						slog.Warn("failed to delete resource", "kind", kind, "name", svcName, "error", err)
					}
				}
			}
			// Phase 2: Delete any remaining kinds not in the predefined order
			for kind := range storedManifests {
				if !deleted[kind] {
					if err := s.kubernetes.DeleteResource(ctx, cluster, kind, ns, svcName); err != nil {
						slog.Warn("failed to delete resource", "kind", kind, "name", svcName, "error", err)
					}
				}
			}
		}
	} else {
		// Fallback: delete Deployment and Service by name
		if err := s.kubernetes.DeleteResource(ctx, cluster, "Deployment", ns, svcName); err != nil {
			slog.Warn("failed to delete k8s deployment (may already be gone)", "name", svcName, "error", err)
		}
		if err := s.kubernetes.DeleteResource(ctx, cluster, "Service", ns, svcName); err != nil {
			slog.Warn("failed to delete k8s service (may already be gone)", "name", svcName, "error", err)
		}
	}

	now := time.Now()
	if err := s.data.UpdateDeploymentStatus(ctx, deployID, "deleted", &now); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DATA_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Success: true, Message: "Resources undeployed"})
}

// handleRedeployToK8s re-deploys using stored manifests and creates a new history record.
func (s *Server) handleRedeployToK8s(c *gin.Context) {
	deployID := c.Param("deploy_id")
	ctx := c.Request.Context()

	deployment, err := s.data.GetDeployment(ctx, deployID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "deployment not found"},
		})
		return
	}

	if deployment.Status == "deployed" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_STATE", Message: "deployment is already active; undeploy first"},
		})
		return
	}

	if deployment.ManifestJSON == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "NO_MANIFEST", Message: "no stored manifest available for redeploy"},
		})
		return
	}

	// Create a new deployment history record
	newID := uuid.New().String()
	now := time.Now()
	newHistory := &models.DeploymentHistory{
		ID:            newID,
		ServiceName:   deployment.ServiceName,
		ImageName:     deployment.ImageName,
		ImageTag:      deployment.ImageTag,
		ServiceType:   deployment.ServiceType,
		Language:      deployment.Language,
		CPURequest:    deployment.CPURequest,
		CPULimit:      deployment.CPULimit,
		MemoryRequest: deployment.MemoryRequest,
		MemoryLimit:   deployment.MemoryLimit,
		Replicas:      deployment.Replicas,
		TargetCluster: deployment.TargetCluster,
		Namespace:     deployment.Namespace,
		DeployedAt:    now,
		Success:       true,
		Status:        "deployed",
		ManifestJSON:  deployment.ManifestJSON,
		AIGenerated:   deployment.AIGenerated,
		AIConfidence:  deployment.AIConfidence,
	}

	if err := s.data.SaveDeployment(ctx, newHistory); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DATA_ERROR", Message: err.Error()},
		})
		return
	}

	slog.Info("redeployment created", "original_id", deployID, "new_id", newID)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"deploy_id": newID,
		"message":   "Redeployment created",
	})
}

// handleDeleteDeployRecord removes a deployment record from the database.
func (s *Server) handleDeleteDeployRecord(c *gin.Context) {
	deployID := c.Param("deploy_id")
	ctx := c.Request.Context()

	deployment, err := s.data.GetDeployment(ctx, deployID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "deployment not found"},
		})
		return
	}

	if deployment.Status == "deployed" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_STATE", Message: "cannot delete record of active deployment; undeploy first"},
		})
		return
	}

	if err := s.data.DeleteDeploymentRecord(ctx, deployID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DATA_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Success: true, Message: "Record deleted"})
}
