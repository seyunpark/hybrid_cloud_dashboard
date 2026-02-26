import { useQuery } from '@tanstack/react-query';
import { deployApi } from '@/api/client';
import { StatusBadge } from '@/components/common/StatusBadge';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { formatDateTime } from '@/utils/formatters';

export function DeployPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['deploy', 'history'],
    queryFn: () => deployApi.getDeployHistory(),
  });

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-gray-900">
          Deployment History
        </h2>
        <span className="text-sm text-gray-500">
          {data?.total ?? 0} deployments
        </span>
      </div>

      {isLoading ? (
        <LoadingSpinner message="Loading deployment history..." />
      ) : data && data.deployments.length > 0 ? (
        <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">
                  Service
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">
                  Image
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">
                  Cluster
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">
                  Status
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">
                  AI
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">
                  Deployed
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {data.deployments.map((deploy) => (
                <tr key={deploy.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm font-medium text-gray-900">
                    {deploy.service_name}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    {deploy.image}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    {deploy.cluster}
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge
                      status={deploy.success ? 'completed' : 'failed'}
                    />
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    {deploy.ai_generated ? (
                      <span className="text-blue-600">
                        AI ({Math.round(deploy.ai_confidence * 100)}%)
                      </span>
                    ) : (
                      'Manual'
                    )}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    {formatDateTime(deploy.deployed_at)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="rounded-lg border border-dashed border-gray-300 p-8 text-center">
          <p className="text-sm text-gray-500">
            No deployment history yet. Deploy a container to get started.
          </p>
        </div>
      )}
    </div>
  );
}
