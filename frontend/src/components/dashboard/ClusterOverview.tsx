import type { Cluster } from '@/api/types';
import { StatusBadge } from '@/components/common/StatusBadge';

interface ClusterOverviewProps {
  cluster: Cluster;
}

export function ClusterOverview({ cluster }: ClusterOverviewProps) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
      <div className="mb-3 flex items-start justify-between">
        <div className="min-w-0 flex-1">
          <h3 className="truncate text-sm font-semibold text-gray-900">
            {cluster.name}
          </h3>
          <p className="text-xs text-gray-500">{cluster.type}</p>
        </div>
        <StatusBadge status={cluster.status} />
      </div>

      <div className="grid grid-cols-3 gap-2">
        <div className="rounded bg-gray-50 p-2 text-center">
          <p className="text-lg font-semibold text-gray-900">
            {cluster.info.nodes}
          </p>
          <p className="text-xs text-gray-500">Nodes</p>
        </div>
        <div className="rounded bg-gray-50 p-2 text-center">
          <p className="text-lg font-semibold text-gray-900">
            {cluster.info.pods}
          </p>
          <p className="text-xs text-gray-500">Pods</p>
        </div>
        <div className="rounded bg-gray-50 p-2 text-center">
          <p className="text-lg font-semibold text-gray-900">
            {cluster.info.namespaces}
          </p>
          <p className="text-xs text-gray-500">Namespaces</p>
        </div>
      </div>

      {cluster.info.version && (
        <p className="mt-3 text-xs text-gray-400">
          Kubernetes {cluster.info.version}
        </p>
      )}
    </div>
  );
}
