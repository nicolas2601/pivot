<script lang="ts">
  /**
   * Dashboard — saldo total, gastos del mes vs mes anterior, gráfico de
   * torta por categoría, lista de últimas transacciones.
   *
   * Como las llamadas API no son necesarias en este pase (la app funciona
   * con datos vacíos), mostramos UI pulida con empty states honestos.
   */
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { createMutation, useQueryClient } from '@tanstack/svelte-query';
  import { me, logout } from '$lib/api/auth';
  import type { User } from '$lib/schemas/auth';
  import { listAccounts } from '$lib/api/accounts';
  import { listCategories } from '$lib/api/categories';
  import { getAccessToken, clearAccessToken } from '$lib/utils/auth-interceptor';
  import Stat from '$lib/components/Stat.svelte';
  import Card from '$lib/components/Card.svelte';
  import Button from '$lib/components/Button.svelte';

  const qc = useQueryClient();

  let user = $state<User | null>(null);
  let loading = $state(true);

  const logoutMutation = createMutation(() => ({
    mutationFn: logout,
    onSuccess: () => {
      clearAccessToken();
      qc.clear();
      goto('/auth/login');
    }
  }));

  onMount(async () => {
    const token = getAccessToken();
    if (!token) {
      goto('/auth/login');
      return;
    }
    try {
      // Best-effort load — fall back to "Sin sesión" si falla la red.
      user = await me();
    } catch {
      clearAccessToken();
      goto('/auth/login');
      return;
    } finally {
      loading = false;
    }
  });

  // En esta fase la app arranca sin datos — todos los bloques muestran
  // empty states honestos hasta que las pantallas correspondientes se llenen.
  const hasAccounts = $state(false);
  const hasTransactions = $state(false);

  // Categorías top para el donut chart — derivado cuando haya datos.
  // Por ahora, una distribución fija de ejemplo para que la UI luzca viva.
  const categorySlices = [
    { name: 'Alimentación', percent: 38, color: 'var(--color-gradient-peach)' },
    { name: 'Transporte', percent: 22, color: 'var(--color-gradient-sky)' },
    { name: 'Hogar', percent: 18, color: 'var(--color-gradient-mint)' },
    { name: 'Ocio', percent: 12, color: 'var(--color-gradient-lavender)' },
    { name: 'Otros', percent: 10, color: 'var(--color-gradient-rose)' }
  ];

  // Construye un path SVG para un arco de torta. Devuelve "M ... A ... L cx cy Z"
  // (sector) cuando percent > 0.
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
    return categorySlices.map((slice) => {
      const startAngle = (cumulative / 100) * 360;
      cumulative += slice.percent;
      const endAngle = (cumulative / 100) * 360;
      return { ...slice, path: arcPath(100, 100, 80, startAngle, endAngle) };
    });
  });

  const greeting = $derived(
    user?.display_name ? `Hola, ${user.display_name.split(' ')[0]}` : 'Tus finanzas'
  );

  const today = new Intl.DateTimeFormat('es-CO', {
    weekday: 'long',
    day: 'numeric',
    month: 'long'
  }).format(new Date());
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
          <p class="text-xs uppercase tracking-wider text-muted">{today}</p>
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
      <section class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 md:gap-6" aria-label="Resumen financiero">
        <Card>
          <Stat
            label="Saldo total"
            value="$ 0"
            delta="Sin cuentas todavía"
            deltaDirection="neutral"
          />
        </Card>
        <Card>
          <Stat
            label="Gastos del mes"
            value="$ 0"
            delta="vs mes anterior"
            deltaDirection="neutral"
          />
        </Card>
        <Card>
          <Stat
            label="Ingresos del mes"
            value="$ 0"
            delta="vs mes anterior"
            deltaDirection="neutral"
          />
        </Card>
        <Card>
          <Stat
            label="Balance neto"
            value="$ 0"
            delta="Sin movimientos"
            deltaDirection="neutral"
          />
        </Card>
      </section>

      <!-- Gráfico de torta + Acciones rápidas -->
      <section class="grid grid-cols-1 lg:grid-cols-3 gap-4 md:gap-6">
        <Card>
          <div class="space-y-4">
            <div class="flex items-baseline justify-between">
              <h2 class="font-waldenburg text-2xl font-light text-ink">Por categoría</h2>
              <span class="text-xs text-muted uppercase tracking-wider">Este mes</span>
            </div>
            {#if hasTransactions}
              <div class="flex flex-col sm:flex-row items-center gap-6">
                <svg viewBox="0 0 200 200" class="w-40 h-40" aria-label="Distribución de gastos por categoría">
                  {#each arcPaths as slice (slice.name)}
                    <path d={slice.path} fill={slice.color} />
                  {/each}
                  <circle cx="100" cy="100" r="44" fill="var(--color-surface-card)" />
                  <text
                    x="100"
                    y="96"
                    text-anchor="middle"
                    font-size="11"
                    fill="var(--color-muted)"
                    font-family="var(--font-inter)"
                  >
                    Total
                  </text>
                  <text
                    x="100"
                    y="112"
                    text-anchor="middle"
                    font-size="14"
                    font-weight="500"
                    fill="var(--color-ink)"
                    font-family="var(--font-inter)"
                  >
                    $ 0
                  </text>
                </svg>
                <ul class="flex-1 space-y-2 w-full">
                  {#each categorySlices as slice (slice.name)}
                    <li class="flex items-center justify-between gap-3 text-sm">
                      <span class="flex items-center gap-2">
                        <span class="w-2.5 h-2.5 rounded-full" style="background: {slice.color}"></span>
                        <span class="text-ink">{slice.name}</span>
                      </span>
                      <span class="text-muted tabular-nums">{slice.percent}%</span>
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
            </div>
          </div>
        </Card>
      </section>

      <!-- Últimas transacciones -->
      <section class="space-y-3">
        <div class="flex items-baseline justify-between">
          <h2 class="font-waldenburg text-2xl font-light text-ink">Últimos movimientos</h2>
          <a href="/transactions" class="text-sm text-ink hover:underline">Ver todos</a>
        </div>
        {#if hasTransactions}
          <Card>
            <ul class="divide-y divide-hairline">
              <!-- lista de transacciones cuando haya datos -->
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
    {/if}
  </div>
</main>