<script lang="ts">
  /**
   * Categories — lista mobile-first con tabs para filtrar y rows limpias.
   * Las categorías se muestran en una lista con avatar (inicial) + tipo.
   */
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
  import { me, logout } from '$lib/api/auth';
  import type { User } from '$lib/schemas/auth';
  import {
    listCategories,
    deleteCategory,
    seedCategories,
    type Category
  } from '$lib/api/categories';
  import Button from '$lib/components/Button.svelte';
  import Card from '$lib/components/Card.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import Tabs from '$lib/components/Tabs.svelte';
  import Avatar from '$lib/components/Avatar.svelte';
  import { getAccessToken, clearAccessToken } from '$lib/utils/auth-interceptor';
  import { toast } from '$lib/stores/toast.svelte';

  const qc = useQueryClient();

  let user = $state<User | null>(null);
  let filter = $state<'all' | 'expense' | 'income'>('all');
  let deleteTarget = $state<Category | null>(null);
  let loading = $state(true);

  const userQuery = createQuery(() => ({ queryKey: ['me'], queryFn: me, retry: false }));
  const categoriesQuery = createQuery(() => ({ queryKey: ['categories'], queryFn: () => listCategories() }));

  const deleteMutation = createMutation(() => ({
    mutationFn: (id: string) => deleteCategory(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['categories'] });
      toast.success('Categoría eliminada');
    },
    onError: (err: Error) => toast.error(err.message, 'No se pudo eliminar')
  }));

  onMount(async () => {
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
    try { await logout(); } catch {}
    clearAccessToken();
    qc.clear();
    goto('/auth/login');
  }

  async function handleSeed() {
    await seedCategories();
    await qc.invalidateQueries({ queryKey: ['categories'] });
  }

  const visible = $derived(
    categoriesQuery.data
      ? filter === 'all'
        ? categoriesQuery.data
        : categoriesQuery.data.filter((c) => c.type === filter)
      : []
  );

  function typeLabel(t: string) {
    return t === 'expense' ? 'Gasto' : 'Ingreso';
  }

  const filterTabs = [
    { id: 'all', label: 'Todas' },
    { id: 'expense', label: 'Gastos' },
    { id: 'income', label: 'Ingresos' }
  ];
</script>

<svelte:head><title>Categorías — Mis finanzas</title></svelte:head>

<main class="bg-canvas min-h-screen">
  <div class="max-w-3xl mx-auto px-4 md:px-6 py-6 md:py-10 space-y-6">
    {#if loading}
      <p class="text-muted py-12 text-center">Cargando...</p>
    {:else}
      <header class="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <p class="text-xs uppercase tracking-wider text-muted">Organización</p>
          <h1 class="font-waldenburg text-4xl md:text-5xl font-light text-ink mt-1">Categorías</h1>
        </div>
        <Button variant="primary" type="button" onclick={() => goto('/categories/new')}>
          Nueva
        </Button>
      </header>

      <div class="flex flex-wrap items-center gap-2 justify-between">
        <Tabs items={filterTabs} bind:active={filter} label="Filtrar categorías" />
        <Button variant="tertiary" type="button" onclick={handleSeed}>
          Cargar defaults
        </Button>
      </div>

      {#if categoriesQuery.isPending}
        <p class="text-muted text-center py-12">Cargando...</p>
      {:else if visible.length === 0}
        <Card>
          <div class="text-center py-10 space-y-4">
            <div>
              <p class="font-waldenburg text-2xl font-light text-ink">Sin categorías</p>
              <p class="text-sm text-muted mt-1">Creá categorías manualmente o cargá las predeterminadas.</p>
            </div>
            <div class="flex justify-center gap-2 flex-wrap">
              <Button variant="outline" type="button" onclick={handleSeed}>
                Cargar predeterminadas
              </Button>
              <Button variant="primary" type="button" onclick={() => goto('/categories/new')}>
                Crear manualmente
              </Button>
            </div>
          </div>
        </Card>
      {:else}
        <Card>
          <ul class="divide-y divide-hairline">
            {#each visible as cat (cat.id)}
              <li class="flex items-center justify-between gap-3 px-4 py-3 first:pt-0 last:pb-0 sm:px-6 sm:py-4">
                <div class="flex items-center gap-3 min-w-0">
                  <Avatar name={cat.name} />
                  <div class="min-w-0">
                    <p class="text-ink font-medium truncate">{cat.name}</p>
                    <p class="text-xs text-muted uppercase tracking-wider mt-0.5">{typeLabel(cat.type)}</p>
                  </div>
                </div>
                <div class="flex gap-2 shrink-0">
                  <Button variant="outline" type="button" onclick={() => goto(`/categories/${cat.id}`)}>
                    Editar
                  </Button>
                  <Button variant="tertiary" type="button" onclick={() => (deleteTarget = cat)}>
                    Eliminar
                  </Button>
                </div>
              </li>
            {/each}
          </ul>
        </Card>
      {/if}

      <div class="pt-2 flex justify-center">
        <Button variant="tertiary" type="button" onclick={handleLogout}>Cerrar sesión</Button>
      </div>
    {/if}
  </div>

  <Modal
    open={deleteTarget !== null}
    title="Eliminar categoría"
    onClose={() => (deleteTarget = null)}
  >
    {#snippet children()}
      {#if deleteTarget}
        <p>¿Eliminar <strong class="text-ink">{deleteTarget.name}</strong>?</p>
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
</main>