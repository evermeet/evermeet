<script>
  import EventBox from "../components/EventBox.svelte";
  import CalendarBox from "../components/CalendarBox.svelte";
  import { config } from "$lib/stores";
  import { getContext } from "svelte";
  import { t } from "$lib/i18n";

  export let data;
</script>

<svelte:head>
  <title>{$t`Explore`} | {$config.sitename || $config.domain}</title>
</svelte:head>

<div class="page-wide">
  <h1 class="heading1">{$t`Explore`}</h1>

  {#if data.featured && data.featured.length > 0}
    <h2 class="heading2">{$t`Featured Calendars`}</h2>

    <div class="mt-6 mb-10 grid grid-cols-3 gap-3">
      {#each data.featured as item}
        <CalendarBox {item} preview />
      {/each}
    </div>
  {/if}

  <h2 class="heading2">{$t`Local Calendars`}</h2>

  <div class="mt-6 grid grid-cols-3 gap-3">
    {#each data.calendars as item}
      <CalendarBox {item} preview />
    {/each}
  </div>
</div>
