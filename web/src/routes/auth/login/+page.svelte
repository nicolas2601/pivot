<script lang="ts">
  import { goto } from '$app/navigation';
  import { superForm } from 'sveltekit-superforms';
  import { zod4Client } from 'sveltekit-superforms/adapters';
  import { LoginInputSchema } from '$lib/schemas/auth';
  import { login } from '$lib/api/auth';
  import { ApiException } from '$lib/utils/api-error';
  import { setAccessToken } from '$lib/utils/auth-interceptor';

  let { data } = $props();

  let serverError = $state<string | null>(null);

  const { form, errors, enhance, submitting } = superForm(data.form, {
    SPA: true,
    validators: zod4Client(LoginInputSchema),
    async onUpdate({ form }) {
      if (!form.valid) return;
      try {
        const result = await login(form.data as { email: string; password: string });
        setAccessToken(result.access_token);
        await goto('/');
      } catch (e) {
        if (e instanceof ApiException) {
          serverError = e.apiError.message;
        } else {
          serverError = 'Error de red';
        }
      }
    }
  });
</script>

<div class="bg-white rounded-lg shadow-sm p-8 space-y-6">
  <header class="text-center space-y-2">
    <h1 class="text-2xl font-bold">Iniciar sesión</h1>
    <p class="text-sm text-slate-600">Bienvenido de vuelta</p>
  </header>

  <form method="POST" use:enhance class="space-y-4" aria-describedby={serverError ? 'login-error' : undefined}>
    <div class="space-y-1">
      <label for="email" class="text-sm font-medium text-slate-700">Email</label>
      <input
        id="email"
        name="email"
        type="email"
        autocomplete="email"
        required
        aria-invalid={$errors.email ? 'true' : undefined}
        aria-describedby={$errors.email ? 'email-error' : undefined}
        bind:value={$form.email}
        class="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
      {#if $errors.email}
        <p id="email-error" class="text-xs text-red-600">{$errors.email[0]}</p>
      {/if}
    </div>

    <div class="space-y-1">
      <label for="password" class="text-sm font-medium text-slate-700">Contraseña</label>
      <input
        id="password"
        name="password"
        type="password"
        autocomplete="current-password"
        required
        aria-invalid={$errors.password ? 'true' : undefined}
        aria-describedby={$errors.password ? 'password-error' : undefined}
        bind:value={$form.password}
        class="w-full px-3 py-2 border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
      {#if $errors.password}
        <p id="password-error" class="text-xs text-red-600">{$errors.password[0]}</p>
      {/if}
    </div>

    {#if serverError}
      <p id="login-error" role="alert" class="text-sm text-red-600 bg-red-50 px-3 py-2 rounded">
        {serverError}
      </p>
    {/if}

    <button
      type="submit"
      disabled={$submitting}
      class="w-full py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium transition-colors"
    >
      {$submitting ? 'Ingresando...' : 'Ingresar'}
    </button>

    <p class="text-sm text-center text-slate-600">
      ¿No tenés cuenta?
      <a href="/auth/register" class="text-blue-600 hover:underline">Registrate</a>
    </p>
  </form>
</div>