import { useState } from 'react';
import { Link } from 'react-router-dom';
import type { StackManifests, StackTopology } from '@/api/types';

interface StackManifestPreviewProps {
  stackName: string;
  topology: StackTopology;
  manifests: StackManifests;
  reasoning?: string;
  confidence?: number;
  onApprove: () => void;
  onCancel: () => void;
  onRefine: (feedback: string) => Promise<void>;
  isRefining?: boolean;
  refineError?: string | null;
  readOnly?: boolean;
  onRetry?: () => void;
  isRetrying?: boolean;
  approveLabel?: string;
}

// Resource kinds display order — known kinds first, then alphabetical
const KIND_ORDER: Record<string, number> = {
  Deployment: 0,
  Service: 1,
  ConfigMap: 2,
  Secret: 3,
  HPA: 4,
  Ingress: 5,
  PersistentVolumeClaim: 6,
};

function sortedKinds(manifests: StackManifests): string[] {
  return Object.keys(manifests).sort((a, b) => {
    const oa = KIND_ORDER[a] ?? 100;
    const ob = KIND_ORDER[b] ?? 100;
    if (oa !== ob) return oa - ob;
    return a.localeCompare(b);
  });
}

function getServiceResources(manifests: StackManifests, serviceName: string) {
  const resources: { kind: string; name: string; yaml: string }[] = [];
  for (const kind of sortedKinds(manifests)) {
    const entries = manifests[kind];
    for (const [name, yaml] of Object.entries(entries)) {
      // _namespace 서비스는 Namespace kind 리소스와 매칭
      if (serviceName === '_namespace') {
        if (kind === 'Namespace') {
          resources.push({ kind, name, yaml });
        }
        continue;
      }
      if (name === serviceName || name.startsWith(serviceName + '-') || name.endsWith('-' + serviceName)) {
        resources.push({ kind, name, yaml });
      }
    }
  }
  return resources;
}

function getUnmatchedResources(manifests: StackManifests, serviceNames: string[]) {
  const resources: { kind: string; name: string; yaml: string }[] = [];
  const hasNamespaceService = serviceNames.includes('_namespace');
  for (const kind of sortedKinds(manifests)) {
    const entries = manifests[kind];
    for (const [name, yaml] of Object.entries(entries)) {
      // _namespace 서비스가 있으면 Namespace kind는 이미 매칭됨
      if (hasNamespaceService && kind === 'Namespace') continue;
      const matched = serviceNames.some(
        (svc) => name === svc || name.startsWith(svc + '-') || name.endsWith('-' + svc),
      );
      if (!matched) {
        resources.push({ kind, name, yaml });
      }
    }
  }
  return resources;
}

export function StackManifestPreview({
  stackName,
  topology,
  manifests,
  reasoning,
  confidence,
  onApprove,
  onCancel,
  onRefine,
  isRefining = false,
  refineError,
  readOnly = false,
  onRetry,
  isRetrying = false,
  approveLabel = 'Approve & Deploy Stack',
}: StackManifestPreviewProps) {
  const [activeTab, setActiveTab] = useState(topology.deploy_order[0] || '');
  const [feedback, setFeedback] = useState('');
  const [showFeedback, setShowFeedback] = useState(false);

  const handleRefine = async () => {
    if (!feedback.trim()) return;
    await onRefine(feedback.trim());
    setFeedback('');
    setShowFeedback(false);
  };

  const unmatchedResources = getUnmatchedResources(manifests, topology.deploy_order);
  const tabs = [
    ...topology.deploy_order,
    ...(unmatchedResources.length > 0 ? ['_shared'] : []),
  ];

  return (
    <div className="space-y-4">
      {/* Stack Header */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-gray-900">
          Stack: {stackName}
        </h3>
        <div className="flex items-center gap-2">
          {confidence !== undefined && (
            <span className="rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-800">
              AI Confidence: {Math.round(confidence * 100)}%
            </span>
          )}
          <span className="rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-600">
            {Object.values(manifests).reduce((acc, r) => acc + Object.keys(r).length, 0)} resources
          </span>
        </div>
      </div>

      {/* Topology View */}
      <div className="rounded-lg border border-blue-200 bg-blue-50 p-4">
        <h4 className="mb-2 text-sm font-semibold text-blue-900">Service Topology</h4>

        {/* Deploy Order */}
        <div className="mb-3 flex items-center gap-1 text-xs text-blue-800">
          <span className="font-medium">Deploy Order:</span>
          {topology.deploy_order.map((name, i) => (
            <span key={name} className="flex items-center gap-1">
              <span className="rounded bg-blue-200 px-1.5 py-0.5 font-mono">
                {i + 1}. {name}
              </span>
              {i < topology.deploy_order.length - 1 && <span className="text-blue-400">&rarr;</span>}
            </span>
          ))}
        </div>

        {/* Connections */}
        {topology.connections.length > 0 && (
          <div className="space-y-1">
            <span className="text-xs font-medium text-blue-700">Connections:</span>
            {topology.connections.map((conn, i) => (
              <div key={i} className="flex items-center gap-2 text-xs text-blue-800">
                <span className="font-mono">{conn.from}</span>
                <span className="text-blue-400">&rarr;</span>
                <span className="font-mono">{conn.to}:{conn.port}</span>
                {conn.env_var && (
                  <span className="rounded bg-blue-100 px-1 text-blue-600">
                    ${conn.env_var}
                  </span>
                )}
              </div>
            ))}
          </div>
        )}

        {topology.connections.length === 0 && (
          <p className="text-xs text-blue-600">No inter-service connections detected.</p>
        )}
      </div>

      {/* AI Fallback Warning */}
      {reasoning?.startsWith('[Fallback]') && (
        <div className="rounded-lg border border-amber-300 bg-amber-50 p-4">
          <div className="flex items-start gap-3">
            <span className="mt-0.5 text-amber-500">&#9888;</span>
            <div className="flex-1">
              <p className="text-sm font-medium text-amber-800">
                기본 템플릿으로 생성됨
              </p>
              <p className="mt-1 text-xs text-amber-700">
                {reasoning.replace('[Fallback] ', '')}
              </p>
              <div className="mt-2 flex items-center gap-2">
                {!readOnly && onRetry && (
                  <button
                    onClick={onRetry}
                    disabled={isRetrying}
                    className="inline-flex items-center gap-1.5 rounded-md bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700 disabled:opacity-50"
                  >
                    {isRetrying ? (
                      <>
                        <span className="inline-block h-3 w-3 animate-spin rounded-full border-2 border-white border-t-transparent" />
                        AI 재생성 중...
                      </>
                    ) : (
                      'AI로 재생성'
                    )}
                  </button>
                )}
                <Link
                  to="/settings"
                  className="inline-block rounded-md bg-amber-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-amber-700"
                >
                  Settings에서 AI 설정 확인
                </Link>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* AI Reasoning */}
      {reasoning && !reasoning.startsWith('[Fallback]') && (
        <div className="rounded-lg border border-gray-200 bg-gray-50 p-3">
          <p className="text-xs text-gray-700">{reasoning}</p>
        </div>
      )}

      {/* Tabbed Manifest View */}
      <div>
        {/* Tab Headers */}
        <div className="flex border-b border-gray-200">
          {tabs.map((name) => (
            <button
              key={name}
              onClick={() => setActiveTab(name)}
              className={`px-4 py-2 text-xs font-medium ${
                activeTab === name
                  ? 'border-b-2 border-blue-500 text-blue-600'
                  : 'text-gray-500 hover:text-gray-700'
              }`}
            >
              {name === '_shared' ? 'Shared' : name}
              {name !== '_shared' && topology.services.find(s => s.service_name === name)?.service_type && (
                <span className="ml-1 text-gray-400">
                  ({topology.services.find(s => s.service_name === name)?.service_type})
                </span>
              )}
            </button>
          ))}
        </div>

        {/* Tab Content — dynamic resource kinds */}
        <div className="space-y-3 pt-3">
          {(activeTab === '_shared' ? unmatchedResources : getServiceResources(manifests, activeTab)).map(
            ({ kind, name, yaml }) => (
              <div key={`${kind}-${name}`}>
                <h5 className="mb-1 flex items-center gap-2 text-xs font-medium text-gray-700">
                  <span className="rounded bg-gray-200 px-1.5 py-0.5 text-xs font-mono">
                    {kind}
                  </span>
                  <span>{name}</span>
                </h5>
                <pre className="max-h-48 overflow-auto rounded bg-gray-900 p-3 text-xs text-green-400">
                  {yaml}
                </pre>
              </div>
            ),
          )}
        </div>
      </div>

      {/* Refine Error */}
      {refineError && (
        <div className="rounded-lg border border-red-200 bg-red-50 p-3">
          <p className="text-sm font-medium text-red-800">수정 요청 실패</p>
          <p className="mt-1 text-xs text-red-600">{refineError}</p>
        </div>
      )}

      {/* Feedback Section */}
      {!readOnly && showFeedback && (
        <div className="rounded-lg border border-amber-200 bg-amber-50 p-4">
          <h4 className="mb-2 text-sm font-medium text-amber-900">
            수정 요청
          </h4>
          <textarea
            value={feedback}
            onChange={(e) => setFeedback(e.target.value)}
            placeholder="예: db를 StatefulSet으로 변경해줘 / backend replicas를 3으로 늘려줘 / frontend에 Ingress 추가해줘"
            className="w-full rounded-md border border-amber-300 bg-white px-3 py-2 text-sm placeholder:text-gray-400 focus:border-amber-500 focus:outline-none"
            rows={3}
            disabled={isRefining}
          />
          <div className="mt-2 flex justify-end gap-2">
            <button
              onClick={() => { setShowFeedback(false); setFeedback(''); }}
              disabled={isRefining}
              className="rounded-md border border-gray-300 px-3 py-1.5 text-xs font-medium text-gray-600 hover:bg-gray-50 disabled:opacity-50"
            >
              취소
            </button>
            <button
              onClick={handleRefine}
              disabled={isRefining || !feedback.trim()}
              className="rounded-md bg-amber-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-amber-700 disabled:opacity-50"
            >
              {isRefining ? (
                <span className="flex items-center gap-1.5">
                  <span className="h-3 w-3 animate-spin rounded-full border-2 border-white border-t-transparent" />
                  AI 수정 중...
                </span>
              ) : (
                'AI에게 수정 요청'
              )}
            </button>
          </div>
        </div>
      )}

      {/* Actions */}
      {!readOnly && (
        <div className="flex justify-between border-t border-gray-200 pt-4">
          <button
            onClick={() => setShowFeedback(!showFeedback)}
            disabled={isRefining}
            className="rounded-md border border-amber-300 px-4 py-2 text-sm font-medium text-amber-700 hover:bg-amber-50 disabled:opacity-50"
          >
            수정 요청
          </button>
          <div className="flex gap-3">
            <button
              onClick={onCancel}
              disabled={isRefining}
              className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
            >
              취소
            </button>
            <button
              onClick={onApprove}
              disabled={isRefining}
              className="rounded-md bg-green-600 px-4 py-2 text-sm font-medium text-white hover:bg-green-700 disabled:opacity-50"
            >
              {approveLabel}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
