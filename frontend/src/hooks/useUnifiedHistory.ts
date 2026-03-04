import { useQuery, keepPreviousData } from '@tanstack/react-query';
import { deployApi } from '@/api/client';

export function useUnifiedHistory(page: number, limit = 20) {
  return useQuery({
    queryKey: ['deploy', 'unified-history', page, limit],
    queryFn: () => deployApi.getUnifiedHistory(page, limit),
    placeholderData: keepPreviousData,
  });
}
