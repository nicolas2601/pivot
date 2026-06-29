<script lang="ts">
  /**
   * /reports — visualizaciones: por categoría, por cuenta, tendencia mensual,
   * cashflow. Período configurable (1/3/6/12 meses).
   */
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { createQuery } from '@tanstack/svelte-query';
  import { getByCategory, getByAccount, getMonthlyTrend, getCashflow } from '$lib/api/reports';
  import { getAccessToken } from '$lib/utils/auth-interceptor';
  import { formatCompactMoney, formatMoney } from '$lib/utils/format';
  import Stat from '$lib/components/Stat.svelte';
  import Card from '$lib/components/Card.svelte';
  import BarChart from '$lib/components/BarChart.svelte';

  type Period = 1 | 3 | 6 | 12;

  let period = $state<Period>(1);

  onMount(() => {
    if (!getAccessToken()) goto('/auth/login');
  });

  // Period → date range
  const dateRange = $derived.by(() => {
    const now = new Date();
    const y = now.getFullYear();
    const m = now.getMonth();
    const last = new Date(y, m + 1, 0).getDate();
    const to = `${y}-${String(m + 1).padStart(2, '0')}-${String(last).padStart(2, '0')}`;
    const fromMonth = new Date(y, m - (period - 1), 1);
    const fromY = fromMonth.getFullYear();
    const fromM = fromMonth.getMonth() + 1;
    const from = `${fromY}-${String(fromM).padStart(2, '0')}-01`;
    return { from, to };
  });

  const PALETTE = [
    '#E8B4A0', '#A8C8E1', '#B8D4B8', '#C5B8D9', '#D9B8B8',
    '#E0C9A6', '#A6C9C0', '#C9A6C9', '#B8C9A6', '#C9C2A6'
  ];

  const byCategoryQuery = createQuery(() => ({
    queryKey: ['reports', 'by-category', dateRange.from, dateRange.to],
    queryFn: () => getByCategory(dateRange)
  }));

  const byAccountQuery = createQuery(() => ({
    queryKey: ['reports', 'by-account', dateRange.from, dateRange.to],
    queryFn: () => getByAccount(dateRange)
  }));

  const trendQuery = createQuery(() => ({
    queryKey: ['reports', 'monthly-trend', dateRange.from, dateRange.to],
    queryFn: () => getMonthlyTrend(dateRange)
  }));

  const cashflowQuery = createQuery(() => ({
    queryKey: ['reports', 'cashflow', dateRange.from, dateRange.to],
    queryFn: () => getCashflow(dateRange)
  }));

  const monthLabel = (y: number, m: number): string =>
    new Date(y, m - 1, 1).toLocaleDateString('es-CO', { month: 'short', year: '2-digit' });

  // For monthly trend bars — normalize to positive/negative for visual.
  const trendBars = $derived.by(() => {
    const items = trendQuery.data?.months ?? [];
    return items.map((it) => {
      const expense = it.expense;
      return {
        label: monthLabel(it.year, it.month),
        value: expense,
        sublabel: `${formatCompactMoney(it.income)} / ${formatCompactMoney(expense)}`,
        color: expense > it.income ? '#D9B8B8' : '#B8D4B8'
      };
    });
  });

  const loading = $derived(
    byCategoryQuery.isPending ||
      byAccountQuery.isPending ||
      trendQuery.isPending ||
      cashflowQuery.isPending
  );
</script>

<svelte:head><title>Reportes — Mis finanzas</title></svelte:head>

<main class="bg-canvas min-h-screen pb-24 md:pb-10">
  <div class="max-w-4xl mx-auto px-4 md:px-6 py-6 md:py-10 space-y-6">
    <header class="flex items-start justify-between gap-4 flex-wrap">
      <div>
        <p class="text-xs uppercase tracking-wider text-muted">Análisis</p>
        <h1 class="font-waldenburg text-4xl md:text-5xl font-light text-ink mt-1">Reportes</h1>
        <p class="text-sm text-muted mt-2">Tu actividad financiera en cifras.</p>
      </div>
      <div class="flex gap-1 p-1 bg-surface-card rounded-md border border-hairline" role="group" aria-label="Período">
        {#each [1, 3, 6, 12] as p (p)}
          <button
            type="button"
            onclick={() => (period = p as Period)}
            class="px-3 py-1.5 text-sm rounded transition-colors {period === p
              ? 'bg-ink text-on-primary'
              : 'text-muted hover:text-ink'}"
          >
            {p === 1 ? 'Este mes' : `${p} meses`}
          </button>
        {/each}
      </div>
    </header>

    {#if loading}
      <p class="text-muted text-center py-12">Cargando...</p>
    {:else}
      <!-- Cashflow headline -->
      {#if cashflowQuery.data}
        <section class="grid grid-cols-1 sm:grid-cols-3 gap-4 md:gap-6" aria-label="Cashflow del período">
          <Card>
            <Stat
              label="Ingresos"
              value={formatCompactMoney(cashflowQuery.data.income)}
              deltaDirection="neutral"
              delta=""
            />
          </Card>
          <Card>
            <Stat
              label="Gastos"
              value={formatCompactMoney(cashflowQuery.data.expense)}
              deltaDirection="neutral"
              delta=""
            />
          </Card>
          <Card>
            <Stat
              label="Tasa de ahorro"
              value={`${cashflowQuery.data.savings_rate.toFixed(1)}%`}
              delta={formatCompactMoney(cashflowQuery.data.savings_total)}
              deltaDirection={cashflowQuery.data.savings_rate >= 20 ? 'better' : cashflowQuery.data.savings_rate < 0 ? 'worse' : 'neutral'}
            />
          </Card>
        </section>
      {/if}

      <!-- By category -->
      <section class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6">
        <Card>
          <div class="space-y-4">
            <h2 class="font-waldenburg text-2xl font-light text-ink">Por categoría</h2>
            {#if byCategoryQuery.data?.categories.length}
              <BarChart
                bars={byCategoryQuery.data.categories.map((c, i) => ({
                  label: c.name,
                  value: c.amount,
                  sublabel: `${formatMoney(c.amount)} (${c.count} ${c.count === 1 ? 'movimiento' : 'movimientos'})`,
                  color: c.color || PALETTE[i % PALETTE.length]
                }))}
              />
            {:else}
              <p class="text-sm text-muted py-6 text-center">Sin gastos en este período.</p>
            {/if}
          </div>
        </Card>

        <!-- By account -->
        <Card>
          <div class="space-y-4">
            <h2 class="font-waldenburg text-2xl font-light text-ink">Por cuenta</h2>
            {#if byAccountQuery.data?.accounts.length}
              <BarChart
                bars={byAccountQuery.data.accounts.map((a) => ({
                  label: a.name,
                  value: a.expense,
                  sublabel: `Saldo: ${formatMoney(a.balance)} · Gastos: ${formatMoney(a.expense)}`,
                  color: PALETTE[a.name.length % PALETTE.length]
                }))}
              />
            {:else}
              <p class="text-sm text-muted py-6 text-center">Sin cuentas con gastos.</p>
            {/if}
          </div>
        </Card>
      </section>

      <!-- Monthly trend -->
      <section>
        <Card>
          <div class="space-y-4">
            <h2 class="font-waldenburg text-2xl font-light text-ink">Tendencia mensual</h2>
            {#if trendBars.length > 0}
              <BarChart
                bars={trendBars}
                height={28}
              />
              <p class="text-xs text-muted">
                Color: rojo cuando gastos superan ingresos, verde cuando es al revés.
              </p>
            {:else}
              <p class="text-sm text-muted py-6 text-center">Sin datos para graficar.</p>
            {/if}
          </div>
        </Card>
      </section>
    {/if}
  </div>
</main>