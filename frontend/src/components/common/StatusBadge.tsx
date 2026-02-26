interface StatusBadgeProps {
  status: string;
}

const statusStyles: Record<string, string> = {
  running: 'bg-green-100 text-green-800',
  Running: 'bg-green-100 text-green-800',
  connected: 'bg-green-100 text-green-800',
  healthy: 'bg-green-100 text-green-800',
  completed: 'bg-green-100 text-green-800',
  exited: 'bg-red-100 text-red-800',
  stopped: 'bg-red-100 text-red-800',
  disconnected: 'bg-red-100 text-red-800',
  error: 'bg-red-100 text-red-800',
  failed: 'bg-red-100 text-red-800',
  paused: 'bg-yellow-100 text-yellow-800',
  pending: 'bg-yellow-100 text-yellow-800',
  Pending: 'bg-yellow-100 text-yellow-800',
  analyzing: 'bg-blue-100 text-blue-800',
  deploying: 'bg-blue-100 text-blue-800',
  in_progress: 'bg-blue-100 text-blue-800',
};

export function StatusBadge({ status }: StatusBadgeProps) {
  const style = statusStyles[status] || 'bg-gray-100 text-gray-800';

  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${style}`}
    >
      {status}
    </span>
  );
}
