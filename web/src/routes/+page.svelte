<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { createQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
  import { me, logout } from '$lib/api/auth';
  import { clearAccessToken, getAccessToken } from '$lib/utils/auth-interceptor';

  const qc = useQueryClient();

  const userQuery = createQuery(() => ({
    queryKey: ['me'],
    queryFn: me,
    enabled: !!getAccessToken(),
    retry: false
  }));

  let isAuthed = $state(!!getAccessToken());

  const logoutMutation = createMutation(() => ({
    mutationFn: logout,
    onSuccess: () => {
      clearAccessToken();
      qc.clear();
      goto('/auth/login');
    }
  }));

  onMount(() => {
    if (!getAccessToken()) {
      goto('/auth/login');
    } else {
      isAuthed = true;
    }
  });
</script>

<main class="min-h-screen bg-slate-50 p-8">
  <div class="max-w-4xl mx-auto space-y-6">
    {#if isAuthed}
      <header class="flex justify-between items-center">
        <h1 class="text-3xl font-bold">Mis finanzas</h1>
        {#if userQuery.data}
          <div class="flex items-center gap-4">
            <span class="text-sm text-slate-600">
              {userQuery.data.display_name || userQuery.data.email}
            </span>
            <button
              type="button"
              onclick={() => logoutMutation.mutate()}
              disabled={logoutMutation.isPending}
              class="px-4 py-2 bg-slate-200 text-slate-700 rounded-md hover:bg-slate-300 text-sm font-medium disabled:opacity-50 transition-colors"
            >
              {logoutMutation.isPending ? 'Saliendo...' : 'Cerrar sesión'}
            </button>
          </div>
        {/if}
      </header>
    {/if}

    {#if isAuthed}
      <div class="bg-white rounded-lg shadow-sm p-12 text-center space-y-4">
        <h2 class="text-2xl font-bold text-slate-700">Bienvenido</h2>
        <p class="text-slate-600">
          Tu dashboard va a aparecer acá cuando construyamos transactions en Fase 3.
        </p>
        <p class="text-sm text-slate-500">
          Auth funcionando. El botón "Cerrar sesión" arriba te redirige a /auth/login.
        </p>
      </div>
    {/if}
  </div>
</main>