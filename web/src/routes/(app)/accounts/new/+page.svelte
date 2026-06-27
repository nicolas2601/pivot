<script lang="ts">
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { createAccount, type AccountType } from '$lib/api/accounts';
  import Button from '$lib/components/Button.svelte';
  import TextInput from '$lib/components/TextInput.svelte';
  import Card from '$lib/components/Card.svelte';
  import { ApiException } from '$lib/utils/api-error';

  let name = $state('');
  let type = $state<AccountType>('debit');
  let currency = $state('COP');
  let openingBalance = $state(0);
  let color = $state('');
  let icon = $state('');

  let serverError = $state<string | null>(null);
  let submitting = $state(false);

  onMount(() => {
    if (!localStorage.getItem('access_token')) goto('/auth/login');
  });

  async function onSubmit(e: Event) {
    e.preventDefault();
    serverError = null;
    submitting = true;
    try {
      await createAccount({
        name,
        type,
        currency,
        opening_balance: openingBalance,
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
</script>

<svelte:head><title>Nueva cuenta — Mis finanzas</title></svelte:head>

<main class="min-h-screen bg-canvas p-8">
  <div class="max-w-md mx-auto space-y-6">
    <header>
      <h1 class="text-3xl font-light text-ink font-waldenburg">Nueva cuenta</h1>
      <p class="text-sm text-muted mt-1">Sumá una cuenta para registrar movimientos.</p>
    </header>

    <Card>
      <form onsubmit={onSubmit} class="space-y-4" novalidate>
        <TextInput
          label="Nombre"
          name="name"
          bind:value={name}
          placeholder="Ej. Bancolombia, Nequi, Efectivo"
          required
        />

        <div class="space-y-1.5">
          <label for="type" class="block text-sm font-medium text-ink">Tipo</label>
          <select
            id="type"
            bind:value={type}
            class="w-full px-4 py-3 h-11 bg-surface-card text-ink rounded-md border border-hairline-strong focus:outline-none focus:border-ink focus:ring-1 focus:ring-ink"
          >
            <option value="cash">Efectivo</option>
            <option value="debit">Débito</option>
            <option value="credit">Crédito</option>
            <option value="savings">Ahorros</option>
          </select>
        </div>

        <TextInput
          label="Moneda (3 letras)"
          name="currency"
          bind:value={currency}
          required
          maxLength={3}
        />

        <TextInput
          label="Saldo inicial (en pesos, sin centavos)"
          name="opening_balance"
          type="number"
          bind:value={openingBalance}
          hint="0 si empezás de cero"
        />

        {#if serverError}
          <p role="alert" class="text-sm text-semantic-error bg-surface-strong px-3 py-2 rounded">
            {serverError}
          </p>
        {/if}

        <div class="flex gap-2 pt-2">
          <Button variant="primary" type="submit" disabled={submitting}>
            {submitting ? 'Creando...' : 'Crear cuenta'}
          </Button>
          <Button variant="outline" type="button" onclick={() => goto('/accounts')}>
            Cancelar
          </Button>
        </div>
      </form>
    </Card>
  </div>
</main>