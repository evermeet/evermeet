<script>
  import { getContext } from "svelte";
  import { format } from "$lib/date";
  import EventBox from "./EventBox.svelte";

  const { events } = $props();
  const { dateLocale: locale, timezone, lang } = getContext("locale");

  function enhanced(arr) {
    for (const e of arr) {
      e.date = format(new Date(e.dateStart), "yyyy-MM-dd");
    }
    return arr;
  }

  let days = enhanced(events).map((e) => e.date);
</script>

{#if events.length > 0}
  <ul class="timeline timeline-vertical timeline-snap-icon">
    {#each days as day}
      <li class="">
        <hr />
        <div class="timeline-start p-2 items-start flex w-full h-full">
          <div>
            <div class="font-semibold">
              {format(new Date(day), "MMM d", { locale })}
            </div>
            <div class="text-base-content/75">
              {format(new Date(day), "EEEE", { locale })}
            </div>
          </div>
        </div>
        <div class="timeline-middle">
          <div class="w-2.5 h-2.5 text-neutral">
            <svg
              height="100%"
              width="100%"
              viewBox="0 0 100 100"
              xmlns="http://www.w3.org/2000/svg"
              ><circle r="45" cx="50" cy="50" fill="currentColor" /></svg
            >
          </div>
        </div>
        <div class="timeline-end ml-4 w-full">
          {#each events.filter((e) => e.date === day) as item}
            <EventBox {item} />
          {/each}
        </div>
        <hr />
      </li>
    {/each}
  </ul>
{:else}
  <div class="text-xl">No events :(</div>
{/if}

<style>
  .timeline-vertical:where(.timeline-snap-icon) > li {
    --timeline-col-start: 8rem;
    --timeline-row-start: 1rem;
  }
  .timeline-vertical li:first-child hr:first-child {
    display: none;
  }
</style>
