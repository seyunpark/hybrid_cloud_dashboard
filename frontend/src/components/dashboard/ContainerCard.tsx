import type { Container } from '@/api/types';
import { StatusBadge } from '@/components/common/StatusBadge';
import { formatBytes, formatCpuPercent } from '@/utils/formatters';

interface ContainerCardProps {
  container: Container;
  onDeploy?: (id: string) => void;
}

export function ContainerCard({ container, onDeploy }: ContainerCardProps) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm transition-shadow hover:shadow-md">
      <div className="mb-3 flex items-start justify-between">
        <div className="min-w-0 flex-1">
          <h3 className="truncate text-sm font-semibold text-gray-900">
            {container.name}
          </h3>
          <p className="truncate text-xs text-gray-500">{container.image}</p>
        </div>
        <StatusBadge status={container.status} />
      </div>

      {/* Resource Stats */}
      {container.stats && (
        <div className="mb-3 grid grid-cols-2 gap-2">
          <div className="rounded bg-gray-50 p-2">
            <p className="text-xs text-gray-500">CPU</p>
            <p className="text-sm font-medium text-gray-900">
              {formatCpuPercent(container.stats.cpu_percent)}
            </p>
          </div>
          <div className="rounded bg-gray-50 p-2">
            <p className="text-xs text-gray-500">Memory</p>
            <p className="text-sm font-medium text-gray-900">
              {formatBytes(container.stats.memory_usage)} /{' '}
              {formatBytes(container.stats.memory_limit)}
            </p>
          </div>
        </div>
      )}

      {/* Ports */}
      {container.ports.length > 0 && (
        <div className="mb-3">
          <p className="mb-1 text-xs text-gray-500">Ports</p>
          <div className="flex flex-wrap gap-1">
            {container.ports.map((port, i) => (
              <span
                key={i}
                className="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600"
              >
                {port.public_port}:{port.private_port}/{port.type}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="flex gap-2 border-t border-gray-100 pt-3">
        <button
          onClick={() => onDeploy?.(container.id)}
          className="flex-1 rounded bg-blue-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-blue-700"
        >
          Deploy to K8s
        </button>
      </div>
    </div>
  );
}
