import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  useKubeContexts,
  useRegisterCluster,
  useUnregisterCluster,
} from '@/hooks/useClusterManagement';
import { useK8sClusters } from '@/hooks/useK8sClusters';
import { configApi } from '@/api/client';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { StatusBadge } from '@/components/common/StatusBadge';
import { useToast } from '@/components/common/Toast';

export function SettingsPage() {
  const { data: contexts, isLoading: contextsLoading, error: contextsError } = useKubeContexts();
  const { data: clusters, isLoading: clustersLoading } = useK8sClusters();
  const registerMutation = useRegisterCluster();
  const unregisterMutation = useUnregisterCluster();
  const { addToast } = useToast();

  const [editingContext, setEditingContext] = useState<string | null>(null);
  const [clusterName, setClusterName] = useState('');

  // AI Config
  const queryClient = useQueryClient();
  const { data: aiConfig, isLoading: aiLoading } = useQuery({
    queryKey: ['config', 'ai'],
    queryFn: () => configApi.getAI(),
  });
  const aiMutation = useMutation({
    mutationFn: configApi.updateAI,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'ai'] });
    },
  });

  const [aiProvider, setAiProvider] = useState('');
  const [aiApiKey, setAiApiKey] = useState('');
  const [aiModel, setAiModel] = useState('');
  const [availableModels, setAvailableModels] = useState<string[]>([]);
  const [modelsLoading, setModelsLoading] = useState(false);
  const [modelsError, setModelsError] = useState('');

  useEffect(() => {
    if (aiConfig) {
      setAiProvider(aiConfig.provider || '');
      setAiModel(aiConfig.model || '');
    }
  }, [aiConfig]);

  // Fetch models when provider changes and we already have a configured key
  useEffect(() => {
    if (aiProvider && aiConfig?.configured) {
      fetchModels(aiProvider);
    }
  }, [aiProvider]); // eslint-disable-line react-hooks/exhaustive-deps

  const fetchModels = async (provider?: string, apiKey?: string) => {
    const p = provider || aiProvider;
    if (!p) return;

    setModelsLoading(true);
    setModelsError('');
    try {
      const models = await configApi.listAIModels(p, apiKey || undefined);
      setAvailableModels(models);
      // Auto-select first model if current model not in list
      if (models.length > 0 && !models.includes(aiModel)) {
        setAiModel(models[0]);
      }
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: { message?: string } } }; message?: string };
      const msg = error.response?.data?.error?.message || 'Failed to fetch models';
      setModelsError(msg);
      setAvailableModels([]);
    } finally {
      setModelsLoading(false);
    }
  };

  const handleAiSave = async () => {
    try {
      await aiMutation.mutateAsync({
        provider: aiProvider || undefined,
        api_key: aiApiKey || undefined,
        model: aiModel || undefined,
      });
      addToast('AI configuration updated', 'success');
      setAiApiKey('');
      // Refresh model list with newly saved key
      if (aiApiKey) {
        fetchModels(aiProvider, aiApiKey);
      }
    } catch {
      addToast('Failed to update AI configuration', 'error');
    }
  };

  const handleRegister = async (contextName: string) => {
    const name = clusterName || contextName;
    try {
      await registerMutation.mutateAsync({ name, context: contextName });
      addToast(`Cluster "${name}" registered successfully`, 'success');
      setEditingContext(null);
      setClusterName('');
    } catch (err: unknown) {
      const error = err as { response?: { data?: { error?: { message?: string } } }; message?: string };
      addToast(error.response?.data?.error?.message || 'Failed to register cluster', 'error');
    }
  };

  const handleUnregister = async (name: string) => {
    try {
      await unregisterMutation.mutateAsync(name);
      addToast(`Cluster "${name}" unregistered`, 'success');
    } catch (err: unknown) {
      const error = err as { message?: string };
      addToast(error.message || 'Failed to unregister cluster', 'error');
    }
  };

  const shortName = (name: string) => {
    const parts = name.split('_');
    return parts.length > 1 ? parts[parts.length - 1] : name;
  };

  return (
    <div className="space-y-8">
      {/* AI Configuration */}
      <section>
        <div className="mb-4">
          <h2 className="text-lg font-semibold text-gray-900">AI Configuration</h2>
          <p className="text-sm text-gray-500">
            Configure LLM provider for K8s manifest generation
          </p>
        </div>

        {aiLoading ? (
          <LoadingSpinner message="Loading AI config..." />
        ) : (
          <div className="rounded-lg border border-gray-200 bg-white p-6">
            {/* Status indicator */}
            <div className="mb-6 flex items-center gap-2">
              <span
                className={`inline-flex h-2.5 w-2.5 rounded-full ${
                  aiConfig?.configured ? 'bg-green-400' : 'bg-yellow-400'
                }`}
              />
              <span className="text-sm text-gray-600">
                {aiConfig?.configured
                  ? `Connected — ${aiConfig.provider} (${aiConfig.model})`
                  : 'Not configured — using template fallback'}
              </span>
              {aiConfig?.api_key && (
                <span className="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-500">
                  Key: {aiConfig.api_key}
                </span>
              )}
            </div>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              {/* Provider */}
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Provider
                </label>
                <select
                  value={aiProvider}
                  onChange={(e) => {
                    setAiProvider(e.target.value);
                    setAvailableModels([]);
                    setAiModel('');
                    setModelsError('');
                  }}
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                >
                  <option value="openai">OpenAI</option>
                  <option value="claude">Claude (Anthropic)</option>
                  <option value="gemini">Gemini (Google)</option>
                </select>
              </div>

              {/* API Key */}
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  API Key
                </label>
                <div className="flex gap-2">
                  <input
                    type="password"
                    value={aiApiKey}
                    onChange={(e) => setAiApiKey(e.target.value)}
                    placeholder={aiConfig?.configured ? 'Enter new key to update' : 'Enter your API key'}
                    className="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                  />
                  <button
                    onClick={() => fetchModels(aiProvider, aiApiKey || undefined)}
                    disabled={modelsLoading}
                    className="whitespace-nowrap rounded-md border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                  >
                    {modelsLoading ? 'Loading...' : 'Fetch Models'}
                  </button>
                </div>
                <p className="mt-1 text-xs text-gray-400">
                  {aiProvider === 'claude'
                    ? 'Anthropic API key (sk-ant-...)'
                    : aiProvider === 'gemini'
                    ? 'Google AI API key (AIza...)'
                    : 'OpenAI API key (sk-...)'}
                </p>
              </div>

              {/* Model */}
              <div className="md:col-span-2">
                <label className="mb-1 block text-sm font-medium text-gray-700">
                  Model
                </label>
                {modelsLoading ? (
                  <div className="flex items-center gap-2 rounded-md border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-500">
                    <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                    </svg>
                    Fetching available models from {aiProvider}...
                  </div>
                ) : modelsError ? (
                  <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600">
                    {modelsError}
                  </div>
                ) : availableModels.length > 0 ? (
                  <select
                    value={aiModel}
                    onChange={(e) => setAiModel(e.target.value)}
                    className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
                  >
                    {availableModels.map((m) => (
                      <option key={m} value={m}>{m}</option>
                    ))}
                  </select>
                ) : (
                  <div className="rounded-md border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-500">
                    {aiConfig?.configured
                      ? 'Click "Fetch Models" to load available models, or enter a new API key first'
                      : 'Enter your API key and click "Fetch Models" to load available models'}
                  </div>
                )}
              </div>
            </div>

            <div className="mt-4 flex justify-end">
              <button
                onClick={handleAiSave}
                disabled={aiMutation.isPending || !aiModel}
                className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                {aiMutation.isPending ? 'Saving...' : 'Save Configuration'}
              </button>
            </div>
          </div>
        )}
      </section>

      {/* Available Kubeconfig Contexts */}
      <section>
        <div className="mb-4">
          <h2 className="text-lg font-semibold text-gray-900">
            Available Kubeconfig Contexts
          </h2>
          <p className="text-sm text-gray-500">
            Select contexts from ~/.kube/config to register as clusters
          </p>
        </div>

        {contextsLoading ? (
          <LoadingSpinner message="Loading kubeconfig..." />
        ) : contextsError ? (
          <div className="rounded-lg border border-red-200 bg-red-50 p-4">
            <p className="text-sm text-red-700">
              Failed to load kubeconfig: {(contextsError as Error).message}
            </p>
          </div>
        ) : contexts && contexts.length > 0 ? (
          <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">Context</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">Cluster</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">User</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">Status</th>
                  <th className="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {contexts.map((ctx) => (
                  <tr key={ctx.name} className="hover:bg-gray-50">
                    <td className="px-4 py-3">
                      <div className="text-sm font-medium text-gray-900">{shortName(ctx.name)}</div>
                      <div className="max-w-xs truncate text-xs text-gray-400" title={ctx.name}>{ctx.name}</div>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600">{shortName(ctx.cluster)}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{shortName(ctx.user)}</td>
                    <td className="px-4 py-3">
                      {ctx.is_active ? (
                        <span className="inline-flex items-center rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">Registered</span>
                      ) : (
                        <span className="inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600">Available</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-right">
                      {ctx.is_active ? (
                        <span className="text-xs text-gray-400">Already registered</span>
                      ) : editingContext === ctx.name ? (
                        <div className="flex items-center justify-end gap-2">
                          <input
                            type="text"
                            value={clusterName}
                            onChange={(e) => setClusterName(e.target.value)}
                            placeholder={shortName(ctx.name)}
                            className="w-40 rounded border border-gray-300 px-2 py-1 text-xs focus:border-blue-500 focus:outline-none"
                          />
                          <button
                            onClick={() => handleRegister(ctx.name)}
                            disabled={registerMutation.isPending}
                            className="rounded bg-blue-600 px-3 py-1 text-xs font-medium text-white hover:bg-blue-700 disabled:opacity-50"
                          >
                            {registerMutation.isPending ? '...' : 'Confirm'}
                          </button>
                          <button
                            onClick={() => { setEditingContext(null); setClusterName(''); }}
                            className="rounded border border-gray-300 px-3 py-1 text-xs font-medium text-gray-600 hover:bg-gray-50"
                          >
                            Cancel
                          </button>
                        </div>
                      ) : (
                        <button
                          onClick={() => { setEditingContext(ctx.name); setClusterName(shortName(ctx.name)); }}
                          className="rounded bg-blue-600 px-3 py-1 text-xs font-medium text-white hover:bg-blue-700"
                        >
                          Register
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="rounded-lg border border-dashed border-gray-300 p-8 text-center">
            <p className="text-sm text-gray-500">No kubeconfig contexts found. Make sure ~/.kube/config exists.</p>
          </div>
        )}
      </section>

      {/* Registered Clusters */}
      <section>
        <div className="mb-4">
          <h2 className="text-lg font-semibold text-gray-900">Registered Clusters</h2>
          <p className="text-sm text-gray-500">Currently active clusters in the dashboard</p>
        </div>

        {clustersLoading ? (
          <LoadingSpinner message="Loading clusters..." />
        ) : clusters && clusters.length > 0 ? (
          <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">Name</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">Type</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">Context</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">Status</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">Info</th>
                  <th className="px-4 py-3 text-right text-xs font-medium uppercase text-gray-500">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {clusters.map((cluster) => (
                  <tr key={cluster.name} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm font-medium text-gray-900">{cluster.name}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{cluster.type}</td>
                    <td className="px-4 py-3">
                      <div className="max-w-xs truncate text-xs text-gray-500" title={cluster.context}>{cluster.context}</div>
                    </td>
                    <td className="px-4 py-3"><StatusBadge status={cluster.status} /></td>
                    <td className="px-4 py-3 text-xs text-gray-500">
                      {cluster.status === 'connected' ? (
                        <span>{cluster.info.nodes} nodes, {cluster.info.pods} pods{cluster.info.version && ` (${cluster.info.version})`}</span>
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => handleUnregister(cluster.name)}
                        disabled={unregisterMutation.isPending}
                        className="rounded border border-red-300 px-3 py-1 text-xs font-medium text-red-600 hover:bg-red-50 disabled:opacity-50"
                      >
                        Unregister
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="rounded-lg border border-dashed border-gray-300 p-8 text-center">
            <p className="text-sm text-gray-500">No clusters registered. Register a context from the list above.</p>
          </div>
        )}
      </section>
    </div>
  );
}
