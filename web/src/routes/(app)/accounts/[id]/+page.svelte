<script lang="ts">
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { getAccount, updateAccount, deleteAccount } from '$lib/api/accounts';
  import type { Account } from '$lib/schemas/account';
  import Button from '$lib/components/Button.svelte';
  import TextInput from '$lib/components/TextInput.svelte';
  import Card from '$lib/components/Card.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import { ApiException } from '$lib/utils/api-error';

  let account = $state<Account | null>(null);
  let name = $state('');
  let color = $state('');
  let icon = $state('');
  let serverError = $state<string | null>(null);
  let submitting = $state(false);
  let deleteOpen = $state(false);
  let loading = $state(true);

  onMount(async () => {
    if (!localStorage.getItem('access_token')) {
      goto('/auth/login');
      return;
    }
    const id = $page.params.id;
    if (!id) {
      goto('/accounts');
      return;
    }
    try {
      const acc = await getAccount(id);
      account = acc;
      name = acc.name;
      color = acc.color ?? '';
      icon = acc.icon ?? '';
    } catch (e) {
      if (e instanceof ApiException && e.status === 404) {
        goto('/accounts');
      }
    } finally {
      loading = false;
    }
  });

  async function onSubmit(e: Event) {
    e.preventDefault();
    if (!account) return;
    serverError = null;
    submitting = true;
    try {
      await updateAccount(account.id, {
        name,
        color: color || undefined,
        icon: icon || undefined
      });
      await goto('/accounts');
    } catch (e) {
      if (e instanceof ApiException) serverError = e.apiError.message;
      else serverError = 'Error de red';
    } finally {
      submitting = false;
    }
  }

  async function onDelete() {
    if (!account) return;
    try {
      await deleteAccount(account.id);
      await goto('/accounts');
    } catch (e) {
      if (e instanceof ApiException) serverError = e.apiError.message;
    }
  }
</script>

<svelte:head><title>Editar cuenta — Mis finanzas</title></svelte:head>

<main class="min-h-screen bg-canvas p-8">
  <div class="max-w-md mx-auto space-y-6">
    <header>
      <h1 class="text-3xl font-light text-ink font-waldenburg">Editar cuenta</h1>
      {#if account}
        <p class="text-sm text-muted mt-1">Modificá los datos de <strong>{account.name}</strong></p>
      {/if}
    </header>

    {#if loading}
      <p class="text-muted text-center py-12">Cargando...</p>
    {:else if account}
      <Card>
        <form onsubmit={onSubmit} class="space-y-4" novalidate>
          <TextInput label="Nombre" name="name" bind:value={name} required />
          <TextInput
            label="Color (hex, opcional)"
            name="color"
            bind:value={color}
            placeholder="#0c0a09"
          />
          <TextInput
            label="Icono (opcional)"
            name="icon"
            bind:value={icon}
            placeholder="wallet, card, cash..."
          />

          {#if serverError}
            <p role="alert" class="text-sm text-semantic-error bg-surface-strong px-3 py-2 rounded">
              {serverError}
            </p>
          {/if}

          <div class="flex gap-2 pt-2">
            <Button variant="primary" type="submit" disabled={submitting}>
              {submitting ? 'Guardando...' : 'Guardar'}
            </Button>
            <Button variant="outline" type="button" onclick={() => goto('/accounts')}>
              Cancelar
            </Button>
            <div class="flex-1"></div>
            <Button variant="danger" type="button" onclick={() => (deleteOpen = true)}>
              Eliminar
            </Button>
          </div>
        </form>
      </Card>
    {/if}
  </div>

  <Modal open={deleteOpen} title="Eliminar cuenta" onClose={() => (deleteOpen = false)}>
    {#snippet children()}
      {#if account}
        <p>¿Eliminar <strong class="text-ink">{account.name}</strong>? Esta acción no se puede deshacer.</p>
      {/if}
    {/snippet}
    {#snippet actions()}
      <Button variant="outline" type="button" onclick={() => (deleteOpen = false)}>
        Cancelar
      </Button>
      <Button variant="danger" type="button" onclick={onDelete}>Eliminar</Button>
    {/snippet}
  </Modal>
</main>