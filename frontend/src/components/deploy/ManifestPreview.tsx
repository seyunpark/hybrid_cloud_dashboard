import type { Manifests, Recommendations } from '@/api/types';

interface ManifestPreviewProps {
  manifests: Manifests;
  recommendations?: Recommendations;
  onApprove: () => void;
  onCancel: () => void;
}

export function ManifestPreview({
  manifests,
  recommendations,
  onApprove,
  onCancel,
}: ManifestPreviewProps) {
  return (
    <div className="space-y-4">
      {/* AI Recommendations */}
      {recommendations && (
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

      {/* Actions */}
      <div className="flex justify-end gap-3 border-t border-gray-200 pt-4">
        <button
          onClick={onCancel}
          className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
        >
          Cancel
        </button>
        <button
          onClick={onApprove}
          className="rounded-md bg-green-600 px-4 py-2 text-sm font-medium text-white hover:bg-green-700"
        >
          Approve & Deploy
        </button>
      </div>
    </div>
  );
}
