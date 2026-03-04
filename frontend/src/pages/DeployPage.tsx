import { useState, useCallback } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { stackDeployApi } from '@/api/client';
import { StatusBadge } from '@/components/common/StatusBadge';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { formatRelativeTime } from '@/utils/formatters';
import { StackDeployModal } from '@/components/deploy/StackDeployModal';
import { useDockerContainers } from '@/hooks/useDockerContainers';
import { useK8sClusters } from '@/hooks/useK8sClusters';
import { useActiveStackDeploys } from '@/hooks/useStackDeploy';
import { useUnifiedHistory } from '@/hooks/useUnifiedHistory';
import type { UnifiedDeployItem } from '@/api/types';

function RecentDeployRow({ item }: { item: UnifiedDeployItem }) {
  const isStack = item.type === 'stack';

  const content = (
    <div className="flex items-center justify-between py-2.5">
      <div className="flex items-center gap-2.5 min-w-0">
        <span
          className={`flex-shrink-0 rounded px-1.5 py-0.5 text-xs font-medium ${
            isStack
              ? 'bg-purple-100 text-purple-700'
              : 'bg-gray-100 text-gray-600'
          }`}
        >
          {isStack ? 'Stack' : 'Single'}
        </span>
        <span className="truncate text-sm font-medium text-gray-900">
          {item.name}
        </span>
        {isStack && item.stack_detail && (
          <span className="hidden text-xs text-gray-400 sm:inline">
            ({item.stack_detail.service_count} services)
          </span>
        )}
        {!isStack && (
          <span className="hidden text-xs text-gray-400 sm:inline">
            {item.image_summary}
          </span>
        )}
      </div>
      <div className="flex flex-shrink-0 items-center gap-2.5">
        <StatusBadge status={item.status} />
        <span className="text-xs text-gray-400">
          {formatRelativeTime(item.deployed_at)}
        </span>
      </div>
    </div>
  );

  if (isStack) {
    return (
      <Link
        to={`/deploy/${item.id}`}
        className="block border-b border-gray-100 last:border-b-0 hover:bg-gray-50 px-3 transition"
      >
        {content}
      </Link>
    );
  }

  return (
    <div className="border-b border-gray-100 last:border-b-0 px-3">
      {content}
    </div>
  );
}

export function DeployPage() {
  const navigate = useNavigate();
  const { data: containers } = useDockerContainers();
  const { data: clusters } = useK8sClusters();
  const { data: activeData } = useActiveStackDeploys();
  const { data: recentData, isLoading: recentLoading } = useUnifiedHistory(1, 5);

  // Modal state
  const [showModal, setShowModal] = useState(false);
  const [stackLoading, setStackLoading] = useState(false);
  const [stackError, setStackError] = useState<string | null>(null);

  const handleStackDeploy = useCallback(
    async (
      containerIds: string[],
      stackName: string,
      clusterName: string,
      namespace: string,
      createNamespace: boolean,
      prompt: string,
    ) => {
      setStackLoading(true);
      setStackError(null);
      try {
        const resp = await stackDeployApi.deployStack({
          container_ids: containerIds,
          stack_name: stackName,
          cluster_name: clusterName,
          namespace,
          create_namespace: createNamespace,
          prompt: prompt || undefined,
          options: { high_availability: false, enable_hpa: false },
        });
        setShowModal(false);
        navigate(`/deploy/${resp.deploy_id}`);
      } catch (err: unknown) {
        const msg =
          err instanceof Error
            ? err.message
            : 'Failed to generate stack manifests';
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
  const recentItems = recentData?.items ?? [];
  const totalCount = recentData?.total ?? 0;

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

      {/* Stack Deploys */}
      {activeDeploys.length > 0 && (
        <div>
          <h3 className="mb-3 text-sm font-semibold text-gray-700">
            Stack Deploys ({activeDeploys.length})
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
                    {deploy.deploy_order
                      .filter((name) => !name.startsWith('_'))
                      .map((name) => (
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

      {/* Recent History */}
      <div>
        <div className="mb-3 flex items-center justify-between">
          <h3 className="text-sm font-semibold text-gray-700">
            Recent History
          </h3>
          {totalCount > 0 && (
            <Link
              to="/history"
              className="text-sm font-medium text-blue-600 hover:text-blue-700"
            >
              View all ({totalCount})
            </Link>
          )}
        </div>

        {recentLoading && !recentData ? (
          <LoadingSpinner message="Loading..." />
        ) : recentItems.length > 0 ? (
          <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
            {recentItems.map((item) => (
              <RecentDeployRow key={item.id} item={item} />
            ))}
          </div>
        ) : (
          <div className="rounded-lg border border-dashed border-gray-300 p-8 text-center">
            <p className="text-sm text-gray-500">
              No deployment history yet.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
