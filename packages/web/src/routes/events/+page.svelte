<script>
    import EventList from '../../components/EventList.svelte';
    export let data;

</script>

<div class="page-wide">
    <div class="flex items-center">
        <h1 class="text-3xl font-semibold grow">Events</h1>
        <label class="flex cursor-pointer gap-2">
            <span class="label-text">Upcoming</span> 
            <input type="checkbox" value="synthwave" class="toggle theme-controller"/>
            <span class="label-text">Past</span> 
        </label>
    </div>

    <div class="mt-10">
        <EventList events={data.events}/>
    </div>
</div>