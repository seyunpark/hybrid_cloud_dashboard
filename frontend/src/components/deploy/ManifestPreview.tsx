import { useState } from 'react';
import { Link } from 'react-router-dom';
import type { Manifests, Recommendations } from '@/api/types';

interface ManifestPreviewProps {
  manifests: Manifests;
  recommendations?: Recommendations;
  onApprove: () => void;
  onCancel: () => void;
  onRefine: (feedback: string) => Promise<void>;
  isRefining?: boolean;
}

export function ManifestPreview({
  manifests,
  recommendations,
  onApprove,
  onCancel,
  onRefine,
  isRefining = false,
}: ManifestPreviewProps) {
  const [feedback, setFeedback] = useState('');
  const [showFeedback, setShowFeedback] = useState(false);

  const handleRefine = async () => {
    if (!feedback.trim()) return;
    await onRefine(feedback.trim());
    setFeedback('');
    setShowFeedback(false);
  };

  return (
    <div className="space-y-4">
      {/* AI Fallback Warning */}
      {recommendations?.reasoning?.startsWith('[Fallback]') && (
        <div className="rounded-lg border border-amber-300 bg-amber-50 p-4">
          <div className="flex items-start gap-3">
            <span className="mt-0.5 text-amber-500">&#9888;</span>
            <div className="flex-1">
              <p className="text-sm font-medium text-amber-800">
                기본 템플릿으로 생성됨
              </p>
              <p className="mt-1 text-xs text-amber-700">
                {recommendations.reasoning.replace('[Fallback] ', '')}
              </p>
              <Link
                to="/settings"
                className="mt-2 inline-block rounded-md bg-amber-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-amber-700"
              >
                Settings에서 AI 설정 확인
              </Link>
            </div>
          </div>
        </div>
      )}

      {/* AI Recommendations */}
      {recommendations && !recommendations.reasoning?.startsWith('[Fallback]') && (
        <div className="rounded-lg border border-blue-200 bg-blue-50 p-4">
          <h3 className="mb-2 text-sm font-semibold text-blue-900">
            AI Recommendations
          </h3>
          <div className="mb-2 grid grid-cols-2 gap-2 text-xs">
            <div>
              <span className="text-blue-700">CPU:</span>{' '}
              {recommendations.cpu_request} / {recommendations.cpu_limit}
            </div>
            <div>
              <span className="text-blue-700">Memory:</span>{' '}
              {recommendations.memory_request} / {recommendations.memory_limit}
            </div>
            <div>
              <span className="text-blue-700">Replicas:</span>{' '}
              {recommendations.replicas}
            </div>
            <div>
              <span className="text-blue-700">HPA:</span>{' '}
              {recommendations.enable_hpa ? 'Enabled' : 'Disabled'}
            </div>
          </div>
          {recommendations.reasoning && (
            <p className="text-xs text-blue-800">{recommendations.reasoning}</p>
          )}
        </div>
      )}

      {/* Manifest YAML */}
      <div className="space-y-3">
        {manifests.deployment && (
          <div>
            <h4 className="mb-1 text-xs font-medium text-gray-700">
              Deployment
            </h4>
            <pre className="max-h-48 overflow-auto rounded bg-gray-900 p-3 text-xs text-green-400">
              {manifests.deployment}
            </pre>
          </div>
        )}
        {manifests.service && (
          <div>
            <h4 className="mb-1 text-xs font-medium text-gray-700">Service</h4>
            <pre className="max-h-48 overflow-auto rounded bg-gray-900 p-3 text-xs text-green-400">
              {manifests.service}
            </pre>
          </div>
        )}
        {manifests.hpa && (
          <div>
            <h4 className="mb-1 text-xs font-medium text-gray-700">HPA</h4>
            <pre className="max-h-48 overflow-auto rounded bg-gray-900 p-3 text-xs text-green-400">
              {manifests.hpa}
            </pre>
          </div>
        )}
      </div>

      {/* Feedback Section */}
      {showFeedback ? (
        <div className="rounded-lg border border-amber-200 bg-amber-50 p-4">
          <h4 className="mb-2 text-sm font-medium text-amber-900">
            수정 요청
          </h4>
          <textarea
            value={feedback}
            onChange={(e) => setFeedback(e.target.value)}
            placeholder="예: replicas를 3으로 변경해줘 / LoadBalancer 타입으로 바꿔줘 / HPA를 추가해줘"
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
      ) : null}

      {/* Actions */}
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
            Cancel
          </button>
          <button
            onClick={onApprove}
            disabled={isRefining}
            className="rounded-md bg-green-600 px-4 py-2 text-sm font-medium text-white hover:bg-green-700 disabled:opacity-50"
          >
            Approve & Deploy
          </button>
        </div>
      </div>
    </div>
  );
}
