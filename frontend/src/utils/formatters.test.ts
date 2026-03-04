import { describe, it, expect } from 'vitest';
import {
  formatBytes,
  formatCpuPercent,
  formatDate,
  formatDateTime,
  formatRelativeTime,
  formatNetworkRate,
} from './formatters';

describe('formatBytes', () => {
  it('should format 0 bytes', () => {
    expect(formatBytes(0)).toBe('0 B');
  });

  it('should format bytes', () => {
    expect(formatBytes(512)).toBe('512 B');
  });

  it('should format kilobytes', () => {
    expect(formatBytes(1024)).toBe('1 KB');
  });

  it('should format megabytes', () => {
    expect(formatBytes(1024 * 1024)).toBe('1 MB');
  });

  it('should format gigabytes', () => {
    expect(formatBytes(1024 * 1024 * 1024)).toBe('1 GB');
  });

  it('should handle custom decimal places', () => {
    expect(formatBytes(1536, 2)).toBe('1.5 KB');
  });

  it('should handle large numbers', () => {
    expect(formatBytes(1024 * 1024 * 1024 * 1024)).toBe('1 TB');
  });
});

describe('formatCpuPercent', () => {
  it('should format CPU percentage', () => {
    expect(formatCpuPercent(5.234)).toBe('5.2%');
  });

  it('should format 0', () => {
    expect(formatCpuPercent(0)).toBe('0.0%');
  });

  it('should format 100', () => {
    expect(formatCpuPercent(100)).toBe('100.0%');
  });

  it('should handle high precision', () => {
    expect(formatCpuPercent(12.3456789)).toBe('12.3%');
  });
});

describe('formatDate', () => {
  it('should format an ISO date string', () => {
    const result = formatDate('2026-02-26T10:30:00Z');
    expect(result).toBeTruthy();
    expect(typeof result).toBe('string');
  });
});

describe('formatDateTime', () => {
  it('should format an ISO datetime string', () => {
    const result = formatDateTime('2026-02-26T10:30:00Z');
    expect(result).toBeTruthy();
    expect(typeof result).toBe('string');
  });
});

describe('formatRelativeTime', () => {
  it('should format seconds ago', () => {
    const now = new Date();
    const fiveSecsAgo = new Date(now.getTime() - 5000).toISOString();
    const result = formatRelativeTime(fiveSecsAgo);
    expect(result).toMatch(/\d+s ago/);
  });

  it('should format minutes ago', () => {
    const now = new Date();
    const fiveMinsAgo = new Date(now.getTime() - 5 * 60 * 1000).toISOString();
    const result = formatRelativeTime(fiveMinsAgo);
    expect(result).toMatch(/\d+m ago/);
  });

  it('should format hours ago', () => {
    const now = new Date();
    const twoHoursAgo = new Date(now.getTime() - 2 * 60 * 60 * 1000).toISOString();
    const result = formatRelativeTime(twoHoursAgo);
    expect(result).toMatch(/\d+h ago/);
  });

  it('should format days ago', () => {
    const now = new Date();
    const fiveDaysAgo = new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000).toISOString();
    const result = formatRelativeTime(fiveDaysAgo);
    expect(result).toMatch(/\d+d ago/);
  });
});

describe('formatNetworkRate', () => {
  it('should format bytes per second', () => {
    expect(formatNetworkRate(500)).toBe('500 B/s');
  });

  it('should format kilobytes per second', () => {
    expect(formatNetworkRate(1500)).toBe('1.5 KB/s');
  });

  it('should format megabytes per second', () => {
    expect(formatNetworkRate(1500000)).toBe('1.4 MB/s');
  });

  it('should format 0', () => {
    expect(formatNetworkRate(0)).toBe('0 B/s');
  });
});
