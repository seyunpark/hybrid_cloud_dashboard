import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { StackDeployProgress } from './StackDeployProgress';
import type { StackDeployStatus } from '@/api/types';

describe('StackDeployProgress', () => {
  const baseStatus: StackDeployStatus = {
    deploy_id: 'stack-1',
    status: 'deploying',
    stack_name: 'my-stack',
    deploy_order: ['db', 'backend', 'frontend'],
    services: {
      db: {
        service_name: 'db',
        status: 'completed',
        steps: [
          { step: 'create_deployment', status: 'completed', message: 'Deployment created' },
          { step: 'create_service', status: 'completed', message: 'Service created' },
        ],
      },
      backend: {
        service_name: 'backend',
        status: 'deploying',
        steps: [
          { step: 'create_deployment', status: 'in_progress', message: 'Creating...' },
          { step: 'create_service', status: 'pending' },
        ],
      },
      frontend: {
        service_name: 'frontend',
        status: 'pending',
        steps: [
          { step: 'create_deployment', status: 'pending' },
          { step: 'create_service', status: 'pending' },
        ],
      },
    },
  };

  it('renders stack name', () => {
    render(<StackDeployProgress status={baseStatus} />);
    expect(screen.getByText(/my-stack/)).toBeInTheDocument();
  });

  it('renders all services in deploy order', () => {
    render(<StackDeployProgress status={baseStatus} />);
    expect(screen.getByText('db')).toBeInTheDocument();
    expect(screen.getByText('backend')).toBeInTheDocument();
    expect(screen.getByText('frontend')).toBeInTheDocument();
  });

  it('shows order badges', () => {
    render(<StackDeployProgress status={baseStatus} />);
    expect(screen.getByText('1')).toBeInTheDocument();
    expect(screen.getByText('2')).toBeInTheDocument();
    expect(screen.getByText('3')).toBeInTheDocument();
  });

  it('shows per-service status badges', () => {
    render(<StackDeployProgress status={baseStatus} />);
    expect(screen.getByText('completed')).toBeInTheDocument();
    expect(screen.getByText('deploying')).toBeInTheDocument();
    expect(screen.getByText('pending')).toBeInTheDocument();
  });

  it('calculates overall progress percentage', () => {
    // 2 out of 6 steps completed = 33%
    render(<StackDeployProgress status={baseStatus} />);
    expect(screen.getByText('33%')).toBeInTheDocument();
  });

  it('shows 100% when all steps completed', () => {
    const completed: StackDeployStatus = {
      ...baseStatus,
      status: 'completed',
      services: {
        db: {
          service_name: 'db',
          status: 'completed',
          steps: [
            { step: 'create_deployment', status: 'completed' },
            { step: 'create_service', status: 'completed' },
          ],
        },
        backend: {
          service_name: 'backend',
          status: 'completed',
          steps: [
            { step: 'create_deployment', status: 'completed' },
            { step: 'create_service', status: 'completed' },
          ],
        },
        frontend: {
          service_name: 'frontend',
          status: 'completed',
          steps: [
            { step: 'create_deployment', status: 'completed' },
            { step: 'create_service', status: 'completed' },
          ],
        },
      },
    };

    render(<StackDeployProgress status={completed} />);
    expect(screen.getByText('100%')).toBeInTheDocument();
  });

  it('renders completed step checkmarks', () => {
    render(<StackDeployProgress status={baseStatus} />);
    const checkmarks = screen.getAllByText('✓');
    expect(checkmarks.length).toBe(2);
  });

  it('renders step messages when present', () => {
    render(<StackDeployProgress status={baseStatus} />);
    expect(screen.getByText(/Deployment created/)).toBeInTheDocument();
    expect(screen.getByText(/Creating\.\.\./)).toBeInTheDocument();
  });

  it('renders failed step icon', () => {
    const failed: StackDeployStatus = {
      ...baseStatus,
      services: {
        ...baseStatus.services,
        backend: {
          service_name: 'backend',
          status: 'failed',
          steps: [
            { step: 'create_deployment', status: 'failed', message: 'OOM error' },
            { step: 'create_service', status: 'pending' },
          ],
        },
      },
    };

    render(<StackDeployProgress status={failed} />);
    expect(screen.getByText('✗')).toBeInTheDocument();
    expect(screen.getByText(/OOM error/)).toBeInTheDocument();
  });
});
