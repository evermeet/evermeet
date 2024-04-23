<script>
    import EventDetail from '../../components/EventDetail.svelte';
    import CalendarBox from '../../components/CalendarBox.svelte';
    import { config } from '$lib/stores';

    export let data;
    const { subscribed, owned } = data.calendars;

</script>

<svelte:head>
    <title>Calendars | {$config.sitename || $config.domain}</title> 
</svelte:head>

<div class="page-wide">
    <h1 class="heading1">Calendars</h1>

    <div class="flex items-center">
        <h2 class="heading2 grow">My Calendars</h2>
        <div>
            <a href="/create-calendar"><button class="btn btn-sm btn-neutral">+ Create Calendar</button></a>
        </div>
    </div>

    <div class="mt-6 grid grid-cols-3 gap-3">
        {#each owned as c}
            <CalendarBox item={c} />
        {/each}
    </div>

    <div class="mt-8 mb-6 w-full h-1 border border-neutral border-l-0 border-r-0 border-b-0">  
    </div>

    <h2 class="heading2 grow">Subscribed Calendars</h2>

    <div class="mt-6">
        {#each subscribed as c}
            <div class="mb-3 itembox flex gap-8">
                <div class="w-[12rem]">
                    <div class="w-12 h-12 mb-2">
                        <img
                            class="{c.personal ? "rounded-full" : "rounded-lg"}"
                            src={c.img} />
                    </div>
                    <div class="text-lg font-medium">{c.name}</div>
                    {#if c._remote}
                        <div class="badge badge-neutral font-mono text-xs my-2">{c._remote}</div>
                    {/if}
                    <a href="{c.baseUrl}" class="btn btn-sm mt-4 btn-neutral">View calendar →</a>
                </div>
                <div class="">
                    <div class="text-neutral-content text-sm">Upcoming Events</div>
                    {#each c.events.slice(0,2) as e}
                        <div class="mt-4">
                            <div class="font-medium hover:underline"><a href={e.baseUrl}>{e.name}</a></div>
                            <div class="text-sm mt-1 text-neutral-content">{e.dateStart}</div>
                        </div>
                    {/each}
                </div>
            </div>
        {/each}
    </div>
</div>