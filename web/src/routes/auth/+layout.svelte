<script lang="ts">
  import '../../app.css';
  let { children } = $props();
</script>

<main class="min-h-screen flex items-center justify-center bg-slate-50 p-4">
  <div class="w-full max-w-md">
    {@render children()}
  </div>
</main>