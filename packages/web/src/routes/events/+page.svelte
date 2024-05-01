<script>
    import EventList from '../../components/EventList.svelte';
    import { t } from '$lib/i18n';
    export let data;

</script>

<div class="page-wide">
    <div class="flex items-center">
        <h1 class="heading1 no-margin grow">{$t`Events`}</h1>
        <label class="flex cursor-pointer gap-2">
            <span class="label-text">{$t`Upcoming`}</span> 
            <input type="checkbox" value="synthwave" class="toggle theme-controller"/>
            <span class="label-text">{$t`Past`}</span> 
        </label>
    </div>

    <div class="mt-10">
        <EventList events={data.events}/>
    </div>
</div>