<script>
  import { getContext } from "svelte";
  import { format, isSameYear } from "$lib/date";
  import EventBox from "./EventBox.svelte";

  const { events, rowClass } = $props();
  const { dateLocale: locale, timezone, lang } = getContext("locale");

  function getDays(arr) {
    const out = [];
    for (const e of arr) {
      const dt = format(new Date(e.dateStart), "yyyy-MM-dd");
      let day = out.find((d) => d.day === dt);
      if (!day) {
        day = { day: dt, items: [] };
        out.push(day);
      }
      day.items.push(e);
    }
    return out;
  }

  let days = $derived(getDays(events));
</script>

{#if events.length > 0}
  <ul class="timeline timeline-vertical timeline-snap-icon">
    {#each days as day}
      <li class={rowClass}>
        <hr />
        <div class="timeline-start p-2 items-start flex w-full h-full">
          <div>
            <div class="font-semibold">
              {format(
                new Date(day.day),
                "MMM d" +
                  (!isSameYear(new Date(day.day), new Date()) ? ", y" : ""),
                { locale },
              )}
            </div>
            <div class="text-base-content/75">
              {format(new Date(day.day), "EEEE", { locale })}
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
          {#each day.items as item}
            <EventBox {item} id={item.id} />
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
