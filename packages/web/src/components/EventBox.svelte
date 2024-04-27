<script>
    import { format } from 'date-fns'
    import { MapPin, Users } from 'svelte-heros-v2';
    import FlagIcon from './FlagIcon.svelte';
    import { config } from '$lib/stores';

    export let item;
    export let virtual = false;
</script>

<a href={item.baseUrl}>
    <div class="mb-3 itembox {!virtual && "itembox-hover"} flex gap-8">
        <div class="grow">
            <div class="text-base-content/75">{format(item.dateStart, "HH:mm")}</div>
            <div class="text-xl font-medium mt-1.5">{item.name}</div>
            <!--div class="text-base-content/75 mt-1.5">by CryptoCanal</div-->
            <div class="flex gap-1.5 mt-1.5 text-base-content/75 items-center"><MapPin /> <div>{item.placeName}</div> <div class="ml-2 opacity-75"><FlagIcon country={item.placeCountry} /></div> <div>{item.placeCity}</div></div>
            {#if item.guestCount}
                <div class="flex gap-1.5 mt-1.5 text-base-content/75"><Users /> <div>{item.guestCount} guests</div></div>
            {/if}
            {#if item._remote}
                <div class="badge badge-neutral font-mono text-xs mt-2.5 opacity-50">{item._remote}</div>
            {/if}
        </div>
        <div class="w-[120px] h-[120px]">
            <img
                class="w-[120px] h-[120px] aspect-square rounded-lg"
                src={item.img} />
        </div>
    </div>
    </a>