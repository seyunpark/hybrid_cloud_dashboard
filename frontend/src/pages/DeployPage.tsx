import { useState, useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, Link } from 'react-router-dom';
import { deployApi, stackDeployApi } from '@/api/client';
import { StatusBadge } from '@/components/common/StatusBadge';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { formatDateTime } from '@/utils/formatters';
import { StackDeployModal } from '@/components/deploy/StackDeployModal';
import { useDockerContainers } from '@/hooks/useDockerContainers';
import { useK8sClusters } from '@/hooks/useK8sClusters';
import { useActiveStackDeploys } from '@/hooks/useStackDeploy';

export function DeployPage() {
  const navigate = useNavigate();
  const { data: historyData, isLoading: historyLoading } = useQuery({
    queryKey: ['deploy', 'history'],
    queryFn: () => deployApi.getDeployHistory(),
  });
  const { data: containers } = useDockerContainers();
  const { data: clusters } = useK8sClusters();
  const { data: activeData } = useActiveStackDeploys();

  // Modal state only
  const [showModal, setShowModal] = useState(false);
  const [stackLoading, setStackLoading] = useState(false);
  const [stackError, setStackError] = useState<string | null>(null);

  const handleStackDeploy = useCallback(
    async (containerIds: string[], stackName: string, clusterName: string, namespace: string, createNamespace: boolean) => {
      setStackLoading(true);
      setStackError(null);
      try {
        const resp = await stackDeployApi.deployStack({
          container_ids: containerIds,
          stack_name: stackName,
          cluster_name: clusterName,
          namespace,
          create_namespace: createNamespace,
          options: { high_availability: false, enable_hpa: false },
        });
        setShowModal(false);
        navigate(`/deploy/${resp.deploy_id}`);
      } catch (err: unknown) {
        const msg = err instanceof Error ? err.message : 'Failed to generate stack manifests';
        setStackError(msg);
      } finally {
        setStackLoading(false);
      }
    },
    [navigate],
  );

  const handleCloseModal = useCallback(() => {
    setShowModal(false);
    setStackError(null);
    setStackLoading(false);
  }, []);

  const activeDeploys = activeData?.deployments ?? [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-gray-900">Deployments</h2>
        <button
          onClick={() => setShowModal(true)}
          className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          Deploy Stack
        </button>
      </div>

      {/* Stack Deploy Modal */}
      <StackDeployModal
        isOpen={showModal}
        onClose={handleCloseModal}
        containers={containers ?? []}
        clusters={clusters ?? []}
        onDeploy={handleStackDeploy}
        isLoading={stackLoading}
        error={stackError}
      />

      {/* Active Deployments */}
      {activeDeploys.length > 0 && (
        <div>
          <h3 className="mb-3 text-sm font-semibold text-gray-700">
            Active Deployments ({activeDeploys.length})
          </h3>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {activeDeploys.map((deploy) => {
              const totalSteps = deploy.deploy_order.reduce((acc, name) => {
                const svc = deploy.services[name];
                return acc + (svc?.steps.length ?? 0);
              }, 0);
              const doneSteps = deploy.deploy_order.reduce((acc, name) => {
                const svc = deploy.services[name];
                return acc + (svc?.steps.filter((s) => s.status === 'completed').length ?? 0);
              }, 0);
              const percent = totalSteps > 0 ? Math.round((doneSteps / totalSteps) * 100) : 0;

              return (
                <Link
                  key={deploy.deploy_id}
                  to={`/deploy/${deploy.deploy_id}`}
                  className="block rounded-lg border border-gray-200 bg-white p-4 shadow-sm transition hover:border-blue-300 hover:shadow-md"
                >
                  <div className="mb-2 flex items-center justify-between">
                    <span className="text-sm font-medium text-gray-900 truncate">
                      {deploy.stack_name}
                    </span>
                    <StatusBadge status={deploy.status} />
                  </div>
                  <div className="mb-2 flex flex-wrap gap-1">
                    {deploy.deploy_order.map((name) => (
                      <span
                        key={name}
                        className="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600"
                      >
                        {name}
                      </span>
                    ))}
                  </div>
                  {deploy.status === 'deploying' && (
                    <div className="h-1.5 w-full overflow-hidden rounded-full bg-gray-200">
                      <div
                        className="h-full rounded-full bg-blue-500 transition-all duration-500"
                        style={{ width: `${percent}%` }}
                      />
                    </div>
                  )}
                </Link>
              );
            })}
          </div>
        </div>
      )}

      {/* History Table */}
      <div>
        <h3 className="mb-3 text-sm font-semibold text-gray-700">
          History ({historyData?.total ?? 0})
        </h3>
        {historyLoading ? (
          <LoadingSpinner message="Loading deployment history..." />
        ) : historyData && historyData.deployments.length > 0 ? (
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
                {historyData.deployments.map((deploy) => (
                  <tr key={deploy.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm font-medium text-gray-900">
                      {deploy.service_name}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500">
                      {deploy.image_name}{deploy.image_tag ? `:${deploy.image_tag}` : ''}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500">
                      {deploy.target_cluster}
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge
                        status={deploy.status || (deploy.success ? 'deployed' : 'failed')}
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
    </div>
  );
}
