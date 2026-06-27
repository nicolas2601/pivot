<script lang="ts">
  import '../app.css';
  import { QueryClientProvider } from '@tanstack/svelte-query';
  import { queryClient } from '$lib/stores/query-client';
  import { installAuthInterceptor } from '$lib/utils/auth-interceptor';
  import { onMount } from 'svelte';

  let { children } = $props();

  onMount(() => {
    installAuthInterceptor();
  });
</script>

<QueryClientProvider client={queryClient}>
  {@render children()}
</QueryClientProvider>