<script lang="ts">
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { getCategory, updateCategory, deleteCategory } from '$lib/api/categories';
  import type { Category, CategoryType } from '$lib/schemas/category';
  import Button from '$lib/components/Button.svelte';
  import TextInput from '$lib/components/TextInput.svelte';
  import Card from '$lib/components/Card.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import { ApiException } from '$lib/utils/api-error';

  let category = $state<Category | null>(null);
  let name = $state('');
  let type = $state<CategoryType>('expense');
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
      goto('/categories');
      return;
    }
    try {
      const cat = await getCategory(id);
      category = cat;
      name = cat.name;
      type = cat.type;
      color = cat.color ?? '';
      icon = cat.icon ?? '';
    } catch (e) {
      if (e instanceof ApiException && e.status === 404) goto('/categories');
    } finally {
      loading = false;
    }
  });

  async function onSubmit(e: Event) {
    e.preventDefault();
    if (!category) return;
    serverError = null;
    submitting = true;
    try {
      await updateCategory(category.id, {
        name,
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

  async function onDelete() {
    if (!category) return;
    try {
      await deleteCategory(category.id);
      await goto('/categories');
    } catch (e) {
      if (e instanceof ApiException) serverError = e.apiError.message;
    }
  }
</script>

<svelte:head><title>Editar categoría — Mis finanzas</title></svelte:head>

<main class="min-h-screen bg-canvas p-8">
  <div class="max-w-md mx-auto space-y-6">
    <header>
      <h1 class="text-3xl font-light text-ink font-waldenburg">Editar categoría</h1>
    </header>

    {#if loading}
      <p class="text-muted text-center py-12">Cargando...</p>
    {:else if category}
      <Card>
        <form onsubmit={onSubmit} class="space-y-4" novalidate>
          <TextInput label="Nombre" name="name" bind:value={name} required />

          <div class="space-y-1.5">
            <label for="type" class="block text-sm font-medium text-ink">Tipo</label>
            <select
              id="type"
              bind:value={type}
              disabled
              class="w-full px-4 py-3 h-11 bg-surface-card text-muted rounded-md border border-hairline-strong cursor-not-allowed"
            >
              <option value="expense">Gasto</option>
              <option value="income">Ingreso</option>
            </select>
            <p class="text-xs text-muted">El tipo no se puede cambiar</p>
          </div>

          <TextInput
            label="Color (hex, opcional)"
            name="color"
            bind:value={color}
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
              {submitting ? 'Guardando...' : 'Guardar'}
            </Button>
            <Button variant="outline" type="button" onclick={() => goto('/categories')}>
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

  <Modal open={deleteOpen} title="Eliminar categoría" onClose={() => (deleteOpen = false)}>
    {#snippet children()}
      {#if category}
        <p>¿Eliminar <strong class="text-ink">{category.name}</strong>?</p>
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