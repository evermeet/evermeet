<script>
  import { Check, ChevronDown, CheckCircle, GlobeAlt, CubeTransparent, LockClosed } from 'svelte-heros-v2';
  import { createSelect, melt } from '@melt-ui/svelte';
  import { fade } from 'svelte/transition';

  export let bind;

  const states = {
    public: { title: 'Public', text: 'Visible to everyone, show on your calendar.', ico: GlobeAlt },
    unlisted: { title: 'Unlisted', text: 'Only people with the link can register.', ico: CubeTransparent },
    private: { title: 'Private', text: 'Only you and calendar managers will see this event.', ico: LockClosed },
  }
  const {
    elements: { trigger, menu, option, group, groupLabel, label },
    states: { selectedLabel, selected, open },
    helpers: { isSelected },
  } = createSelect({
    forceVisible: true,
    positioning: {
      placement: 'bottom-end',
      fitViewport: true,
    },
    defaultSelected: { value: $bind, label: states[$bind].title },
  });

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
    <div class="text-neutral-content flex gap-1.5 items-center">
        <svelte:component this={states[$selected.value].ico} class="w-5"/>
        {$selectedLabel || 'Select visibility'}
    </div>
    <ChevronDown class="size-5 opacity-30" />
  </button>

  {#if $open}
    <ul
      class="dropdown-content popup-menu w-[18.5rem]"
      use:melt={$menu}
      transition:fade={{ duration: 150 }}
    >
      {#each Object.keys(states) as state}
        <li use:melt={$option({ value: state, label: states[state].title })}>
          <div class="item">
            <div class="flex gap-4 items-center">
              <div>
                <svelte:component this={states[state].ico} class="w-5" />
              </div>
              <div>
                <div>{states[state].title}</div>
                <div class="text-neutral-content text-sm">{states[state].text}</div>
              </div>
              <div>
                <CheckCircle variation="solid" class="size-4 {$isSelected(state) ? 'block' : 'hidden'}" />
              </div>
            </div>
          </div>
        </li>
      {/each}
    </ul>
  {/if}

</div>