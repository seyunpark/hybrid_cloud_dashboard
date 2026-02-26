import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { dockerApi } from '@/api/client';

export function useDockerContainers(all = false) {
  return useQuery({
    queryKey: ['docker', 'containers', { all }],
    queryFn: () => dockerApi.listContainers(all),
    refetchInterval: 10000,
  });
}

export function useDockerContainer(id: string) {
  return useQuery({
    queryKey: ['docker', 'containers', id],
    queryFn: () => dockerApi.getContainer(id),
    enabled: !!id,
  });
}

export function useRestartContainer() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => dockerApi.restartContainer(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['docker', 'containers'] });
    },
  });
}

export function useStopContainer() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => dockerApi.stopContainer(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['docker', 'containers'] });
    },
  });
}

export function useDeleteContainer() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, force }: { id: string; force?: boolean }) =>
      dockerApi.deleteContainer(id, force),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['docker', 'containers'] });
    },
  });
}
