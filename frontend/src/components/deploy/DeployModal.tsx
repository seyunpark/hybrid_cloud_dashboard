import { useState, useEffect } from 'react';
import { useK8sNamespaces } from '@/hooks/useK8sClusters';
import type { Container, Cluster } from '@/api/types';

interface DeployModalProps {
  isOpen: boolean;
  onClose: () => void;
  containers: Container[];
  clusters: Cluster[];
  onDeploy: (containerId: string, clusterName: string, namespace: string) => void;
  preselectedContainerId?: string;
  isLoading?: boolean;
  error?: string | null;
}

export function DeployModal({
  isOpen,
  onClose,
  containers,
  clusters,
  onDeploy,
  preselectedContainerId = '',
  isLoading = false,
  error,
}: DeployModalProps) {
  const [selectedContainer, setSelectedContainer] = useState('');
  const [selectedCluster, setSelectedCluster] = useState('');
  const [namespace, setNamespace] = useState('default');
  const [isNewNamespace, setIsNewNamespace] = useState(false);
  const [newNamespaceInput, setNewNamespaceInput] = useState('');

  const { data: namespaces, isLoading: nsLoading } = useK8sNamespaces(selectedCluster);

  const isDuplicateNamespace = isNewNamespace && namespaces?.includes(newNamespaceInput.trim());
  const effectiveNamespace = isNewNamespace ? newNamespaceInput.trim() : namespace;

  useEffect(() => {
    if (preselectedContainerId) {
      setSelectedContainer(preselectedContainerId);
    }
  }, [preselectedContainerId]);

  // Reset namespace when cluster changes
  useEffect(() => {
    setNamespace('default');
    setIsNewNamespace(false);
    setNewNamespaceInput('');
  }, [selectedCluster]);

  if (!isOpen) return null;

  const handleDeploy = () => {
    if (selectedContainer && selectedCluster && effectiveNamespace) {
      onDeploy(selectedContainer, selectedCluster, effectiveNamespace);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">
            Deploy to Kubernetes
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            ✕
          </button>
        </div>

        {error && (
          <div className="mb-4 rounded bg-red-50 p-3 text-sm text-red-700">
            {error}
          </div>
        )}

        <div className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              Docker Container
            </label>
            <select
              value={selectedContainer}
              onChange={(e) => setSelectedContainer(e.target.value)}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              disabled={isLoading}
            >
              <option value="">Select a container...</option>
              {containers.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name} ({c.image})
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              Target Cluster
            </label>
            <select
              value={selectedCluster}
              onChange={(e) => setSelectedCluster(e.target.value)}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              disabled={isLoading}
            >
              <option value="">Select a cluster...</option>
              {clusters.map((cl) => (
                <option key={cl.name} value={cl.name}>
                  {cl.name} ({cl.status})
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              Namespace
            </label>
            {!selectedCluster ? (
              <select
                disabled
                className="w-full rounded-md border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-400"
              >
                <option>Select a cluster first</option>
              </select>
            ) : nsLoading ? (
              <div className="flex items-center gap-2 rounded-md border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-500">
                <span className="h-3 w-3 animate-spin rounded-full border-2 border-gray-400 border-t-transparent" />
                Loading namespaces...
              </div>
            ) : isNewNamespace ? (
              <div className="space-y-1">
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={newNamespaceInput}
                    onChange={(e) => setNewNamespaceInput(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))}
                    placeholder="new-namespace"
                    className={`flex-1 rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-1 ${
                      isDuplicateNamespace
                        ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                        : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
                    }`}
                    disabled={isLoading}
                    autoFocus
                  />
                  <button
                    type="button"
                    onClick={() => { setIsNewNamespace(false); setNewNamespaceInput(''); }}
                    className="rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-600 hover:bg-gray-50"
                  >
                    Cancel
                  </button>
                </div>
                {isDuplicateNamespace && (
                  <p className="text-xs text-red-600">This namespace already exists</p>
                )}
              </div>
            ) : (
              <div className="flex gap-2">
                <select
                  value={namespace}
                  onChange={(e) => setNamespace(e.target.value)}
                  className="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  disabled={isLoading}
                >
                  {namespaces?.map((ns) => (
                    <option key={ns} value={ns}>{ns}</option>
                  ))}
                </select>
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
        </div>

        <div className="mt-6 flex justify-end gap-3">
          <button
            onClick={onClose}
            className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
            disabled={isLoading}
          >
            Cancel
          </button>
          <button
            onClick={handleDeploy}
            disabled={!selectedContainer || !selectedCluster || !effectiveNamespace || isDuplicateNamespace || isLoading}
            className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {isLoading ? (
              <span className="flex items-center gap-2">
                <span className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
                Generating...
              </span>
            ) : (
              'Generate Manifest (AI)'
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
