<script>
    import { MapPin } from 'svelte-heros-v2';
    import { format } from 'date-fns';
    import countries from 'i18n-iso-countries';
    import { config } from '$lib/config';
    import enLocale from 'i18n-iso-countries/langs/en.json';
    import { parse, parseInline } from 'marked';
    import { user } from '$lib/stores';

    countries.registerLocale(enLocale);

    export let item;

    $: countryName = item.placeCountry ? countries.getName(item.placeCountry, 'en') : ''
</script>

<svelte:head>
    <title>{item.name} | {config.sitename || config.domain}</title> 
</svelte:head>

<div class="flex gap-8">
    <div class="w-[330px]">
        <div class="w-[330px]">
            <img
                class="w-full aspect-square rounded-xl"
                src={item.img} alt={item.name} />
        </div>

        {#if $user}
            <div class="mt-6">
                <div class="itembox text-neutral-content flex gap-3">
                    <div>You have manage access for this event.</div>
                    <a href="/manage/event/{item.id}"><button class="btn btn-secondary">Manage&nbsp;↗</button></a>
                </div>
            </div>
        {/if}

        {#if item.calendar}
            <div class="mt-6">
                <div class="flex gap-4 items-center">
                    <div class="w-10 h-10 aspect-square">
                        <img class="w-10 h-10 aspect-square rounded-lg" src={item.calendar.img} />
                    </div>
                    <div>
                        <div class="text-sm">Presented by</div>
                        <div class="font-medium"><a href="/{item.calendar.slug}">{item.calendar.name}</a></div>
                    </div>
                </div>
                {#if item.calendar.description}
                    <div class="mt-4 text-sm text-neutral-content">
                        {item.calendar.description}
                    </div>
                {/if}
            </div>
        {/if}

        {#if item.hosts}
            <div class="mt-6 border-t-0 border-l-0 border-r-0 border border-neutral pb-2">
                <div class="font-mono text-sm">Hosted by</div>
            </div>
            <div class="mt-4 mb-8">
                {#each item.hosts as host}
                    <div class="flex gap-2 mb-2">
                        <div class="w-6 h-6">
                            <img
                                class="rounded-full w-6 h-6 aspect-square"
                                src={host.img} />
                        </div>
                        <div>{host.name}</div>
                    </div>
                {/each}
            </div>
        {/if}

        {#if item.guestCount}
            <div class="mt-6 border-t-0 border-l-0 border-r-0 border border-neutral pb-2">
                <div class="font-mono text-sm">{item.guestCount} going</div>
            </div>
            <div class="mt-4 mb-8 text-sm text-neutral-content">
                {item.guests.map(g => g.name).slice(0,2).join(', ')} and {item.guestCount-2} others
            </div>
        {/if}

        <div class="mt-8 text-sm text-accent">
            <a href="#contact">Contact the host</a>
        </div>
    </div>
    <div>
        <h1 class="text-5xl font-semibold font-mono mb-6">{item.name}</h1>
        <div class="flex gap-4 items-center">
            <div class="w-10 h-10 border rounded-lg border-neutral">
                <div class="text-center">
                    <div class="text-xs uppercase bg-neutral">{format(item.dateStart, 'MMM')}</div>
                    <div class="">{format(item.dateStart, 'd')}</div>
                </div>
            </div>
            <div>
                <div class="text-lg font-mono">{format(item.dateStart, 'eeee, MMMM d')}</div>
                <div class="text-sm text-neutral-content">6:00 PM - 10:00 PM GMT+2</div>
            </div>
        </div>
        <div class="flex gap-4 items-center mt-4">
            <div class="w-10 h-10 border rounded-lg border-neutral flex justify-center items-center">
                    <MapPin />
            </div>
            <div>
                <div class="font-mono text-lg">{item.placeName}</div>
                {#if item.placeCity && countryName}
                    <div class="text-sm text-neutral-content">{item.placeCity}, {countryName}</div>
                {/if}
            </div>
        </div>

        {#if item.description}
            <div class="mt-6 border-t-0 border-l-0 border-r-0 border border-neutral pb-2">
                <div class="font-mono text-sm">About Event</div>
            </div>

            <div class="mt-4 prose">
                {@html parse(item.description)}
            </div>
        {/if}
    </div>
</div>