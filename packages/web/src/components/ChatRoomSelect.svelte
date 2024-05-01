<script>
  import { LockClosed, PaperAirplane } from "svelte-heros-v2";

  export let rooms;
  export let baseUrl;
  export let currentRoom;
</script>

<div class="flex gap-10 items-center">
  <h2 class="text-2xl font-medium">Chat</h2>
  <ul class="mt-2 menu menu-horizontal menu-sm bg-base-200 rounded-box">
    {#each rooms as r}
      <li>
        <a
          href="{baseUrl}/chat{r.slug !== 'general' ? '?room=' + r.slug : ''}"
          class:active={currentRoom === r.slug}
        >
          #{r.slug}
          {#if r.private}
            <LockClosed size="14" class="mr-0 outline-none" tabindex="-1" />
          {/if}
        </a>
      </li>
    {/each}
  </ul>
</div>
