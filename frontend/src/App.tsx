import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Layout } from '@/components/layout/Layout';
import { Dashboard } from '@/components/dashboard/Dashboard';
import { DeployPage } from '@/pages/DeployPage';
import { HistoryPage } from '@/pages/HistoryPage';
import { ContainerDetailPage } from '@/pages/ContainerDetailPage';
import { ClusterDetailPage } from '@/pages/ClusterDetailPage';
import { SettingsPage } from '@/pages/SettingsPage';
import { StackDeployDetailPage } from '@/pages/StackDeployDetailPage';
import { ErrorBoundary } from '@/components/common/ErrorBoundary';
import { ToastProvider } from '@/components/common/Toast';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 5000,
      refetchOnWindowFocus: false,
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <ErrorBoundary>
          <BrowserRouter>
            <Routes>
              <Route element={<Layout />}>
                <Route path="/" element={<Dashboard />} />
                <Route path="/container/:id" element={<ContainerDetailPage />} />
                <Route path="/cluster/:name" element={<ClusterDetailPage />} />
                <Route path="/deploy" element={<DeployPage />} />
                <Route path="/deploy/:deployId" element={<StackDeployDetailPage />} />
                <Route path="/history" element={<HistoryPage />} />
                <Route path="/settings" element={<SettingsPage />} />
              </Route>
            </Routes>
          </BrowserRouter>
        </ErrorBoundary>
      </ToastProvider>
    </QueryClientProvider>
  );
}

export default App;
