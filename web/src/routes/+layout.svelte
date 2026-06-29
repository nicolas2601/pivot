<script lang="ts">
  import '../app.css';
  import { QueryClientProvider } from '@tanstack/svelte-query';
  import { queryClient } from '$lib/stores/query-client';
  import { installAuthInterceptor } from '$lib/utils/auth-interceptor';
  import Toast from '$lib/components/Toast.svelte';
  import PWAInfo from '$lib/components/PWAInfo.svelte';

  let { children } = $props();

  installAuthInterceptor();
</script>

<PWAInfo />
<QueryClientProvider client={queryClient}>
  {@render children()}
  <Toast />
</QueryClientProvider>