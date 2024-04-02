<script>
    import EventDetail from '../../components/EventDetail.svelte';
    import CalendarDetail from "../../components/CalendarDetail.svelte";

    export let data;
</script>

{#if data.type === 'event'}
    <div class="page-wide">
        <EventDetail item={data.item} />
    </div>
{/if}

{#if data.type === 'calendar'}
    <CalendarDetail item={data.item} />
{/if}