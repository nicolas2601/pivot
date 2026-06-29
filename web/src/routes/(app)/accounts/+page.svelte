<script lang="ts">
  /**
   * Accounts — lista mobile-first con cards balanceadas.
   * Top header tiene saludo + botón "Nueva cuenta" (no nav inline — el
   * BottomNav ya provee navegación).
   */
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
  import { me, logout } from '$lib/api/auth';
  import type { User } from '$lib/schemas/auth';
  import {
    listAccounts,
    deleteAccount,
    type Account
  } from '$lib/api/accounts';
  import { seedCategories, listCategories } from '$lib/api/categories';
  import Button from '$lib/components/Button.svelte';
  import Card from '$lib/components/Card.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import { getAccessToken, clearAccessToken } from '$lib/utils/auth-interceptor';
  import { toast } from '$lib/stores/toast.svelte';

  const qc = useQueryClient();

  let user = $state<User | null>(null);
  let deleteTarget = $state<Account | null>(null);
  let loading = $state(true);

  const userQuery = createQuery(() => ({ queryKey: ['me'], queryFn: me, retry: false }));
  const accountsQuery = createQuery(() => ({ queryKey: ['accounts'], queryFn: listAccounts }));
  const categoriesQuery = createQuery(() => ({ queryKey: ['categories'], queryFn: () => listCategories() }));

  const deleteMutation = createMutation(() => ({
    mutationFn: (id: string) => deleteAccount(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['accounts'] });
      toast.success('Cuenta eliminada');
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

<main class="bg-canvas min-h-screen">
  <div class="max-w-5xl mx-auto px-4 md:px-6 py-6 md:py-10 space-y-6">
    {#if loading}
      <p class="text-muted py-12 text-center">Cargando...</p>
    {:else}
      <header class="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <p class="text-xs uppercase tracking-wider text-muted">Tu dinero</p>
          <h1 class="font-waldenburg text-4xl md:text-5xl font-light text-ink mt-1">Cuentas</h1>
        </div>
        <Button variant="primary" type="button" onclick={() => goto('/accounts/new')}>
          Nueva cuenta
        </Button>
      </header>

      {#if accountsQuery.isPending}
        <p class="text-muted text-center py-12">Cargando...</p>
      {:else if !accountsQuery.data || accountsQuery.data.length === 0}
        <Card>
          <div class="text-center py-10 space-y-4">
            <div>
              <p class="font-waldenburg text-2xl font-light text-ink">Empezá por acá</p>
              <p class="text-sm text-muted mt-1">Creá tu primera cuenta para registrar movimientos.</p>
            </div>
            <div class="flex justify-center">
              <Button variant="primary" type="button" onclick={() => goto('/accounts/new')}>
                Crear cuenta
              </Button>
            </div>
          </div>
        </Card>
      {:else}
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {#each accountsQuery.data as acc (acc.id)}
            <Card>
              <div class="space-y-4">
                <div class="flex items-start justify-between gap-2">
                  <div>
                    <h2 class="text-base font-medium text-ink">{acc.name}</h2>
                    <p class="text-xs text-muted uppercase tracking-wider mt-1">{typeLabel(acc.type)}</p>
                  </div>
                  <span class="text-xs text-muted bg-surface-strong px-2.5 py-1 rounded-pill">
                    {acc.currency}
                  </span>
                </div>
                <p class="font-waldenburg text-3xl font-light text-ink tabular-nums">
                  {formatBalance(acc.opening_balance, acc.currency)}
                </p>
                <div class="flex gap-2 pt-1">
                  <Button variant="outline" type="button" onclick={() => goto(`/accounts/${acc.id}`)} class="flex-1">
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

      <div class="pt-2 flex justify-center">
        <Button variant="tertiary" type="button" onclick={handleLogout}>Cerrar sesión</Button>
      </div>
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