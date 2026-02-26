// --- Docker Models ---

export interface Container {
  id: string;
  name: string;
  image: string;
  status: string;
  state: string;
  created_at: string;
  ports: ContainerPort[];
  stats?: ContainerStats;
}

export interface ContainerPort {
  private_port: number;
  public_port: number;
  type: string;
}

export interface ContainerStats {
  cpu_percent: number;
  memory_usage: number;
  memory_limit: number;
  memory_percent: number;
  network_rx: number;
  network_tx: number;
}

export interface ContainerDetail extends Container {
  config: ContainerConfig;
  mounts: Mount[];
  network: NetworkInfo;
}

export interface ContainerConfig {
  env: string[];
  cmd: string[];
  working_dir: string;
  exposed_ports: string[];
}

export interface Mount {
  type: string;
  source: string;
  destination: string;
}

export interface NetworkInfo {
  ip_address: string;
  gateway: string;
  mac_address: string;
}

// --- Kubernetes Models ---

export interface Cluster {
  name: string;
  type: string;
  context: string;
  status: string;
  info: ClusterInfo;
}

export interface ClusterInfo {
  nodes: number;
  pods: number;
  namespaces: number;
  version: string;
}

export interface Pod {
  name: string;
  namespace: string;
  status: string;
  phase: string;
  node: string;
  ip: string;
  created_at: string;
  containers: PodContainer[];
  resources: PodResources;
  conditions: Condition[];
}

export interface PodContainer {
  name: string;
  image: string;
  ready: boolean;
  restart_count: number;
  state: string;
}

export interface PodResources {
  cpu_request: string;
  cpu_limit: string;
  memory_request: string;
  memory_limit: string;
}

export interface Condition {
  type: string;
  status: string;
  reason: string;
  message: string;
}

export interface Deployment {
  name: string;
  namespace: string;
  replicas: number;
  ready_replicas: number;
  available_replicas: number;
  updated_replicas: number;
  image: string;
  created_at: string;
  conditions: Condition[];
  selector: Record<string, string>;
}

export interface Service {
  name: string;
  namespace: string;
  type: string;
  cluster_ip: string;
  ports: ServicePort[];
  selector: Record<string, string>;
}

export interface ServicePort {
  name: string;
  port: number;
  target_port: number;
  protocol: string;
}

// --- Deploy Models ---

export interface DeployRequest {
  container_id: string;
  cluster_name: string;
  namespace: string;
  options: DeployOptions;
}

export interface DeployOptions {
  high_availability: boolean;
  enable_hpa: boolean;
}

export interface DeployResponse {
  deploy_id: string;
  status: string;
  ai_analysis?: AIAnalysis;
  recommendations?: Recommendations;
  manifests?: Manifests;
  estimated_cost?: EstimatedCost;
}

export interface AIAnalysis {
  service_type: string;
  detected_language: string;
  similar_deployments: number;
}

export interface Recommendations {
  cpu_request: string;
  cpu_limit: string;
  memory_request: string;
  memory_limit: string;
  replicas: number;
  enable_hpa: boolean;
  reasoning: string;
}

export interface Manifests {
  deployment: string;
  service: string;
  hpa?: string;
  configmap?: string;
}

export interface EstimatedCost {
  monthly_usd: number;
  breakdown: string;
}

export interface DeployStatus {
  deploy_id: string;
  status: string;
  started_at?: string;
  completed_at?: string;
  steps: DeployStep[];
  result?: DeployResult;
}

export interface DeployStep {
  step: string;
  status: string;
  message?: string;
  completed_at?: string;
}

export interface DeployResult {
  deployment_name: string;
  namespace: string;
  replicas: number;
  pods_ready: string;
  service_url: string;
}

export interface DeploymentHistory {
  id: string;
  service_name: string;
  image: string;
  cluster: string;
  namespace: string;
  deployed_at: string;
  success: boolean;
  ai_generated: boolean;
  ai_confidence: number;
  resources: {
    cpu_request: string;
    memory_request: string;
  };
}

// --- Common Response Models ---

export interface ErrorResponse {
  error: {
    code: string;
    message: string;
    details?: unknown;
  };
}

export interface SuccessResponse {
  success: boolean;
  message: string;
}

export interface HealthResponse {
  status: string;
  timestamp: string;
}

export interface ReadyResponse {
  status: string;
  checks: Record<string, string>;
  timestamp: string;
}

// --- WebSocket Message Types ---

export interface DockerStatsMessage {
  type: 'docker_stats';
  timestamp: string;
  containers: ContainerStats[];
}

export interface K8sMetricsMessage {
  type: 'k8s_metrics';
  timestamp: string;
  cluster: string;
  pods: {
    name: string;
    namespace: string;
    cpu_usage: string;
    memory_usage: string;
    status: string;
  }[];
}

export interface LogMessage {
  type: 'log';
  timestamp: string;
  log: string;
}

export interface DeployStatusMessage {
  type: 'deploy_status';
  deploy_id: string;
  timestamp: string;
  step: string;
  status: string;
  progress: number;
  message: string;
}
