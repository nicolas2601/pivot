<script lang="ts">
  /**
   * Dashboard — saldo total, gastos del mes, ingresos, balance neto,
   * distribución por categoría, últimas 5 transacciones.
   * Datos en vivo desde el backend vía @tanstack/svelte-query.
   */
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
  import { me, logout } from '$lib/api/auth';
  import { listAccounts } from '$lib/api/accounts';
  import { listTransactions } from '$lib/api/transactions';
  import { getSummary, getByCategory } from '$lib/api/reports';
  import { toast } from '$lib/stores/toast.svelte';
  import { inViewport } from '$lib/motion/transitions';
  import type { User } from '$lib/schemas/auth';
  import type { Transaction } from '$lib/schemas/transaction';
  import type { CategoryReportItem } from '$lib/schemas/report';
  import { getAccessToken, clearAccessToken } from '$lib/utils/auth-interceptor';
  import { authStore } from '$lib/stores/auth.svelte.ts';
  import {
    formatCompactMoney,
    formatMoney,
    formatDateShort,
    monthLabel,
    firstAndLastOfCurrentMonth,
    pctDelta,
    deltaDirection
  } from '$lib/utils/format';
  import Stat from '$lib/components/Stat.svelte';
  import Card from '$lib/components/Card.svelte';
  import Button from '$lib/components/Button.svelte';

  const qc = useQueryClient();

  let user = $state<User | null>(null);
  let userLoading = $state(true);

  const logoutMutation = createMutation(() => ({
    mutationFn: logout,
    onSuccess: () => {
      clearAccessToken();
      authStore.clearUser();
      qc.clear();
      toast.success('Sesión cerrada');
      goto('/auth/login');
    },
    onError: () => toast.warning('No se pudo cerrar la sesión en el servidor, pero la local sí.')
  }));

  onMount(async () => {
    const token = getAccessToken();
    if (!token) {
      goto('/auth/login');
      return;
    }
    try {
      user = await me();
      authStore.setUser(user);
    } catch {
      clearAccessToken();
      authStore.clearUser();
      goto('/auth/login');
      return;
    } finally {
      userLoading = false;
    }
  });

  const today = new Date();
  const { from: monthFrom, to: monthTo } = firstAndLastOfCurrentMonth(today);
  // Previous month range — for delta computation
  const prevMonthDate = new Date(today.getFullYear(), today.getMonth() - 1, 15);
  const prevFrom = firstAndLastOfCurrentMonth(prevMonthDate).from;
  const prevTo = firstAndLastOfCurrentMonth(prevMonthDate).to;

  // ─── Queries ───────────────────────────────────────────────────────────
  const accountsQuery = createQuery(() => ({
    queryKey: ['accounts'],
    queryFn: listAccounts
  }));

  const transactionsQuery = createQuery(() => ({
    queryKey: ['transactions', { limit: 5 }],
    queryFn: () => listTransactions({ limit: 5 })
  }));

  const summaryQuery = createQuery(() => ({
    queryKey: ['reports', 'summary', { from: monthFrom, to: monthTo }],
    queryFn: () => getSummary({ from: monthFrom, to: monthTo })
  }));

  const prevSummaryQuery = createQuery(() => ({
    queryKey: ['reports', 'summary', { from: prevFrom, to: prevTo }],
    queryFn: () => getSummary({ from: prevFrom, to: prevTo })
  }));

  const byCategoryQuery = createQuery(() => ({
    queryKey: ['reports', 'by-category', { from: monthFrom, to: monthTo }],
    queryFn: () => getByCategory({ from: monthFrom, to: monthTo })
  }));

  // ─── Derived state ────────────────────────────────────────────────────
  const hasAccounts = $derived((accountsQuery.data?.length ?? 0) > 0);
  const hasTransactions = $derived((transactionsQuery.data?.transactions.length ?? 0) > 0);
  const totalBalance = $derived(
    (accountsQuery.data ?? []).reduce((sum, a) => sum + (a.opening_balance ?? 0), 0)
  );

  const summary = $derived(summaryQuery.data);
  const prevSummary = $derived(prevSummaryQuery.data);
  const incomeDelta = $derived(pctDelta(summary?.total_income ?? 0, prevSummary?.total_income ?? 0));
  const expenseDelta = $derived(pctDelta(summary?.total_expense ?? 0, prevSummary?.total_expense ?? 0));

  // For the donut chart we need categories that actually have spending.
  // Fallback palette when the backend returns zero categories.
  const FALLBACK_PALETTE = [
    '#E8B4A0',
    '#A8C8E1',
    '#B8D4B8',
    '#C5B8D9',
    '#D9B8B8'
  ];

  const donutSlices = $derived.by(() => {
    const cats = byCategoryQuery.data?.categories ?? [];
    if (cats.length === 0) return [];
    return cats.map((c: CategoryReportItem, i: number) => ({
      name: c.name,
      percent: c.percent,
      amount: c.amount,
      color: c.color || FALLBACK_PALETTE[i % FALLBACK_PALETTE.length]
    }));
  });

  // ─── SVG helpers (donut chart) ────────────────────────────────────────
  function arcPath(cx: number, cy: number, r: number, startAngle: number, endAngle: number): string {
    const start = polar(cx, cy, r, endAngle);
    const end = polar(cx, cy, r, startAngle);
    const largeArc = endAngle - startAngle > 180 ? 1 : 0;
    return `M ${cx} ${cy} L ${start.x} ${start.y} A ${r} ${r} 0 ${largeArc} 0 ${end.x} ${end.y} Z`;
  }
  function polar(cx: number, cy: number, r: number, deg: number) {
    const rad = ((deg - 90) * Math.PI) / 180;
    return { x: cx + r * Math.cos(rad), y: cy + r * Math.sin(rad) };
  }
  const arcPaths = $derived.by(() => {
    let cumulative = 0;
    return donutSlices.map((slice) => {
      const startAngle = (cumulative / 100) * 360;
      cumulative += slice.percent;
      const endAngle = (cumulative / 100) * 360;
      return { ...slice, path: arcPath(100, 100, 80, startAngle, endAngle) };
    });
  });

  // ─── UI text ──────────────────────────────────────────────────────────
  const greeting = $derived(
    user?.display_name ? `Hola, ${user.display_name.split(' ')[0]}` : 'Tus finanzas'
  );
  const todayLabel = new Intl.DateTimeFormat('es-CO', {
    weekday: 'long',
    day: 'numeric',
    month: 'long'
  }).format(today);

  function txAccent(type: Transaction['type']): string {
    if (type === 'income') return 'text-ink';
    if (type === 'expense') return 'text-ink';
    return 'text-muted';
  }
  function txPrefix(type: Transaction['type']): string {
    if (type === 'income') return '+';
    if (type === 'expense') return '−';
    return '';
  }

  const loading = $derived(
    userLoading ||
      accountsQuery.isPending ||
      transactionsQuery.isPending ||
      summaryQuery.isPending
  );
</script>

<svelte:head><title>Dashboard — Mis finanzas</title></svelte:head>

<main class="bg-canvas min-h-screen">
  <div class="max-w-5xl mx-auto px-4 md:px-6 py-6 md:py-10 space-y-8">
    {#if loading}
      <p class="text-muted py-12 text-center">Cargando...</p>
    {:else}
      <!-- Encabezado -->
      <header class="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <p class="text-xs uppercase tracking-wider text-muted">{todayLabel}</p>
          <h1 class="font-waldenburg text-4xl md:text-5xl font-light text-ink mt-1">{greeting}</h1>
        </div>
        <div class="flex items-center gap-3">
          {#if user?.display_name || user?.email}
            <span class="text-sm text-body hidden sm:inline">{user?.display_name || user?.email}</span>
          {/if}
          <Button
            variant="outline"
            type="button"
            disabled={logoutMutation.isPending}
            onclick={() => logoutMutation.mutate()}
          >
            {logoutMutation.isPending ? 'Saliendo...' : 'Cerrar sesión'}
          </Button>
        </div>
      </header>

      <!-- Stats grid -->
      <section class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 md:gap-6 stagger" aria-label="Resumen financiero" use:inViewport>
        <Card>
          <Stat
            label="Saldo total"
            value={hasAccounts ? formatCompactMoney(totalBalance) : '—'}
            delta={hasAccounts
              ? `${accountsQuery.data?.length ?? 0} cuenta${(accountsQuery.data?.length ?? 0) === 1 ? '' : 's'}`
              : 'Sin cuentas todavía'}
            deltaDirection="neutral"
          />
        </Card>
        <Card>
          <Stat
            label="Gastos del mes"
            value={formatCompactMoney(summary?.total_expense ?? 0)}
            delta={expenseDelta === null
              ? 'vs mes anterior'
              : `${expenseDelta > 0 ? '+' : ''}${expenseDelta.toFixed(0)}% vs mes anterior`}
            deltaDirection={deltaDirection(expenseDelta) === 'up' ? 'worse' : deltaDirection(expenseDelta) === 'down' ? 'better' : 'neutral'}
          />
        </Card>
        <Card>
          <Stat
            label="Ingresos del mes"
            value={formatCompactMoney(summary?.total_income ?? 0)}
            delta={incomeDelta === null
              ? 'vs mes anterior'
              : `${incomeDelta > 0 ? '+' : ''}${incomeDelta.toFixed(0)}% vs mes anterior`}
            deltaDirection={deltaDirection(incomeDelta) === 'up' ? 'better' : deltaDirection(incomeDelta) === 'down' ? 'worse' : 'neutral'}
          />
        </Card>
        <Card>
          <Stat
            label="Balance neto"
            value={formatCompactMoney(summary?.net ?? 0)}
            delta={summary?.net === 0 ? 'Sin movimientos' : summary && summary.net > 0 ? 'Superávit' : 'Déficit'}
            deltaDirection={summary && summary.net > 0 ? 'better' : summary && summary.net < 0 ? 'worse' : 'neutral'}
          />
        </Card>
      </section>

      <!-- Gráfico de torta + Acciones rápidas -->
      <section class="grid grid-cols-1 lg:grid-cols-3 gap-4 md:gap-6 stagger" use:inViewport>
        <Card>
          <div class="space-y-4">
            <div class="flex items-baseline justify-between">
              <h2 class="font-waldenburg text-2xl font-light text-ink">Por categoría</h2>
              <span class="text-xs text-muted uppercase tracking-wider">{monthLabel(today)}</span>
            </div>
            {#if byCategoryQuery.isLoading}
              <p class="text-sm text-muted py-6 text-center">Cargando...</p>
            {:else if donutSlices.length > 0}
              <div class="flex flex-col sm:flex-row items-center gap-6">
                <svg viewBox="0 0 200 200" class="w-40 h-40" aria-label="Distribución de gastos por categoría">
                  {#each arcPaths as slice (slice.name)}
                    <path d={slice.path} fill={slice.color} />
                  {/each}
                  <circle cx="100" cy="100" r="44" fill="var(--color-surface-card)" />
                  <text x="100" y="96" text-anchor="middle" font-size="11" fill="var(--color-muted)" font-family="var(--font-inter)">
                    Total
                  </text>
                  <text x="100" y="112" text-anchor="middle" font-size="14" font-weight="500" fill="var(--color-ink)" font-family="var(--font-inter)">
                    {formatCompactMoney(summary?.total_expense ?? 0)}
                  </text>
                </svg>
                <ul class="flex-1 space-y-2 w-full">
                  {#each donutSlices as slice (slice.name)}
                    <li class="flex items-center justify-between gap-3 text-sm">
                      <span class="flex items-center gap-2 min-w-0">
                        <span class="w-2.5 h-2.5 rounded-full shrink-0" style="background: {slice.color}"></span>
                        <span class="text-ink truncate">{slice.name}</span>
                      </span>
                      <span class="text-muted tabular-nums shrink-0">{slice.percent.toFixed(0)}%</span>
                    </li>
                  {/each}
                </ul>
              </div>
            {:else}
              <div class="py-8 text-center space-y-2">
                <p class="text-ink">Sin gastos este mes</p>
                <p class="text-sm text-muted">Empezá a registrar movimientos para ver tu distribución por categoría.</p>
              </div>
            {/if}
          </div>
        </Card>

        <Card>
          <div class="space-y-4">
            <h2 class="font-waldenburg text-2xl font-light text-ink">Acciones rápidas</h2>
            <div class="space-y-2">
              <Button variant="primary" type="button" onclick={() => goto('/transactions/new')} class="w-full">
                Nuevo movimiento
              </Button>
              <Button variant="outline" type="button" onclick={() => goto('/accounts/new')} class="w-full">
                Nueva cuenta
              </Button>
              <Button variant="outline" type="button" onclick={() => goto('/budgets/new')} class="w-full">
                Nuevo presupuesto
              </Button>
              <Button variant="tertiary" type="button" onclick={() => goto('/travel/new')} class="w-full">
                Nuevo viaje
              </Button>
              <Button variant="tertiary" type="button" onclick={() => goto('/goals/new')} class="w-full">
                Nueva meta
              </Button>
            </div>
          </div>
        </Card>
      </section>

      <!-- Últimas transactions -->
      <section class="space-y-3 fade-on-view" use:inViewport>
        <div class="flex items-baseline justify-between">
          <h2 class="font-waldenburg text-2xl font-light text-ink">Últimos movimientos</h2>
          <a href="/transactions" class="text-sm text-ink hover:underline">Ver todos</a>
        </div>
        {#if hasTransactions && transactionsQuery.data}
          <Card>
            <ul class="divide-y divide-hairline">
              {#each transactionsQuery.data.transactions as tx (tx.id)}
                <li>
                  <a
                    href="/transactions/{tx.id}"
                    class="flex items-center justify-between gap-4 py-3 -mx-2 px-2 rounded hover:bg-surface-strong transition-colors"
                  >
                    <div class="min-w-0">
                      <p class="text-sm text-ink truncate">{tx.description || tx.type}</p>
                      <p class="text-xs text-muted">{formatDateShort(tx.date)}</p>
                    </div>
                    <span class="text-sm tabular-nums shrink-0 {txAccent(tx.type)}">
                      {txPrefix(tx.type)}{formatMoney(tx.amount, tx.currency)}
                    </span>
                  </a>
                </li>
              {/each}
            </ul>
          </Card>
        {:else}
          <Card>
            <div class="py-8 text-center space-y-2">
              <p class="text-ink">Sin movimientos todavía</p>
              <p class="text-sm text-muted">Creá tu primera cuenta y registrá un movimiento para empezar.</p>
              <div class="pt-3 flex justify-center gap-2 flex-wrap">
                <Button variant="outline" type="button" onclick={() => goto('/accounts/new')}>
                  Crear cuenta
                </Button>
                <Button variant="primary" type="button" onclick={() => goto('/transactions/new')}>
                  Registrar movimiento
                </Button>
              </div>
            </div>
          </Card>
        {/if}
      </section>

      <!-- Metas activas — bonus -->
      <section>
        <div class="flex items-baseline justify-between mb-3">
          <h2 class="font-waldenburg text-2xl font-light text-ink">Metas activas</h2>
          <a href="/goals" class="text-sm text-ink hover:underline">Ver todas</a>
        </div>
        <Card>
          <p class="text-sm text-muted text-center py-4">
            <a href="/goals" class="text-ink hover:underline">Tus metas de ahorro</a> aparecen acá cuando las creás.
          </p>
        </Card>
      </section>
    {/if}
  </div>
</main>