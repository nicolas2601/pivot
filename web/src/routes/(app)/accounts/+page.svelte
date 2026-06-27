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
    listAccounts,
    deleteAccount,
    type Account
  } from '$lib/api/accounts';
  import { seedCategories, listCategories, type Category } from '$lib/api/categories';
  import Button from '$lib/components/Button.svelte';
  import Card from '$lib/components/Card.svelte';
  import Modal from '$lib/components/Modal.svelte';

  const qc = useQueryClient();

  let user = $state<User | null>(null);
  let deleteTarget = $state<Account | null>(null);
  let loading = $state(true);

  const userQuery = createQuery(() => ({ queryKey: ['me'], queryFn: me, retry: false }));
  const accountsQuery = createQuery(() => ({ queryKey: ['accounts'], queryFn: listAccounts }));
  const categoriesQuery = createQuery(() => ({ queryKey: ['categories'], queryFn: () => listCategories() }));

  const deleteMutation = createMutation(() => ({
    mutationFn: (id: string) => deleteAccount(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['accounts'] })
  }));

  onMount(async () => {
    const token = localStorage.getItem('access_token');
    if (!token) {
      goto('/auth/login');
      return;
    }
    try {
      user = await me();
    } catch {
      localStorage.removeItem('access_token');
      goto('/auth/login');
      return;
    }
    loading = false;
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

  function formatBalance(cents: number, currency: string) {
    return new Intl.NumberFormat('es-CO', {
      style: 'currency',
      currency,
      maximumFractionDigits: 0
    }).format(cents / 100);
  }

  function typeLabel(t: string) {
    return ({ cash: 'Efectivo', debit: 'Débito', credit: 'Crédito', savings: 'Ahorros' } as Record<string, string>)[t] || t;
  }
</script>

<svelte:head><title>Cuentas — Mis finanzas</title></svelte:head>

<main class="min-h-screen bg-canvas p-8">
  <div class="max-w-5xl mx-auto space-y-8">
    <header class="flex justify-between items-center">
      <div>
        <h1 class="text-3xl font-light text-ink font-waldenburg">Cuentas</h1>
        <p class="text-sm text-muted mt-1">Hola, {user?.display_name || user?.email}</p>
      </div>
      <div class="flex items-center gap-3">
        <Button variant="outline" type="button" onclick={() => goto('/')}>Dashboard</Button>
        <Button variant="tertiary" type="button" onclick={handleLogout}>Cerrar sesión</Button>
      </div>
    </header>

    <nav class="flex gap-6 border-b border-hairline pb-3 text-sm">
      <a href="/accounts" class="text-ink font-medium border-b-2 border-ink pb-2 -mb-3">Cuentas</a>
      <a href="/categories" class="text-muted hover:text-ink pb-2">Categorías</a>
    </nav>

    <div class="flex justify-between items-center">
      <h2 class="text-lg font-medium text-ink">Tus cuentas</h2>
      <Button variant="primary" type="button" onclick={() => goto('/accounts/new')}>
        Nueva cuenta
      </Button>
    </div>

    {#if accountsQuery.isPending}
      <p class="text-muted text-center py-12">Cargando...</p>
    {:else if !accountsQuery.data || accountsQuery.data.length === 0}
      <Card>
        <div class="text-center py-8 space-y-3">
          <p class="text-ink">Aún no tenés cuentas</p>
          <p class="text-sm text-muted">Creá tu primera cuenta para empezar a registrar movimientos.</p>
          <div class="pt-2">
            <Button variant="primary" type="button" onclick={() => goto('/accounts/new')}>
              Crear cuenta
            </Button>
          </div>
        </div>
      </Card>
    {:else}
      <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {#each accountsQuery.data as acc (acc.id)}
          <Card>
            <div class="space-y-3">
              <div class="flex items-start justify-between">
                <div>
                  <h3 class="text-lg font-medium text-ink">{acc.name}</h3>
                  <p class="text-xs text-muted uppercase tracking-wide mt-1">{typeLabel(acc.type)}</p>
                </div>
                <span class="text-xs text-muted bg-surface-strong px-2 py-1 rounded-pill">
                  {acc.currency}
                </span>
              </div>
              <p class="text-2xl font-light text-ink font-waldenburg">
                {formatBalance(acc.opening_balance, acc.currency)}
              </p>
              <div class="flex gap-2 pt-2">
                <Button variant="outline" type="button" onclick={() => goto(`/accounts/${acc.id}`)}>
                  Editar
                </Button>
                <Button variant="tertiary" type="button" onclick={() => (deleteTarget = acc)}>
                  Eliminar
                </Button>
              </div>
            </div>
          </Card>
        {/each}
      </div>
    {/if}

    {#if categoriesQuery.data && categoriesQuery.data.length === 0 && !accountsQuery.isPending}
      <Card>
        <div class="text-center py-6 space-y-3">
          <p class="text-sm text-ink">No tenés categorías todavía</p>
          <p class="text-xs text-muted">Creá las categorías predeterminadas (Alimentación, Transporte, Salario...) con un click.</p>
          <Button variant="outline" type="button" onclick={handleSeed}>
            Cargar categorías predeterminadas
          </Button>
        </div>
      </Card>
    {/if}
  </div>

  <Modal
    open={deleteTarget !== null}
    title="Eliminar cuenta"
    onClose={() => (deleteTarget = null)}
  >
    {#snippet children()}
      {#if deleteTarget}
        <p>¿Estás seguro de eliminar <strong class="text-ink">{deleteTarget.name}</strong>? Esta acción no se puede deshacer.</p>
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