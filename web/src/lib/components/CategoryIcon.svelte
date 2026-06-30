<script lang="ts">
  /**
   * CategoryIcon — mapea el string `icon` que la API guarda por categoría
   * (heart, utensils, car, shopping-bag, etc.) al NavIcon correspondiente.
   * Fallback: círculo con la inicial.
   *
   * Mantener un solo set de iconos en la app evita el clásico bug de
   * iconos faltantes que aparecen como rectángulos vacíos.
   */
  import NavIcon from './NavIcon.svelte';

  type IconName =
    | 'utensils' | 'car' | 'shopping-bag' | 'home-modern' | 'film'
    | 'heart' | 'lightning' | 'banknotes' | 'piggy-bank' | 'gift'
    | 'stethoscope' | 'book' | 'tools' | 'plane-departure' | 'receipt'
    | 'cash' | 'briefcase' | 'trending-up' | 'edit' | 'wallet'
    | 'tag' | 'home' | 'sparkles' | 'gift' | 'bell';

  let {
    icon = '',
    name = '',
    color = '#e7e5e4',
    class: klass = 'w-5 h-5',
    accent = false
  }: {
    icon?: string | null;
    name?: string;
    color?: string | null;
    class?: string;
    accent?: boolean;
  } = $props();

  // Whitelist — un icono desconocido cae a wallet siempre
  const KNOWN: ReadonlyArray<IconName> = [
    'utensils', 'car', 'shopping-bag', 'home-modern', 'film',
    'heart', 'lightning', 'banknotes', 'piggy-bank', 'gift',
    'stethoscope', 'book', 'tools', 'plane-departure', 'receipt',
    'cash', 'briefcase', 'trending-up', 'edit', 'wallet',
    'tag', 'home', 'sparkles', 'bell'
  ];

  const iconName: IconName = $derived(
    KNOWN.includes(icon as IconName) ? (icon as IconName) : 'tag'
  );

  const initial = $derived(name?.charAt(0).toUpperCase() || '?');
</script>

<span
  class="inline-flex items-center justify-center rounded-xl shrink-0
         {accent ? 'ring-1 ring-ink/5 shadow-sm' : ''}"
  style="background: {color}; width: 2.5rem; height: 2.5rem;"
  aria-hidden="true"
>
  <NavIcon icon={iconName} class="w-4 h-4 {accent ? 'text-ink/90' : 'text-ink'}" />
</span>
