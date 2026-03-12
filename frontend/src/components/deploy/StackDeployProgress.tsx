import { useState } from 'react';
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
    case 'skipped':
      return <span className="text-gray-400">&#8722;</span>;
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
    skipped: 'bg-gray-100 text-gray-500',
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

  const isComplete = percent === 100;
  const hasFailed = status.deploy_order.some(name => status.services[name]?.status === 'failed');
  const [expanded, setExpanded] = useState(!isComplete || hasFailed);

  const barColor = hasFailed ? 'bg-red-500' : isComplete ? 'bg-green-500' : 'bg-blue-500';

  return (
    <div className="space-y-4">
      {/* Overall Progress */}
      <div>
        <div className="mb-1.5 flex items-center justify-between text-xs">
          <span className="font-medium text-gray-700">배포 진행률</span>
          <div className="flex items-center gap-2">
            <span className={`font-semibold ${hasFailed ? 'text-red-600' : isComplete ? 'text-green-600' : 'text-blue-600'}`}>
              {percent}%
            </span>
            {(isComplete || hasFailed) && (
              <button
                onClick={() => setExpanded(!expanded)}
                className="text-gray-400 hover:text-gray-600 transition-colors"
              >
                <svg
                  className={`h-4 w-4 transition-transform duration-200 ${expanded ? 'rotate-180' : ''}`}
                  fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}
                >
                  <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
                </svg>
              </button>
            )}
          </div>
        </div>
        <div className="h-2.5 w-full overflow-hidden rounded-full bg-gray-200">
          <div
            className={`h-full rounded-full ${barColor} transition-all duration-500`}
            style={{ width: `${percent}%` }}
          />
        </div>
        {isComplete && !expanded && !hasFailed && (
          <p className="mt-1.5 text-xs text-green-600">
            모든 서비스가 성공적으로 배포되었습니다.
            <button onClick={() => setExpanded(true)} className="ml-1 text-green-700 underline hover:text-green-800">
              상세보기
            </button>
          </p>
        )}
        {hasFailed && !expanded && (
          <p className="mt-1.5 text-xs text-red-600">
            일부 서비스 배포에 실패했습니다.
            <button onClick={() => setExpanded(true)} className="ml-1 text-red-700 underline hover:text-red-800">
              상세보기
            </button>
          </p>
        )}
      </div>

      {/* Per-Service Progress (collapsible) */}
      {expanded && (
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
                    <div key={i} className="flex items-start gap-2 text-xs">
                      <span className="mt-0.5 shrink-0">
                        <StepIcon status={step.status} />
                      </span>
                      <div className="min-w-0">
                        <span className={step.status === 'in_progress' ? 'font-medium text-blue-700' : 'text-gray-600'}>
                          {step.step.startsWith('apply:')
                            ? `Apply ${step.step.slice(6)}`
                            : step.step.replace(/^create_/, 'Create ').replace(/^\w/, (c) => c.toUpperCase())}
                        </span>
                        {step.message && step.status === 'failed' && (
                          <p className="mt-0.5 text-red-500 break-all">{step.message}</p>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}