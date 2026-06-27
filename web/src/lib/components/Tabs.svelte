<script lang="ts">
  /**
   * Tabs — pill tabs para filtros (Tipo, Cuenta, etc.).
   * Estilo: pill geometry, surface-strong base, ink fill cuando activo.
   * Accesibilidad: role="tablist" / role="tab" con aria-selected.
   */

  import type { Snippet } from 'svelte';

  interface TabItem {
    id: string;
    label: string;
  }

  interface Props {
    items: TabItem[];
    active: string;
    onChange?: (id: string) => void;
    label?: string;
    children?: Snippet;
  }

  let { items, active = $bindable(), onChange, label }: Props = $props();

  function handleClick(id: string) {
    active = id;
    onChange?.(id);
  }
</script>

<div role="tablist" aria-label={label} class="flex flex-wrap gap-2">
  {#each items as item (item.id)}
    <button
      type="button"
      role="tab"
      aria-selected={active === item.id}
      onclick={() => handleClick(item.id)}
      class="px-3 py-1.5 rounded-pill text-sm font-medium transition-colors {active === item.id ? 'bg-ink text-on-primary' : 'bg-surface-strong text-ink hover:bg-hairline'}"
    >
      {item.label}
    </button>
  {/each}
</div>