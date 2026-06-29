import { describe, it, expect, beforeEach, vi } from 'vitest';
import { toast } from './toast.svelte';

describe('toast store', () => {
  beforeEach(() => {
    toast.clear();
    vi.useFakeTimers();
  });

  it('adds success toast', () => {
    toast.success('Saved');
    expect(toast.items.length).toBe(1);
    expect(toast.items[0].variant).toBe('success');
    expect(toast.items[0].message).toBe('Saved');
  });

  it('adds error with title', () => {
    toast.error('Could not save', 'Network');
    expect(toast.items.length).toBe(1);
    expect(toast.items[0].variant).toBe('error');
    expect(toast.items[0].title).toBe('Network');
  });

  it('dismisses a toast', () => {
    const id = toast.success('Hi');
    expect(toast.items.length).toBe(1);
    toast.dismiss(id);
    expect(toast.items.length).toBe(0);
  });

  it('clears all', () => {
    toast.error('a');
    toast.info('b');
    toast.warning('c');
    expect(toast.items.length).toBe(3);
    toast.clear();
    expect(toast.items.length).toBe(0);
  });

  it('auto-dismisses after ttl', () => {
    toast.success('short', undefined, 1000);
    expect(toast.items.length).toBe(1);
    vi.advanceTimersByTime(1100);
    expect(toast.items.length).toBe(0);
  });

  it('error uses longer default ttl', () => {
    toast.error('persist');
    expect(toast.items[0].ttl).toBe(6000);
    expect(toast.items.length).toBe(1);
    vi.advanceTimersByTime(5999);
    expect(toast.items.length).toBe(1);
    vi.advanceTimersByTime(2);
    expect(toast.items.length).toBe(0);
  });
});