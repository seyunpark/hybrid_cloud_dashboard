import type { DeployStep } from '@/api/types';

interface DeployProgressProps {
  steps: DeployStep[];
  status: string;
}

const stepLabels: Record<string, string> = {
  push_image: 'Push Image to Registry',
  create_configmap: 'Create ConfigMap',
  create_deployment: 'Create Deployment',
  create_service: 'Create Service',
  create_hpa: 'Create HPA',
};

function StepIcon({ status }: { status: string }) {
  switch (status) {
    case 'completed':
      return (
        <span className="flex h-6 w-6 items-center justify-center rounded-full bg-green-100 text-xs text-green-600">
          ✓
        </span>
      );
    case 'in_progress':
      return (
        <span className="flex h-6 w-6 items-center justify-center rounded-full bg-blue-100">
          <span className="h-3 w-3 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" />
        </span>
      );
    case 'failed':
      return (
        <span className="flex h-6 w-6 items-center justify-center rounded-full bg-red-100 text-xs text-red-600">
          ✕
        </span>
      );
    default:
      return (
        <span className="flex h-6 w-6 items-center justify-center rounded-full bg-gray-100 text-xs text-gray-400">
          ○
        </span>
      );
  }
}

export function DeployProgress({ steps, status }: DeployProgressProps) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-sm font-semibold text-gray-900">
          Deployment Progress
        </h3>
        <span className="text-xs text-gray-500">{status}</span>
      </div>

      <div className="space-y-3">
        {steps.map((step, index) => (
          <div key={index} className="flex items-start gap-3">
            <StepIcon status={step.status} />
            <div className="flex-1">
              <p className="text-sm font-medium text-gray-900">
                {stepLabels[step.step] || step.step}
              </p>
              {step.message && (
                <p className="text-xs text-gray-500">{step.message}</p>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
