<script lang="ts">
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { createCategory, type CategoryType } from '$lib/api/categories';
  import Button from '$lib/components/Button.svelte';
  import TextInput from '$lib/components/TextInput.svelte';
  import Card from '$lib/components/Card.svelte';
  import { ApiException } from '$lib/utils/api-error';

  let name = $state('');
  let type = $state<CategoryType>('expense');
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
      await createCategory({
        name,
        type,
        color: color || undefined,
        icon: icon || undefined
      });
      await goto('/categories');
    } catch (e) {
      if (e instanceof ApiException) serverError = e.apiError.message;
      else serverError = 'Error de red';
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head><title>Nueva categoría — Mis finanzas</title></svelte:head>

<main class="min-h-screen bg-canvas p-8">
  <div class="max-w-md mx-auto space-y-6">
    <header>
      <h1 class="text-3xl font-light text-ink font-waldenburg">Nueva categoría</h1>
      <p class="text-sm text-muted mt-1">Agrupá tus movimientos por tipo.</p>
    </header>

    <Card>
      <form onsubmit={onSubmit} class="space-y-4" novalidate>
        <TextInput
          label="Nombre"
          name="name"
          bind:value={name}
          required
          placeholder="Ej. Supermercado, Salario, Freelance"
        />

        <div class="space-y-1.5">
          <label for="type" class="block text-sm font-medium text-ink">Tipo</label>
          <select
            id="type"
            bind:value={type}
            class="w-full px-4 py-3 h-11 bg-surface-card text-ink rounded-md border border-hairline-strong focus:outline-none focus:border-ink focus:ring-1 focus:ring-ink"
          >
            <option value="expense">Gasto</option>
            <option value="income">Ingreso</option>
          </select>
        </div>

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
        />

        {#if serverError}
          <p role="alert" class="text-sm text-semantic-error bg-surface-strong px-3 py-2 rounded">
            {serverError}
          </p>
        {/if}

        <div class="flex gap-2 pt-2">
          <Button variant="primary" type="submit" disabled={submitting}>
            {submitting ? 'Creando...' : 'Crear categoría'}
          </Button>
          <Button variant="outline" type="button" onclick={() => goto('/categories')}>
            Cancelar
          </Button>
        </div>
      </form>
    </Card>
  </div>
</main>