import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { StackManifestPreview } from './StackManifestPreview';
import type { StackTopology, StackManifests } from '@/api/types';

describe('StackManifestPreview', () => {
  const mockTopology: StackTopology = {
    services: [
      { container_id: 'c1', service_name: 'db', service_type: 'database', image: 'postgres:15' },
      { container_id: 'c2', service_name: 'backend', service_type: 'api-server', image: 'myapp:latest' },
    ],
    connections: [
      { from: 'backend', to: 'db', port: 5432, env_var: 'DATABASE_URL' },
    ],
    deploy_order: ['db', 'backend'],
  };

  const mockManifests: StackManifests = {
    deployments: {
      db: 'apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: db',
      backend: 'apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: backend',
    },
    services: {
      db: 'apiVersion: v1\nkind: Service\nmetadata:\n  name: db',
      backend: 'apiVersion: v1\nkind: Service\nmetadata:\n  name: backend',
    },
  };

  const mockOnRefine = vi.fn().mockResolvedValue(undefined);

  it('renders stack name', () => {
    render(
      <StackManifestPreview
        stackName="my-stack"
        topology={mockTopology}
        manifests={mockManifests}
        onApprove={() => {}}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    expect(screen.getByText(/my-stack/)).toBeInTheDocument();
  });

  it('renders deploy order with arrows', () => {
    render(
      <StackManifestPreview
        stackName="my-stack"
        topology={mockTopology}
        manifests={mockManifests}
        onApprove={() => {}}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    expect(screen.getByText(/1\. db/)).toBeInTheDocument();
    expect(screen.getByText(/2\. backend/)).toBeInTheDocument();
  });

  it('renders service connections', () => {
    render(
      <StackManifestPreview
        stackName="my-stack"
        topology={mockTopology}
        manifests={mockManifests}
        onApprove={() => {}}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    // Connection shows "backend → db:5432" with env var
    expect(screen.getByText(/db:5432/)).toBeInTheDocument();
    expect(screen.getByText(/DATABASE_URL/)).toBeInTheDocument();
  });

  it('renders AI confidence badge', () => {
    render(
      <StackManifestPreview
        stackName="my-stack"
        topology={mockTopology}
        manifests={mockManifests}
        confidence={0.85}
        onApprove={() => {}}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    expect(screen.getByText(/85%/)).toBeInTheDocument();
  });

  it('renders reasoning when provided', () => {
    render(
      <StackManifestPreview
        stackName="my-stack"
        topology={mockTopology}
        manifests={mockManifests}
        reasoning="DB를 먼저 배포하고 backend가 연결하도록 구성했습니다."
        onApprove={() => {}}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    expect(screen.getByText(/DB를 먼저 배포하고/)).toBeInTheDocument();
  });

  it('renders manifest tabs for each service', () => {
    render(
      <StackManifestPreview
        stackName="my-stack"
        topology={mockTopology}
        manifests={mockManifests}
        onApprove={() => {}}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    // Tab buttons should exist
    const tabs = screen.getAllByRole('button');
    const tabLabels = tabs.map((t) => t.textContent);
    expect(tabLabels.some((l) => l?.includes('db'))).toBe(true);
    expect(tabLabels.some((l) => l?.includes('backend'))).toBe(true);
  });

  it('switches manifest content when tab clicked', () => {
    render(
      <StackManifestPreview
        stackName="my-stack"
        topology={mockTopology}
        manifests={mockManifests}
        onApprove={() => {}}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    // Default tab is first in deploy_order ('db') — manifest pre shows "name: db"
    const pres = document.querySelectorAll('pre');
    expect(pres[0].textContent).toContain('name: db');

    // Click backend tab (find tab button that contains "backend" + "api-server")
    const backendTab = screen.getAllByRole('button').find((b) => b.textContent?.includes('api-server'));
    if (backendTab) fireEvent.click(backendTab);

    const presAfter = document.querySelectorAll('pre');
    expect(presAfter[0].textContent).toContain('name: backend');
  });

  it('calls onApprove when approve button clicked', () => {
    const onApprove = vi.fn();
    render(
      <StackManifestPreview
        stackName="my-stack"
        topology={mockTopology}
        manifests={mockManifests}
        onApprove={onApprove}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    fireEvent.click(screen.getByText('Approve & Deploy Stack'));
    expect(onApprove).toHaveBeenCalledOnce();
  });

  it('calls onCancel when cancel button clicked', () => {
    const onCancel = vi.fn();
    render(
      <StackManifestPreview
        stackName="my-stack"
        topology={mockTopology}
        manifests={mockManifests}
        onApprove={() => {}}
        onCancel={onCancel}
        onRefine={mockOnRefine}
      />,
    );

    fireEvent.click(screen.getByText('취소'));
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it('shows feedback section when 수정 요청 button clicked', () => {
    render(
      <StackManifestPreview
        stackName="my-stack"
        topology={mockTopology}
        manifests={mockManifests}
        onApprove={() => {}}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    fireEvent.click(screen.getByText('수정 요청'));
    expect(screen.getByPlaceholderText(/db를 StatefulSet으로/)).toBeInTheDocument();
  });
});
