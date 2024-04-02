<script>
    import { format } from 'date-fns'
    import EventBox from './EventBox.svelte';
    export let events;

    for (const e of events) {
        e.date = format(new Date(e.dateStart), "yyyy-MM-dd")
    }

    const days = events.map(e => e.date)
</script>

<ul class="">
    {#each days as day}
        <li class="flex gap-4 mb-8">
            <div class="w-32 p-2">
                <div class="font-semibold">{format(new Date(day), 'MMM d')}</div>
                <div class="text-neutral-content">{format(new Date(day), 'EEEE')}</div>
            </div>
            <div></div>
            <div class="timeline-end grow">
                {#each events.filter(e => e.date === day) as item}
                    <EventBox {item} />
                {/each}
            </div>
        </li>
    {/each}
</ul>