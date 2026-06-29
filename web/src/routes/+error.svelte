<script lang="ts">
  import { page } from '$app/stores';
  import { fade, fly } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
</script>

<main class="min-h-screen bg-canvas flex items-center justify-center px-6">
  <div class="max-w-md text-center space-y-6" in:fade={{ duration: 220 }}>
    <p
      class="font-waldenburg text-[8rem] md:text-[10rem] font-light text-ink/15 leading-none"
      in:fly={{ y: -20, duration: 360, delay: 60, easing: cubicOut }}
    >
      {$page.status}
    </p>
    <div class="space-y-3" in:fly={{ y: 16, duration: 380, delay: 160, easing: cubicOut }}>
      <h1 class="font-waldenburg text-3xl md:text-4xl font-light text-ink">
        {$page.status === 404 ? 'No encontramos esa página' : 'Algo salió mal'}
      </h1>
      <p class="text-muted text-sm md:text-base">
        {$page.error?.message ?? 'La página que buscás no existe o fue movida.'}
      </p>
      <div class="flex gap-2 justify-center pt-4">
        <a
          href="/"
          class="inline-flex items-center justify-center px-5 py-2 rounded-md bg-ink text-on-primary text-sm font-medium hover:opacity-90 transition-opacity"
        >
          Volver al inicio
        </a>
        <a
          href="/accounts"
          class="inline-flex items-center justify-center px-5 py-2 rounded-md border border-hairline text-sm font-medium text-ink hover:bg-surface-card transition-colors"
        >
          Ver cuentas
        </a>
      </div>
    </div>
  </div>
</main>