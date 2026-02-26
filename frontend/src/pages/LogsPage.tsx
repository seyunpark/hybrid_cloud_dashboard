import { LogViewer } from '@/components/logs/LogViewer';

export function LogsPage() {
  return (
    <div className="flex h-full flex-col gap-4">
      <h2 className="text-lg font-semibold text-gray-900">Log Viewer</h2>
      <div className="min-h-0 flex-1">
        <LogViewer title="Container Logs" />
      </div>
    </div>
  );
}
