<script lang="ts">
  /**
   * NavIcon — iconos SVG inline del BottomNav. Sin dependencias externas.
   * Stroke 1.6, currentColor, viewBox 24x24 — siguen el DESIGN.md.
   */
  type IconName = 'home' | 'wallet' | 'tag' | 'list' | 'plane' | 'target' | 'repeat';
  let { icon }: { icon: IconName } = $props();
</script>

<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true">
  {#if icon === 'home'}
    <path d="M3 11l9-8 9 8v9a2 2 0 0 1-2 2h-4v-7h-6v7H5a2 2 0 0 1-2-2v-9z" stroke-linecap="round" stroke-linejoin="round" />
  {:else if icon === 'wallet'}
    <path d="M3 7a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7z" stroke-linecap="round" stroke-linejoin="round" />
    <path d="M3 8h18" stroke-linecap="round" /><circle cx="17" cy="13" r="1.2" fill="currentColor" />
  {:else if icon === 'tag'}
    <path d="M20.59 13.41L13.42 20.58a2 2 0 0 1-2.83 0L3 13V3h10l7.59 7.59a2 2 0 0 1 0 2.82z" stroke-linecap="round" stroke-linejoin="round" />
    <circle cx="7.5" cy="7.5" r="1.2" fill="currentColor" />
  {:else if icon === 'list'}
    <path d="M8 6h13M8 12h13M8 18h13" stroke-linecap="round" /><circle cx="4" cy="6" r="1" fill="currentColor" /><circle cx="4" cy="12" r="1" fill="currentColor" /><circle cx="4" cy="18" r="1" fill="currentColor" />
  {:else if icon === 'plane'}
    <path d="M2 16l20-7L4 4l3 8-2 4z" stroke-linecap="round" stroke-linejoin="round" /><path d="M9 14l5 7" stroke-linecap="round" />
  {:else if icon === 'target'}
    <circle cx="12" cy="12" r="9" stroke-linecap="round" /><circle cx="12" cy="12" r="5" /><circle cx="12" cy="12" r="1.5" fill="currentColor" />
  {:else if icon === 'repeat'}
    <path d="M17 1l4 4-4 4" stroke-linecap="round" stroke-linejoin="round" /><path d="M3 11V9a4 4 0 0 1 4-4h14" stroke-linecap="round" stroke-linejoin="round" />
    <path d="M7 23l-4-4 4-4" stroke-linecap="round" stroke-linejoin="round" /><path d="M21 13v2a4 4 0 0 1-4 4H3" stroke-linecap="round" stroke-linejoin="round" />
  {/if}
</svg>