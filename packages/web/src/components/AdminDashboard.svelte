<script>
  import { apiCall } from "../lib/api.js";
  import { onMount } from "svelte";

  let base = null;

  onMount(async () => {
    base = await apiCall(fetch, `admin`);
  });
</script>

<h2 class="text-xl font-semibold mb-4">Collections</h2>

<div class="grid grid-cols-6 gap-2">
  {#if base}
    {#each base.collections as c}
      <div class="itembox">
        <div class="text-2xl font-bold font-mono">{c.size}</div>
        <div class="text-base-content/75">{c.name}</div>
      </div>
    {/each}
  {:else}
    <div class="skeleton h-24"></div>
    <div class="skeleton h-24"></div>
    <div class="skeleton h-24"></div>
  {/if}
</div>
