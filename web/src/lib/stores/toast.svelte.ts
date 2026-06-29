/**
 * Toast store — Svelte 5 runes, no external deps.
 * Usage:
 *   import { toast } from '$lib/stores/toast.svelte';
 *   toast.success('Guardado');
 *   toast.error('No se pudo guardar', 'Reintentá');
 */
export type ToastVariant = 'success' | 'error' | 'info' | 'warning';

export interface Toast {
  id: string;
  variant: ToastVariant;
  message: string;
  title?: string;
  ttl: number; // ms
}

interface ToastState {
  items: Toast[];
}

const DEFAULT_TTL = 4000;
let counter = 0;
const timers = new Map<string, ReturnType<typeof setTimeout>>();

const state = $state<ToastState>({ items: [] });

function create(
  variant: ToastVariant,
  message: string,
  title?: string,
  ttl: number = DEFAULT_TTL
): string {
  const id = `t${++counter}`;
  const toast: Toast = { id, variant, message, title, ttl };
  state.items.push(toast);
  const timer = setTimeout(() => dismiss(id), ttl);
  timers.set(id, timer);
  return id;
}

function dismiss(id: string): void {
  const idx = state.items.findIndex((t) => t.id === id);
  if (idx >= 0) state.items.splice(idx, 1);
  const timer = timers.get(id);
  if (timer) {
    clearTimeout(timer);
    timers.delete(id);
  }
}

function clear(): void {
  for (const t of timers.values()) clearTimeout(t);
  timers.clear();
  state.items.splice(0);
}

export const toast = {
  get items() {
    return state.items;
  },
  success: (message: string, title?: string, ttl?: number) => create('success', message, title, ttl),
  error: (message: string, title?: string, ttl?: number) => create('error', message, title, ttl ?? 6000),
  info: (message: string, title?: string, ttl?: number) => create('info', message, title, ttl),
  warning: (message: string, title?: string, ttl?: number) => create('warning', message, title, ttl ?? 5500),
  dismiss,
  clear
};