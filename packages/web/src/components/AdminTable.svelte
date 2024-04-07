<script>
    import { onMount } from "svelte";
    import { apiCall } from "../lib/api";
    import { writable } from "svelte/store";

    export let collection;
    let items = [];

    const layouts = {
        users: [
            { name: 'Name', col: 'name' }
        ],
        calendars: [
            { name: 'Name', col: 'name', type: 'nameslug' },
            { name: 'Subs', col: 'subs' },  
        ],
        events: [
            { name: 'Date', col: 'dateStart' },     
            { name: 'Name', col: 'name', type: 'nameslug' },
            { name: 'Location', col: 'placeCity', type: 'location' },
            //{ name: 'Country', col: 'placeCountry' },
            { name: 'Guests', col: 'guestCount' },
            { name: 'Calendar', col: 'calendarId' },
        ]
    }

    onMount(async () => {
        const resp = await apiCall(fetch, 'admin/collection/' + collection)
        if (resp) {
            items = resp.items
        }
    })

</script>

<table class="table bg-base-300">
    <thead>
        <tr>
            <th></th>
            {#each layouts[collection] as li}
                <th>{li.name}</th>
            {/each}
        </tr>
    </thead>
    <tbody>
        {#each items as item}
            <tr class='hover'>
                <td><input type="checkbox" class="checkbox checkbox-sm" /></td>
                {#each layouts[collection] as li}
                    <td class="text-neutral-content">
                        {#if li.type === 'nameslug'}
                            {item[li.col]} <div class="ml-2 badge badge-neutral text-xs font-mono">{item.slug}</div>
                        {:else if li.type === 'location'}
                            {#if item.placeCity && item.placeCountry}
                                {item.placeCity}, {item.placeCountry.toUpperCase()}
                            {/if}
                        {:else}
                            {item[li.col] || ''}
                        {/if}
                    </td>
                {/each}
            </tr>
        {/each}
    </tbody>
</table>