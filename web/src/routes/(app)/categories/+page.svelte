<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import {
    createQuery,
    createMutation,
    useQueryClient
  } from '@tanstack/svelte-query';
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

  const qc = useQueryClient();

  let user = $state<User | null>(null);
  let filter = $state<'all' | 'expense' | 'income'>('all');
  let deleteTarget = $state<Category | null>(null);

  const userQuery = createQuery(() => ({ queryKey: ['me'], queryFn: me, retry: false }));
  const categoriesQuery = createQuery(() => ({ queryKey: ['categories'], queryFn: () => listCategories() }));

  const deleteMutation = createMutation(() => ({
    mutationFn: (id: string) => deleteCategory(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] })
  }));

  onMount(async () => {
    if (!localStorage.getItem('access_token')) {
      goto('/auth/login');
      return;
    }
    try {
      user = await me();
    } catch {
      localStorage.removeItem('access_token');
      goto('/auth/login');
    }
  });

  async function handleLogout() {
    try { await logout(); } catch {}
    localStorage.removeItem('access_token');
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
</script>

<svelte:head><title>Categorías — Mis finanzas</title></svelte:head>

<main class="min-h-screen bg-canvas p-8">
  <div class="max-w-5xl mx-auto space-y-8">
    <header class="flex justify-between items-center">
      <div>
        <h1 class="text-3xl font-light text-ink font-waldenburg">Categorías</h1>
        <p class="text-sm text-muted mt-1">Hola, {user?.display_name || user?.email}</p>
      </div>
      <div class="flex items-center gap-3">
        <Button variant="outline" type="button" onclick={() => goto('/')}>Dashboard</Button>
        <Button variant="tertiary" type="button" onclick={handleLogout}>Cerrar sesión</Button>
      </div>
    </header>

    <nav class="flex gap-6 border-b border-hairline pb-3 text-sm">
      <a href="/accounts" class="text-muted hover:text-ink pb-2">Cuentas</a>
      <a href="/categories" class="text-ink font-medium border-b-2 border-ink pb-2 -mb-3">Categorías</a>
    </nav>

    <div class="flex flex-wrap gap-2 items-center">
      <button
        type="button"
        onclick={() => (filter = 'all')}
        class="px-3 py-1.5 rounded-pill text-sm {filter === 'all' ? 'bg-ink text-on-primary' : 'bg-surface-strong text-ink hover:bg-hairline'}"
      >
        Todas
      </button>
      <button
        type="button"
        onclick={() => (filter = 'expense')}
        class="px-3 py-1.5 rounded-pill text-sm {filter === 'expense' ? 'bg-ink text-on-primary' : 'bg-surface-strong text-ink hover:bg-hairline'}"
      >
        Gastos
      </button>
      <button
        type="button"
        onclick={() => (filter = 'income')}
        class="px-3 py-1.5 rounded-pill text-sm {filter === 'income' ? 'bg-ink text-on-primary' : 'bg-surface-strong text-ink hover:bg-hairline'}"
      >
        Ingresos
      </button>
      <div class="flex-1"></div>
      <Button variant="outline" type="button" onclick={handleSeed}>
        Cargar defaults
      </Button>
      <Button variant="primary" type="button" onclick={() => goto('/categories/new')}>
        Nueva
      </Button>
    </div>

    {#if categoriesQuery.isPending}
      <p class="text-muted text-center py-12">Cargando...</p>
    {:else if visible.length === 0}
      <Card>
        <div class="text-center py-8 space-y-3">
          <p class="text-ink">No hay categorías</p>
          <p class="text-sm text-muted">Creá categorías o cargá las predeterminadas en español.</p>
          <div class="pt-2 flex justify-center gap-2">
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
      <div class="bg-surface-card border border-hairline rounded-xl overflow-hidden">
        {#each visible as cat (cat.id)}
          <div class="flex items-center justify-between px-6 py-4 border-b border-hairline last:border-b-0">
            <div class="flex items-center gap-3">
              <span
                class="w-9 h-9 rounded-full bg-surface-strong flex items-center justify-center text-sm"
              >
                {cat.name.charAt(0).toUpperCase()}
              </span>
              <div>
                <p class="text-ink font-medium">{cat.name}</p>
                <p class="text-xs text-muted uppercase tracking-wide">{typeLabel(cat.type)}</p>
              </div>
            </div>
            <div class="flex gap-2">
              <Button variant="outline" type="button" onclick={() => goto(`/categories/${cat.id}`)}>
                Editar
              </Button>
              <Button variant="tertiary" type="button" onclick={() => (deleteTarget = cat)}>
                Eliminar
              </Button>
            </div>
          </div>
        {/each}
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