<script lang="ts">
  /**
   * ProgressBar — barra horizontal para presupuestos y goals.
   * Estilo: track hairline, fill primary (ink), radius pill.
   * Accesibilidad: role="progressbar" con aria-valuenow/min/max.
   */

  interface Props {
    value: number;
    max: number;
    label?: string;
    showValues?: boolean;
    unit?: string;
  }

  let { value, max, label, showValues = true, unit = '' }: Props = $props();

  const percent = $derived(max > 0 ? Math.min(100, Math.round((value / max) * 100)) : 0);
  const isOver = $derived(value > max);
</script>

<div class="space-y-2">
  {#if label}
    <div class="flex justify-between items-baseline">
      <span class="text-sm text-ink">{label}</span>
      {#if showValues}
        <span class="text-xs text-muted tabular-nums">
          {unit}{value.toLocaleString('es-CO')} / {unit}{max.toLocaleString('es-CO')}
        </span>
      {/if}
    </div>
  {/if}
  <div
    role="progressbar"
    aria-valuenow={value}
    aria-valuemin={0}
    aria-valuemax={max}
    aria-label={label}
    class="w-full h-2 bg-hairline rounded-pill overflow-hidden"
  >
    <div
      class="h-full rounded-pill transition-all duration-300 {isOver ? 'bg-semantic-error' : 'bg-ink'}"
      style="width: {percent}%"
    ></div>
  </div>
  <div class="flex justify-end">
    <span class="text-xs {isOver ? 'text-semantic-error' : 'text-muted'} tabular-nums">
      {percent}%
    </span>
  </div>
</div>