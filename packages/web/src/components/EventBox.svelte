<script>
    import { format } from 'date-fns'
    import { MapPin, Users } from 'svelte-heros-v2';

    export let item;
</script>

<a href="/{item.slug}">
    <div class="mb-3 itembox itembox-hover flex gap-8">
        <div class="grow">
            <div class="text-neutral-content">{format(item.dateStart, "HH:mm")}</div>
            <div class="text-xl font-medium mt-1.5">{item.name}</div>
            <!--div class="text-neutral-content mt-1.5">by CryptoCanal</div-->
            <div class="flex gap-1.5 mt-1.5 text-neutral-content"><MapPin /> <div>{item.placeName}</div></div>
            {#if item.guestCount}
                <div class="flex gap-1.5 mt-1.5 text-neutral-content"><Users /> <div>{item.guestCount} guests</div></div>
            {/if}
        </div>
        <div class="w-[120px] h-[120px]">
            <img
                class="w-[120px] h-[120px] aspect-square rounded-lg"
                src={item.img} />
        </div>
    </div>
    </a>