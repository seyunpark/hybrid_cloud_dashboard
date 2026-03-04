import { useState } from 'react';
import { Link } from 'react-router-dom';
import { StatusBadge } from '@/components/common/StatusBadge';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { formatRelativeTime } from '@/utils/formatters';
import { useUnifiedHistory } from '@/hooks/useUnifiedHistory';
import type { UnifiedDeployItem } from '@/api/types';

function DeployItemRow({ item }: { item: UnifiedDeployItem }) {
  const isStack = item.type === 'stack';

  return (
    <Link
      to={isStack ? `/deploy/${item.id}` : '#'}
      className={`block rounded-lg border border-gray-200 bg-white p-4 transition ${
        isStack
          ? 'hover:border-blue-300 hover:shadow-md cursor-pointer'
          : 'cursor-default'
      }`}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3 min-w-0">
          <span
            className={`flex-shrink-0 rounded px-1.5 py-0.5 text-xs font-medium ${
              isStack
                ? 'bg-purple-100 text-purple-700'
                : 'bg-gray-100 text-gray-600'
            }`}
          >
            {isStack ? 'Stack' : 'Single'}
          </span>

          <span className="truncate text-sm font-medium text-gray-900">
            {item.name}
          </span>

          {isStack && item.stack_detail && (
            <div className="hidden items-center gap-1 sm:flex">
              {item.stack_detail.services.slice(0, 4).map((svc) => (
                <span
                  key={svc}
                  className="rounded bg-blue-50 px-1.5 py-0.5 text-xs text-blue-600"
                >
                  {svc}
                </span>
              ))}
              {item.stack_detail.services.length > 4 && (
                <span className="text-xs text-gray-400">
                  +{item.stack_detail.services.length - 4}
                </span>
              )}
            </div>
          )}

          {!isStack && item.single_detail && (
            <span className="hidden text-xs text-gray-500 sm:inline">
              {item.image_summary}
            </span>
          )}
        </div>

        <div className="flex flex-shrink-0 items-center gap-3">
          {item.cluster && (
            <span className="hidden text-xs text-gray-400 lg:inline">
              {item.cluster}
            </span>
          )}

          {item.ai_generated && (
            <span className="hidden rounded-full bg-blue-50 px-2 py-0.5 text-xs text-blue-600 md:inline">
              AI {Math.round(item.confidence * 100)}%
            </span>
          )}

          <StatusBadge status={item.status} />

          <span className="w-20 text-right text-xs text-gray-400">
            {formatRelativeTime(item.deployed_at)}
          </span>
        </div>
      </div>

      <div className="mt-2 flex flex-wrap items-center gap-2 sm:hidden">
        {isStack && item.stack_detail && (
          <span className="text-xs text-gray-500">
            {item.stack_detail.service_count} services
          </span>
        )}
        {!isStack && (
          <span className="text-xs text-gray-500">{item.image_summary}</span>
        )}
        {item.cluster && (
          <span className="text-xs text-gray-400">{item.cluster}</span>
        )}
      </div>
    </Link>
  );
}

function Pagination({
  page,
  totalPages,
  onChange,
}: {
  page: number;
  totalPages: number;
  onChange: (page: number) => void;
}) {
  return (
    <div className="flex items-center justify-between border-t border-gray-200 pt-4">
      <button
        onClick={() => onChange(page - 1)}
        disabled={page <= 1}
        className="rounded-md border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-600 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
      >
        Previous
      </button>
      <span className="text-sm text-gray-500">
        Page {page} of {totalPages}
      </span>
      <button
        onClick={() => onChange(page + 1)}
        disabled={page >= totalPages}
        className="rounded-md border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-600 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
      >
        Next
      </button>
    </div>
  );
}

export function HistoryPage() {
  const [page, setPage] = useState(1);
  const { data, isLoading } = useUnifiedHistory(page, 20);

  const items = data?.items ?? [];
  const totalPages = data?.total_pages ?? 1;

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-lg font-semibold text-gray-900">
          Deployment History
        </h2>
        {data && (
          <p className="text-xs text-gray-500">{data.total} total</p>
        )}
      </div>

      {isLoading && !data ? (
        <LoadingSpinner message="Loading deployment history..." />
      ) : items.length > 0 ? (
        <div className="space-y-2">
          {items.map((item) => (
            <DeployItemRow key={item.id} item={item} />
          ))}
        </div>
      ) : (
        <div className="rounded-lg border border-dashed border-gray-300 p-8 text-center">
          <p className="text-sm text-gray-500">
            No deployments yet.
          </p>
        </div>
      )}

      {totalPages > 1 && (
        <Pagination page={page} totalPages={totalPages} onChange={setPage} />
      )}
    </div>
  );
}
