import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDockerContainers } from '@/hooks/useDockerContainers';
import { useK8sClusters } from '@/hooks/useK8sClusters';
import { ContainerCard } from './ContainerCard';
import { ClusterOverview } from './ClusterOverview';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { DeployModal } from '@/components/deploy/DeployModal';
import { ManifestPreview } from '@/components/deploy/ManifestPreview';
import { DeployProgress } from '@/components/deploy/DeployProgress';
import { StackDeployModal } from '@/components/deploy/StackDeployModal';
import { deployApi, stackDeployApi } from '@/api/client';
import type { DeployResponse, DeployStatus } from '@/api/types';
import { useWebSocket } from '@/hooks/useWebSocket';
import { MetricChart } from '@/components/common/MetricChart';

type DeployPhase = 'select' | 'preview' | 'progress';

interface MetricDataPoint {
  time: string;
  value: number;
}

export function Dashboard() {
  const navigate = useNavigate();
  const { data: containers, isLoading: containersLoading } =
    useDockerContainers();
  const { data: clusters, isLoading: clustersLoading } = useK8sClusters();

  // Single deploy flow state
  const [deployPhase, setDeployPhase] = useState<DeployPhase | null>(null);
  const [preselectedContainerId, setPreselectedContainerId] = useState<string>('');
  const [deployResponse, setDeployResponse] = useState<DeployResponse | null>(null);
  const [deployStatus, setDeployStatus] = useState<DeployStatus | null>(null);
  const [deployError, setDeployError] = useState<string | null>(null);
  const [isGenerating, setIsGenerating] = useState(false);
  const [isRefining, setIsRefining] = useState(false);

  // Stack deploy modal state
  const [showStackModal, setShowStackModal] = useState(false);
  const [isStackGenerating, setIsStackGenerating] = useState(false);
  const [stackError, setStackError] = useState<string | null>(null);

  // Real-time stats via WebSocket
  const [cpuHistory, setCpuHistory] = useState<MetricDataPoint[]>([]);
  const [memHistory, setMemHistory] = useState<MetricDataPoint[]>([]);

  useWebSocket({
    url: '/ws/docker/stats',
    onMessage: useCallback((data: unknown) => {
      const msg = data as { containers?: Array<{ stats?: { cpu_percent: number; memory_percent: number } }> };
      if (msg.containers && msg.containers.length > 0) {
        const now = new Date().toLocaleTimeString();
        let totalCpu = 0;
        let totalMem = 0;
        let count = 0;
        for (const c of msg.containers) {
          if (c.stats) {
            totalCpu += c.stats.cpu_percent;
            totalMem += c.stats.memory_percent;
            count++;
          }
        }
        if (count > 0) {
          setCpuHistory(prev => [...prev.slice(-29), { time: now, value: totalCpu / count }]);
          setMemHistory(prev => [...prev.slice(-29), { time: now, value: totalMem / count }]);
        }
      }
    }, []),
  });

  // Deploy WebSocket (only when deploying)
  useWebSocket({
    url: deployResponse?.deploy_id ? `/ws/deploy/${deployResponse.deploy_id}/status` : '',
    onMessage: useCallback((data: unknown) => {
      const msg = data as { data?: DeployStatus };
      if (msg.data) {
        setDeployStatus(msg.data);
      }
    }, []),
    shouldReconnect: false,
  });

  const handleContainerDeploy = (containerId: string) => {
    setPreselectedContainerId(containerId);
    setDeployPhase('select');
    setDeployError(null);
  };

  const handleDeploySubmit = async (containerId: string, clusterName: string, namespace: string) => {
    setIsGenerating(true);
    setDeployError(null);
    try {
      const resp = await deployApi.deployDockerToK8s({
        container_id: containerId,
        cluster_name: clusterName,
        namespace,
        options: { high_availability: false, enable_hpa: false },
      });
      setDeployResponse(resp);
      setDeployPhase('preview');
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: { message?: string } } }; message?: string };
      setDeployError(error.response?.data?.error?.message || error.message || 'Failed to generate manifest');
    } finally {
      setIsGenerating(false);
    }
  };

  const handleApprove = async () => {
    if (!deployResponse) return;
    try {
      const status = await deployApi.executeDeploy(deployResponse.deploy_id, true);
      setDeployStatus(status as DeployStatus);
      setDeployPhase('progress');
    } catch (err: unknown) {
      const error = err as { message?: string };
      setDeployError(error.message || 'Failed to execute deployment');
    }
  };

  const handleRefine = async (feedback: string) => {
    if (!deployResponse) return;
    setIsRefining(true);
    setDeployError(null);
    try {
      const refined = await deployApi.refineDeploy(deployResponse.deploy_id, feedback);
      setDeployResponse(refined);
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: { message?: string } } }; message?: string };
      setDeployError(error.response?.data?.error?.message || error.message || 'Failed to refine manifest');
    } finally {
      setIsRefining(false);
    }
  };

  const handleCloseModal = () => {
    setDeployPhase(null);
    setDeployResponse(null);
    setDeployStatus(null);
    setDeployError(null);
    setPreselectedContainerId('');
    setIsGenerating(false);
  };

  // --- Stack Deploy Handlers ---

  const handleStackDeploySubmit = async (
    containerIds: string[],
    stackName: string,
    clusterName: string,
    namespace: string,
    createNamespace: boolean,
  ) => {
    setIsStackGenerating(true);
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
      setShowStackModal(false);
      navigate(`/deploy/${resp.deploy_id}`);
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: { message?: string } } }; message?: string };
      setStackError(error.response?.data?.error?.message || error.message || 'Failed to generate stack manifests');
    } finally {
      setIsStackGenerating(false);
    }
  };

  const handleCloseStackModal = () => {
    setShowStackModal(false);
    setStackError(null);
    setIsStackGenerating(false);
  };

  return (
    <div className="space-y-8">
      {/* Real-time Metrics */}
      {cpuHistory.length > 0 && (
        <section>
          <h2 className="mb-4 text-lg font-semibold text-gray-900">Real-time Metrics</h2>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <MetricChart data={cpuHistory} title="Average CPU Usage" color="#3b82f6" unit="%" height={200} />
            <MetricChart data={memHistory} title="Average Memory Usage" color="#10b981" unit="%" height={200} />
          </div>
        </section>
      )}

      {/* Docker Containers Section */}
      <section>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">
            Docker Containers
          </h2>
          <div className="flex items-center gap-3">
            <button
              onClick={() => { setShowStackModal(true); setStackError(null); }}
              disabled={!containers || containers.filter(c => c.status.startsWith('Up')).length < 2}
              className="rounded-md border border-purple-300 bg-purple-50 px-3 py-1.5 text-xs font-medium text-purple-700 hover:bg-purple-100 disabled:cursor-not-allowed disabled:opacity-50"
            >
              Deploy Stack
            </button>
            <span className="text-sm text-gray-500">
              {containers?.length ?? 0} containers
            </span>
          </div>
        </div>

        {containersLoading ? (
          <LoadingSpinner message="Loading containers..." />
        ) : containers && containers.length > 0 ? (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            {containers.map((container) => (
              <ContainerCard
                key={container.id}
                container={container}
                onDeploy={handleContainerDeploy}
              />
            ))}
          </div>
        ) : (
          <div className="rounded-lg border border-dashed border-gray-300 p-8 text-center">
            <p className="text-sm text-gray-500">
              No Docker containers found. Start a container to see it here.
            </p>
          </div>
        )}
      </section>

      {/* Kubernetes Clusters Section */}
      <section>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">
            Kubernetes Clusters
          </h2>
          <span className="text-sm text-gray-500">
            {clusters?.length ?? 0} clusters
          </span>
        </div>

        {clustersLoading ? (
          <LoadingSpinner message="Loading clusters..." />
        ) : clusters && clusters.length > 0 ? (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            {clusters.map((cluster) => (
              <ClusterOverview key={cluster.name} cluster={cluster} />
            ))}
          </div>
        ) : (
          <div className="rounded-lg border border-dashed border-gray-300 p-8 text-center">
            <p className="text-sm text-gray-500">
              No Kubernetes clusters configured. Add clusters in config.yaml.
            </p>
          </div>
        )}
      </section>

      {/* Deploy Modal - Phase: Select */}
      {deployPhase === 'select' && (
        <DeployModal
          isOpen={true}
          onClose={handleCloseModal}
          containers={containers ?? []}
          clusters={clusters ?? []}
          onDeploy={handleDeploySubmit}
          preselectedContainerId={preselectedContainerId}
          isLoading={isGenerating}
          error={deployError}
        />
      )}

      {/* Deploy Modal - Phase: Preview */}
      {deployPhase === 'preview' && deployResponse?.manifests && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="max-h-[90vh] w-full max-w-2xl overflow-auto rounded-lg bg-white p-6 shadow-xl">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-lg font-semibold text-gray-900">
                AI-Generated Manifest
              </h2>
              <button onClick={handleCloseModal} className="text-gray-400 hover:text-gray-600">
                ✕
              </button>
            </div>
            {deployError && (
              <div className="mb-4 rounded bg-red-50 p-3 text-sm text-red-700">{deployError}</div>
            )}
            <ManifestPreview
              manifests={deployResponse.manifests}
              recommendations={deployResponse.recommendations}
              onApprove={handleApprove}
              onCancel={handleCloseModal}
              onRefine={handleRefine}
              isRefining={isRefining}
            />
          </div>
        </div>
      )}

      {/* Deploy Modal - Phase: Progress */}
      {deployPhase === 'progress' && deployStatus && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-lg font-semibold text-gray-900">
                Deploying...
              </h2>
              {(deployStatus.status === 'completed' || deployStatus.status === 'failed') && (
                <button onClick={handleCloseModal} className="text-gray-400 hover:text-gray-600">
                  ✕
                </button>
              )}
            </div>
            <DeployProgress steps={deployStatus.steps} status={deployStatus.status} />
            {deployStatus.result && (
              <div className="mt-4 rounded-lg border border-green-200 bg-green-50 p-4">
                <h3 className="mb-2 text-sm font-semibold text-green-900">Deployment Successful</h3>
                <div className="space-y-1 text-xs text-green-800">
                  <p>Name: {deployStatus.result.deployment_name}</p>
                  <p>Namespace: {deployStatus.result.namespace}</p>
                  <p>Pods: {deployStatus.result.pods_ready}</p>
                  <p>URL: {deployStatus.result.service_url}</p>
                </div>
              </div>
            )}
            {deployStatus.status === 'completed' && (
              <div className="mt-4 flex justify-end">
                <button
                  onClick={handleCloseModal}
                  className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
                >
                  Done
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Stack Deploy Modal */}
      <StackDeployModal
        isOpen={showStackModal}
        onClose={handleCloseStackModal}
        containers={containers ?? []}
        clusters={clusters ?? []}
        onDeploy={handleStackDeploySubmit}
        isLoading={isStackGenerating}
        error={stackError}
      />
    </div>
  );
}
