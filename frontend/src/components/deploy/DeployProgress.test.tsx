import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { DeployProgress } from './DeployProgress';
import type { DeployStep } from '@/api/types';

describe('DeployProgress', () => {
  it('renders all steps', () => {
    const steps: DeployStep[] = [
      { step: 'push_image', status: 'completed', message: 'Done' },
      { step: 'create_deployment', status: 'in_progress', message: 'Creating...' },
      { step: 'create_service', status: 'pending' },
    ];

    render(<DeployProgress steps={steps} status="deploying" />);

    expect(screen.getByText('Push Image to Registry')).toBeInTheDocument();
    expect(screen.getByText('Create Deployment')).toBeInTheDocument();
    expect(screen.getByText('Create Service')).toBeInTheDocument();
  });

  it('shows step messages', () => {
    const steps: DeployStep[] = [
      { step: 'push_image', status: 'completed', message: 'Image pushed successfully' },
    ];

    render(<DeployProgress steps={steps} status="deploying" />);
    expect(screen.getByText('Image pushed successfully')).toBeInTheDocument();
  });

  it('shows status label', () => {
    render(<DeployProgress steps={[]} status="completed" />);
    expect(screen.getByText('completed')).toBeInTheDocument();
  });

  it('renders completed step icon', () => {
    const steps: DeployStep[] = [
      { step: 'push_image', status: 'completed' },
    ];

    render(<DeployProgress steps={steps} status="deploying" />);
    expect(screen.getByText('✓')).toBeInTheDocument();
  });

  it('renders failed step icon', () => {
    const steps: DeployStep[] = [
      { step: 'push_image', status: 'failed', message: 'Push failed' },
    ];

    render(<DeployProgress steps={steps} status="failed" />);
    expect(screen.getByText('✕')).toBeInTheDocument();
  });

  it('handles unknown step names gracefully', () => {
    const steps: DeployStep[] = [
      { step: 'custom_step', status: 'pending' },
    ];

    render(<DeployProgress steps={steps} status="deploying" />);
    expect(screen.getByText('custom_step')).toBeInTheDocument();
  });
});
