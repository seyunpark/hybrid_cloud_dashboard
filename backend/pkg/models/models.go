package models

import "time"

// --- Docker Models ---

type Container struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Image     string           `json:"image"`
	Status    string           `json:"status"`
	State     string           `json:"state"`
	CreatedAt time.Time        `json:"created_at"`
	Ports     []ContainerPort  `json:"ports"`
	Stats     *ContainerStats  `json:"stats,omitempty"`
}

type ContainerPort struct {
	PrivatePort int    `json:"private_port"`
	PublicPort  int    `json:"public_port"`
	Type        string `json:"type"`
}

type ContainerStats struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsage   int64   `json:"memory_usage"`
	MemoryLimit   int64   `json:"memory_limit"`
	MemoryPercent float64 `json:"memory_percent"`
	NetworkRx     int64   `json:"network_rx"`
	NetworkTx     int64   `json:"network_tx"`
}

type ContainerDetail struct {
	Container
	Config  ContainerConfig  `json:"config"`
	Mounts  []Mount          `json:"mounts"`
	Network NetworkInfo      `json:"network"`
}

type ContainerConfig struct {
	Env          []string `json:"env"`
	Cmd          []string `json:"cmd"`
	WorkingDir   string   `json:"working_dir"`
	ExposedPorts []string `json:"exposed_ports"`
}

type Mount struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type NetworkInfo struct {
	IPAddress  string `json:"ip_address"`
	Gateway    string `json:"gateway"`
	MACAddress string `json:"mac_address"`
}

// --- Kubernetes Models ---

type Cluster struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Context string      `json:"context"`
	Status  string      `json:"status"`
	Info    ClusterInfo `json:"info"`
}

type ClusterInfo struct {
	Nodes      int    `json:"nodes"`
	Pods       int    `json:"pods"`
	Namespaces int    `json:"namespaces"`
	Version    string `json:"version"`
}

type Pod struct {
	Name       string         `json:"name"`
	Namespace  string         `json:"namespace"`
	Status     string         `json:"status"`
	Phase      string         `json:"phase"`
	Node       string         `json:"node"`
	IP         string         `json:"ip"`
	CreatedAt  time.Time      `json:"created_at"`
	Containers []PodContainer `json:"containers"`
	Resources  PodResources   `json:"resources"`
	Conditions []Condition    `json:"conditions"`
}

type PodContainer struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Ready        bool   `json:"ready"`
	RestartCount int    `json:"restart_count"`
	State        string `json:"state"`
}

type PodResources struct {
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
}

type Condition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type Deployment struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int               `json:"replicas"`
	ReadyReplicas     int               `json:"ready_replicas"`
	AvailableReplicas int               `json:"available_replicas"`
	UpdatedReplicas   int               `json:"updated_replicas"`
	Image             string            `json:"image"`
	CreatedAt         time.Time         `json:"created_at"`
	Conditions        []Condition       `json:"conditions"`
	Selector          map[string]string `json:"selector"`
}

type Service struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	ClusterIP string            `json:"cluster_ip"`
	Ports     []ServicePort     `json:"ports"`
	Selector  map[string]string `json:"selector"`
}

type ServicePort struct {
	Name       string `json:"name"`
	Port       int    `json:"port"`
	TargetPort int    `json:"target_port"`
	Protocol   string `json:"protocol"`
}

// --- Deploy Models ---

type DeployRequest struct {
	ContainerID string        `json:"container_id" binding:"required"`
	ClusterName string        `json:"cluster_name" binding:"required"`
	Namespace   string        `json:"namespace"`
	Options     DeployOptions `json:"options"`
}

type DeployOptions struct {
	HighAvailability bool `json:"high_availability"`
	EnableHPA        bool `json:"enable_hpa"`
}

type DeployResponse struct {
	DeployID        string           `json:"deploy_id"`
	Status          string           `json:"status"`
	AIAnalysis      *AIAnalysis      `json:"ai_analysis,omitempty"`
	Recommendations *Recommendations `json:"recommendations,omitempty"`
	Manifests       *Manifests       `json:"manifests,omitempty"`
	EstimatedCost   *EstimatedCost   `json:"estimated_cost,omitempty"`
}

type AIAnalysis struct {
	ServiceType        string `json:"service_type"`
	DetectedLanguage   string `json:"detected_language"`
	SimilarDeployments int    `json:"similar_deployments"`
}

type Recommendations struct {
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
	Replicas      int    `json:"replicas"`
	EnableHPA     bool   `json:"enable_hpa"`
	Reasoning     string `json:"reasoning"`
}

type Manifests struct {
	Deployment string `json:"deployment"`
	Service    string `json:"service"`
	HPA        string `json:"hpa,omitempty"`
	ConfigMap  string `json:"configmap,omitempty"`
}

type EstimatedCost struct {
	MonthlyUSD float64 `json:"monthly_usd"`
	Breakdown  string  `json:"breakdown"`
}

type ExecuteRequest struct {
	Approved      bool              `json:"approved"`
	Modifications map[string]string `json:"modifications,omitempty"`
}

type DeployStatus struct {
	DeployID    string       `json:"deploy_id"`
	Status      string       `json:"status"`
	StartedAt   *time.Time   `json:"started_at,omitempty"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	Steps       []DeployStep `json:"steps"`
	Result      *DeployResult `json:"result,omitempty"`
}

type DeployStep struct {
	Step        string     `json:"step"`
	Status      string     `json:"status"`
	Message     string     `json:"message,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type DeployResult struct {
	DeploymentName string `json:"deployment_name"`
	Namespace      string `json:"namespace"`
	Replicas       int    `json:"replicas"`
	PodsReady      string `json:"pods_ready"`
	ServiceURL     string `json:"service_url"`
}

// --- Deployment History (Data Layer) ---

type DeploymentHistory struct {
	ID            string    `json:"id"`
	ServiceName   string    `json:"service_name"`
	ImageName     string    `json:"image_name"`
	ImageTag      string    `json:"image_tag"`
	ServiceType   string    `json:"service_type"`
	Language      string    `json:"language"`
	CPURequest    string    `json:"cpu_request"`
	CPULimit      string    `json:"cpu_limit"`
	MemoryRequest string    `json:"memory_request"`
	MemoryLimit   string    `json:"memory_limit"`
	Replicas      int       `json:"replicas"`
	ActualCPU     string    `json:"actual_cpu"`
	ActualMemory  string    `json:"actual_memory"`
	TargetCluster string    `json:"target_cluster"`
	Namespace     string    `json:"namespace"`
	DeployedAt    time.Time `json:"deployed_at"`
	Success       bool      `json:"success"`
	Status        string    `json:"status"`                    // "deployed", "deleted", "failed"
	ManifestJSON  string    `json:"manifest_json,omitempty"`   // stored manifest for redeploy
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
	OOMEvents     int       `json:"oom_events"`
	ThrottleEvents int      `json:"throttle_events"`
	AIGenerated   bool      `json:"ai_generated"`
	AIConfidence  float64   `json:"ai_confidence"`
}

// --- AI Models ---

type ManifestResult struct {
	Deployment string  `json:"deployment"`
	Service    string  `json:"service"`
	ConfigMap  string  `json:"configmap,omitempty"`
	HPA        string  `json:"hpa,omitempty"`
	Reasoning  string  `json:"reasoning"`
	Confidence float64 `json:"confidence"`
}

// --- Cluster Management Models ---

type KubeContext struct {
	Name      string `json:"name"`
	Cluster   string `json:"cluster"`
	User      string `json:"user"`
	Namespace string `json:"namespace,omitempty"`
	IsActive  bool   `json:"is_active"`
}

type RegisterClusterRequest struct {
	Name       string `json:"name" binding:"required"`
	Context    string `json:"context" binding:"required"`
	Type       string `json:"type"`
	Kubeconfig string `json:"kubeconfig"`
	Registry   string `json:"registry"`
}

type UpdateAIConfigRequest struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
}

// RegisteredCluster represents a cluster saved in the database for persistence.
type RegisteredCluster struct {
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Kubeconfig string    `json:"kubeconfig"`
	Context    string    `json:"context"`
	Registry   string    `json:"registry"`
	CreatedAt  time.Time `json:"created_at"`
}

// --- Stack Deploy Models ---

// StackDeployRequest represents a request to deploy multiple containers as a connected stack.
type StackDeployRequest struct {
	ContainerIDs    []string      `json:"container_ids" binding:"required,min=2"`
	ClusterName     string        `json:"cluster_name"`
	Namespace       string        `json:"namespace"`
	StackName       string        `json:"stack_name"`
	CreateNamespace bool          `json:"create_namespace"`
	Prompt          string        `json:"prompt"`
	Options         DeployOptions `json:"options"`
}

// ServiceConnection represents a detected connection between services.
type ServiceConnection struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Port   int    `json:"port"`
	EnvVar string `json:"env_var"`
}

// StackTopology represents the AI-detected service topology.
type StackTopology struct {
	Services    []StackServiceInfo  `json:"services"`
	Connections []ServiceConnection `json:"connections"`
	DeployOrder []string            `json:"deploy_order"`
}

// StackServiceInfo represents one service in the stack.
type StackServiceInfo struct {
	ContainerID string `json:"container_id"`
	ServiceName string `json:"service_name"`
	ServiceType string `json:"service_type"`
	Image       string `json:"image"`
}

// StackManifests maps resource kind (e.g. "Deployment", "Service", "ConfigMap", "Secret")
// to a map of resource name → YAML string. This allows dynamic resource types.
type StackManifests map[string]map[string]string

// StackDeployResponse is the response for a stack deployment request.
type StackDeployResponse struct {
	DeployID   string          `json:"deploy_id"`
	Status     string          `json:"status"`
	StackName  string          `json:"stack_name"`
	Topology   *StackTopology  `json:"topology,omitempty"`
	Manifests  StackManifests  `json:"manifests,omitempty"`
	Reasoning  string          `json:"reasoning,omitempty"`
	Confidence float64         `json:"confidence,omitempty"`
}

// ServiceDeployStatus tracks individual service progress within a stack.
type ServiceDeployStatus struct {
	ServiceName string       `json:"service_name"`
	Status      string       `json:"status"`
	Steps       []DeployStep `json:"steps"`
}

// StackDeployStatus tracks per-service deployment progress.
type StackDeployStatus struct {
	DeployID    string                         `json:"deploy_id"`
	Status      string                         `json:"status"`
	StackName   string                         `json:"stack_name"`
	StartedAt   *time.Time                     `json:"started_at,omitempty"`
	CompletedAt *time.Time                     `json:"completed_at,omitempty"`
	Services    map[string]*ServiceDeployStatus `json:"services"`
	DeployOrder []string                       `json:"deploy_order"`
}

// StackExecuteRequest is the request for executing a stack deployment.
type StackExecuteRequest struct {
	Approved        bool   `json:"approved"`
	ClusterName     string `json:"cluster_name"`
	Namespace       string `json:"namespace"`
	CreateNamespace bool   `json:"create_namespace"`
}

// StackDeployRecord is the DB-persisted representation of a stack deployment.
type StackDeployRecord struct {
	DeployID        string     `json:"deploy_id"`
	StackName       string     `json:"stack_name"`
	ClusterName     string     `json:"cluster_name"`
	Namespace       string     `json:"namespace"`
	ContainerIDs    []string   `json:"container_ids"`
	CreateNamespace bool       `json:"create_namespace"`
	Prompt          string     `json:"prompt,omitempty"`
	Status          string     `json:"status"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	TopologyJSON  string     `json:"topology_json,omitempty"`
	ManifestsJSON string     `json:"manifests_json,omitempty"`
	Reasoning     string     `json:"reasoning,omitempty"`
	Confidence    float64    `json:"confidence"`
	DeployOrder   []string   `json:"deploy_order"`
	ServicesJSON  string     `json:"services_json,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// --- Unified Deploy History ---

// UnifiedDeployItem represents either a single deploy or a stack deploy
// in a unified, chronologically-ordered history view.
type UnifiedDeployItem struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`                        // "single" or "stack"
	Name         string    `json:"name"`                        // service_name or stack_name
	ImageSummary string    `json:"image_summary"`               // "nginx:latest" or "3 services"
	Cluster      string    `json:"cluster"`
	Namespace    string    `json:"namespace"`
	Status       string    `json:"status"`
	AIGenerated  bool      `json:"ai_generated"`
	Confidence   float64   `json:"confidence"`
	DeployedAt   time.Time `json:"deployed_at"`
	StackDetail  *StackDeployBrief  `json:"stack_detail,omitempty"`
	SingleDetail *SingleDeployBrief `json:"single_detail,omitempty"`
}

// StackDeployBrief contains summary info for a stack deploy in the unified list.
type StackDeployBrief struct {
	ServiceCount int      `json:"service_count"`
	Services     []string `json:"services"`
	DeployOrder  []string `json:"deploy_order"`
}

// SingleDeployBrief contains summary info for a single deploy in the unified list.
type SingleDeployBrief struct {
	ImageName string `json:"image_name"`
	ImageTag  string `json:"image_tag"`
	Replicas  int    `json:"replicas"`
}

// PaginatedResponse wraps paginated query results.
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// --- Common API Response Models ---

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type ReadyResponse struct {
	Status    string            `json:"status"`
	Checks    map[string]string `json:"checks"`
	Timestamp string            `json:"timestamp"`
}
