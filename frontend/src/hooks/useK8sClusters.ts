import { useQuery } from '@tanstack/react-query';
import { k8sApi } from '@/api/client';

export function useK8sClusters() {
  return useQuery({
    queryKey: ['k8s', 'clusters'],
    queryFn: () => k8sApi.listClusters(),
    refetchInterval: 30000,
  });
}

export function useK8sPods(cluster: string, namespace = 'default') {
  return useQuery({
    queryKey: ['k8s', 'pods', cluster, namespace],
    queryFn: () => k8sApi.listPods(cluster, namespace),
    enabled: !!cluster,
    refetchInterval: 10000,
  });
}

export function useK8sDeployments(cluster: string, namespace = 'default') {
  return useQuery({
    queryKey: ['k8s', 'deployments', cluster, namespace],
    queryFn: () => k8sApi.listDeployments(cluster, namespace),
    enabled: !!cluster,
    refetchInterval: 10000,
  });
}

export function useK8sServices(cluster: string, namespace = 'default') {
  return useQuery({
    queryKey: ['k8s', 'services', cluster, namespace],
    queryFn: () => k8sApi.listServices(cluster, namespace),
    enabled: !!cluster,
    refetchInterval: 30000,
  });
}
