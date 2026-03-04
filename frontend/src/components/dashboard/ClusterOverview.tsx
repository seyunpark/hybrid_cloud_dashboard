import { Link } from 'react-router-dom';
import type { Cluster } from '@/api/types';
import { StatusBadge } from '@/components/common/StatusBadge';

interface ClusterOverviewProps {
  cluster: Cluster;
}

export function ClusterOverview({ cluster }: ClusterOverviewProps) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm transition-shadow hover:shadow-md">
      <div className="mb-3 flex items-start justify-between">
        <div className="min-w-0 flex-1">
          <Link
            to={`/cluster/${cluster.name}`}
            className="truncate text-sm font-semibold text-gray-900 hover:text-blue-600"
          >
            {cluster.name}
          </Link>
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

      <div className="mt-3 border-t border-gray-100 pt-3">
        <Link
          to={`/cluster/${cluster.name}`}
          className="block w-full rounded border border-gray-300 px-3 py-1.5 text-center text-xs font-medium text-gray-700 transition-colors hover:bg-gray-50"
        >
          View Details
        </Link>
      </div>
    </div>
  );
}
