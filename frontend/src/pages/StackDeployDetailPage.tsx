import { useCallback, useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { stackDeployApi } from '@/api/client';
import type { StackManifests } from '@/api/types';
import { StatusBadge } from '@/components/common/StatusBadge';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { StackManifestPreview } from '@/components/deploy/StackManifestPreview';
import { StackDeployProgress } from '@/components/deploy/StackDeployProgress';
import {
  useStackDeployDetail,
  useRefineStackDeploy,
  useRegenerateStackDeploy,
  useExecuteStackDeploy,
  useReopenStackDeploy,
} from '@/hooks/useStackDeploy';
import { useK8sClusters } from '@/hooks/useK8sClusters';

// --- Utility ---

function downloadManifests(stackName: string, manifests: StackManifests) {
  const parts: string[] = [];
  for (const [kind, resources] of Object.entries(manifests)) {
    for (const [name, yaml] of Object.entries(resources)) {
      parts.push(`# ${kind}: ${name}\n${yaml}`);
    }
  }
  const content = parts.join('\n---\n');
  const blob = new Blob([content], { type: 'application/x-yaml' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `${stackName}-manifests.yaml`;
  a.click();
  URL.revokeObjectURL(url);
}

// --- Stepper ---

const STEPS = [
  { id: 1, label: 'K8s 매니페스트 변환' },
  { id: 2, label: '매니페스트 확인 & 피드백' },
  { id: 3, label: '배포 실행' },
] as const;

function getActiveStep(status: string): number {
  switch (status) {
    case 'generating':
      return 1;
    case 'pending':
    case 'analyzing':
      return 2;
    case 'deploying':
      return 3;
    case 'deployed':
    case 'completed':
    case 'undeployed':
      return 4;
    case 'failed':
    case 'cancelled':
      return -1;
    default:
      return 0;
  }
}

function Stepper({ activeStep, status }: { activeStep: number; status: string }) {
  return (
    <div className="flex items-center gap-0">
      {STEPS.map((step, idx) => {
        const isCompleted = activeStep > step.id;
        const isActive = activeStep === step.id;
        const isFailed = status === 'failed' && isActive;
        const isCancelled = status === 'cancelled' && step.id === 2;

        let circleClass = 'bg-gray-200 text-gray-500';
        if (isCompleted) circleClass = 'bg-green-500 text-white';
        else if (isFailed) circleClass = 'bg-red-500 text-white';
        else if (isCancelled) circleClass = 'bg-gray-400 text-white';
        else if (isActive) circleClass = 'bg-blue-500 text-white';

        let labelClass = 'text-gray-400';
        if (isCompleted) labelClass = 'text-green-700';
        else if (isFailed) labelClass = 'text-red-700';
        else if (isCancelled) labelClass = 'text-gray-500';
        else if (isActive) labelClass = 'text-blue-700 font-semibold';

        let lineClass = 'bg-gray-200';
        if (activeStep > step.id + 1 || (activeStep === 4 && idx < STEPS.length - 1)) {
          lineClass = 'bg-green-500';
        } else if (activeStep === step.id + 1) {
          lineClass = 'bg-blue-300';
        }

        return (
          <div key={step.id} className="flex items-center">
            <div className="flex flex-col items-center">
              <div
                className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-medium transition-colors ${circleClass}`}
              >
                {isCompleted ? (
                  <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                ) : (
                  step.id
                )}
              </div>
              <span className={`mt-1.5 text-xs whitespace-nowrap ${labelClass}`}>
                {step.label}
              </span>
            </div>
            {idx < STEPS.length - 1 && (
              <div className={`mx-2 h-0.5 w-16 rounded ${lineClass} -mt-5`} />
            )}
          </div>
        );
      })}
    </div>
  );
}

// --- Cluster Selection for Redeploy (namespace is already in manifests) ---

function RedeployClusterPanel({
  onConfirm,
  onBack,
  isPending,
}: {
  onConfirm: (clusterName: string) => void;
  onBack: () => void;
  isPending: boolean;
}) {
  const [selectedCluster, setSelectedCluster] = useState('');
  const { data: clusters } = useK8sClusters();

  return (
    <div className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
      <h3 className="mb-2 text-base font-semibold text-gray-900">재배포 클러스터 선택</h3>
      <p className="mb-4 text-xs text-gray-500">
        매니페스트에 설정된 네임스페이스로 배포됩니다. 배포할 클러스터를 선택하세요.
      </p>

      <div>
        <label className="mb-1 block text-sm font-medium text-gray-700">클러스터</label>
        <select
          value={selectedCluster}
          onChange={(e) => setSelectedCluster(e.target.value)}
          className="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
          disabled={isPending}
        >
          <option value="">클러스터를 선택하세요</option>
          {(clusters ?? []).map((cluster) => (
            <option key={cluster.name} value={cluster.name}>
              {cluster.name} ({cluster.status})
            </option>
          ))}
        </select>
      </div>

      <div className="mt-4 flex justify-end gap-3">
        <button
          onClick={onBack}
          disabled={isPending}
          className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
        >
          뒤로
        </button>
        <button
          onClick={() => selectedCluster && onConfirm(selectedCluster)}
          disabled={!selectedCluster || isPending}
          className="rounded-md bg-green-600 px-4 py-2 text-sm font-medium text-white hover:bg-green-700 disabled:opacity-50"
        >
          {isPending ? (
            <span className="flex items-center gap-2">
              <span className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
              재배포 중...
            </span>
          ) : (
            '재배포 실행'
          )}
        </button>
      </div>
    </div>
  );
}

// --- Main Page ---

export function StackDeployDetailPage() {
  const { deployId } = useParams<{ deployId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { data, isLoading, error } = useStackDeployDetail(deployId);
  const refineMutation = useRefineStackDeploy(deployId);
  const regenerateMutation = useRegenerateStackDeploy(deployId);
  const executeMutation = useExecuteStackDeploy(deployId);
  const reopenMutation = useReopenStackDeploy(deployId);

  // 'redeploy' = 재배포 클러스터 선택 중
  const [pendingAction, setPendingAction] = useState<'redeploy' | null>(null);

  const invalidateAll = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ['deploy', 'active'] });
    queryClient.invalidateQueries({ queryKey: ['deploy', 'stack', deployId] });
  }, [queryClient, deployId]);

  const undeployMutation = useMutation({
    mutationFn: () => stackDeployApi.undeployStack(deployId!),
    onSuccess: invalidateAll,
  });

  const redeployMutation = useMutation({
    mutationFn: (params: { cluster_name: string }) =>
      stackDeployApi.redeployStack(deployId!, params),
    onSuccess: () => {
      setPendingAction(null);
      invalidateAll();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => stackDeployApi.deleteStackDeploy(deployId!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['deploy', 'active'] });
      navigate('/deploy');
    },
  });

  const deployStatus = data?.status?.status ?? '';
  const activeStep = getActiveStep(deployStatus);
  const isActionable = deployStatus === 'pending' || deployStatus === 'analyzing';
  const isDeploying = deployStatus === 'deploying';
  const isDone = ['deployed', 'completed', 'failed', 'cancelled', 'undeployed'].includes(deployStatus);
  const anyMutationPending = undeployMutation.isPending || redeployMutation.isPending || deleteMutation.isPending || reopenMutation.isPending;

  // --- Handlers ---

  const handleRefine = useCallback(
    async (feedback: string) => {
      await refineMutation.mutateAsync(feedback);
    },
    [refineMutation],
  );

  const handleApprove = useCallback(async () => {
    await executeMutation.mutateAsync({ approved: true });
  }, [executeMutation]);

  const handleCancel = useCallback(async () => {
    await executeMutation.mutateAsync({ approved: false });
  }, [executeMutation]);

  const handleConfirmRedeploy = useCallback(
    async (clusterName: string) => {
      await redeployMutation.mutateAsync({ cluster_name: clusterName });
      setPendingAction(null);
    },
    [redeployMutation],
  );

  // --- Render ---

  if (isLoading) {
    return <LoadingSpinner message="Loading deployment details..." />;
  }

  if (error || !data) {
    return (
      <div className="space-y-4">
        <Link to="/deploy" className="text-sm text-blue-600 hover:text-blue-800">&larr; Deployments</Link>
        <div className="rounded-lg border border-red-200 bg-red-50 p-6 text-center">
          <p className="text-sm text-red-700">Deployment not found or failed to load.</p>
        </div>
      </div>
    );
  }

  const { response, status } = data;
  const hasManifests = !!(response.topology && response.manifests);

  // Extract refine error message
  const refineErrorMsg = refineMutation.error
    ? (refineMutation.error as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error?.message
      || refineMutation.error.message
      || '알 수 없는 오류가 발생했습니다'
    : null;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link to="/deploy" className="text-sm text-blue-600 hover:text-blue-800">&larr; Deployments</Link>
          <h2 className="text-lg font-semibold text-gray-900">Stack: {response.stack_name}</h2>
          <StatusBadge status={status.status} />
        </div>
        {status.started_at && (
          <span className="text-xs text-gray-500">
            Started: {new Date(status.started_at).toLocaleString()}
          </span>
        )}
      </div>

      {/* Stepper */}
      <div className="flex justify-center rounded-lg border border-gray-200 bg-white px-6 py-4 shadow-sm">
        <Stepper activeStep={activeStep} status={status.status} />
      </div>

      {/* Phase: Generating */}
      {status.status === 'generating' && (
        <div className="rounded-lg border border-blue-200 bg-blue-50 p-8 text-center">
          <div className="mx-auto mb-3 h-10 w-10 animate-spin rounded-full border-4 border-blue-500 border-t-transparent" />
          <p className="text-base font-semibold text-blue-800">K8s 매니페스트 변환 중...</p>
          <p className="mt-2 text-sm text-blue-600">
            컨테이너 정보를 분석하고 최적의 Kubernetes 매니페스트를 생성하고 있습니다.
          </p>
          <p className="mt-1 text-xs text-blue-400">이 과정은 약 10~30초 정도 소요됩니다.</p>
        </div>
      )}

      {/* Manifest Preview Card */}
      {hasManifests && status.status !== 'cancelled' && (
        <div className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
          {/* Download */}
          <div className="mb-4 flex justify-end">
            <button
              onClick={() => downloadManifests(response.stack_name, response.manifests!)}
              className="flex items-center gap-1.5 rounded-md border border-gray-300 px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50"
            >
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              YAML 다운로드
            </button>
          </div>

          {/* Review hint */}
          {isActionable && !pendingAction && (
            <div className="mb-4 rounded-md bg-amber-50 border border-amber-200 px-4 py-3">
              <p className="text-sm font-medium text-amber-800">생성된 매니페스트를 확인하세요</p>
              <p className="text-xs text-amber-600 mt-1">
                아래 매니페스트를 검토한 후 수정 요청 또는 배포 승인을 해주세요.
              </p>
            </div>
          )}

          <StackManifestPreview
            stackName={response.stack_name}
            topology={response.topology!}
            manifests={response.manifests!}
            reasoning={response.reasoning}
            confidence={response.confidence}
            onApprove={handleApprove}
            onCancel={handleCancel}
            onRefine={handleRefine}
            isRefining={refineMutation.isPending || executeMutation.isPending}
            refineError={refineErrorMsg}
            readOnly={!isActionable || !!pendingAction}
            onRetry={isActionable && !pendingAction ? () => regenerateMutation.mutate() : undefined}
            isRetrying={regenerateMutation.isPending}
          />
        </div>
      )}

      {/* Redeploy Cluster Selection — appears BELOW manifests */}
      {pendingAction === 'redeploy' && (
        <RedeployClusterPanel
          onConfirm={handleConfirmRedeploy}
          onBack={() => setPendingAction(null)}
          isPending={redeployMutation.isPending}
        />
      )}

      {/* Deploy Progress */}
      {(isDeploying || isDone) && status.status !== 'cancelled' && status.status !== 'undeployed' && (
        <div className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
          <StackDeployProgress status={status} />

          {/* Success */}
          {(status.status === 'deployed' || status.status === 'completed') && status.completed_at && (
            <div className="mt-4 rounded-lg border border-green-200 bg-green-50 p-4">
              <p className="text-sm font-medium text-green-800">배포가 성공적으로 완료되었습니다</p>
              <p className="text-xs text-green-600">
                완료 시각: {new Date(status.completed_at).toLocaleString()}
              </p>
            </div>
          )}

          {/* Deploy Summary */}
          {(status.status === 'deployed' || status.status === 'completed') && (
            <div className="mt-4 rounded-lg border border-gray-200 bg-white p-4">
              <h4 className="mb-3 text-sm font-semibold text-gray-800">배포 요약</h4>
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div>
                  <span className="text-gray-500">클러스터</span>
                  <p className="font-medium text-gray-900">{data.cluster_name || '-'}</p>
                </div>
                <div>
                  <span className="text-gray-500">네임스페이스</span>
                  <p className="font-medium text-gray-900">{data.namespace || 'default'}</p>
                </div>
                <div>
                  <span className="text-gray-500">서비스 수</span>
                  <p className="font-medium text-gray-900">
                    {status.deploy_order.filter(n => n !== '_namespace').length}개
                  </p>
                </div>
                <div>
                  <span className="text-gray-500">배포된 서비스</span>
                  <div className="mt-0.5 flex flex-wrap gap-1">
                    {status.deploy_order
                      .filter(n => n !== '_namespace')
                      .map(name => (
                        <span key={name} className="rounded bg-green-100 px-1.5 py-0.5 text-xs font-medium text-green-800">
                          {name}
                        </span>
                      ))}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Failed */}
          {status.status === 'failed' && (
            <div className="mt-4 rounded-lg border border-red-200 bg-red-50 p-4">
              <p className="text-sm font-medium text-red-800">배포에 실패했습니다</p>
            </div>
          )}
        </div>
      )}

      {/* Cancelled */}
      {status.status === 'cancelled' && (
        <div className="rounded-lg border border-gray-200 bg-gray-50 p-6 text-center">
          <p className="text-sm text-gray-600">이 배포가 취소되었습니다.</p>
        </div>
      )}

      {/* Undeployed */}
      {status.status === 'undeployed' && (
        <div className="rounded-lg border border-gray-300 bg-gray-50 p-6 text-center">
          <p className="text-sm font-medium text-gray-700">배포가 중지되었습니다</p>
          <p className="mt-1 text-xs text-gray-500">
            K8s 리소스가 삭제되었습니다. 매니페스트를 수정하거나 재배포할 수 있습니다.
          </p>
        </div>
      )}

      {/* Management Panel */}
      {isDone && !pendingAction && (
        <div className="rounded-lg border border-gray-200 bg-white p-4 shadow-sm">
          <h4 className="mb-3 text-sm font-semibold text-gray-700">관리</h4>

          {/* Mutation errors */}
          {(undeployMutation.error || redeployMutation.error || deleteMutation.error || reopenMutation.error) && (
            <div className="mb-3 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700">
              {(undeployMutation.error || redeployMutation.error || deleteMutation.error || reopenMutation.error)?.message || '작업에 실패했습니다'}
            </div>
          )}

          <div className="flex flex-wrap gap-3">
            {/* Deployed/Completed → Undeploy */}
            {(deployStatus === 'deployed' || deployStatus === 'completed') && (
              <button
                onClick={() => {
                  if (window.confirm('배포를 중지하시겠습니까? K8s에 배포된 리소스가 삭제됩니다.')) {
                    undeployMutation.mutate();
                  }
                }}
                disabled={anyMutationPending}
                className="rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
              >
                {undeployMutation.isPending ? '배포 중지 중...' : '배포 중지'}
              </button>
            )}

            {/* Undeployed → Reopen / Redeploy / Delete */}
            {deployStatus === 'undeployed' && (
              <>
                <button
                  onClick={() => reopenMutation.mutate()}
                  disabled={anyMutationPending}
                  className="rounded-md bg-amber-500 px-4 py-2 text-sm font-medium text-white hover:bg-amber-600 disabled:opacity-50"
                >
                  {reopenMutation.isPending ? '복구 중...' : '매니페스트 수정'}
                </button>
                <button
                  onClick={() => setPendingAction('redeploy')}
                  disabled={anyMutationPending}
                  className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
                >
                  재배포
                </button>
                <button
                  onClick={() => {
                    if (window.confirm('이 배포 기록을 완전히 삭제하시겠습니까?')) {
                      deleteMutation.mutate();
                    }
                  }}
                  disabled={anyMutationPending}
                  className="rounded-md border border-red-300 px-4 py-2 text-sm font-medium text-red-700 hover:bg-red-50 disabled:opacity-50"
                >
                  {deleteMutation.isPending ? '삭제 중...' : '기록 삭제'}
                </button>
              </>
            )}

            {/* Failed → Redeploy / Delete */}
            {deployStatus === 'failed' && (
              <>
                {response.manifests && (
                  <button
                    onClick={() => setPendingAction('redeploy')}
                    disabled={anyMutationPending}
                    className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
                  >
                    재배포
                  </button>
                )}
                <button
                  onClick={() => {
                    if (window.confirm('이 배포 기록을 완전히 삭제하시겠습니까?')) {
                      deleteMutation.mutate();
                    }
                  }}
                  disabled={anyMutationPending}
                  className="rounded-md border border-red-300 px-4 py-2 text-sm font-medium text-red-700 hover:bg-red-50 disabled:opacity-50"
                >
                  {deleteMutation.isPending ? '삭제 중...' : '기록 삭제'}
                </button>
              </>
            )}

            {/* Cancelled → Delete */}
            {deployStatus === 'cancelled' && (
              <button
                onClick={() => {
                  if (window.confirm('이 배포 기록을 완전히 삭제하시겠습니까?')) {
                    deleteMutation.mutate();
                  }
                }}
                disabled={anyMutationPending}
                className="rounded-md border border-red-300 px-4 py-2 text-sm font-medium text-red-700 hover:bg-red-50 disabled:opacity-50"
              >
                {deleteMutation.isPending ? '삭제 중...' : '기록 삭제'}
              </button>
            )}
          </div>

          {/* Guidance */}
          {(deployStatus === 'deployed' || deployStatus === 'completed') && (
            <p className="mt-2 text-xs text-gray-500">
              배포된 상태에서는 먼저 배포를 중지해야 합니다. 배포 중지 후 재배포 또는 삭제가 가능합니다.
            </p>
          )}
        </div>
      )}
    </div>
  );
}
