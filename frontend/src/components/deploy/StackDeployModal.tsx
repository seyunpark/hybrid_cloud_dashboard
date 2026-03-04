import { useState } from 'react';
import type { Container, Cluster } from '@/api/types';
import { useK8sNamespaces } from '@/hooks/useK8sClusters';

interface StackDeployModalProps {
  isOpen: boolean;
  onClose: () => void;
  containers: Container[];
  clusters: Cluster[];
  onDeploy: (containerIds: string[], stackName: string, clusterName: string, namespace: string, createNamespace: boolean) => void;
  isLoading?: boolean;
  error?: string | null;
}

export function StackDeployModal({
  isOpen,
  onClose,
  containers,
  clusters,
  onDeploy,
  isLoading = false,
  error,
}: StackDeployModalProps) {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [stackName, setStackName] = useState('');
  const [selectedCluster, setSelectedCluster] = useState('');
  const [namespace, setNamespace] = useState('default');
  const [isNewNamespace, setIsNewNamespace] = useState(false);
  const [newNamespaceInput, setNewNamespaceInput] = useState('');

  const { data: nsData } = useK8sNamespaces(selectedCluster);
  const namespaceList = nsData ?? [];
  const effectiveNamespace = isNewNamespace ? newNamespaceInput.trim() : namespace;

  if (!isOpen) return null;

  const toggleContainer = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const handleSubmit = () => {
    const ids = Array.from(selectedIds);
    const name = stackName.trim() || `${containers.find(c => c.id === ids[0])?.name || 'app'}-stack`;
    onDeploy(ids, name, selectedCluster, effectiveNamespace, isNewNamespace);
  };

  const runningContainers = containers.filter((c) => c.status.startsWith('Up'));
  const canSubmit = selectedIds.size >= 2 && selectedCluster && effectiveNamespace && !isLoading;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="max-h-[90vh] w-full max-w-2xl overflow-auto rounded-lg bg-white p-6 shadow-xl">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">Deploy Stack</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">✕</button>
        </div>

        {error && (
          <div className="mb-4 rounded bg-red-50 p-3 text-sm text-red-700">{error}</div>
        )}

        {/* Stack Name */}
        <div className="mb-4">
          <label className="mb-1 block text-sm font-medium text-gray-700">Stack Name</label>
          <input
            type="text"
            value={stackName}
            onChange={(e) => setStackName(e.target.value)}
            placeholder="Auto-generated if empty"
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
            disabled={isLoading}
          />
        </div>

        {/* Container Selection */}
        <div className="mb-4">
          <label className="mb-1 block text-sm font-medium text-gray-700">
            Containers ({selectedIds.size} selected, minimum 2)
          </label>
          <div className="max-h-48 overflow-auto rounded-md border border-gray-300">
            {runningContainers.length === 0 ? (
              <p className="p-3 text-sm text-gray-500">No running containers found</p>
            ) : (
              runningContainers.map((container) => (
                <label
                  key={container.id}
                  className={`flex cursor-pointer items-center gap-3 border-b border-gray-100 px-3 py-2 last:border-b-0 hover:bg-gray-50 ${
                    selectedIds.has(container.id) ? 'bg-blue-50' : ''
                  }`}
                >
                  <input
                    type="checkbox"
                    checked={selectedIds.has(container.id)}
                    onChange={() => toggleContainer(container.id)}
                    className="h-4 w-4 rounded border-gray-300 text-blue-600"
                    disabled={isLoading}
                  />
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-sm font-medium text-gray-900">{container.name}</div>
                    <div className="truncate text-xs text-gray-500">
                      {container.image}
                      {container.ports && container.ports.length > 0 && (
                        <span className="ml-2">Ports: {container.ports.map(p => p.private_port).join(', ')}</span>
                      )}
                    </div>
                  </div>
                </label>
              ))
            )}
          </div>
        </div>

        {/* Cluster */}
        <div className="mb-4">
          <label className="mb-1 block text-sm font-medium text-gray-700">Target Cluster</label>
          <select
            value={selectedCluster}
            onChange={(e) => { setSelectedCluster(e.target.value); setNamespace('default'); setIsNewNamespace(false); }}
            className="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
            disabled={isLoading}
          >
            <option value="">클러스터를 선택하세요</option>
            {clusters.map((cluster) => (
              <option key={cluster.name} value={cluster.name}>
                {cluster.name} ({cluster.status})
              </option>
            ))}
          </select>
        </div>

        {/* Namespace */}
        <div className="mb-4">
          <label className="mb-1 block text-sm font-medium text-gray-700">Namespace</label>
          {isNewNamespace ? (
            <div className="space-y-1">
              <div className="flex gap-2">
                <input
                  type="text"
                  value={newNamespaceInput}
                  onChange={(e) => setNewNamespaceInput(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))}
                  placeholder="new-namespace"
                  className="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                  disabled={isLoading}
                  autoFocus
                />
                <button
                  type="button"
                  onClick={() => { setIsNewNamespace(false); setNewNamespaceInput(''); }}
                  className="rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-600 hover:bg-gray-50"
                  disabled={isLoading}
                >
                  취소
                </button>
              </div>
              <p className="text-xs text-gray-500">배포 시 이 네임스페이스가 자동으로 생성됩니다</p>
            </div>
          ) : (
            <div className="flex gap-2">
              {selectedCluster && namespaceList.length > 0 ? (
                <select
                  value={namespace}
                  onChange={(e) => setNamespace(e.target.value)}
                  className="flex-1 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                  disabled={isLoading}
                >
                  {namespaceList.map((ns) => (
                    <option key={ns} value={ns}>{ns}</option>
                  ))}
                </select>
              ) : (
                <input
                  type="text"
                  value={namespace}
                  onChange={(e) => setNamespace(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))}
                  className="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                  disabled={isLoading}
                  placeholder={selectedCluster ? 'default' : '먼저 클러스터를 선택하세요'}
                />
              )}
              <button
                type="button"
                onClick={() => setIsNewNamespace(true)}
                className="whitespace-nowrap rounded-md border border-dashed border-blue-300 px-3 py-2 text-sm font-medium text-blue-600 hover:bg-blue-50"
                disabled={isLoading}
              >
                + New
              </button>
            </div>
          )}
        </div>

        {/* Actions */}
        <div className="flex justify-end gap-3">
          <button
            onClick={onClose}
            disabled={isLoading}
            className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={!canSubmit}
            className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {isLoading ? (
              <span className="flex items-center gap-2">
                <span className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
                Generating...
              </span>
            ) : (
              'Generate Stack Manifests (AI)'
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
