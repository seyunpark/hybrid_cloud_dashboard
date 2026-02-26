import { useState } from 'react';
import type { Container, Cluster } from '@/api/types';

interface DeployModalProps {
  isOpen: boolean;
  onClose: () => void;
  containers: Container[];
  clusters: Cluster[];
  onDeploy: (containerId: string, clusterName: string, namespace: string) => void;
}

export function DeployModal({
  isOpen,
  onClose,
  containers,
  clusters,
  onDeploy,
}: DeployModalProps) {
  const [selectedContainer, setSelectedContainer] = useState('');
  const [selectedCluster, setSelectedCluster] = useState('');
  const [namespace, setNamespace] = useState('default');

  if (!isOpen) return null;

  const handleDeploy = () => {
    if (selectedContainer && selectedCluster) {
      onDeploy(selectedContainer, selectedCluster, namespace);
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

        <div className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              Docker Container
            </label>
            <select
              value={selectedContainer}
              onChange={(e) => setSelectedContainer(e.target.value)}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
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
            >
              <option value="">Select a cluster...</option>
              {clusters.map((cl) => (
                <option key={cl.name} value={cl.name}>
                  {cl.name}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              Namespace
            </label>
            <input
              type="text"
              value={namespace}
              onChange={(e) => setNamespace(e.target.value)}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>
        </div>

        <div className="mt-6 flex justify-end gap-3">
          <button
            onClick={onClose}
            className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            onClick={handleDeploy}
            disabled={!selectedContainer || !selectedCluster}
            className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            Generate Manifest (AI)
          </button>
        </div>
      </div>
    </div>
  );
}
