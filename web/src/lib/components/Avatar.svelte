<script lang="ts">
  /**
   * Avatar — initials en círculo para miembros de un grupo de viaje o personas.
   * Estilo: surface-strong (f0efed), texto ink, radius full.
   * Tamaño: configurable vía size prop.
   */

  interface Props {
    name: string;
    size?: 'sm' | 'md' | 'lg';
  }

  let { name, size = 'md' }: Props = $props();

  const initials = $derived(
    name
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part.charAt(0).toUpperCase())
      .join('') || '?'
  );

  const sizeClass = $derived(
    {
      sm: 'w-8 h-8 text-xs',
      md: 'w-10 h-10 text-sm',
      lg: 'w-12 h-12 text-base'
    }[size]
  );
</script>

<span
  class="inline-flex items-center justify-center rounded-full bg-surface-strong text-ink font-medium {sizeClass}"
  aria-label={name}
>
  {initials}
</span>