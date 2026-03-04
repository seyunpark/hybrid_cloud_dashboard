package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/ai"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"
)

// namespaceRegex matches "namespace: <value>" lines in YAML (indented, i.e. under metadata).
var namespaceRegex = regexp.MustCompile(`(?m)^(\s+namespace:\s+)\S+`)

// replaceNamespaceInManifests updates metadata.namespace in all manifest YAMLs.
func replaceNamespaceInManifests(manifests map[string]map[string]string, newNamespace string) {
	for kind, resources := range manifests {
		if kind == "Namespace" {
			continue // Don't modify Namespace resource itself
		}
		for name, yamlStr := range resources {
			resources[name] = namespaceRegex.ReplaceAllString(yamlStr, "${1}"+newNamespace)
		}
	}
}

// stackDeployState holds in-memory state for an active stack deployment.
type stackDeployState struct {
	Status         *models.StackDeployStatus
	Response       *models.StackDeployResponse
	Request        *models.StackDeployRequest
	Manifests      *ai.StackManifestResult
	ContainerInfos []ai.ContainerInfo
}

// --- helpers ---

// saveStackDeployToDB persists the current in-memory state to the DB.
func (s *Server) saveStackDeployToDB(ctx context.Context, state *stackDeployState) {
	record := s.stateToRecord(state)
	if err := s.data.SaveStackDeploy(ctx, record); err != nil {
		slog.Error("failed to save stack deploy to DB", "deploy_id", record.DeployID, "error", err)
	}
}

// updateStackDeployInDB updates an existing DB record from in-memory state.
func (s *Server) updateStackDeployInDB(ctx context.Context, state *stackDeployState) {
	record := s.stateToRecord(state)
	if err := s.data.UpdateStackDeploy(ctx, record); err != nil {
		slog.Error("failed to update stack deploy in DB", "deploy_id", record.DeployID, "error", err)
	}
}

func (s *Server) stateToRecord(state *stackDeployState) *models.StackDeployRecord {
	containerIDs := []string{}
	clusterName := ""
	namespace := "default"
	if state.Request != nil {
		containerIDs = state.Request.ContainerIDs
		clusterName = state.Request.ClusterName
		namespace = state.Request.Namespace
	}

	topologyJSON := ""
	manifestsJSON := ""
	reasoning := ""
	confidence := 0.0
	if state.Response != nil {
		if state.Response.Topology != nil {
			if b, err := json.Marshal(state.Response.Topology); err == nil {
				topologyJSON = string(b)
			}
		}
		if state.Response.Manifests != nil {
			if b, err := json.Marshal(state.Response.Manifests); err == nil {
				manifestsJSON = string(b)
			}
		}
		reasoning = state.Response.Reasoning
		confidence = state.Response.Confidence
	}

	servicesJSON := ""
	if state.Status.Services != nil {
		if b, err := json.Marshal(state.Status.Services); err == nil {
			servicesJSON = string(b)
		}
	}

	return &models.StackDeployRecord{
		DeployID:      state.Status.DeployID,
		StackName:     state.Status.StackName,
		ClusterName:   clusterName,
		Namespace:     namespace,
		ContainerIDs:  containerIDs,
		Status:        state.Status.Status,
		StartedAt:     state.Status.StartedAt,
		CompletedAt:   state.Status.CompletedAt,
		TopologyJSON:  topologyJSON,
		ManifestsJSON: manifestsJSON,
		Reasoning:     reasoning,
		Confidence:    confidence,
		DeployOrder:   state.Status.DeployOrder,
		ServicesJSON:  servicesJSON,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// recordToDetailResponse converts a DB record to the API detail response format.
func recordToDetailResponse(record *models.StackDeployRecord) gin.H {
	resp := &models.StackDeployResponse{
		DeployID:   record.DeployID,
		Status:     record.Status,
		StackName:  record.StackName,
		Reasoning:  record.Reasoning,
		Confidence: record.Confidence,
	}

	if record.TopologyJSON != "" {
		var topo models.StackTopology
		if err := json.Unmarshal([]byte(record.TopologyJSON), &topo); err == nil {
			resp.Topology = &topo
		}
	}
	if record.ManifestsJSON != "" {
		var manifests models.StackManifests
		if err := json.Unmarshal([]byte(record.ManifestsJSON), &manifests); err == nil {
			resp.Manifests = manifests
		}
	}

	status := &models.StackDeployStatus{
		DeployID:    record.DeployID,
		Status:      record.Status,
		StackName:   record.StackName,
		StartedAt:   record.StartedAt,
		CompletedAt: record.CompletedAt,
		DeployOrder: record.DeployOrder,
		Services:    map[string]*models.ServiceDeployStatus{},
	}

	if record.ServicesJSON != "" {
		var svcs map[string]*models.ServiceDeployStatus
		if err := json.Unmarshal([]byte(record.ServicesJSON), &svcs); err == nil {
			status.Services = svcs
		}
	}

	return gin.H{
		"response":     resp,
		"status":       status,
		"cluster_name": record.ClusterName,
		"namespace":    record.Namespace,
	}
}

// RestoreStackDeploy restores an in-memory state from a DB record (used at startup).
func (s *Server) RestoreStackDeploy(record *models.StackDeployRecord) {
	resp := &models.StackDeployResponse{
		DeployID:   record.DeployID,
		Status:     record.Status,
		StackName:  record.StackName,
		Reasoning:  record.Reasoning,
		Confidence: record.Confidence,
	}
	if record.TopologyJSON != "" {
		var topo models.StackTopology
		if err := json.Unmarshal([]byte(record.TopologyJSON), &topo); err == nil {
			resp.Topology = &topo
		}
	}
	if record.ManifestsJSON != "" {
		var manifests models.StackManifests
		if err := json.Unmarshal([]byte(record.ManifestsJSON), &manifests); err == nil {
			resp.Manifests = manifests
		}
	}

	status := &models.StackDeployStatus{
		DeployID:    record.DeployID,
		Status:      record.Status,
		StackName:   record.StackName,
		StartedAt:   record.StartedAt,
		CompletedAt: record.CompletedAt,
		DeployOrder: record.DeployOrder,
		Services:    map[string]*models.ServiceDeployStatus{},
	}
	if record.ServicesJSON != "" {
		var svcs map[string]*models.ServiceDeployStatus
		if err := json.Unmarshal([]byte(record.ServicesJSON), &svcs); err == nil {
			status.Services = svcs
		}
	}

	req := &models.StackDeployRequest{
		ContainerIDs: record.ContainerIDs,
		ClusterName:  record.ClusterName,
		Namespace:    record.Namespace,
		StackName:    record.StackName,
	}

	s.mu.Lock()
	s.stackDeployStates[record.DeployID] = &stackDeployState{
		Status:   status,
		Response: resp,
		Request:  req,
	}
	s.mu.Unlock()
}

// --- handlers ---

func (s *Server) handleDeployStack(c *gin.Context) {
	var req models.StackDeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	if len(req.ContainerIDs) < 2 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: "at least 2 containers required for stack deployment"},
		})
		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	ctx := c.Request.Context()

	// Gather container info synchronously (fast Docker API calls)
	containerInfos := make([]ai.ContainerInfo, 0, len(req.ContainerIDs))
	for _, containerID := range req.ContainerIDs {
		container, err := s.docker.GetContainer(ctx, containerID)
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{Code: "CONTAINER_NOT_FOUND", Message: fmt.Sprintf("container %s not found: %v", containerID, err)},
			})
			return
		}

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

		containerInfos = append(containerInfos, ai.ContainerInfo{
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
		})
	}

	// Auto-generate stack name if not provided
	if req.StackName == "" {
		req.StackName = containerInfos[0].Name + "-stack"
	}

	deployID := uuid.New().String()
	now := time.Now()

	// Create initial state with "generating" status — respond immediately
	resp := &models.StackDeployResponse{
		DeployID:  deployID,
		Status:    "generating",
		StackName: req.StackName,
	}

	state := &stackDeployState{
		Status: &models.StackDeployStatus{
			DeployID:    deployID,
			Status:      "generating",
			StackName:   req.StackName,
			StartedAt:   &now,
			Services:    map[string]*models.ServiceDeployStatus{},
			DeployOrder: []string{},
		},
		Response:       resp,
		Request:        &req,
		ContainerInfos: containerInfos,
	}

	s.mu.Lock()
	s.stackDeployStates[deployID] = state
	s.mu.Unlock()

	// Persist to DB
	s.saveStackDeployToDB(ctx, state)

	// Launch AI manifest generation in background
	go s.generateStackManifestAsync(deployID, req, containerInfos)

	c.JSON(http.StatusOK, resp)
}

// generateStackManifestAsync runs AI manifest generation in a goroutine.
func (s *Server) generateStackManifestAsync(deployID string, req models.StackDeployRequest, containerInfos []ai.ContainerInfo) {
	stackInfo := ai.StackContainerInfo{
		StackName:  req.StackName,
		Containers: containerInfos,
		Namespace:  req.Namespace,
	}

	similar, _ := s.data.FindSimilar(context.Background(), "", "", 5)

	aiCtx, aiCancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer aiCancel()

	manifest, err := s.ai.GenerateStackManifest(aiCtx, stackInfo, similar)
	if err != nil {
		slog.Error("async stack manifest generation failed", "deploy_id", deployID, "error", err)
		s.mu.Lock()
		state, ok := s.stackDeployStates[deployID]
		if ok {
			state.Status.Status = "failed"
			now := time.Now()
			state.Status.CompletedAt = &now
			state.Response.Status = "failed"
		}
		s.mu.Unlock()
		if ok {
			s.updateStackDeployInDB(context.Background(), state)
		}
		return
	}

	s.mu.Lock()
	state, exists := s.stackDeployStates[deployID]
	if !exists {
		s.mu.Unlock()
		return
	}

	state.Manifests = manifest
	state.Response.Status = "analyzing"
	state.Response.Topology = &manifest.Topology
	state.Response.Manifests = models.StackManifests(manifest.Manifests)
	state.Response.Reasoning = manifest.Reasoning
	state.Response.Confidence = manifest.Confidence

	// Inject Namespace manifest if createNamespace was requested at modal time
	if req.CreateNamespace && req.Namespace != "" && req.Namespace != "default" {
		nsYAML := fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    app.kubernetes.io/managed-by: hybrid-cloud-dashboard
    stack: %s`, req.Namespace, req.StackName)

		if manifest.Manifests == nil {
			manifest.Manifests = make(map[string]map[string]string)
		}
		if manifest.Manifests["Namespace"] == nil {
			manifest.Manifests["Namespace"] = make(map[string]string)
		}
		manifest.Manifests["Namespace"][req.Namespace] = nsYAML

		// Prepend _namespace to deploy order
		hasNs := false
		for _, n := range manifest.Topology.DeployOrder {
			if n == "_namespace" {
				hasNs = true
				break
			}
		}
		if !hasNs {
			manifest.Topology.DeployOrder = append([]string{"_namespace"}, manifest.Topology.DeployOrder...)
			manifest.Topology.Services = append([]models.StackServiceInfo{{
				ContainerID: "",
				ServiceName: "_namespace",
				ServiceType: "namespace",
				Image:       "",
			}}, manifest.Topology.Services...)
		}

		state.Response.Topology = &manifest.Topology
		state.Response.Manifests = models.StackManifests(manifest.Manifests)
	}

	state.Status.Status = "pending"
	state.Status.DeployOrder = manifest.Topology.DeployOrder
	svcStatuses := make(map[string]*models.ServiceDeployStatus)
	for _, svcName := range manifest.Topology.DeployOrder {
		svcStatuses[svcName] = &models.ServiceDeployStatus{
			ServiceName: svcName,
			Status:      "pending",
			Steps:       []models.DeployStep{},
		}
	}
	state.Status.Services = svcStatuses
	s.mu.Unlock()

	// Persist AI results to DB
	s.updateStackDeployInDB(context.Background(), state)

	slog.Info("stack manifest generated", "deploy_id", deployID, "services", len(manifest.Topology.DeployOrder))
}

func (s *Server) handleRefineStackDeploy(c *gin.Context) {
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
	state, exists := s.stackDeployStates[deployID]
	s.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "stack deployment not found"},
		})
		return
	}

	// Reconstruct Manifests if nil (can happen after DB restore)
	if state.Manifests == nil && state.Response != nil && state.Response.Manifests != nil {
		s.mu.Lock()
		state.Manifests = &ai.StackManifestResult{
			Manifests:  map[string]map[string]string(state.Response.Manifests),
			Reasoning:  state.Response.Reasoning,
			Confidence: state.Response.Confidence,
		}
		if state.Response.Topology != nil {
			state.Manifests.Topology = *state.Response.Topology
		}
		s.mu.Unlock()
	}

	if state.Manifests == nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "NO_MANIFESTS", Message: "매니페스트가 없어 수정 요청을 처리할 수 없습니다"},
		})
		return
	}

	refined, err := s.ai.RefineStackManifest(c.Request.Context(), state.Manifests, req.Feedback)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := err.Error()
		if strings.Contains(errMsg, "API key") || strings.Contains(errMsg, "not configured") {
			statusCode = http.StatusUnprocessableEntity
		}
		c.JSON(statusCode, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "AI_ERROR", Message: errMsg},
		})
		return
	}

	s.mu.Lock()
	state.Manifests = refined
	state.Response.Topology = &refined.Topology
	state.Response.Manifests = models.StackManifests(refined.Manifests)
	state.Response.Reasoning = refined.Reasoning
	state.Response.Confidence = refined.Confidence

	// Update deploy order in status
	state.Status.DeployOrder = refined.Topology.DeployOrder
	svcStatuses := make(map[string]*models.ServiceDeployStatus)
	for _, svcName := range refined.Topology.DeployOrder {
		svcStatuses[svcName] = &models.ServiceDeployStatus{
			ServiceName: svcName,
			Status:      "pending",
			Steps:       []models.DeployStep{},
		}
	}
	state.Status.Services = svcStatuses
	s.mu.Unlock()

	// Persist refined manifests to DB
	s.updateStackDeployInDB(c.Request.Context(), state)

	c.JSON(http.StatusOK, state.Response)
}

// handleRegenerateStackDeploy re-runs AI manifest generation from scratch.
func (s *Server) handleRegenerateStackDeploy(c *gin.Context) {
	deployID := c.Param("deploy_id")

	s.mu.RLock()
	state, exists := s.stackDeployStates[deployID]
	s.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "stack deployment not found"},
		})
		return
	}

	if state.Status.Status != "pending" && state.Status.Status != "analyzing" {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_STATUS", Message: fmt.Sprintf("재생성은 pending 상태에서만 가능합니다 (현재: %s)", state.Status.Status)},
		})
		return
	}

	// If container infos are missing (e.g. restored from DB), re-fetch from Docker
	if len(state.ContainerInfos) == 0 && state.Request != nil && len(state.Request.ContainerIDs) > 0 {
		var infos []ai.ContainerInfo
		for _, id := range state.Request.ContainerIDs {
			container, err := s.docker.GetContainer(c.Request.Context(), id)
			if err != nil {
				continue
			}
			envVars := make(map[string]string)
			for _, env := range container.Config.Env {
				parts := strings.SplitN(env, "=", 2)
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
			infos = append(infos, ai.ContainerInfo{
				Name:       container.Name,
				Image:      imageName,
				ImageTag:   imageTag,
				EnvVars:    envVars,
				Ports:      ports,
				Volumes:    volumes,
				Command:    container.Config.Cmd,
				WorkingDir: container.Config.WorkingDir,
			})
		}
		if len(infos) > 0 {
			state.ContainerInfos = infos
		}
	}

	if len(state.ContainerInfos) == 0 {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "NO_CONTAINER_INFO", Message: "컨테이너 정보가 없어 재생성할 수 없습니다. 원본 컨테이너가 실행 중인지 확인해주세요."},
		})
		return
	}

	// Reset state to generating
	s.mu.Lock()
	state.Status.Status = "generating"
	state.Status.CompletedAt = nil
	state.Status.Services = map[string]*models.ServiceDeployStatus{}
	state.Status.DeployOrder = []string{}
	state.Response.Status = "generating"
	state.Response.Topology = nil
	state.Response.Manifests = nil
	state.Response.Reasoning = ""
	state.Response.Confidence = 0
	state.Manifests = nil
	s.mu.Unlock()

	s.updateStackDeployInDB(c.Request.Context(), state)

	// Re-launch AI generation with original request and container infos
	go s.generateStackManifestAsync(deployID, *state.Request, state.ContainerInfos)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "AI 매니페스트 재생성을 시작합니다"})
}

// handleReopenStackDeploy restores an undeployed stack to pending state for manifest editing.
func (s *Server) handleReopenStackDeploy(c *gin.Context) {
	deployID := c.Param("deploy_id")
	ctx := c.Request.Context()

	// Try in-memory first, then DB
	s.mu.RLock()
	state, inMemory := s.stackDeployStates[deployID]
	s.mu.RUnlock()

	if inMemory {
		if state.Status.Status != "undeployed" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorDetail{Code: "INVALID_STATUS", Message: fmt.Sprintf("매니페스트 수정은 undeployed 상태에서만 가능합니다 (현재: %s)", state.Status.Status)},
			})
			return
		}

		// Reset to pending
		s.mu.Lock()
		state.Status.Status = "pending"
		state.Status.CompletedAt = nil
		state.Response.Status = "pending"
		// Reset service statuses
		for _, svcName := range state.Status.DeployOrder {
			if svc, ok := state.Status.Services[svcName]; ok {
				svc.Status = "pending"
				svc.Steps = []models.DeployStep{}
			}
		}
		// Keep cluster/namespace — namespace is baked into manifests, cluster is reused on re-approve
		s.mu.Unlock()

		s.updateStackDeployInDB(ctx, state)

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "매니페스트 수정 모드로 복귀합니다"})
		return
	}

	// DB path
	record, err := s.data.GetStackDeploy(ctx, deployID)
	if err != nil || record == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "stack deployment not found"},
		})
		return
	}

	if record.Status != "undeployed" {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_STATUS", Message: fmt.Sprintf("매니페스트 수정은 undeployed 상태에서만 가능합니다 (현재: %s)", record.Status)},
		})
		return
	}

	// Update DB record — keep cluster/namespace (baked into manifests)
	record.Status = "pending"
	record.CompletedAt = nil
	if err := s.data.UpdateStackDeploy(ctx, record); err != nil {
		slog.Error("failed to update stack deploy for reopen", "deploy_id", deployID, "error", err)
	}

	// Restore to in-memory so it's poll-able
	s.RestoreStackDeploy(record)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "매니페스트 수정 모드로 복귀합니다"})
}

func (s *Server) handleExecuteStackDeploy(c *gin.Context) {
	deployID := c.Param("deploy_id")

	var req models.StackExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: err.Error()},
		})
		return
	}

	s.mu.RLock()
	state, exists := s.stackDeployStates[deployID]
	s.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "stack deployment not found"},
		})
		return
	}

	if !req.Approved {
		s.mu.Lock()
		state.Status.Status = "undeployed"
		now := time.Now()
		state.Status.CompletedAt = &now
		state.Response.Status = "undeployed"
		s.mu.Unlock()

		s.updateStackDeployInDB(c.Request.Context(), state)

		c.JSON(http.StatusOK, state.Status)
		return
	}

	// Use stored cluster/namespace from initial request (set at modal time)
	// Override only if explicitly provided in execute request
	s.mu.Lock()
	if state.Request == nil {
		state.Request = &models.StackDeployRequest{}
	}
	if req.ClusterName != "" {
		state.Request.ClusterName = req.ClusterName
	}
	s.mu.Unlock()

	if state.Request.ClusterName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: "cluster_name is required — select a cluster when creating the stack deploy"},
		})
		return
	}

	// Initialize steps for each service based on actual manifest resource kinds
	s.mu.Lock()
	state.Status.Status = "deploying"
	for _, svcName := range state.Status.DeployOrder {
		if svc, ok := state.Status.Services[svcName]; ok {
			steps := []models.DeployStep{}

			// Special handling for _namespace
			if svcName == "_namespace" {
				steps = []models.DeployStep{
					{Step: "create_namespace", Status: "pending"},
				}
				svc.Steps = steps
				continue
			}

			if state.Manifests != nil {
				for kind, resources := range state.Manifests.Manifests {
					if kind == "Namespace" {
						continue // Namespace is handled by _namespace step
					}
					for resName := range resources {
						if resName == svcName || strings.HasPrefix(resName, svcName+"-") || strings.HasSuffix(resName, "-"+svcName) {
							steps = append(steps, models.DeployStep{
								Step:   fmt.Sprintf("create_%s", strings.ToLower(kind)),
								Status: "pending",
							})
							break
						}
					}
				}
			}
			if len(steps) == 0 {
				steps = []models.DeployStep{
					{Step: "create_deployment", Status: "pending"},
					{Step: "create_service", Status: "pending"},
				}
			}
			svc.Steps = steps
		}
	}
	s.mu.Unlock()

	// Persist deploying status to DB
	s.updateStackDeployInDB(c.Request.Context(), state)

	go s.executeStackDeployAsync(deployID)

	c.JSON(http.StatusOK, state.Status)
}

func (s *Server) handleListActiveStackDeploys(c *gin.Context) {
	// Merge in-memory (real-time) + DB (persisted) deploys
	s.mu.RLock()
	inMemoryIDs := make(map[string]bool)
	result := make([]*models.StackDeployStatus, 0, len(s.stackDeployStates))
	for id, state := range s.stackDeployStates {
		inMemoryIDs[id] = true
		result = append(result, state.Status)
	}
	s.mu.RUnlock()

	// Load from DB to include persisted deploys not in memory
	dbRecords, err := s.data.ListStackDeploys(c.Request.Context(), 100)
	if err != nil {
		slog.Error("failed to list stack deploys from DB", "error", err)
	} else {
		for _, rec := range dbRecords {
			if inMemoryIDs[rec.DeployID] {
				continue // in-memory version is more up-to-date
			}
			status := &models.StackDeployStatus{
				DeployID:    rec.DeployID,
				Status:      rec.Status,
				StackName:   rec.StackName,
				StartedAt:   rec.StartedAt,
				CompletedAt: rec.CompletedAt,
				DeployOrder: rec.DeployOrder,
				Services:    map[string]*models.ServiceDeployStatus{},
			}
			if rec.ServicesJSON != "" {
				var svcs map[string]*models.ServiceDeployStatus
				if err := json.Unmarshal([]byte(rec.ServicesJSON), &svcs); err == nil {
					status.Services = svcs
				}
			}
			result = append(result, status)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"deployments": result,
		"total":       len(result),
	})
}

func (s *Server) handleGetStackDeployDetail(c *gin.Context) {
	deployID := c.Param("deploy_id")

	// Try in-memory first (real-time data)
	s.mu.RLock()
	state, exists := s.stackDeployStates[deployID]
	s.mu.RUnlock()

	if exists {
		resp := gin.H{
			"response": state.Response,
			"status":   state.Status,
		}
		if state.Request != nil {
			resp["cluster_name"] = state.Request.ClusterName
			resp["namespace"] = state.Request.Namespace
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	// Fallback to DB (persisted data — survives restart)
	record, err := s.data.GetStackDeploy(c.Request.Context(), deployID)
	if err != nil || record == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "stack deployment not found"},
		})
		return
	}

	c.JSON(http.StatusOK, recordToDetailResponse(record))
}

func (s *Server) handleDeleteStackDeploy(c *gin.Context) {
	deployID := c.Param("deploy_id")

	// Check in-memory first
	s.mu.Lock()
	state, inMemory := s.stackDeployStates[deployID]
	if inMemory {
		status := state.Status.Status
		// Cannot delete while in progress
		if status == "generating" || status == "deploying" {
			s.mu.Unlock()
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorDetail{Code: "DEPLOY_IN_PROGRESS", Message: "cannot delete a deployment that is in progress"},
			})
			return
		}
		// Cannot delete deployed (must undeploy first)
		if status == "deployed" || status == "completed" {
			s.mu.Unlock()
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorDetail{Code: "UNDEPLOY_REQUIRED", Message: "배포된 상태에서는 먼저 배포 중지를 해주세요"},
			})
			return
		}
		delete(s.stackDeployStates, deployID)
		s.mu.Unlock()
	} else {
		s.mu.Unlock()
		// Check DB record status
		record, err := s.data.GetStackDeploy(c.Request.Context(), deployID)
		if err != nil || record == nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "stack deployment not found"},
			})
			return
		}
		if record.Status == "deployed" || record.Status == "completed" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorDetail{Code: "UNDEPLOY_REQUIRED", Message: "배포된 상태에서는 먼저 배포 중지를 해주세요"},
			})
			return
		}
		if record.Status == "generating" || record.Status == "deploying" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorDetail{Code: "DEPLOY_IN_PROGRESS", Message: "cannot delete a deployment that is in progress"},
			})
			return
		}
	}

	// Delete from DB
	if err := s.data.DeleteStackDeploy(c.Request.Context(), deployID); err != nil {
		slog.Error("failed to delete stack deploy from DB", "deploy_id", deployID, "error", err)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "stack deployment record removed"})
}

func (s *Server) handleGetStackDeployStatus(c *gin.Context) {
	deployID := c.Param("deploy_id")

	// Try in-memory first
	s.mu.RLock()
	state, exists := s.stackDeployStates[deployID]
	s.mu.RUnlock()

	if exists {
		c.JSON(http.StatusOK, state.Status)
		return
	}

	// Fallback to DB
	record, err := s.data.GetStackDeploy(c.Request.Context(), deployID)
	if err != nil || record == nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "stack deployment not found"},
		})
		return
	}

	status := &models.StackDeployStatus{
		DeployID:    record.DeployID,
		Status:      record.Status,
		StackName:   record.StackName,
		StartedAt:   record.StartedAt,
		CompletedAt: record.CompletedAt,
		DeployOrder: record.DeployOrder,
		Services:    map[string]*models.ServiceDeployStatus{},
	}
	if record.ServicesJSON != "" {
		var svcs map[string]*models.ServiceDeployStatus
		if err := json.Unmarshal([]byte(record.ServicesJSON), &svcs); err == nil {
			status.Services = svcs
		}
	}
	c.JSON(http.StatusOK, status)
}

// handleUndeployStack removes K8s resources for a completed stack, then marks it as undeployed.
func (s *Server) handleUndeployStack(c *gin.Context) {
	deployID := c.Param("deploy_id")
	ctx := c.Request.Context()

	// Try in-memory first, fallback to DB
	var clusterName, namespace string
	var deployOrder []string
	var manifests map[string]map[string]string
	var currentStatus string

	s.mu.RLock()
	state, inMemory := s.stackDeployStates[deployID]
	s.mu.RUnlock()

	if inMemory {
		currentStatus = state.Status.Status
		if state.Request != nil {
			clusterName = state.Request.ClusterName
			namespace = state.Request.Namespace
		}
		deployOrder = state.Status.DeployOrder
		if state.Manifests != nil {
			manifests = state.Manifests.Manifests
		}
	} else {
		record, err := s.data.GetStackDeploy(ctx, deployID)
		if err != nil || record == nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "stack deployment not found"},
			})
			return
		}
		currentStatus = record.Status
		clusterName = record.ClusterName
		namespace = record.Namespace
		deployOrder = record.DeployOrder
		if record.ManifestsJSON != "" {
			json.Unmarshal([]byte(record.ManifestsJSON), &manifests)
		}
	}

	if currentStatus != "deployed" && currentStatus != "completed" {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_STATUS", Message: fmt.Sprintf("배포 중지는 deployed 상태에서만 가능합니다 (현재: %s)", currentStatus)},
		})
		return
	}

	// Delete K8s resources for each service (best-effort, reverse order)
	for i := len(deployOrder) - 1; i >= 0; i-- {
		svcName := deployOrder[i]
		slog.Info("undeploying stack service", "deploy_id", deployID, "service", svcName)

		// Delete Deployment resource
		if err := s.kubernetes.DeleteDeployment(ctx, clusterName, namespace, svcName); err != nil {
			slog.Warn("failed to delete deployment", "service", svcName, "error", err)
		}
		// Delete Service resource
		if err := s.kubernetes.DeleteService(ctx, clusterName, namespace, svcName); err != nil {
			slog.Warn("failed to delete service", "service", svcName, "error", err)
		}

		// Delete other resource kinds (ConfigMap, Secret, HPA, etc.) if manifests available
		if manifests != nil {
			for kind, resources := range manifests {
				if kind == "Deployment" || kind == "Service" {
					continue // already handled
				}
				for resName := range resources {
					if resName == svcName || strings.HasPrefix(resName, svcName+"-") || strings.HasSuffix(resName, "-"+svcName) {
						slog.Info("would delete additional resource", "kind", kind, "name", resName)
						// Additional resource deletion would go here when K8s service supports it
					}
				}
			}
		}
	}

	now := time.Now()

	// Update in-memory state
	if inMemory {
		s.mu.Lock()
		state.Status.Status = "undeployed"
		state.Status.CompletedAt = &now
		state.Response.Status = "undeployed"
		s.mu.Unlock()
		s.updateStackDeployInDB(ctx, state)
	} else {
		// Update DB directly
		record, _ := s.data.GetStackDeploy(ctx, deployID)
		if record != nil {
			record.Status = "undeployed"
			record.CompletedAt = &now
			s.data.UpdateStackDeploy(ctx, record)
		}
	}

	// Also mark related deployment_history entries as deleted
	for _, svcName := range deployOrder {
		historyID := fmt.Sprintf("%s_%s", deployID, svcName)
		if err := s.data.UpdateDeploymentStatus(ctx, historyID, "deleted", &now); err != nil {
			slog.Warn("failed to update deployment history status", "id", historyID, "error", err)
		}
	}

	// Remove from in-memory (no longer active)
	if inMemory {
		s.mu.Lock()
		delete(s.stackDeployStates, deployID)
		s.mu.Unlock()
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "스택 배포가 중지되었습니다"})
}

// handleRedeployStack creates a new stack deployment from stored manifests.
func (s *Server) handleRedeployStack(c *gin.Context) {
	deployID := c.Param("deploy_id")
	ctx := c.Request.Context()

	// Only cluster_name is accepted — namespace is already baked into manifests
	var req struct {
		ClusterName string `json:"cluster_name"`
	}
	_ = c.ShouldBindJSON(&req) // ignore errors — field optional

	// Load original deployment (in-memory or DB)
	var record *models.StackDeployRecord

	s.mu.RLock()
	state, inMemory := s.stackDeployStates[deployID]
	s.mu.RUnlock()

	if inMemory {
		record = s.stateToRecord(state)
	} else {
		var err error
		record, err = s.data.GetStackDeploy(ctx, deployID)
		if err != nil || record == nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{Code: "DEPLOY_NOT_FOUND", Message: "stack deployment not found"},
			})
			return
		}
	}

	// Only allow redeploy from undeployed or failed
	if record.Status != "undeployed" && record.Status != "failed" {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_STATUS", Message: fmt.Sprintf("재배포는 undeployed 또는 failed 상태에서만 가능합니다 (현재: %s)", record.Status)},
		})
		return
	}

	if record.ManifestsJSON == "" {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "NO_MANIFESTS", Message: "저장된 매니페스트가 없어 재배포할 수 없습니다"},
		})
		return
	}

	// Cluster is required for redeploy
	clusterName := req.ClusterName
	if clusterName == "" {
		clusterName = record.ClusterName
	}
	if clusterName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "INVALID_REQUEST", Message: "cluster_name is required for redeployment"},
		})
		return
	}
	// Namespace is already baked into manifests from generation time
	ns := record.Namespace
	if ns == "" {
		ns = "default"
	}

	now := time.Now()

	// Parse stored manifests (namespace already set correctly)
	var manifests models.StackManifests
	json.Unmarshal([]byte(record.ManifestsJSON), &manifests)
	var topology models.StackTopology
	json.Unmarshal([]byte(record.TopologyJSON), &topology)

	// Reset the existing deployment state (reuse same deploy ID)
	svcStatuses := make(map[string]*models.ServiceDeployStatus)
	for _, svcName := range record.DeployOrder {
		svcStatuses[svcName] = &models.ServiceDeployStatus{
			ServiceName: svcName,
			Status:      "pending",
			Steps:       []models.DeployStep{},
		}
	}

	reusedState := &stackDeployState{
		Status: &models.StackDeployStatus{
			DeployID:    deployID,
			Status:      "deploying",
			StackName:   record.StackName,
			StartedAt:   &now,
			CompletedAt: nil,
			DeployOrder: record.DeployOrder,
			Services:    svcStatuses,
		},
		Response: &models.StackDeployResponse{
			DeployID:   deployID,
			Status:     "deploying",
			StackName:  record.StackName,
			Topology:   &topology,
			Manifests:  manifests,
			Reasoning:  record.Reasoning,
			Confidence: record.Confidence,
		},
		Request: &models.StackDeployRequest{
			ContainerIDs: record.ContainerIDs,
			ClusterName:  clusterName,
			Namespace:    ns,
			StackName:    record.StackName,
		},
		Manifests: &ai.StackManifestResult{
			Topology:   topology,
			Manifests:  map[string]map[string]string(manifests),
			Reasoning:  record.Reasoning,
			Confidence: record.Confidence,
		},
	}

	// Initialize steps for each service
	for _, svcName := range record.DeployOrder {
		if svc, ok := reusedState.Status.Services[svcName]; ok {
			steps := []models.DeployStep{}
			for kind, resources := range manifests {
				for resName := range resources {
					if resName == svcName || strings.HasPrefix(resName, svcName+"-") || strings.HasSuffix(resName, "-"+svcName) {
						steps = append(steps, models.DeployStep{
							Step:   fmt.Sprintf("create_%s", strings.ToLower(kind)),
							Status: "pending",
						})
						break
					}
				}
			}
			if len(steps) == 0 {
				steps = []models.DeployStep{
					{Step: "create_deployment", Status: "pending"},
					{Step: "create_service", Status: "pending"},
				}
			}
			svc.Steps = steps
		}
	}

	s.mu.Lock()
	s.stackDeployStates[deployID] = reusedState
	s.mu.Unlock()

	// Update DB
	s.updateStackDeployInDB(ctx, reusedState)

	// Execute deployment
	go s.executeStackDeployAsync(deployID)

	c.JSON(http.StatusOK, gin.H{
		"deploy_id":  deployID,
		"status":     "deploying",
		"stack_name": record.StackName,
		"message":    "재배포가 시작되었습니다",
	})
}

func (s *Server) executeStackDeployAsync(deployID string) {
	s.mu.RLock()
	state, exists := s.stackDeployStates[deployID]
	s.mu.RUnlock()
	if !exists {
		return
	}

	updateStep := func(svcName, stepName, status, message string) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if svc, ok := state.Status.Services[svcName]; ok {
			for i, step := range svc.Steps {
				if step.Step == stepName {
					svc.Steps[i].Status = status
					svc.Steps[i].Message = message
					if status == "completed" || status == "failed" {
						now := time.Now()
						svc.Steps[i].CompletedAt = &now
					}
					break
				}
			}
			if status == "in_progress" {
				svc.Status = "deploying"
			} else if status == "completed" {
				allDone := true
				for _, step := range svc.Steps {
					if step.Status != "completed" {
						allDone = false
						break
					}
				}
				if allDone {
					svc.Status = "deployed"
				}
			} else if status == "failed" {
				svc.Status = "failed"
			}
		}
	}

	ctx := context.Background()

	// Deploy each service in order — execute steps dynamically
	for _, svcName := range state.Status.DeployOrder {
		slog.Info("deploying stack service", "deploy_id", deployID, "service", svcName)

		s.mu.RLock()
		svc := state.Status.Services[svcName]
		steps := make([]models.DeployStep, len(svc.Steps))
		copy(steps, svc.Steps)
		s.mu.RUnlock()

		for _, step := range steps {
			label := strings.TrimPrefix(step.Step, "create_")
			updateStep(svcName, step.Step, "in_progress", fmt.Sprintf("Creating %s...", label))
			time.Sleep(1 * time.Second) // Simulated
			updateStep(svcName, step.Step, "completed", fmt.Sprintf("%s created", label))
		}

		// Update services_json in DB after each service completes
		s.updateStackDeployInDB(ctx, state)

		// Skip deployment_history for _namespace (it's not a real service)
		if svcName == "_namespace" {
			continue
		}

		// Extract this service's manifests for storage
		svcManifest := make(map[string]string)
		if state.Manifests != nil {
			for kind, resources := range state.Manifests.Manifests {
				for resName, yaml := range resources {
					if resName == svcName || strings.HasPrefix(resName, svcName+"-") || strings.HasSuffix(resName, "-"+svcName) {
						svcManifest[kind] = yaml
					}
				}
			}
		}
		manifestJSON, _ := json.Marshal(svcManifest)

		var targetCluster, ns string
		if state.Request != nil {
			targetCluster = state.Request.ClusterName
			ns = state.Request.Namespace
		}
		history := &models.DeploymentHistory{
			ID:            fmt.Sprintf("%s_%s", deployID, svcName),
			ServiceName:   svcName,
			TargetCluster: targetCluster,
			Namespace:     ns,
			DeployedAt:    time.Now(),
			Success:       true,
			Status:        "deployed",
			ManifestJSON:  string(manifestJSON),
			AIGenerated:   true,
			AIConfidence:  state.Response.Confidence,
			Replicas:      1,
		}
		if err := s.data.SaveDeployment(ctx, history); err != nil {
			slog.Error("failed to save stack service deployment history", "service", svcName, "error", err)
		}
	}

	// Mark entire stack as deployed
	s.mu.Lock()
	now := time.Now()
	state.Status.Status = "deployed"
	state.Status.CompletedAt = &now
	state.Response.Status = "deployed"
	s.mu.Unlock()

	// Persist final status to DB
	s.updateStackDeployInDB(ctx, state)

	slog.Info("stack deployment completed", "deploy_id", deployID)
}
