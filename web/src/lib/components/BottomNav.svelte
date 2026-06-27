<script lang="ts">
  /**
   * BottomNav — navegación principal responsive.
   * Mobile (<768px): bottom tab bar con 5 items (Dashboard, Cuentas, Categorías, Movimientos, Viajes).
   * Desktop (≥768px): top nav horizontal con los mismos items + label del usuario a la derecha.
   * Estilo: canvas background, ink text, surface-strong para activo.
   * Accesibilidad: role="navigation", aria-current="page" en el activo.
   */

  import { page } from '$app/stores';

  interface NavItem {
    href: string;
    label: string;
    // Iconos SVG inline — sin dependencias externas.
    icon: 'home' | 'wallet' | 'tag' | 'list' | 'plane';
  }

  const items: NavItem[] = [
    { href: '/', label: 'Dashboard', icon: 'home' },
    { href: '/accounts', label: 'Cuentas', icon: 'wallet' },
    { href: '/categories', label: 'Categorías', icon: 'tag' },
    { href: '/transactions', label: 'Movimientos', icon: 'list' },
    { href: '/travel', label: 'Viajes', icon: 'plane' }
  ];

  function isActive(href: string, pathname: string): boolean {
    if (href === '/') return pathname === '/';
    return pathname === href || pathname.startsWith(`${href}/`);
  }
</script>

<nav aria-label="Navegación principal">
  <!-- Desktop: top nav horizontal -->
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
  </div>

  <!-- Mobile: bottom tab bar -->
  <div
    class="md:hidden fixed bottom-0 left-0 right-0 z-40 bg-surface-card border-t border-hairline pb-[env(safe-area-inset-bottom)]"
  >
    <div class="grid grid-cols-5 max-w-md mx-auto">
      {#each items as item (item.href)}
        <a
          href={item.href}
          aria-current={isActive(item.href, $page.url.pathname) ? 'page' : undefined}
          aria-label={item.label}
          class="flex flex-col items-center justify-center gap-1 py-2 min-h-[56px] text-xs transition-colors {isActive(item.href, $page.url.pathname) ? 'text-ink' : 'text-muted'}"
        >
          {#if item.icon === 'home'}
            <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true">
              <path d="M3 11l9-8 9 8v9a2 2 0 0 1-2 2h-4v-7h-6v7H5a2 2 0 0 1-2-2v-9z" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          {:else if item.icon === 'wallet'}
            <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true">
              <path d="M3 7a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7z" stroke-linecap="round" stroke-linejoin="round" />
              <path d="M3 8h18" stroke-linecap="round" /><circle cx="17" cy="13" r="1.2" fill="currentColor" />
            </svg>
          {:else if item.icon === 'tag'}
            <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true">
              <path d="M20.59 13.41L13.42 20.58a2 2 0 0 1-2.83 0L3 13V3h10l7.59 7.59a2 2 0 0 1 0 2.82z" stroke-linecap="round" stroke-linejoin="round" />
              <circle cx="7.5" cy="7.5" r="1.2" fill="currentColor" />
            </svg>
          {:else if item.icon === 'list'}
            <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true">
              <path d="M8 6h13M8 12h13M8 18h13" stroke-linecap="round" /><circle cx="4" cy="6" r="1" fill="currentColor" /><circle cx="4" cy="12" r="1" fill="currentColor" /><circle cx="4" cy="18" r="1" fill="currentColor" />
            </svg>
          {:else if item.icon === 'plane'}
            <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true">
              <path d="M2 16l20-7L4 4l3 8-2 4z" stroke-linecap="round" stroke-linejoin="round" /><path d="M9 14l5 7" stroke-linecap="round" />
            </svg>
          {/if}
          <span class="text-[10px] font-medium leading-none">{item.label}</span>
        </a>
      {/each}
    </div>
  </div>
</nav>