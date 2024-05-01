<script>
  import { apiCall } from "../lib/api.js";
  import { onMount } from "svelte";

  export let item;

  let logs = null;

  onMount(async () => {
    logs = await apiCall(fetch, `event/${item.id}/audit`);
  });
</script>

<h3 id="hosts" class="text-2xl mb-8 font-semibold">Latest logs</h3>

{#if logs}
  <table class="table bg-base-200">
    <thead>
      <tr><th>Date</th><th>Action</th><th>Author</th></tr>
    </thead>
    <tbody>
      {#each logs as log}
        <tr
          ><td>{log.time}</td><td>{log.type}</td><td
            >{log.author || "System"}</td
          ></tr
        >
      {/each}
    </tbody>
  </table>
{:else}{/if}

<h3 id="hosts" class="text-2xl my-8 font-semibold">Source</h3>

<pre class="overflow-scroll itembox"><code
    class="mono pre text-xs text-base-content/75"
    >{JSON.stringify(item, null, 2)}</code
  ></pre>
