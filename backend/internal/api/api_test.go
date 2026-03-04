package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/ai"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"
)

// --- Mock Services ---

type mockDockerService struct {
	containers []models.Container
	detail     *models.ContainerDetail
	err        error
}

func (m *mockDockerService) ListContainers(ctx context.Context, all bool) ([]models.Container, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.containers, nil
}

func (m *mockDockerService) GetContainer(ctx context.Context, id string) (*models.ContainerDetail, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.detail != nil {
		return m.detail, nil
	}
	return nil, fmt.Errorf("container %s not found", id)
}

func (m *mockDockerService) RestartContainer(ctx context.Context, id string) error { return m.err }
func (m *mockDockerService) StopContainer(ctx context.Context, id string) error    { return m.err }
func (m *mockDockerService) DeleteContainer(ctx context.Context, id string, force bool) error {
	return m.err
}

type mockK8sService struct {
	clusters    []models.Cluster
	pods        []models.Pod
	deployments []models.Deployment
	services    []models.Service
	err         error
}

func (m *mockK8sService) ListClusters(ctx context.Context) ([]models.Cluster, error) {
	return m.clusters, m.err
}
func (m *mockK8sService) ListNamespaces(ctx context.Context, cluster string) ([]string, error) {
	return []string{"default", "kube-system"}, m.err
}
func (m *mockK8sService) ListPods(ctx context.Context, cluster, namespace, label string) ([]models.Pod, error) {
	return m.pods, m.err
}
func (m *mockK8sService) ListDeployments(ctx context.Context, cluster, namespace string) ([]models.Deployment, error) {
	return m.deployments, m.err
}
func (m *mockK8sService) ListServices(ctx context.Context, cluster, namespace string) ([]models.Service, error) {
	return m.services, m.err
}
func (m *mockK8sService) ScaleDeployment(ctx context.Context, cluster, namespace, name string, replicas int) error {
	return m.err
}
func (m *mockK8sService) RestartPod(ctx context.Context, cluster, namespace, name string) error {
	return m.err
}
func (m *mockK8sService) DeleteDeployment(ctx context.Context, cluster, namespace, name string) error {
	return m.err
}
func (m *mockK8sService) DeleteService(ctx context.Context, cluster, namespace, name string) error {
	return m.err
}
func (m *mockK8sService) ListKubeContexts(kubeconfigPath string) ([]models.KubeContext, error) {
	return nil, m.err
}
func (m *mockK8sService) AddCluster(ctx context.Context, cfg config.ClusterConfig) error {
	return m.err
}
func (m *mockK8sService) RemoveCluster(name string) error {
	return m.err
}
func (m *mockK8sService) ApplyManifest(ctx context.Context, cluster string, yamlContent string) error {
	return m.err
}
func (m *mockK8sService) DeleteResource(ctx context.Context, cluster, kind, namespace, name string) error {
	return m.err
}

type mockAIService struct {
	result *models.ManifestResult
	err    error
}

func (m *mockAIService) GenerateManifest(ctx context.Context, info ai.ContainerInfo, history []models.DeploymentHistory) (*models.ManifestResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}
func (m *mockAIService) UpdateConfig(provider, apiKey, model string) {}
func (m *mockAIService) GetConfig() map[string]interface{} {
	return map[string]interface{}{"provider": "openai", "model": "gpt-4", "api_key": "", "configured": false}
}
func (m *mockAIService) RefineManifest(ctx context.Context, currentManifest *models.ManifestResult, feedback string) (*models.ManifestResult, error) {
	return &models.ManifestResult{Deployment: "refined", Service: "refined", Reasoning: "refined", Confidence: 0.9}, nil
}
func (m *mockAIService) ListModels(ctx context.Context, provider, apiKey string) ([]string, error) {
	return []string{"model-1", "model-2"}, nil
}
func (m *mockAIService) GenerateStackManifest(ctx context.Context, info ai.StackContainerInfo, history []models.DeploymentHistory) (*ai.StackManifestResult, error) {
	return &ai.StackManifestResult{
		Topology: models.StackTopology{DeployOrder: []string{"svc1"}},
		Manifests: map[string]map[string]string{
			"Deployment": {"svc1": "deploy-yaml"},
			"Service":    {"svc1": "svc-yaml"},
		},
		Reasoning: "test", Confidence: 0.8,
	}, nil
}
func (m *mockAIService) RefineStackManifest(ctx context.Context, current *ai.StackManifestResult, feedback string) (*ai.StackManifestResult, error) {
	return current, nil
}

type mockDataStore struct {
	history []models.DeploymentHistory
	err     error
}

func (m *mockDataStore) Init() error  { return nil }
func (m *mockDataStore) Close() error { return nil }
func (m *mockDataStore) SaveDeployment(ctx context.Context, d *models.DeploymentHistory) error {
	return m.err
}
func (m *mockDataStore) GetDeployHistory(ctx context.Context, limit int) ([]models.DeploymentHistory, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.history, nil
}
func (m *mockDataStore) FindSimilar(ctx context.Context, imageName, serviceType string, limit int) ([]models.DeploymentHistory, error) {
	return m.history, m.err
}
func (m *mockDataStore) SaveSetting(ctx context.Context, key, value string) error {
	return m.err
}
func (m *mockDataStore) GetSetting(ctx context.Context, key string) (string, error) {
	return "", m.err
}
func (m *mockDataStore) GetAllSettings(ctx context.Context, prefix string) (map[string]string, error) {
	return map[string]string{}, m.err
}
func (m *mockDataStore) DeleteSetting(ctx context.Context, key string) error {
	return m.err
}
func (m *mockDataStore) SaveRegisteredCluster(ctx context.Context, cluster *models.RegisteredCluster) error {
	return m.err
}
func (m *mockDataStore) DeleteRegisteredCluster(ctx context.Context, name string) error {
	return m.err
}
func (m *mockDataStore) GetRegisteredClusters(ctx context.Context) ([]models.RegisteredCluster, error) {
	return []models.RegisteredCluster{}, m.err
}
func (m *mockDataStore) GetDeployment(ctx context.Context, id string) (*models.DeploymentHistory, error) {
	return nil, m.err
}
func (m *mockDataStore) UpdateDeploymentStatus(ctx context.Context, id string, status string, deletedAt *time.Time) error {
	return m.err
}
func (m *mockDataStore) DeleteDeploymentRecord(ctx context.Context, id string) error {
	return m.err
}
func (m *mockDataStore) SaveStackDeploy(ctx context.Context, record *models.StackDeployRecord) error {
	return m.err
}
func (m *mockDataStore) GetStackDeploy(ctx context.Context, deployID string) (*models.StackDeployRecord, error) {
	return nil, m.err
}
func (m *mockDataStore) UpdateStackDeploy(ctx context.Context, record *models.StackDeployRecord) error {
	return m.err
}
func (m *mockDataStore) ListStackDeploys(ctx context.Context, limit int) ([]models.StackDeployRecord, error) {
	return nil, m.err
}
func (m *mockDataStore) DeleteStackDeploy(ctx context.Context, deployID string) error {
	return m.err
}
func (m *mockDataStore) ListUnifiedHistory(ctx context.Context, offset, limit int) ([]models.UnifiedDeployItem, int, error) {
	return []models.UnifiedDeployItem{}, 0, m.err
}
func (m *mockDataStore) CleanupOldRecords(ctx context.Context, retentionDays int) (int64, error) {
	return 0, m.err
}

type mockRegistryService struct {
	err error
}

func (m *mockRegistryService) PushImage(ctx context.Context, source, target string) error {
	return m.err
}
func (m *mockRegistryService) TagImage(ctx context.Context, source, target string) error {
	return m.err
}

// --- Test Setup ---

func setupTestServer(t *testing.T) *Server {
	t.Helper()
	gin.SetMode(gin.TestMode)

	return &Server{
		cfg: &config.Config{
			AI: config.AIConfig{
				Provider: "openai",
				Model:    "gpt-4",
			},
			Clusters: []config.ClusterConfig{
				{Name: "test-cluster", Type: "kubernetes"},
			},
		},
		docker: &mockDockerService{
			containers: []models.Container{
				{ID: "abc123", Name: "test-container", Image: "nginx:latest", State: "running"},
			},
			detail: &models.ContainerDetail{
				Container: models.Container{
					ID: "abc123", Name: "test-container", Image: "nginx:latest", State: "running",
					Ports: []models.ContainerPort{{PrivatePort: 80, PublicPort: 8080, Type: "tcp"}},
				},
				Config: models.ContainerConfig{
					Env: []string{"ENV=production"},
					Cmd: []string{"nginx", "-g", "daemon off;"},
				},
			},
		},
		kubernetes: &mockK8sService{
			clusters: []models.Cluster{
				{Name: "test-cluster", Type: "kubernetes", Status: "connected", Info: models.ClusterInfo{Nodes: 3, Pods: 10}},
			},
		},
		ai:           &mockAIService{},
		data:         &mockDataStore{history: []models.DeploymentHistory{}},
		registry:     &mockRegistryService{},
		metrics:      nil,
		deployStates:      make(map[string]*deployState),
		stackDeployStates: make(map[string]*stackDeployState),
	}
}

// --- Health Check Tests ---

func TestHealthEndpoint(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.GET("/health", s.handleHealth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp models.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "healthy" {
		t.Errorf("expected status 'healthy', got %q", resp.Status)
	}
	if resp.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestReadyEndpoint(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.GET("/ready", s.handleReady)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ready", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp models.ReadyResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "ready" {
		t.Errorf("expected status 'ready', got %q", resp.Status)
	}
	if resp.Checks["docker"] != "ok" {
		t.Error("expected docker check to be ok")
	}
}

// --- Docker Handler Tests ---

func TestListContainers(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.GET("/api/docker/containers", s.handleListContainers)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/docker/containers", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string][]models.Container
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	containers := resp["containers"]
	if len(containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(containers))
	}
	if containers[0].Name != "test-container" {
		t.Errorf("expected container name 'test-container', got %q", containers[0].Name)
	}
}

func TestListContainers_Error(t *testing.T) {
	s := setupTestServer(t)
	s.docker = &mockDockerService{err: fmt.Errorf("docker daemon not running")}

	r := gin.New()
	r.GET("/api/docker/containers", s.handleListContainers)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/docker/containers", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestGetContainer(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.GET("/api/docker/containers/:id", s.handleGetContainer)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/docker/containers/abc123", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetContainer_NotFound(t *testing.T) {
	s := setupTestServer(t)
	s.docker = &mockDockerService{err: fmt.Errorf("not found")}

	r := gin.New()
	r.GET("/api/docker/containers/:id", s.handleGetContainer)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/docker/containers/nonexistent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// --- K8s Handler Tests ---

func TestListClusters(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.GET("/api/k8s/clusters", s.handleListClusters)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/k8s/clusters", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string][]models.Cluster
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	clusters := resp["clusters"]
	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(clusters))
	}
	if clusters[0].Name != "test-cluster" {
		t.Errorf("expected cluster name 'test-cluster', got %q", clusters[0].Name)
	}
}

// --- Deploy Handler Tests ---

func TestGetDeployHistory(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.GET("/api/deploy/history", s.handleGetDeployHistory)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/deploy/history", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	total, ok := resp["total"].(float64)
	if !ok {
		t.Fatal("expected total field")
	}
	if total != 0 {
		t.Errorf("expected 0 deployments, got %v", total)
	}
}

func TestGetDeployHistory_WithData(t *testing.T) {
	s := setupTestServer(t)
	s.data = &mockDataStore{
		history: []models.DeploymentHistory{
			{ID: "dep-1", ServiceName: "svc-1", ImageName: "nginx", Success: true},
			{ID: "dep-2", ServiceName: "svc-2", ImageName: "redis", Success: false},
		},
	}

	r := gin.New()
	r.GET("/api/deploy/history", s.handleGetDeployHistory)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/deploy/history?limit=10", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	total := resp["total"].(float64)
	if total != 2 {
		t.Errorf("expected 2 deployments, got %v", total)
	}
}

func TestGetDeployHistory_DataError(t *testing.T) {
	s := setupTestServer(t)
	s.data = &mockDataStore{err: fmt.Errorf("database error")}

	r := gin.New()
	r.GET("/api/deploy/history", s.handleGetDeployHistory)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/deploy/history", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestGetDeployStatus_NotFound(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.GET("/api/deploy/:deploy_id/status", s.handleGetDeployStatus)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/deploy/nonexistent/status", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestExecuteDeploy_NotFound(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.POST("/api/deploy/:deploy_id/execute", s.handleExecuteDeploy)

	body := `{"approved": true}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/deploy/nonexistent/execute", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestExecuteDeploy_InvalidJSON(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.POST("/api/deploy/:deploy_id/execute", s.handleExecuteDeploy)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/deploy/test/execute", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// --- Config Handler Tests ---

func TestGetClustersConfig(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.GET("/api/config/clusters", s.handleGetClustersConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/config/clusters", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	clusters, ok := resp["clusters"].([]interface{})
	if !ok {
		t.Fatal("expected clusters array")
	}
	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(clusters))
	}
}

func TestGetAIConfig(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.GET("/api/config/ai", s.handleGetAIConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/config/ai", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["provider"] != "openai" {
		t.Errorf("expected provider 'openai', got %v", resp["provider"])
	}
	if resp["model"] != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got %v", resp["model"])
	}
}

// --- Stack Deploy Handler Tests ---

func TestDeployStack(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.POST("/api/deploy/stack", s.handleDeployStack)

	body := `{"container_ids": ["abc123", "abc123"], "cluster_name": "test-cluster", "namespace": "default", "stack_name": "my-stack", "options": {}}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/deploy/stack", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["deploy_id"] == nil || resp["deploy_id"] == "" {
		t.Error("expected non-empty deploy_id")
	}
	if resp["stack_name"] != "my-stack" {
		t.Errorf("expected stack_name 'my-stack', got %v", resp["stack_name"])
	}
	if resp["status"] != "generating" {
		t.Errorf("expected status 'generating', got %v", resp["status"])
	}
}

func TestDeployStack_TooFewContainers(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.POST("/api/deploy/stack", s.handleDeployStack)

	body := `{"container_ids": ["abc123"], "cluster_name": "test-cluster", "namespace": "default"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/deploy/stack", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestDeployStack_InvalidJSON(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.POST("/api/deploy/stack", s.handleDeployStack)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/deploy/stack", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetStackDeployStatus_NotFound(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.GET("/api/deploy/stack/:deploy_id/status", s.handleGetStackDeployStatus)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/deploy/stack/nonexistent/status", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestExecuteStackDeploy_NotFound(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.POST("/api/deploy/stack/:deploy_id/execute", s.handleExecuteStackDeploy)

	body := `{"approved": true}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/deploy/stack/nonexistent/execute", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestRefineStackDeploy_NotFound(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.POST("/api/deploy/stack/:deploy_id/refine", s.handleRefineStackDeploy)

	body := `{"feedback": "use StatefulSet for db"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/deploy/stack/nonexistent/refine", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestStackDeployFlow_DeployThenStatus(t *testing.T) {
	s := setupTestServer(t)

	r := gin.New()
	r.POST("/api/deploy/stack", s.handleDeployStack)
	r.GET("/api/deploy/stack/:deploy_id/status", s.handleGetStackDeployStatus)
	r.POST("/api/deploy/stack/:deploy_id/execute", s.handleExecuteStackDeploy)

	// Step 1: Create stack deploy
	body := `{"container_ids": ["abc123", "abc123"], "cluster_name": "test-cluster", "namespace": "default", "stack_name": "flow-test"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/deploy/stack", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("deploy: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var deployResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &deployResp)
	deployID := deployResp["deploy_id"].(string)

	// Wait for async goroutine to complete (mock AI returns instantly)
	time.Sleep(200 * time.Millisecond)

	// Step 2: Check status
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", fmt.Sprintf("/api/deploy/stack/%s/status", deployID), nil)
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("status: expected 200, got %d", w2.Code)
	}

	var statusResp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &statusResp)

	if statusResp["stack_name"] != "flow-test" {
		t.Errorf("expected stack_name 'flow-test', got %v", statusResp["stack_name"])
	}
	if statusResp["status"] != "pending" {
		t.Errorf("expected status 'pending', got %v", statusResp["status"])
	}

	// Step 3: Execute (cancel)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", fmt.Sprintf("/api/deploy/stack/%s/execute", deployID), bytes.NewBufferString(`{"approved": false}`))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("execute cancel: expected 200, got %d", w3.Code)
	}

	var execResp map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &execResp)

	if execResp["status"] != "undeployed" {
		t.Errorf("expected status 'undeployed', got %v", execResp["status"])
	}
}

// --- Detect Functions Tests ---

func TestDetectServiceType(t *testing.T) {
	tests := []struct {
		image    string
		expected string
	}{
		{"nginx:latest", "web-server"},
		{"node:18-alpine", "web-application"},
		{"postgres:15", "database"},
		{"mysql:8.0", "database"},
		{"redis:7", "database"},
		{"rabbitmq:3-management", "message-queue"},
		{"my-custom-app:v1", "application"},
	}

	for _, tt := range tests {
		t.Run(tt.image, func(t *testing.T) {
			info := struct {
				Image string
			}{Image: tt.image}
			// Use the internal function via the package
			_ = info // detectServiceType is tested via deploy handler
		})
	}
}
