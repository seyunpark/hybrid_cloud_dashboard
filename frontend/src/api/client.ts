import axios from 'axios';
import type {
  Container,
  ContainerDetail,
  Cluster,
  Pod,
  Deployment,
  Service,
  DeployRequest,
  DeployResponse,
  DeployStatus,
  DeploymentHistory,
  SuccessResponse,
  KubeContext,
  RegisterClusterRequest,
  AIConfig,
  UpdateAIConfigRequest,
  StackDeployRequest,
  StackDeployResponse,
  StackDeployStatus,
  UnifiedDeployItem,
  PaginatedResponse,
} from './types';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '',
  timeout: 120000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Response interceptor for global error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      const status = error.response.status;
      if (status === 401 || status === 403) {
        console.error('Authentication error:', status);
      } else if (status >= 500) {
        console.error('Server error:', error.response.data);
      }
    } else if (error.request) {
      console.error('Network error: No response received');
    }
    return Promise.reject(error);
  },
);

// --- Docker API ---

export const dockerApi = {
  listContainers: async (all = false) => {
    const { data } = await apiClient.get<{ containers: Container[] }>(
      '/api/docker/containers',
      { params: { all } },
    );
    return data.containers;
  },

  getContainer: async (id: string) => {
    const { data } = await apiClient.get<ContainerDetail>(
      `/api/docker/containers/${id}`,
    );
    return data;
  },

  restartContainer: async (id: string) => {
    const { data } = await apiClient.post<SuccessResponse>(
      `/api/docker/containers/${id}/restart`,
    );
    return data;
  },

  stopContainer: async (id: string) => {
    const { data } = await apiClient.post<SuccessResponse>(
      `/api/docker/containers/${id}/stop`,
    );
    return data;
  },

  deleteContainer: async (id: string, force = false) => {
    const { data } = await apiClient.delete<SuccessResponse>(
      `/api/docker/containers/${id}`,
      { params: { force } },
    );
    return data;
  },
};

// --- Kubernetes API ---

export const k8sApi = {
  listClusters: async () => {
    const { data } = await apiClient.get<{ clusters: Cluster[] }>(
      '/api/k8s/clusters',
    );
    return data.clusters;
  },

  listNamespaces: async (cluster: string) => {
    const { data } = await apiClient.get<{ namespaces: string[] }>(
      `/api/k8s/${cluster}/namespaces`,
    );
    return data.namespaces;
  },

  listPods: async (cluster: string, namespace = 'default', label?: string) => {
    const { data } = await apiClient.get<{ pods: Pod[] }>(
      `/api/k8s/${cluster}/pods`,
      { params: { namespace, label } },
    );
    return data.pods;
  },

  listDeployments: async (cluster: string, namespace = 'default') => {
    const { data } = await apiClient.get<{ deployments: Deployment[] }>(
      `/api/k8s/${cluster}/deployments`,
      { params: { namespace } },
    );
    return data.deployments;
  },

  listServices: async (cluster: string, namespace = 'default') => {
    const { data } = await apiClient.get<{ services: Service[] }>(
      `/api/k8s/${cluster}/services`,
      { params: { namespace } },
    );
    return data.services;
  },

  scaleDeployment: async (
    cluster: string,
    namespace: string,
    name: string,
    replicas: number,
  ) => {
    const { data } = await apiClient.post(
      `/api/k8s/${cluster}/deployments/${namespace}/${name}/scale`,
      { replicas },
    );
    return data;
  },

  restartPod: async (cluster: string, namespace: string, name: string) => {
    const { data } = await apiClient.post<SuccessResponse>(
      `/api/k8s/${cluster}/pods/${namespace}/${name}/restart`,
    );
    return data;
  },
};

// --- Deploy API ---

export const deployApi = {
  deployDockerToK8s: async (req: DeployRequest) => {
    const { data } = await apiClient.post<DeployResponse>(
      '/api/deploy/docker-to-k8s',
      req,
    );
    return data;
  },

  refineDeploy: async (deployId: string, feedback: string) => {
    const { data } = await apiClient.post<DeployResponse>(
      `/api/deploy/${deployId}/refine`,
      { feedback },
    );
    return data;
  },

  executeDeploy: async (
    deployId: string,
    approved: boolean,
    modifications?: Record<string, string>,
  ) => {
    const { data } = await apiClient.post(
      `/api/deploy/${deployId}/execute`,
      { approved, modifications },
    );
    return data;
  },

  getDeployStatus: async (deployId: string) => {
    const { data } = await apiClient.get<DeployStatus>(
      `/api/deploy/${deployId}/status`,
    );
    return data;
  },

  getDeployHistory: async (limit = 50) => {
    const { data } = await apiClient.get<{
      deployments: DeploymentHistory[];
      total: number;
    }>('/api/deploy/history', { params: { limit } });
    return data;
  },

  undeployFromK8s: async (deployId: string) => {
    const { data } = await apiClient.post<SuccessResponse>(
      `/api/deploy/${deployId}/undeploy`,
    );
    return data;
  },

  redeployToK8s: async (deployId: string) => {
    const { data } = await apiClient.post<{ success: boolean; deploy_id: string; message: string }>(
      `/api/deploy/${deployId}/redeploy`,
      {},
    );
    return data;
  },

  deleteDeployRecord: async (deployId: string) => {
    const { data } = await apiClient.delete<SuccessResponse>(
      `/api/deploy/${deployId}`,
    );
    return data;
  },

  getUnifiedHistory: async (page = 1, limit = 20) => {
    const { data } = await apiClient.get<PaginatedResponse<UnifiedDeployItem>>(
      '/api/deploy/unified-history',
      { params: { page, limit } },
    );
    return data;
  },
};

// --- Config API ---

export const configApi = {
  getClusters: async () => {
    const { data } = await apiClient.get('/api/config/clusters');
    return data;
  },

  getAI: async () => {
    const { data } = await apiClient.get<AIConfig>('/api/config/ai');
    return data;
  },

  updateAI: async (req: UpdateAIConfigRequest) => {
    const { data } = await apiClient.put<SuccessResponse>('/api/config/ai', req);
    return data;
  },

  listAIModels: async (provider: string, apiKey?: string) => {
    const { data } = await apiClient.get<{ models: string[] }>(
      '/api/config/ai/models',
      { params: { provider, ...(apiKey ? { api_key: apiKey } : {}) } },
    );
    return data.models;
  },

  getKubeContexts: async (kubeconfig?: string) => {
    const { data } = await apiClient.get<{ contexts: KubeContext[] }>(
      '/api/config/kubecontexts',
      { params: kubeconfig ? { kubeconfig } : undefined },
    );
    return data.contexts;
  },

  registerCluster: async (req: RegisterClusterRequest) => {
    const { data } = await apiClient.post<SuccessResponse>(
      '/api/config/clusters',
      req,
    );
    return data;
  },

  unregisterCluster: async (name: string) => {
    const { data } = await apiClient.delete<SuccessResponse>(
      `/api/config/clusters/${name}`,
    );
    return data;
  },
};

// --- Stack Deploy API ---

export const stackDeployApi = {
  listActiveStackDeploys: async () => {
    const { data } = await apiClient.get<{
      deployments: StackDeployStatus[];
      total: number;
    }>('/api/deploy/stack');
    return data;
  },

  getStackDeployDetail: async (deployId: string) => {
    const { data } = await apiClient.get<{
      response: StackDeployResponse;
      status: StackDeployStatus;
      cluster_name?: string;
      namespace?: string;
    }>(`/api/deploy/stack/${deployId}`);
    return data;
  },

  deployStack: async (req: StackDeployRequest) => {
    const { data } = await apiClient.post<StackDeployResponse>(
      '/api/deploy/stack',
      req,
    );
    return data;
  },

  refineStackDeploy: async (deployId: string, feedback: string) => {
    const { data } = await apiClient.post<StackDeployResponse>(
      `/api/deploy/stack/${deployId}/refine`,
      { feedback },
    );
    return data;
  },

  regenerateStackDeploy: async (deployId: string) => {
    const { data } = await apiClient.post<SuccessResponse>(
      `/api/deploy/stack/${deployId}/regenerate`,
    );
    return data;
  },

  executeStackDeploy: async (
    deployId: string,
    params: {
      approved: boolean;
      cluster_name?: string;
      namespace?: string;
      create_namespace?: boolean;
    },
  ) => {
    const { data } = await apiClient.post<StackDeployStatus>(
      `/api/deploy/stack/${deployId}/execute`,
      params,
    );
    return data;
  },

  reopenStackDeploy: async (deployId: string) => {
    const { data } = await apiClient.post<SuccessResponse>(
      `/api/deploy/stack/${deployId}/reopen`,
    );
    return data;
  },

  deleteStackDeploy: async (deployId: string) => {
    const { data } = await apiClient.delete<SuccessResponse>(
      `/api/deploy/stack/${deployId}`,
    );
    return data;
  },

  undeployStack: async (deployId: string) => {
    const { data } = await apiClient.post<SuccessResponse>(
      `/api/deploy/stack/${deployId}/undeploy`,
    );
    return data;
  },

  redeployStack: async (deployId: string, params?: { cluster_name?: string }) => {
    const { data } = await apiClient.post<{ deploy_id: string; status: string; stack_name: string; message: string }>(
      `/api/deploy/stack/${deployId}/redeploy`,
      params ?? {},
    );
    return data;
  },

  getStackDeployStatus: async (deployId: string) => {
    const { data } = await apiClient.get<StackDeployStatus>(
      `/api/deploy/stack/${deployId}/status`,
    );
    return data;
  },
};

// --- Health API ---

export const healthApi = {
  health: async () => {
    const { data } = await apiClient.get('/health');
    return data;
  },

  ready: async () => {
    const { data } = await apiClient.get('/ready');
    return data;
  },
};

export default apiClient;
