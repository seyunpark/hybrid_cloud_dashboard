import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ManifestPreview } from './ManifestPreview';
import type { Manifests, Recommendations } from '@/api/types';

describe('ManifestPreview', () => {
  const mockManifests: Manifests = {
    deployment: 'apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: test',
    service: 'apiVersion: v1\nkind: Service\nmetadata:\n  name: test',
  };

  const mockRecommendations: Recommendations = {
    cpu_request: '100m',
    cpu_limit: '500m',
    memory_request: '128Mi',
    memory_limit: '512Mi',
    replicas: 2,
    enable_hpa: false,
    reasoning: 'Based on similar deployments, these resources are recommended.',
  };

  const mockOnRefine = vi.fn().mockResolvedValue(undefined);

  it('renders manifests', () => {
    render(
      <ManifestPreview
        manifests={mockManifests}
        onApprove={() => {}}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    expect(screen.getByText('Deployment')).toBeInTheDocument();
    expect(screen.getByText('Service')).toBeInTheDocument();
  });

  it('renders recommendations when provided', () => {
    render(
      <ManifestPreview
        manifests={mockManifests}
        recommendations={mockRecommendations}
        onApprove={() => {}}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    expect(screen.getByText('AI Recommendations')).toBeInTheDocument();
    expect(screen.getByText(/100m/)).toBeInTheDocument();
    expect(screen.getByText(/512Mi/)).toBeInTheDocument();
  });

  it('shows reasoning text', () => {
    render(
      <ManifestPreview
        manifests={mockManifests}
        recommendations={mockRecommendations}
        onApprove={() => {}}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    expect(screen.getByText(/Based on similar deployments/)).toBeInTheDocument();
  });

  it('calls onApprove when approve button clicked', () => {
    const onApprove = vi.fn();
    render(
      <ManifestPreview
        manifests={mockManifests}
        onApprove={onApprove}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    fireEvent.click(screen.getByText('Approve & Deploy'));
    expect(onApprove).toHaveBeenCalledOnce();
  });

  it('calls onCancel when cancel button clicked', () => {
    const onCancel = vi.fn();
    render(
      <ManifestPreview
        manifests={mockManifests}
        onApprove={() => {}}
        onCancel={onCancel}
        onRefine={mockOnRefine}
      />,
    );

    fireEvent.click(screen.getByText('Cancel'));
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it('renders HPA when provided', () => {
    const manifests: Manifests = {
      ...mockManifests,
      hpa: 'apiVersion: autoscaling/v2\nkind: HorizontalPodAutoscaler',
    };

    render(
      <ManifestPreview
        manifests={manifests}
        onApprove={() => {}}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    expect(screen.getByText('HPA')).toBeInTheDocument();
  });

  it('does not render HPA section when not provided', () => {
    render(
      <ManifestPreview
        manifests={mockManifests}
        onApprove={() => {}}
        onCancel={() => {}}
        onRefine={mockOnRefine}
      />,
    );

    expect(screen.queryByText('HPA')).not.toBeInTheDocument();
  });
});
