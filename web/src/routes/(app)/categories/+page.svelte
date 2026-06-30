<script lang="ts">
  /**
   * CategoriesPage — rediseño premium.
   *
   * Mejoras vs versión anterior:
   * - Lista → grid de cards (2 cols mobile, 3+ desktop) que no trunca nombres
   * - Iconos reales por categoría (vienen del backend: heart, utensils, etc.)
   * - Empty state con moodboard pastel en lugar de un card pelado
   * - List items entran con stagger via Svelte transition
   * - Botones de acción compactos sin borde (icon-only) en mobile, con label en desktop
   * - Header se separa del filtro (no se rompen en mobile)
   */
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
  import { fly, fade } from 'svelte/transition';
  import { quintOut } from 'svelte/easing';
  import { me, logout } from '$lib/api/auth';
  import type { User } from '$lib/schemas/auth';
  import {
    listCategories,
    deleteCategory,
    seedCategories,
    type Category
  } from '$lib/api/categories';
  import Button from '$lib/components/Button.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import Tabs from '$lib/components/Tabs.svelte';
  import NavIcon from '$lib/components/NavIcon.svelte';
  import CategoryIcon from '$lib/components/CategoryIcon.svelte';
  import { getAccessToken, clearAccessToken } from '$lib/utils/auth-interceptor';
  import { toast } from '$lib/stores/toast.svelte';

  const qc = useQueryClient();

  let user = $state<User | null>(null);
  let filter = $state<'all' | 'expense' | 'income'>('all');
  let deleteTarget = $state<Category | null>(null);
  let loading = $state(true);
  let mounted = $state(false);

  const userQuery = createQuery(() => ({ queryKey: ['me'], queryFn: me, retry: false }));
  const categoriesQuery = createQuery(() => ({
    queryKey: ['categories'],
    queryFn: () => listCategories(),
    staleTime: 60_000
  }));

  const deleteMutation = createMutation(() => ({
    mutationFn: (id: string) => deleteCategory(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['categories'] });
      toast.success('Categoría eliminada');
    },
    onError: (err: Error) => toast.error(err.message, 'No se pudo eliminar')
  }));

  onMount(async () => {
    mounted = true;
    const token = getAccessToken();
    if (!token) {
      goto('/auth/login');
      return;
    }
    try {
      user = await me();
    } catch {
      clearAccessToken();
      goto('/auth/login');
      return;
    } finally {
      loading = false;
    }
  });

  async function handleLogout() {
    try {
      await logout();
    } catch {}
    clearAccessToken();
    qc.clear();
    goto('/auth/login');
  }

  async function handleSeed() {
    await seedCategories();
    await qc.invalidateQueries({ queryKey: ['categories'] });
    toast.success('Categorías predeterminadas cargadas');
  }

  const visible = $derived(
    categoriesQuery.data
      ? filter === 'all'
        ? categoriesQuery.data
        : categoriesQuery.data.filter((c) => c.type === filter)
      : []
  );

  const totals = $derived(() => {
    const data = categoriesQuery.data ?? [];
    return {
      total: data.length,
      expense: data.filter((c) => c.type === 'expense').length,
      income: data.filter((c) => c.type === 'income').length
    };
  });

  type TotalKey = 'total' | 'expense' | 'income';
  const statTiles: ReadonlyArray<{ k: TotalKey; l: string }> = [
    { k: 'total', l: 'Total' },
    { k: 'expense', l: 'Gastos' },
    { k: 'income', l: 'Ingresos' }
  ];

  const filterTabs = [
    { id: 'all', label: 'Todas' },
    { id: 'expense', label: 'Gastos' },
    { id: 'income', label: 'Ingresos' }
  ];

  // Soft pastel que se le da a cada card de categoría como tinte de fondo
  // (10% alpha). Si la API no devuelve color caemos al hairline del design.
  function bgFor(color?: string | null): string {
    if (!color) return '#f0efed';
    return color + '26'; // ~15% alpha
  }
</script>

<svelte:head><title>Categorías — Pivot</title></svelte:head>

<main class="bg-canvas min-h-screen pb-32 md:pb-16">
  <div class="max-w-5xl mx-auto px-4 md:px-8 py-8 md:py-14 space-y-8 md:space-y-12">
    {#if loading}
      <div class="space-y-4" in:fade={{ duration: 200 }}>
        <div class="h-4 w-24 bg-surface-strong rounded-full"></div>
        <div class="h-12 w-48 bg-surface-strong rounded-md"></div>
      </div>
    {:else}
      <header class="space-y-3" in:fly={{ y: 12, duration: 500, easing: quintOut }}>
        <p class="text-xs uppercase tracking-[0.2em] text-muted">Organización</p>
        <div class="flex items-end justify-between gap-4 flex-wrap">
          <h1 class="font-waldenburg text-5xl md:text-6xl font-light text-ink leading-none">Categorías</h1>
          <Button variant="primary" type="button" onclick={() => goto('/categories/new')}>
            <NavIcon icon="plus" class="w-4 h-4 -ml-1" />
            <span>Nueva</span>
          </Button>
        </div>
      </header>

      <!-- Stat strip -->
      {#if mounted}
        <div
          class="grid grid-cols-3 gap-3"
          in:fly={{ y: 12, duration: 500, easing: quintOut, delay: 80 }}
        >
          {#each statTiles as s (s.k)}
            <div class="bg-surface-card border border-hairline rounded-2xl px-4 py-3 text-center">
              <p class="text-xs uppercase tracking-wider text-muted">{s.l}</p>
              <p class="font-waldenburg text-3xl text-ink leading-none mt-1">{totals()[s.k]}</p>
            </div>
          {/each}
        </div>
      {/if}

      <div
        class="flex flex-wrap items-center gap-2 justify-between"
        in:fly={{ y: 12, duration: 500, easing: quintOut, delay: 160 }}
      >
        <Tabs items={filterTabs} bind:active={filter} label="Filtrar categorías" />
        <Button variant="ghost" type="button" onclick={handleSeed}>
          <NavIcon icon="sparkles" class="w-4 h-4 -ml-0.5" />
          <span>Cargar predeterminadas</span>
        </Button>
      </div>

      {#if categoriesQuery.isPending}
        <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
          {#each Array(8) as _, i}
            <div class="bg-surface-card border border-hairline rounded-2xl p-4 animate-pulse">
              <div class="h-10 w-10 bg-surface-strong rounded-xl mb-3"></div>
              <div class="h-3 w-3/4 bg-surface-strong rounded mb-2"></div>
              <div class="h-2 w-1/3 bg-surface-strong rounded"></div>
            </div>
          {/each}
        </div>
      {:else if visible.length === 0}
        <div
          class="relative overflow-hidden bg-gradient-to-br from-canvas to-canvas-soft border border-hairline rounded-3xl p-10 md:p-16 text-center"
          in:fade={{ duration: 400 }}
        >
          <!-- Decorative pastel orbs (editorial style) -->
          <div class="absolute -top-20 -right-20 w-64 h-64 rounded-full bg-mint/20 blur-3xl pointer-events-none"></div>
          <div class="absolute -bottom-20 -left-20 w-64 h-64 rounded-full bg-peach/20 blur-3xl pointer-events-none"></div>

          <div class="relative max-w-md mx-auto space-y-5">
            <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-surface-card border border-hairline shadow-sm">
              <NavIcon icon="tag" class="w-7 h-7 text-muted" />
            </div>
            <h2 class="font-waldenburg text-3xl md:text-4xl font-light text-ink">
              {filter === 'all' ? 'Sin categorías todavía' : `Sin categorías de tipo ${filter === 'expense' ? 'gasto' : 'ingreso'}`}
            </h2>
            <p class="text-body text-sm md:text-base">
              {filter === 'all'
                ? 'Empezá cargando las 13 categorías predeterminadas con un click — las reconocés al instante gracias a sus íconos.'
                : 'Probá cambiando el filtro o cargá las predeterminadas.'}
            </p>
            <div class="flex justify-center gap-3 flex-wrap pt-2">
              <Button variant="outline" type="button" onclick={handleSeed}>
                <NavIcon icon="sparkles" class="w-4 h-4 -ml-0.5" />
                <span>Cargar predeterminadas</span>
              </Button>
              {#if filter !== 'all'}
                <Button variant="outline" type="button" onclick={() => (filter = 'all')}>
                  Ver todas
                </Button>
              {/if}
              <Button variant="primary" type="button" onclick={() => goto('/categories/new')}>
                <NavIcon icon="plus" class="w-4 h-4 -ml-1" />
                <span>Crear manualmente</span>
              </Button>
            </div>
          </div>
        </div>
      {:else}
        <ul class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
          {#each visible as cat, i (cat.id)}
            <li
              class="group relative bg-surface-card border border-hairline rounded-2xl p-4 hover:border-ink/20 hover:shadow-md transition-all duration-200"
              in:fly={{ y: 16, duration: 320, easing: quintOut, delay: i * 35 }}
              style="background: linear-gradient(to bottom right, {bgFor(cat.color)}, #ffffff 60%);"
            >
              <div class="flex items-start gap-3">
                <CategoryIcon icon={cat.icon} name={cat.name} color={cat.color ?? '#e7e5e4'} accent />
                <div class="min-w-0 flex-1">
                  <p class="font-medium text-ink leading-tight break-words">{cat.name}</p>
                  <div class="flex items-center gap-1.5 mt-1.5">
                    <span
                      class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-[10px] uppercase tracking-wider font-medium
                             {cat.type === 'expense' ? 'bg-peach/30 text-ink' : 'bg-mint/30 text-ink'}"
                    >
                      <span class="w-1 h-1 rounded-full {cat.type === 'expense' ? 'bg-ink/60' : 'bg-ink/60'}"></span>
                      {cat.type === 'expense' ? 'Gasto' : 'Ingreso'}
                    </span>
                    {#if cat.is_default}
                      <span class="text-[10px] uppercase tracking-wider text-muted">Default</span>
                    {/if}
                  </div>
                </div>
              </div>

              <!-- Action row: icons on mobile, full buttons on desktop -->
              <div class="mt-3 pt-3 border-t border-hairline/60 flex items-center justify-end gap-1">
                <button
                  type="button"
                  aria-label="Editar"
                  onclick={() => goto(`/categories/${cat.id}`)}
                  class="sm:hidden inline-flex items-center justify-center w-9 h-9 rounded-lg text-body hover:bg-hairline hover:text-ink transition-colors"
                >
                  <NavIcon icon="edit" class="w-4 h-4" />
                </button>
                <button
                  type="button"
                  aria-label="Eliminar"
                  onclick={() => (deleteTarget = cat)}
                  class="sm:hidden inline-flex items-center justify-center w-9 h-9 rounded-lg text-body hover:bg-semantic-error/10 hover:text-semantic-error transition-colors"
                >
                  <NavIcon icon="trash" class="w-4 h-4" />
                </button>
                <div class="hidden sm:flex gap-2">
                  <button
                    type="button"
                    onclick={() => goto(`/categories/${cat.id}`)}
                    class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium text-body hover:text-ink hover:bg-hairline transition-colors"
                  >
                    <NavIcon icon="edit" class="w-3.5 h-3.5" /> Editar
                  </button>
                  <button
                    type="button"
                    onclick={() => (deleteTarget = cat)}
                    class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium text-body hover:text-semantic-error hover:bg-semantic-error/10 transition-colors"
                  >
                    <NavIcon icon="trash" class="w-3.5 h-3.5" /> Eliminar
                  </button>
                </div>
              </div>
            </li>
          {/each}
        </ul>
      {/if}

      <div class="pt-2 flex justify-center">
        <Button variant="ghost" type="button" onclick={handleLogout}>
          <NavIcon icon="logout" class="w-4 h-4" />
          <span>Cerrar sesión</span>
        </Button>
      </div>
    {/if}
  </div>
</main>

<Modal
  open={deleteTarget !== null}
  title="Eliminar categoría"
  onClose={() => (deleteTarget = null)}
>
  {#snippet children()}
    {#if deleteTarget}
      <p>¿Eliminar <strong class="text-ink">{deleteTarget.name}</strong>? Si tiene transacciones asociadas, primero tendrás que reasignarlas.</p>
    {/if}
  {/snippet}
  {#snippet actions()}
    <Button variant="outline" type="button" onclick={() => (deleteTarget = null)}>
      Cancelar
    </Button>
    <Button
      variant="danger"
      type="button"
      onclick={async () => {
        if (deleteTarget) {
          await deleteMutation.mutateAsync(deleteTarget.id);
          deleteTarget = null;
        }
      }}
    >
      Eliminar
    </Button>
  {/snippet}
</Modal>