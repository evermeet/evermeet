<script>
    import { format } from 'date-fns';
    import { createDatePicker, melt } from '@melt-ui/svelte';
    import { ChevronRight, ChevronLeft, Calendar } from 'svelte-heros-v2';
    import { CalendarDate } from '@internationalized/date';

    import { fade } from 'svelte/transition';

    export let bind;

    const {
      elements: {
        calendar,
        cell,
        content,
        field,
        grid,
        heading,
        label,
        nextButton,
        prevButton,
        segment,
        trigger,
      },
      states: { months, headingValue, weekdays, segmentContents, open },
      helpers: { isDateDisabled, isDateUnavailable },
      options: { locale },
    } = createDatePicker({
      locale: 'en-GB',
      forceVisible: true,
      defaultValue: new CalendarDate(2024, 1, 11),
      preventDeselect: true,
    });
</script>


<div class="h-8 -mt-1 -mb-1 bg-base-300 rounded flex items-center w-[10rem]">
  <div use:melt={$field} class="">
    {#key $locale}
        {#each $segmentContents as seg}
          <div use:melt={$segment(seg.part)}>
            {seg.value}
          </div>
        {/each} 
      {/key}
      <div class="grow w-full justify-end">
        <button use:melt={$trigger}>
          <Calendar size={16} />
        </button>
      </div>
  </div>

  {#if $open}
    <div
      transition:fade={{ duration: 100 }}
      use:melt={$content}
      class=""
    >
      <div use:melt={$calendar}>
        <header>
          <button use:melt={$prevButton}>
            <ChevronLeft size={24} />
          </button>
          <div use:melt={$heading}>
            {$headingValue}
          </div>
          <button use:melt={$nextButton}>
            <ChevronRight size={24} />
          </button>
        </header>
        <div>
          {#each $months as month}
            <table use:melt={$grid}>
              <thead aria-hidden="true">
                <tr>
                  {#each $weekdays as day}
                    <th>
                      <div>
                        {day}
                      </div>
                    </th>
                  {/each}
                </tr>
              </thead>
              <tbody>
                {#each month.weeks as weekDates}
                  <tr>
                    {#each weekDates as date}
                      <td
                        role="gridcell"
                        aria-disabled={$isDateDisabled(date) ||
                          $isDateUnavailable(date)}
                      >
                        <div use:melt={$cell(date, month.value)}>
                          {date.day}
                        </div>
                      </td>
                    {/each}
                  </tr>
                {/each}
              </tbody>
            </table>
          {/each}
        </div>
      </div>
    </div>
  {/if}
</div>

<style lang="postcss">

  [data-melt-datefield-field] div:last-of-type {
    @apply ml-4 flex w-full;
  }

  [data-melt-popover-content] {
    @apply z-10 min-w-[320px] rounded-lg bg-base-300 shadow-sm;
  }

  [data-melt-popover-trigger] {
    @apply rounded-md bg-base-300 p-1 transition-all;
  }

  [data-melt-datefield-label] {
    @apply select-none font-medium;
  }

  [data-melt-datefield-label][data-invalid] {
    @apply text-red-500;
  }

  [data-melt-datefield-field] {
    @apply mt-0.5 flex w-full items-center rounded-lg bg-base-300 p-1.5;
  }

  [data-melt-datefield-field][data-invalid] {
    @apply border-red-400;
  }

  [data-melt-datefield-segment][data-invalid] {
    @apply text-red-500;
  }

  [data-melt-datefield-segment]:not([data-segment='literal']) {
    @apply px-0.5;
  }

  [data-melt-datefield-validation] {
    @apply self-start text-red-500;
  }

  [data-melt-calendar] {
    @apply w-full rounded-lg p-1.5 shadow-lg;
  }

  header {
    @apply flex items-center justify-between pb-2;
  }

  header + div {
    @apply flex items-center gap-6;
  }

  [data-melt-calendar-prevbutton] {
    @apply rounded-lg p-1 transition-all hover:bg-neutral;
  }

  [data-melt-calendar-nextbutton] {
    @apply rounded-lg p-1 transition-all hover:bg-neutral;
  }

  [data-melt-calendar-prevbutton][data-disabled] {
    @apply pointer-events-none rounded-lg p-1 opacity-40;
  }

  [data-melt-calendar-nextbutton][data-disabled] {
    @apply pointer-events-none rounded-lg p-1 opacity-40;
  }

  [data-melt-calendar-heading] {
    @apply font-semibold;
  }

  th {
    @apply text-sm font-semibold;

    & div {
      @apply flex h-6 w-6 items-center justify-center p-4;
    }
  }

  [data-melt-calendar-grid] {
    @apply w-full;
  }

  [data-melt-calendar-cell] {
    @apply flex h-6 w-6 cursor-pointer select-none items-center justify-center rounded-lg p-4 hover:bg-neutral focus:ring focus:ring-neutral/75;
  }

  [data-melt-calendar-cell][data-disabled] {
    @apply pointer-events-none opacity-40;
  }
  [data-melt-calendar-cell][data-unavailable] {
    @apply pointer-events-none text-red-400 line-through;
  }

  [data-melt-calendar-cell][data-selected] {
    @apply bg-primary text-base-100;
  }

  [data-melt-calendar-cell][data-outside-visible-months] {
    @apply pointer-events-none cursor-default opacity-40 hover:bg-transparent;
  }

  [data-melt-calendar-cell][data-outside-month] {
    @apply pointer-events-none cursor-default opacity-25 hover:bg-transparent;
  }
</style>