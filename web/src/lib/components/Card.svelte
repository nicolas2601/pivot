<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    href?: string;
    onclick?: (e: MouseEvent) => void;
    children: Snippet;
  }

  let { href, onclick, children }: Props = $props();

  const baseClass =
    'block bg-surface-card text-ink rounded-xl border border-hairline p-6 lift';
</script>

{#if href}
  <a {href} class="{baseClass} no-underline">
    {@render children()}
  </a>
{:else if onclick}
  <button type="button" {onclick} class="{baseClass} text-left w-full cursor-pointer">
    {@render children()}
  </button>
{:else}
  <div class="{baseClass}">
    {@render children()}
  </div>
{/if}