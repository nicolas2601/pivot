<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    variant?: 'primary' | 'outline' | 'tertiary' | 'danger';
    type?: 'button' | 'submit';
    disabled?: boolean;
    href?: string;
    onclick?: (e: MouseEvent) => void;
    class?: string;
    children: Snippet;
  }

  let {
    variant = 'primary',
    type = 'button',
    disabled = false,
    href,
    onclick,
    class: className = '',
    children
  }: Props = $props();

  const baseClass =
    'inline-flex items-center justify-center font-medium transition-colors duration-150 h-10 px-5 rounded-pill text-button focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-ink disabled:opacity-50 disabled:cursor-not-allowed';

  const variantClass = $derived(
    {
      primary: 'bg-primary text-on-primary hover:bg-primary-active',
      outline: 'bg-transparent text-ink border border-hairline-strong hover:bg-surface-strong',
      tertiary: 'bg-transparent text-ink hover:underline',
      danger: 'bg-semantic-error text-on-primary hover:opacity-90'
    }[variant]
  );
</script>

{#if href}
  <a {href} class="{baseClass} {variantClass} {className}" aria-disabled={disabled}>
    {@render children()}
  </a>
{:else}
  <button {type} {disabled} {onclick} class="{baseClass} {variantClass} {className}">
    {@render children()}
  </button>
{/if}