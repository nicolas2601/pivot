/**
 * Motion primitives — Svelte 5 compatible.
 * Use these for entrance/exit/interaction transitions that respect
 * prefers-reduced-motion automatically.
 *
 * Import:  import { fade, fly, slide } from '$lib/motion/transitions';
 *
 * Each wrapper checks globalThis.matchMedia in the browser only;
 * degraded fallback is instant + opacity-0 (no layout shift).
 */
import { cubicOut } from 'svelte/easing';

function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined') return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

// --- fade -----------------------------------------------------------------

export interface FadeOptions {
  duration?: number;
  delay?: number;
}

export function fade(
  node: Element,
  { duration = 200, delay = 0 }: FadeOptions = {}
): { duration: number; delay: number; css?: (t: number) => string } {
  if (prefersReducedMotion()) return { duration: 0, delay: 0 };
  const o = Number(getComputedStyle(node).opacity);
  return {
    duration,
    delay,
    css: (t) => `opacity: ${t * o}`
  };
}

// --- fly (entrance + exit) -----------------------------------------------

export interface FlyOptions {
  x?: number;
  y?: number;
  duration?: number;
  delay?: number;
  easing?: typeof cubicOut;
}

export function fly(
  node: Element,
  { x = 0, y = 12, duration = 240, delay = 0, easing = cubicOut }: FlyOptions = {}
): {
  duration: number;
  delay: number;
  easing: typeof cubicOut;
  css: (t: number) => string;
} {
  if (prefersReducedMotion()) return { duration: 0, delay: 0, easing, css: () => '' };
  return {
    duration,
    delay,
    easing,
    css: (t) =>
      `transform: translate3d(${(1 - t) * x}px, ${(1 - t) * y}px, 0); opacity: ${t};`
  };
}

// --- scale (cards, modals) ----------------------------------------------

export interface ScaleOptions {
  start?: number;
  duration?: number;
  delay?: number;
  easing?: typeof cubicOut;
}

export function scale(
  node: Element,
  { start = 0.96, duration = 180, delay = 0, easing = cubicOut }: ScaleOptions = {}
): {
  duration: number;
  delay: number;
  easing: typeof cubicOut;
  css: (t: number) => string;
} {
  if (prefersReducedMotion()) return { duration: 0, delay: 0, easing, css: () => '' };
  const s = (t: number) =>
    `transform: scale(${start + (1 - start) * t}); opacity: ${t};`;
  return { duration, delay, easing, css: s };
}

// --- inViewport: animate first time element scrolls into view ----------

/**
 * Svelte action — adds `data-in-view=true` once the element enters the
 * viewport. CSS keys off `[data-in-view]` to trigger entrance animations.
 * Respects prefers-reduced-motion (immediately marks true).
 */
export function inViewport(
  node: HTMLElement,
  { rootMargin = '0px 0px -10% 0px', once = true }: { rootMargin?: string; once?: boolean } = {}
): { destroy?: () => void } {
  if (prefersReducedMotion()) {
    node.dataset.inView = 'true';
    return {};
  }
  if (typeof IntersectionObserver === 'undefined') {
    node.dataset.inView = 'true';
    return {};
  }
  const io = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        const target = entry.target as HTMLElement;
        if (entry.isIntersecting) {
          target.dataset.inView = 'true';
          if (once) io.unobserve(target);
        } else if (!once) {
          target.dataset.inView = 'false';
        }
      }
    },
    { rootMargin }
  );
  io.observe(node);
  return {
    destroy() {
      io.disconnect();
    }
  };
}

// --- stagger: apply --i CSS var to children so CSS can stagger them ---

/**
 * Action: assign each child an --i index for CSS-driven stagger animations.
 * Pair with `transition-delay: calc(var(--i) * 60ms)` in CSS.
 */
export function stagger(
  node: HTMLElement,
  { step = 60, max = 12 }: { step?: number; max?: number } = {}
): { destroy?: () => void } {
  const children = Array.from(node.children) as HTMLElement[];
  children.forEach((child, i) => {
    child.style.setProperty('--i', String(Math.min(i, max)));
    child.style.transitionDelay = `${Math.min(i, max) * step}ms`;
  });
  return {};
}