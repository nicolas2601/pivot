<script lang="ts">
  /**
   * (app) layout — wraps every authenticated page with the BottomNav.
   * Mobile: bottom tab bar (fixed) + padding-bottom para que el contenido
   * no quede tapado. Desktop: top nav horizontal.
   *
   * No tocamos /+layout.svelte (root) — esto es un layout group separado
   * que solo aplica a las rutas dentro de (app)/. /auth/* queda fuera.
   */
  import BottomNav from '$lib/components/BottomNav.svelte';

  let { children } = $props();
</script>

<div class="min-h-screen bg-canvas pb-20 md:pb-0">
  <BottomNav />
  {@render children()}
</div>