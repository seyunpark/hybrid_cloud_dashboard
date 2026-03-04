import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query';
import { stackDeployApi } from '@/api/client';

export function useStackDeployDetail(deployId: string | undefined) {
  return useQuery({
    queryKey: ['deploy', 'stack', deployId],
    queryFn: () => stackDeployApi.getStackDeployDetail(deployId!),
    enabled: !!deployId,
    refetchInterval: (query) => {
      const status = query.state.data?.status?.status;
      if (status === 'generating') return 2000;
      if (status === 'deploying') return 2000;
      if (status === 'pending' || status === 'analyzing') return 5000;
      return false;
    },
  });
}

export function useActiveStackDeploys() {
  return useQuery({
    queryKey: ['deploy', 'active'],
    queryFn: () => stackDeployApi.listActiveStackDeploys(),
    refetchInterval: 5000,
  });
}

export function useRefineStackDeploy(deployId: string | undefined) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (feedback: string) =>
      stackDeployApi.refineStackDeploy(deployId!, feedback),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['deploy', 'stack', deployId] });
    },
  });
}

export function useRegenerateStackDeploy(deployId: string | undefined) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => stackDeployApi.regenerateStackDeploy(deployId!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['deploy', 'stack', deployId] });
    },
  });
}

export function useExecuteStackDeploy(deployId: string | undefined) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (params: {
      approved: boolean;
      cluster_name?: string;
      namespace?: string;
      create_namespace?: boolean;
    }) => stackDeployApi.executeStackDeploy(deployId!, params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['deploy', 'stack', deployId] });
      queryClient.invalidateQueries({ queryKey: ['deploy', 'active'] });
      queryClient.invalidateQueries({ queryKey: ['deploy', 'history'] });
    },
  });
}

export function useReopenStackDeploy(deployId: string | undefined) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => stackDeployApi.reopenStackDeploy(deployId!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['deploy', 'stack', deployId] });
      queryClient.invalidateQueries({ queryKey: ['deploy', 'active'] });
    },
  });
}
