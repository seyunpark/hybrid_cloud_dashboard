import { useParams, Link } from 'react-router-dom';
import { useDockerContainer, useRestartContainer, useStopContainer } from '@/hooks/useDockerContainers';
import { useWebSocket } from '@/hooks/useWebSocket';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { StatusBadge } from '@/components/common/StatusBadge';
import { MetricChart } from '@/components/common/MetricChart';
import { formatBytes, formatCpuPercent, formatDateTime } from '@/utils/formatters';
import { useState, useCallback } from 'react';

interface MetricDataPoint {
  time: string;
  value: number;
}

export function ContainerDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { data: container, isLoading, error } = useDockerContainer(id ?? '');
  const restartMutation = useRestartContainer();
  const stopMutation = useStopContainer();

  const [cpuHistory, setCpuHistory] = useState<MetricDataPoint[]>([]);
  const [memHistory, setMemHistory] = useState<MetricDataPoint[]>([]);

  useWebSocket({
    url: id ? `/ws/docker/${id}/logs` : '',
    onMessage: useCallback(() => {
      // Logs handled separately in LogViewer
    }, []),
    shouldReconnect: false,
  });

  // Use stats WS and filter for this container
  useWebSocket({
    url: '/ws/docker/stats',
    onMessage: useCallback((data: unknown) => {
      const msg = data as { containers?: Array<{ container_id: string; stats?: { cpu_percent: number; memory_percent: number } }> };
      if (msg.containers) {
        const c = msg.containers.find(c => c.container_id === id);
        if (c?.stats) {
          const now = new Date().toLocaleTimeString();
          setCpuHistory(prev => [...prev.slice(-59), { time: now, value: c.stats!.cpu_percent }]);
          setMemHistory(prev => [...prev.slice(-59), { time: now, value: c.stats!.memory_percent }]);
        }
      }
    }, [id]),
  });

  if (isLoading) return <LoadingSpinner message="Loading container details..." />;
  if (error || !container) {
    return (
      <div className="rounded-lg border border-red-200 bg-red-50 p-8 text-center">
        <p className="text-sm text-red-600">Container not found or failed to load.</p>
        <Link to="/" className="mt-2 inline-block text-sm text-blue-600 hover:underline">Back to Dashboard</Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link to="/" className="text-sm text-gray-500 hover:text-gray-700">&larr; Dashboard</Link>
        <h1 className="text-xl font-bold text-gray-900">{container.name}</h1>
        <StatusBadge status={container.state} />
      </div>

      {/* Overview */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
        <InfoCard label="Image" value={container.image} />
        <InfoCard label="Status" value={container.status} />
        <InfoCard label="ID" value={container.id} />
        <InfoCard label="Created" value={formatDateTime(container.created_at)} />
      </div>

      {/* Actions */}
      <div className="flex gap-2">
        <button
          onClick={() => id && restartMutation.mutate(id)}
          disabled={restartMutation.isPending}
          className="rounded bg-yellow-500 px-3 py-1.5 text-xs font-medium text-white hover:bg-yellow-600 disabled:opacity-50"
        >
          {restartMutation.isPending ? 'Restarting...' : 'Restart'}
        </button>
        <button
          onClick={() => id && stopMutation.mutate(id)}
          disabled={stopMutation.isPending || container.state !== 'running'}
          className="rounded bg-red-500 px-3 py-1.5 text-xs font-medium text-white hover:bg-red-600 disabled:opacity-50"
        >
          {stopMutation.isPending ? 'Stopping...' : 'Stop'}
        </button>
      </div>

      {/* Real-time Metrics */}
      {cpuHistory.length > 0 && (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <MetricChart data={cpuHistory} title="CPU Usage" color="#3b82f6" unit="%" height={250} />
          <MetricChart data={memHistory} title="Memory Usage" color="#10b981" unit="%" height={250} />
        </div>
      )}

      {/* Current Stats */}
      {container.stats && (
        <section>
          <h2 className="mb-3 text-sm font-semibold text-gray-900">Resource Usage</h2>
          <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
            <StatCard label="CPU" value={formatCpuPercent(container.stats.cpu_percent)} />
            <StatCard label="Memory" value={`${formatBytes(container.stats.memory_usage)} / ${formatBytes(container.stats.memory_limit)}`} />
            <StatCard label="Network Rx" value={formatBytes(container.stats.network_rx)} />
            <StatCard label="Network Tx" value={formatBytes(container.stats.network_tx)} />
          </div>
        </section>
      )}

      {/* Config */}
      <section>
        <h2 className="mb-3 text-sm font-semibold text-gray-900">Configuration</h2>
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          {container.config.cmd && container.config.cmd.length > 0 && (
            <div className="mb-3">
              <p className="text-xs font-medium text-gray-500">Command</p>
              <code className="text-sm text-gray-700">{container.config.cmd.join(' ')}</code>
            </div>
          )}
          {container.config.working_dir && (
            <div className="mb-3">
              <p className="text-xs font-medium text-gray-500">Working Directory</p>
              <code className="text-sm text-gray-700">{container.config.working_dir}</code>
            </div>
          )}
          {container.config.env && container.config.env.length > 0 && (
            <div>
              <p className="mb-1 text-xs font-medium text-gray-500">Environment Variables</p>
              <div className="max-h-40 overflow-auto rounded bg-gray-50 p-2">
                {container.config.env.map((e, i) => (
                  <p key={i} className="truncate text-xs text-gray-600">{e}</p>
                ))}
              </div>
            </div>
          )}
        </div>
      </section>

      {/* Ports */}
      {container.ports && container.ports.length > 0 && (
        <section>
          <h2 className="mb-3 text-sm font-semibold text-gray-900">Ports</h2>
          <div className="flex flex-wrap gap-2">
            {container.ports.map((port, i) => (
              <span key={i} className="rounded-full bg-gray-100 px-3 py-1 text-xs text-gray-700">
                {port.public_port}:{port.private_port}/{port.type}
              </span>
            ))}
          </div>
        </section>
      )}

      {/* Mounts */}
      {container.mounts && container.mounts.length > 0 && (
        <section>
          <h2 className="mb-3 text-sm font-semibold text-gray-900">Mounts</h2>
          <div className="rounded-lg border border-gray-200 bg-white">
            {container.mounts.map((mount, i) => (
              <div key={i} className="flex border-b border-gray-100 px-4 py-2 last:border-0">
                <span className="w-20 text-xs font-medium text-gray-500">{mount.type}</span>
                <span className="flex-1 truncate text-xs text-gray-700">{mount.source} → {mount.destination}</span>
              </div>
            ))}
          </div>
        </section>
      )}

      {/* Network */}
      {container.network && container.network.ip_address && (
        <section>
          <h2 className="mb-3 text-sm font-semibold text-gray-900">Network</h2>
          <div className="grid grid-cols-3 gap-4">
            <InfoCard label="IP Address" value={container.network.ip_address} />
            <InfoCard label="Gateway" value={container.network.gateway} />
            <InfoCard label="MAC Address" value={container.network.mac_address} />
          </div>
        </section>
      )}
    </div>
  );
}

function InfoCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-3">
      <p className="text-xs text-gray-500">{label}</p>
      <p className="truncate text-sm font-medium text-gray-900">{value}</p>
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg bg-gray-50 p-3">
      <p className="text-xs text-gray-500">{label}</p>
      <p className="text-sm font-semibold text-gray-900">{value}</p>
    </div>
  );
}
