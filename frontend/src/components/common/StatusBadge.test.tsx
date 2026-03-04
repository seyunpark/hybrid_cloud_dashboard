import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { StatusBadge } from './StatusBadge';

describe('StatusBadge', () => {
  it('renders running status with green color', () => {
    render(<StatusBadge status="running" />);
    const badge = screen.getByText('running');
    expect(badge).toBeInTheDocument();
    expect(badge.className).toContain('green');
  });

  it('renders exited status with red color', () => {
    render(<StatusBadge status="exited" />);
    const badge = screen.getByText('exited');
    expect(badge).toBeInTheDocument();
    expect(badge.className).toContain('red');
  });

  it('renders connected status with green color', () => {
    render(<StatusBadge status="connected" />);
    const badge = screen.getByText('connected');
    expect(badge).toBeInTheDocument();
    expect(badge.className).toContain('green');
  });

  it('renders disconnected status with red color', () => {
    render(<StatusBadge status="disconnected" />);
    const badge = screen.getByText('disconnected');
    expect(badge).toBeInTheDocument();
    expect(badge.className).toContain('red');
  });

  it('renders pending status with yellow color', () => {
    render(<StatusBadge status="pending" />);
    const badge = screen.getByText('pending');
    expect(badge).toBeInTheDocument();
    expect(badge.className).toContain('yellow');
  });

  it('renders deploying status with blue color', () => {
    render(<StatusBadge status="deploying" />);
    const badge = screen.getByText('deploying');
    expect(badge).toBeInTheDocument();
    expect(badge.className).toContain('blue');
  });

  it('renders unknown status with default gray color', () => {
    render(<StatusBadge status="unknown-status" />);
    const badge = screen.getByText('unknown-status');
    expect(badge).toBeInTheDocument();
    expect(badge.className).toContain('gray');
  });
});
