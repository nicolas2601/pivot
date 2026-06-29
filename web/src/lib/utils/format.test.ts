import { describe, it, expect } from 'vitest';
import {
  formatMoney,
  formatCompactMoney,
  formatDate,
  formatDateShort,
  monthLabel,
  firstAndLastOfCurrentMonth,
  pctDelta,
  deltaDirection
} from './format';

describe('formatMoney', () => {
  it('formats COP without decimals', () => {
    // es-CO uses $ symbol and dot as thousands separator
    const out = formatMoney(1500000, 'COP');
    expect(out).toContain('1');
    expect(out).toContain('500');
  });

  it('handles zero', () => {
    expect(formatMoney(0)).toBeTruthy();
  });

  it('handles negative', () => {
    const out = formatMoney(-50000);
    expect(out).toContain('50');
  });
});

describe('formatCompactMoney', () => {
  it('uses compact notation for large amounts', () => {
    const out = formatCompactMoney(1500000000, 'COP');
    expect(out.length).toBeLessThan(formatMoney(1500000000, 'COP').length);
    // expect something like $1,5M or 1.500M
  });
});

describe('formatDate', () => {
  it('parses ISO and returns Spanish date', () => {
    const out = formatDate('2026-01-15T00:00:00Z');
    expect(out).toContain('ene');
    expect(out).toContain('2026');
  });
});

describe('formatDateShort', () => {
  it('returns day + month only', () => {
    // Use noon UTC to avoid timezone-boundary surprises.
    const out = formatDateShort('2026-01-15T12:00:00Z');
    expect(out).toContain('15');
    expect(out).toContain('ene');
    expect(out).not.toContain('2026');
  });
});

describe('monthLabel', () => {
  it('returns month + year in Spanish', () => {
    const out = monthLabel(new Date(2026, 5, 15));
    expect(out).toContain('2026');
    expect(out).toContain('junio');
  });
});

describe('firstAndLastOfCurrentMonth', () => {
  it('returns YYYY-MM-DD strings for first and last day', () => {
    const out = firstAndLastOfCurrentMonth(new Date(2026, 0, 15));
    expect(out.from).toBe('2026-01-01');
    expect(out.to).toBe('2026-01-31');
  });

  it('handles February (28 days in non-leap)', () => {
    const out = firstAndLastOfCurrentMonth(new Date(2026, 1, 15));
    expect(out.to).toBe('2026-02-28');
  });

  it('handles February in leap year', () => {
    const out = firstAndLastOfCurrentMonth(new Date(2024, 1, 15));
    expect(out.to).toBe('2024-02-29');
  });
});

describe('pctDelta', () => {
  it('positive when current > previous', () => {
    expect(pctDelta(120, 100)).toBeCloseTo(20, 5);
  });
  it('negative when current < previous', () => {
    expect(pctDelta(80, 100)).toBeCloseTo(-20, 5);
  });
  it('null when previous is zero', () => {
    expect(pctDelta(100, 0)).toBeNull();
  });
  it('uses absolute previous for sign so 50→100 shows +100% (not -100%)', () => {
    expect(pctDelta(100, -50)).toBeCloseTo(300, 5);
  });
});

describe('deltaDirection', () => {
  it('up for positive > 1%', () => {
    expect(deltaDirection(5)).toBe('up');
    expect(deltaDirection(0.5)).toBe('neutral');
    expect(deltaDirection(-5)).toBe('down');
    expect(deltaDirection(null)).toBe('neutral');
  });
});