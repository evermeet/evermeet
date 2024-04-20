<script>
    import { Check, ChevronDown, CheckCircle, GlobeAlt, CubeTransparent, LockClosed, Cube, Server, Cloud } from 'svelte-heros-v2';
    import { createSelect, melt } from '@melt-ui/svelte';
    import { fade } from 'svelte/transition';
  
    export let bind;
    export let instances;
  
    const {
      elements: { trigger, menu, option, group, groupLabel, label },
      states: { selectedLabel, selected, select, open },
      helpers: { isSelected },
    } = createSelect({
      forceVisible: true,
      positioning: {
        placement: 'bottom',
        fitViewport: true,
      },
      defaultSelected: { value: $bind, label: $bind },
    });

    bind.subscribe((x) => {
        if (x !== $selected.value) {
            selected.set({ value: x, label: x })
        }
    })
  
    selected.subscribe((x) => {
      bind.set(x.value);
    })
  
  </script>
  
  <div class="dropdown">
    <button
      class="btn btn-sm no-animation justify-between {$open ? 'bg-base-300' : ''} min-w-[9rem]"
      use:melt={$trigger}
      aria-label="Visibility"
    >
      <div class="text-neutral-content flex gap-1.5 items-center text-xs">
          <svelte:component this={instances.find(x => x.domain === $selectedLabel).local ? Server : GlobeAlt} class="w-4" tabindex="-1" />
          {$selectedLabel || 'Select visibility'}
      </div>
      <ChevronDown tabIndex="-1" class="size-5 opacity-30" />
    </button>
  
    {#if $open}
      <ul
        class="dropdown-content popup-menu bg-base-200 min-w-[16rem]"
        use:melt={$menu}
        transition:fade={{ duration: 150 }}
      >
        {#each instances as i}
          <li use:melt={$option({ value: i.domain, label: i.domain })}>
            <div class="item">
              <div class="flex gap-4 items-center">
                <div>
                  <svelte:component this={i.local ? Server : GlobeAlt} class="w-5" tabindex="-1" />
                </div>
                <div class="grow">
                  <div>{i.domain}</div>
                  {#if i.local}
                    <div class="text-neutral-content text-sm">Local instance</div>
                  {/if}
                </div>
                <div class="">
                  <CheckCircle tabindex="-1" variation="solid" class="size-4 {$isSelected(i.domain) ? 'block' : 'hidden'}" />
                </div>
              </div>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  
  </div>