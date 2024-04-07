<script>
    import EventBox from '../../components/EventBox.svelte';
    import CalendarBox from '../../components/CalendarBox.svelte';
    import { config } from '$lib/config';

    export let data;

    const days = [
        "2022-04-01",
        "2022-04-02"
    ]

</script>

<svelte:head>
    <title>Explore | {config.sitename || config.domain}</title> 
</svelte:head>

<div class="page-wide">
    <h1 class="heading1">Explore</h1>

    <h2 class="heading2">Featured Calendars</h2>

    <div class="mt-6 grid grid-cols-3 gap-3">
        {#each data.calendars as item}
            <CalendarBox {item} preview />
        {/each}
    </div>

</div>