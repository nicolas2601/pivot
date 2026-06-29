<script lang="ts">
  /**
   * BarChart — minimal SVG bar chart, no deps.
   * Used by /reports to show category/account breakdowns and monthly trend.
   * Horizontal layout for labels (works great with Spanish category names).
   */
  interface Bar {
    label: string;
    value: number;
    color?: string;
    sublabel?: string;
  }
  let {
    bars,
    height = 24,
    gap = 8,
    showValues = true
  }: {
    bars: Bar[];
    height?: number;
    gap?: number;
    showValues?: boolean;
  } = $props();

  const max = $derived(Math.max(1, ...bars.map((b) => b.value)));
  const total = $derived(bars.reduce((s, b) => s + b.value, 0));
  const defaultColor = 'var(--color-ink)';
</script>

<div class="space-y-2" role="list">
  {#each bars as bar (bar.label)}
    <div role="listitem" class="space-y-1">
      <div class="flex items-baseline justify-between gap-2 text-sm">
        <span class="text-ink truncate">{bar.label}</span>
        {#if showValues}
          <span class="text-muted tabular-nums shrink-0">
            {bar.sublabel ?? bar.value.toLocaleString('es-CO')}
            {#if total > 0 && bar.sublabel === undefined}
              <span class="text-muted/70 ml-1">({((bar.value / total) * 100).toFixed(0)}%)</span>
            {/if}
          </span>
        {/if}
      </div>
      <div
        class="rounded-full bg-surface-strong overflow-hidden"
        style="height: {height}px"
      >
        <div
          class="h-full rounded-full transition-all duration-300"
          style="width: {(bar.value / max) * 100}%; background: {bar.color ?? defaultColor};"
          aria-label="{bar.label}: {bar.value}"
        ></div>
      </div>
    </div>
  {/each}
</div>