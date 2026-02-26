import { useDockerContainers } from '@/hooks/useDockerContainers';
import { useK8sClusters } from '@/hooks/useK8sClusters';
import { ContainerCard } from './ContainerCard';
import { ClusterOverview } from './ClusterOverview';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';

export function Dashboard() {
  const { data: containers, isLoading: containersLoading } =
    useDockerContainers();
  const { data: clusters, isLoading: clustersLoading } = useK8sClusters();

  return (
    <div className="space-y-8">
      {/* Docker Containers Section */}
      <section>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">
            Docker Containers
          </h2>
          <span className="text-sm text-gray-500">
            {containers?.length ?? 0} containers
          </span>
        </div>

        {containersLoading ? (
          <LoadingSpinner message="Loading containers..." />
        ) : containers && containers.length > 0 ? (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            {containers.map((container) => (
              <ContainerCard key={container.id} container={container} />
            ))}
          </div>
        ) : (
          <div className="rounded-lg border border-dashed border-gray-300 p-8 text-center">
            <p className="text-sm text-gray-500">
              No Docker containers found. Start a container to see it here.
            </p>
          </div>
        )}
      </section>

      {/* Kubernetes Clusters Section */}
      <section>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">
            Kubernetes Clusters
          </h2>
          <span className="text-sm text-gray-500">
            {clusters?.length ?? 0} clusters
          </span>
        </div>

        {clustersLoading ? (
          <LoadingSpinner message="Loading clusters..." />
        ) : clusters && clusters.length > 0 ? (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            {clusters.map((cluster) => (
              <ClusterOverview key={cluster.name} cluster={cluster} />
            ))}
          </div>
        ) : (
          <div className="rounded-lg border border-dashed border-gray-300 p-8 text-center">
            <p className="text-sm text-gray-500">
              No Kubernetes clusters configured. Add clusters in config.yaml.
            </p>
          </div>
        )}
      </section>
    </div>
  );
}
