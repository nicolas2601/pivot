import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { SvelteKitPWA } from '@vite-pwa/sveltekit';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [
    tailwindcss(),
    sveltekit(),
    SvelteKitPWA({
      strategies: 'generateSW',
      registerType: 'autoUpdate',
      injectRegister: 'auto',
      manifest: false, // we serve the static one
      workbox: {
        // Pre-cache app shell + dashboard + reports for offline shell.
        globPatterns: ['client/**/*.{js,css,ico,png,svg,webmanifest,woff2}'],
        navigateFallback: '/',
        navigateFallbackDenylist: [/^\/api/, /^\/auth/],
        // Runtime caching for font/CSS to make repeat visits instant.
        runtimeCaching: [
          {
            urlPattern: ({ url }) => url.origin === self.location.origin && url.pathname.startsWith('/_app/'),
            handler: 'StaleWhileRevalidate',
            options: { cacheName: 'app-shell' }
          },
          {
            urlPattern: ({ request }) => request.destination === 'image',
            handler: 'CacheFirst',
            options: { cacheName: 'images', expiration: { maxEntries: 50 } }
          }
        ]
      },
      devOptions: {
        enabled: true,
        type: 'module',
        navigateFallback: '/'
      }
    })
  ],
  server: { port: 5173, strictPort: false },
  test: {
    include: ['src/**/*.{test,spec}.{js,ts}'],
    environment: 'jsdom'
  }
});