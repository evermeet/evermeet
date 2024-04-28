<script>
    import { page } from '$app/stores'

    import EventDetail from './EventDetail.svelte';
    import CalendarDetail from "./CalendarDetail.svelte";

    export let data;
    $: params = paramsToObject($page.url.searchParams)
    //$: expand = Boolean(['', true].includes($page.url.searchParams.get('expand')))

    function paramsToObject(entries) {
        const result = {}
        for(const [key, value] of entries) { // each 'entry' is a [key, value] tupple
            result[key] = value === "" ? true : value;
        }
        return result;
    }
</script>

{#if data.result.type === 'event'}
    <div class="page-wide">
        <EventDetail item={data.result.item} id={data.result.item.id} selectedTab={data.selectedTab} {params} />
    </div>
{/if}

{#if data.result.type === 'calendar'}
    <CalendarDetail item={data.result.item} id={data.result.item.id} selectedTab={data.selectedTab} {params} />
{/if}