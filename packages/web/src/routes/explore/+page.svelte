<script>
    import EventBox from '../../components/EventBox.svelte';
    import CalendarBox from '../../components/CalendarBox.svelte';
    import { config, user } from '$lib/stores';

    export let data;

</script>

<svelte:head>
    <title>Explore | {$config.sitename || $config.domain}</title> 
</svelte:head>

<div class="page-wide">
    <h1 class="heading1">Explore</h1>

    {#if data.featured && data.featured.length > 0}
        <h2 class="heading2">Featured Calendars</h2>

        <div class="mt-6 mb-10 grid grid-cols-3 gap-3">
            {#each data.featured as item}
                <CalendarBox {item} preview />
            {/each}
        </div>
    {/if}

    <h2 class="heading2">Local Calendars</h2>

    <div class="mt-6 grid grid-cols-3 gap-3">
        {#each data.calendars as item}
            <CalendarBox {item} preview />
        {/each}
    </div>

</div>