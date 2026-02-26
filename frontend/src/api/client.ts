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
} from './types';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

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
};

// --- Config API ---

export const configApi = {
  getClusters: async () => {
    const { data } = await apiClient.get('/api/config/clusters');
    return data;
  },

  getAI: async () => {
    const { data } = await apiClient.get('/api/config/ai');
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
