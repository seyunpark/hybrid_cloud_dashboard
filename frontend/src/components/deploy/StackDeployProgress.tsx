import type { StackDeployStatus } from '@/api/types';

interface StackDeployProgressProps {
  status: StackDeployStatus;
}

function StepIcon({ status }: { status: string }) {
  switch (status) {
    case 'completed':
      return <span className="text-green-500">&#10003;</span>;
    case 'in_progress':
      return (
        <span className="inline-block h-4 w-4 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
      );
    case 'failed':
      return <span className="text-red-500">&#10007;</span>;
    default:
      return <span className="text-gray-300">&#9675;</span>;
  }
}

function ServiceStatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    pending: 'bg-gray-100 text-gray-600',
    deploying: 'bg-blue-100 text-blue-700',
    completed: 'bg-green-100 text-green-700',
    deployed: 'bg-green-100 text-green-700',
    failed: 'bg-red-100 text-red-700',
  };
  return (
    <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${colors[status] || colors.pending}`}>
      {status}
    </span>
  );
}

export function StackDeployProgress({ status }: StackDeployProgressProps) {
  const overallProgress = status.deploy_order.reduce(
    (acc, name) => {
      const svc = status.services[name];
      if (!svc) return acc;
      const total = svc.steps.length;
      const done = svc.steps.filter(s => s.status === 'completed').length;
      return { total: acc.total + total, done: acc.done + done };
    },
    { total: 0, done: 0 },
  );

  const percent = overallProgress.total > 0
    ? Math.round((overallProgress.done / overallProgress.total) * 100)
    : 0;

  return (
    <div className="space-y-4">
      {/* Overall Progress */}
      <div>
        <div className="mb-1 flex items-center justify-between text-xs text-gray-600">
          <span>Stack: {status.stack_name}</span>
          <span>{percent}%</span>
        </div>
        <div className="h-2 w-full overflow-hidden rounded-full bg-gray-200">
          <div
            className="h-full rounded-full bg-blue-500 transition-all duration-500"
            style={{ width: `${percent}%` }}
          />
        </div>
      </div>

      {/* Per-Service Progress */}
      <div className="space-y-3">
        {status.deploy_order.map((name, idx) => {
          const svc = status.services[name];
          if (!svc) return null;

          return (
            <div key={name} className="rounded-lg border border-gray-200 p-3">
              <div className="mb-2 flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="flex h-5 w-5 items-center justify-center rounded-full bg-gray-100 text-xs font-medium text-gray-600">
                    {idx + 1}
                  </span>
                  <span className="text-sm font-medium text-gray-900">{name}</span>
                </div>
                <ServiceStatusBadge status={svc.status} />
              </div>

              <div className="space-y-1.5 pl-7">
                {svc.steps.map((step, i) => (
                  <div key={i} className="flex items-center gap-2 text-xs">
                    <StepIcon status={step.status} />
                    <span className={step.status === 'in_progress' ? 'font-medium text-blue-700' : 'text-gray-600'}>
                      {step.step.replace(/^create_/, 'Create ').replace(/^\w/, (c) => c.toUpperCase())}
                    </span>
                    {step.message && (
                      <span className="text-gray-400">- {step.message}</span>
                    )}
                  </div>
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
