<script lang="ts">
  /**
   * BottomNav — navegación principal responsive.
   * Mobile (<768px): bottom tab bar 2x4 grid con 5 items principales + 3 secundarios
   *   (Metas, Recurrentes, Presupuestos, Viajes — distribuidos en 2 filas).
   * Desktop (≥768px): top nav horizontal con todos los items separados por un divisor.
   * Estilo: canvas background, ink text, surface-strong para activo.
   * Accesibilidad: role="navigation", aria-current="page" en el activo.
   */

  import { page } from '$app/stores';
  import NavIcon from './NavIcon.svelte';

  interface NavItem {
    href: string;
    label: string;
    icon: 'home' | 'wallet' | 'tag' | 'list' | 'plane' | 'target' | 'repeat';
  }

  const items: NavItem[] = [
    { href: '/', label: 'Dashboard', icon: 'home' },
    { href: '/accounts', label: 'Cuentas', icon: 'wallet' },
    { href: '/categories', label: 'Categorías', icon: 'tag' },
    { href: '/transactions', label: 'Movimientos', icon: 'list' },
    { href: '/travel', label: 'Viajes', icon: 'plane' }
  ];

  // Items secundarios: features que no son navegación diaria pero necesitan
  // un lugar accesible. En mobile viven en una segunda fila del bottom bar.
  const secondaryItems: NavItem[] = [
    { href: '/goals', label: 'Metas', icon: 'target' },
    { href: '/recurring', label: 'Recurrentes', icon: 'repeat' },
    { href: '/budgets', label: 'Presupuestos', icon: 'list' }
  ];

  function isActive(href: string, pathname: string): boolean {
    if (href === '/') return pathname === '/';
    return pathname === href || pathname.startsWith(`${href}/`);
  }
</script>

<nav aria-label="Navegación principal">
  <!-- Desktop: top nav horizontal con dos secciones separadas por divisor -->
  <div class="hidden md:flex items-center gap-1 border-b border-hairline bg-canvas px-6 h-16">
    {#each items as item (item.href)}
      <a
        href={item.href}
        aria-current={isActive(item.href, $page.url.pathname) ? 'page' : undefined}
        class="px-3 py-2 text-sm font-medium rounded-md transition-colors {isActive(item.href, $page.url.pathname) ? 'text-ink bg-surface-strong' : 'text-body hover:text-ink hover:bg-surface-strong'}"
      >
        {item.label}
      </a>
    {/each}
    <div class="mx-2 h-5 w-px bg-hairline" aria-hidden="true"></div>
    {#each secondaryItems as item (item.href)}
      <a
        href={item.href}
        aria-current={isActive(item.href, $page.url.pathname) ? 'page' : undefined}
        class="px-3 py-2 text-sm font-medium rounded-md transition-colors {isActive(item.href, $page.url.pathname) ? 'text-ink bg-surface-strong' : 'text-body hover:text-ink hover:bg-surface-strong'}"
      >
        {item.label}
      </a>
    {/each}
  </div>

  <!-- Mobile: bottom tab bar 2x4 grid. 4 principales + Viajes fila 1;
       Metas + Recurrentes + Presupuestos + (slot vacío) fila 2. -->
  <div
    class="md:hidden fixed bottom-0 left-0 right-0 z-40 bg-surface-card border-t border-hairline pb-[env(safe-area-inset-bottom)]"
  >
    <div class="grid grid-cols-4 max-w-md mx-auto">
      {#each [...items.slice(0, 4)] as item (item.href)}
        <a
          href={item.href}
          aria-current={isActive(item.href, $page.url.pathname) ? 'page' : undefined}
          aria-label={item.label}
          class="flex flex-col items-center justify-center gap-1 py-2 min-h-[56px] text-xs transition-colors {isActive(item.href, $page.url.pathname) ? 'text-ink' : 'text-muted'}"
        >
          <NavIcon icon={item.icon} />
          <span class="text-[10px] font-medium leading-none">{item.label}</span>
        </a>
      {/each}
    </div>
    <div class="grid grid-cols-4 max-w-md mx-auto border-t border-hairline">
      {#each [...items.slice(4), ...secondaryItems] as item (item.href)}
        <a
          href={item.href}
          aria-current={isActive(item.href, $page.url.pathname) ? 'page' : undefined}
          aria-label={item.label}
          class="flex flex-col items-center justify-center gap-1 py-2 min-h-[56px] text-xs transition-colors {isActive(item.href, $page.url.pathname) ? 'text-ink' : 'text-muted'}"
        >
          <NavIcon icon={item.icon} />
          <span class="text-[10px] font-medium leading-none">{item.label}</span>
        </a>
      {/each}
    </div>
  </div>
</nav>