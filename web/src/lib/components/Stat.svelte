<script lang="ts">
  /**
   * Stat — display numérico para el dashboard.
   * Jerarquía: número grande (font-waldenburg 300), label caption-uppercase, delta opcional.
   * El delta se muestra con color semántico solo si se pasa `direction: better | worse | neutral`.
   */

  interface Props {
    label: string;
    value: string;
    delta?: string;
    deltaDirection?: 'better' | 'worse' | 'neutral';
  }

  let { label, value, delta, deltaDirection = 'neutral' }: Props = $props();

  const deltaColor = $derived(
    {
      better: 'text-semantic-success',
      worse: 'text-semantic-error',
      neutral: 'text-muted'
    }[deltaDirection]
  );
</script>

<div class="space-y-2">
  <p class="text-xs uppercase tracking-wider text-muted">{label}</p>
  <p class="font-waldenburg text-4xl font-light text-ink tabular-nums leading-none">{value}</p>
  {#if delta}
    <p class="text-xs {deltaColor} tabular-nums">{delta}</p>
  {/if}
</div>