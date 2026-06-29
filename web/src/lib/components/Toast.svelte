<script lang="ts">
  import { fly } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { toast } from '$lib/stores/toast.svelte';

  const variantStyles = {
    success: 'bg-semantic-success/10 border-semantic-success/30 text-semantic-success',
    error: 'bg-semantic-error/10 border-semantic-error/30 text-semantic-error',
    info: 'bg-surface-card border-hairline text-ink',
    warning: 'bg-semantic-warning/10 border-semantic-warning/30 text-semantic-warning'
  } as const;

  const icons = {
    success: '✓',
    error: '!',
    info: 'i',
    warning: '⚠'
  } as const;
</script>

<!--
  Fixed top-right stack of toasts. No portal needed in Svelte 5.
  Keyboard accessible: each toast has a close button; tab through them.
-->
<aside
  class="fixed top-4 right-4 z-50 flex flex-col gap-2 w-full max-w-sm pointer-events-none"
  aria-live="polite"
  aria-label="Notificaciones"
>
  {#each toast.items as t (t.id)}
    <div
      role="status"
      class="pointer-events-auto rounded-md border px-4 py-3 shadow-lg backdrop-blur-sm flex gap-3 items-start {variantStyles[t.variant]}"
      transition:fly={{ x: 300, duration: 220, easing: cubicOut }}
    >
      <span
        class="inline-flex items-center justify-center w-6 h-6 rounded-full text-xs font-bold shrink-0 border {variantStyles[t.variant]}"
        aria-hidden="true"
      >
        {icons[t.variant]}
      </span>
      <div class="flex-1 min-w-0 space-y-0.5">
        {#if t.title}
          <p class="font-waldenburg text-sm font-semibold leading-tight">{t.title}</p>
        {/if}
        <p class="text-sm leading-snug">{t.message}</p>
      </div>
      <button
        type="button"
        onclick={() => toast.dismiss(t.id)}
        class="shrink-0 text-xs opacity-70 hover:opacity-100"
        aria-label="Cerrar notificación"
      >
        ×
      </button>
    </div>
  {/each}
</aside>