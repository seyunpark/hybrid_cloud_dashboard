import { useParams, Link } from 'react-router-dom';
import { useK8sPods, useK8sDeployments, useK8sServices, useK8sNamespaces } from '@/hooks/useK8sClusters';
import { useQuery } from '@tanstack/react-query';
import { k8sApi } from '@/api/client';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { StatusBadge } from '@/components/common/StatusBadge';
import { formatRelativeTime } from '@/utils/formatters';
import { useState } from 'react';

export function ClusterDetailPage() {
  const { name } = useParams<{ name: string }>();
  const [namespace, setNamespace] = useState('default');

  const { data: clusters } = useQuery({
    queryKey: ['k8s', 'clusters'],
    queryFn: () => k8sApi.listClusters(),
  });

  const cluster = clusters?.find(c => c.name === name);
  const { data: namespaces, isLoading: nsLoading } = useK8sNamespaces(name ?? '');

  const { data: pods, isLoading: podsLoading } = useK8sPods(name ?? '', namespace);
  const { data: deployments, isLoading: deploymentsLoading } = useK8sDeployments(name ?? '', namespace);
  const { data: services, isLoading: servicesLoading } = useK8sServices(name ?? '', namespace);

  if (!name) {
    return (
      <div className="rounded-lg border border-red-200 bg-red-50 p-8 text-center">
        <p className="text-sm text-red-600">Cluster name is required.</p>
        <Link to="/" className="mt-2 inline-block text-sm text-blue-600 hover:underline">Back to Dashboard</Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link to="/" className="text-sm text-gray-500 hover:text-gray-700">&larr; Dashboard</Link>
        <h1 className="text-xl font-bold text-gray-900">{name}</h1>
        {cluster && <StatusBadge status={cluster.status} />}
      </div>

      {/* Cluster Info */}
      {cluster && (
        <div className="grid grid-cols-2 gap-4 md:grid-cols-5">
          <InfoCard label="Type" value={cluster.type} />
          <InfoCard label="Nodes" value={String(cluster.info.nodes)} />
          <InfoCard label="Pods" value={String(cluster.info.pods)} />
          <InfoCard label="Namespaces" value={String(cluster.info.namespaces)} />
          <InfoCard label="Version" value={cluster.info.version || 'N/A'} />
        </div>
      )}

      {/* Namespace Selector */}
      <div className="flex items-center gap-2">
        <label className="text-sm font-medium text-gray-700">Namespace:</label>
        {nsLoading ? (
          <span className="text-sm text-gray-400">Loading...</span>
        ) : (
          <select
            value={namespace}
            onChange={(e) => setNamespace(e.target.value)}
            className="rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none"
          >
            {namespaces?.map((ns) => (
              <option key={ns} value={ns}>{ns}</option>
            ))}
          </select>
        )}
      </div>

      {/* Deployments */}
      <section>
        <h2 className="mb-3 text-sm font-semibold text-gray-900">
          Deployments {deployments && `(${deployments.length})`}
        </h2>
        {deploymentsLoading ? (
          <LoadingSpinner size="sm" message="Loading deployments..." />
        ) : deployments && deployments.length > 0 ? (
          <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Name</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Image</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Replicas</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Created</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {deployments.map(d => (
                  <tr key={d.name} className="hover:bg-gray-50">
                    <td className="px-4 py-2 text-sm font-medium text-gray-900">{d.name}</td>
                    <td className="px-4 py-2 text-sm text-gray-500">{d.image}</td>
                    <td className="px-4 py-2 text-sm text-gray-500">{d.ready_replicas}/{d.replicas}</td>
                    <td className="px-4 py-2 text-sm text-gray-500">{formatRelativeTime(d.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-sm text-gray-500">No deployments in this namespace.</p>
        )}
      </section>

      {/* Pods */}
      <section>
        <h2 className="mb-3 text-sm font-semibold text-gray-900">
          Pods {pods && `(${pods.length})`}
        </h2>
        {podsLoading ? (
          <LoadingSpinner size="sm" message="Loading pods..." />
        ) : pods && pods.length > 0 ? (
          <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Name</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Status</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Node</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">IP</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Restarts</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {pods.map(p => (
                  <tr key={p.name} className="hover:bg-gray-50">
                    <td className="px-4 py-2 text-sm font-medium text-gray-900">{p.name}</td>
                    <td className="px-4 py-2"><StatusBadge status={p.status.toLowerCase()} /></td>
                    <td className="px-4 py-2 text-sm text-gray-500">{p.node || '-'}</td>
                    <td className="px-4 py-2 text-sm text-gray-500">{p.ip || '-'}</td>
                    <td className="px-4 py-2 text-sm text-gray-500">
                      {p.containers.reduce((sum, c) => sum + c.restart_count, 0)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-sm text-gray-500">No pods in this namespace.</p>
        )}
      </section>

      {/* Services */}
      <section>
        <h2 className="mb-3 text-sm font-semibold text-gray-900">
          Services {services && `(${services.length})`}
        </h2>
        {servicesLoading ? (
          <LoadingSpinner size="sm" message="Loading services..." />
        ) : services && services.length > 0 ? (
          <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Name</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Type</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Cluster IP</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Ports</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {services.map(svc => (
                  <tr key={svc.name} className="hover:bg-gray-50">
                    <td className="px-4 py-2 text-sm font-medium text-gray-900">{svc.name}</td>
                    <td className="px-4 py-2 text-sm text-gray-500">{svc.type}</td>
                    <td className="px-4 py-2 text-sm text-gray-500">{svc.cluster_ip}</td>
                    <td className="px-4 py-2 text-sm text-gray-500">
                      {svc.ports.map(p => `${p.port}:${p.target_port}/${p.protocol}`).join(', ')}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-sm text-gray-500">No services in this namespace.</p>
        )}
      </section>
    </div>
  );
}

function InfoCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-3">
      <p className="text-xs text-gray-500">{label}</p>
      <p className="text-sm font-medium text-gray-900">{value}</p>
    </div>
  );
}
