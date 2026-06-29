<script lang="ts">
  import { pwaInfo } from 'virtual:pwa-info';
</script>

<svelte:head>
  {#if pwaInfo}
    <link rel="manifest" href={pwaInfo.manifestPath} />
    {#if pwaInfo.registerType === 'autoUpdate'}
      <script>
        // Auto-update SW without prompting the user — appropriate for a
        // single-user personal finance app where the user always wants
        // the freshest fix. Falls back to prompt mode if type is 'prompt'.
        if ('serviceWorker' in navigator) {
          navigator.serviceWorker.getRegistration().then((reg) => {
            reg?.update();
          });
        }
      </script>
    {/if}
  {/if}
</svelte:head>