<script>
  import { Check, ChevronDown, CheckCircle } from "svelte-heros-v2";
  import { createSelect, melt } from "@melt-ui/svelte";
  import { fade } from "svelte/transition";

  export let items;
  export let bind;

  const {
    elements: { trigger, menu, option, group, groupLabel, label },
    states: { selectedLabel, selected, open },
    helpers: { isSelected },
  } = createSelect({
    positioning: {
      placement: "bottom-start",
      fitViewport: true,
    },
    selected: bind,
  });
</script>

<div class="dropdown w-full">
  <button
    class="btn btn-sm no-animation justify-between {$open ? 'bg-base-300' : ''}"
    use:melt={$trigger}
    aria-label="Calendar"
  >
    <div class="text-base-content/75">
      {#if $selectedLabel}
        {#each [items.find((x) => x.id === $selectedLabel)] as item}
          <div class="flex gap-2 items-center">
            <img src={item.img} class="w-5 h-5 rounded" />
            {item.name}
          </div>
        {/each}
      {:else}
        Select a calendar
      {/if}
    </div>
    <ChevronDown class="size-5 opacity-30" />
  </button>
  {#if $open}
    <ul
      class="dropdown-content popup-menu"
      use:melt={$menu}
      transition:fade={{ duration: 150 }}
    >
      <li class="text-xs p-2 text-base-content/75">
        Choose the calendar of the event:
      </li>
      {#each items as item}
        <li use:melt={$option({ value: item.id, label: item.id })}>
          <div class="flex gap-2 justify-between items-center">
            <div>
              <a class="flex gap-2 text-base justify-left"
                ><img src={item.img} class="w-5 h-5 rounded" />{item.name}</a
              >
            </div>
            <CheckCircle
              variation="solid"
              class="size-4 mr-2 {$isSelected(item.id) ? 'block' : 'hidden'}"
            />
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</div>
