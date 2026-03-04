import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { configApi } from '@/api/client';
import type { RegisterClusterRequest } from '@/api/types';

export function useKubeContexts() {
  return useQuery({
    queryKey: ['config', 'kubecontexts'],
    queryFn: () => configApi.getKubeContexts(),
  });
}

export function useRegisterCluster() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: RegisterClusterRequest) => configApi.registerCluster(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'kubecontexts'] });
      queryClient.invalidateQueries({ queryKey: ['k8s', 'clusters'] });
    },
  });
}

export function useUnregisterCluster() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => configApi.unregisterCluster(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'kubecontexts'] });
      queryClient.invalidateQueries({ queryKey: ['k8s', 'clusters'] });
    },
  });
}
