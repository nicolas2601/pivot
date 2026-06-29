/**
 * formatMoney — convierte centavos (int64) a string con moneda formateada.
 * Locale-aware: usa el formato del país del usuario por defecto.
 */
export function formatMoney(cents: number, currency = 'COP', locale = 'es-CO'): string {
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    maximumFractionDigits: 0
  }).format(cents);
}

/**
 * formatCompactMoney — para stats en dashboard: usa notación abreviada
 * si el número es grande (ej. "$ 1.2M" en vez de "$ 1.234.567").
 */
export function formatCompactMoney(cents: number, currency = 'COP', locale = 'es-CO'): string {
  const value = cents;
  const abs = Math.abs(value);

  if (abs >= 1_000_000_000) {
    return new Intl.NumberFormat(locale, {
      notation: 'compact',
      style: 'currency',
      currency,
      maximumFractionDigits: 1
    }).format(value);
  }
  if (abs >= 100_000) {
    // For millions in COP-style amounts without decimals
    return new Intl.NumberFormat(locale, {
      notation: 'compact',
      style: 'currency',
      currency,
      maximumFractionDigits: 1
    }).format(value);
  }
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
    maximumFractionDigits: 0
  }).format(value);
}

/**
 * formatDate — ISO string → fecha legible en es-CO.
 */
export function formatDate(iso: string, opts: Intl.DateTimeFormatOptions = {}): string {
  return new Date(iso).toLocaleDateString('es-CO', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    ...opts
  });
}

/**
 * formatDateShort — solo día y mes, sin año (para listas de transacciones).
 */
export function formatDateShort(iso: string): string {
  return new Date(iso).toLocaleDateString('es-CO', {
    day: 'numeric',
    month: 'short'
  });
}

/**
 * monthLabel — devuelve el nombre del mes en es-CO para un Date.
 */
export function monthLabel(date: Date): string {
  return new Intl.DateTimeFormat('es-CO', {
    month: 'long',
    year: 'numeric'
  }).format(date);
}

/**
 * firstAndLastOfCurrentMonth — para queries de reports del mes actual.
 */
export function firstAndLastOfCurrentMonth(now = new Date()): {
  from: string;
  to: string;
} {
  const y = now.getFullYear();
  const m = now.getMonth();
  const last = new Date(y, m + 1, 0).getDate();
  const from = `${y}-${String(m + 1).padStart(2, '0')}-01`;
  const to = `${y}-${String(m + 1).padStart(2, '0')}-${String(last).padStart(2, '0')}`;
  return { from, to };
}

/**
 * pctDelta — calcula el cambio porcentual entre un valor actual y uno previo.
 * Devuelve null si el valor previo es 0 (no se puede calcular %).
 * Positivo = subió, negativo = bajó.
 */
export function pctDelta(current: number, previous: number): number | null {
  if (previous === 0) return null;
  return ((current - previous) / Math.abs(previous)) * 100;
}

/**
 * deltaDirection — clasifica un delta para el Stat component.
 * > 1% → up/down, dentro de ±1% → neutral.
 */
export type DeltaDirection = 'up' | 'down' | 'neutral';
export function deltaDirection(delta: number | null): DeltaDirection {
  if (delta === null || Math.abs(delta) < 1) return 'neutral';
  return delta > 0 ? 'up' : 'down';
}