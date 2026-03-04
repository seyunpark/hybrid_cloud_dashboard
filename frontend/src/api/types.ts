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

export interface KubeContext {
  name: string;
  cluster: string;
  user: string;
  namespace?: string;
  is_active: boolean;
}

export interface RegisterClusterRequest {
  name: string;
  context: string;
  type?: string;
  kubeconfig?: string;
  registry?: string;
}

export interface AIConfig {
  provider: string;
  model: string;
  api_key: string;
  temperature: number;
  configured: boolean;
}

export interface UpdateAIConfigRequest {
  provider?: string;
  api_key?: string;
  model?: string;
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
  estimated_cost?: { monthly_usd: number; breakdown: string };
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
  image_name: string;
  image_tag: string;
  service_type: string;
  language: string;
  cpu_request: string;
  cpu_limit: string;
  memory_request: string;
  memory_limit: string;
  replicas: number;
  actual_cpu: string;
  actual_memory: string;
  target_cluster: string;
  namespace: string;
  deployed_at: string;
  success: boolean;
  status: string;
  manifest_json?: string;
  deleted_at?: string;
  oom_events: number;
  throttle_events: number;
  ai_generated: boolean;
  ai_confidence: number;
}

// --- Unified Deploy History ---

export interface StackDeployBrief {
  service_count: number;
  services: string[];
  deploy_order: string[];
}

export interface SingleDeployBrief {
  image_name: string;
  image_tag: string;
  replicas: number;
}

export interface UnifiedDeployItem {
  id: string;
  type: 'single' | 'stack';
  name: string;
  image_summary: string;
  cluster: string;
  namespace: string;
  status: string;
  ai_generated: boolean;
  confidence: number;
  deployed_at: string;
  stack_detail?: StackDeployBrief;
  single_detail?: SingleDeployBrief;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

// --- Stack Deploy Models ---

export interface StackDeployRequest {
  container_ids: string[];
  cluster_name?: string;
  namespace?: string;
  stack_name: string;
  create_namespace?: boolean;
  options: DeployOptions;
}

export interface ServiceConnection {
  from: string;
  to: string;
  port: number;
  env_var: string;
}

export interface StackTopology {
  services: StackServiceInfo[];
  connections: ServiceConnection[];
  deploy_order: string[];
}

export interface StackServiceInfo {
  container_id: string;
  service_name: string;
  service_type: string;
  image: string;
}

// Dynamic manifest map: resource kind (e.g. "Deployment", "Service", "ConfigMap", "Secret")
// → resource name → YAML string
export type StackManifests = Record<string, Record<string, string>>;

export interface StackDeployResponse {
  deploy_id: string;
  status: string;
  stack_name: string;
  topology?: StackTopology;
  manifests?: StackManifests;
  reasoning?: string;
  confidence?: number;
}

export interface ServiceDeployStatus {
  service_name: string;
  status: string;
  steps: DeployStep[];
}

export interface StackDeployStatus {
  deploy_id: string;
  status: string;
  stack_name: string;
  started_at?: string;
  completed_at?: string;
  services: Record<string, ServiceDeployStatus>;
  deploy_order: string[];
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

