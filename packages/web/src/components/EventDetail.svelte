<script>
    import { MapPin, CheckCircle } from 'svelte-heros-v2';
    import { format } from 'date-fns';
    import countries from 'i18n-iso-countries';
    import enLocale from 'i18n-iso-countries/langs/en.json';
    import { parse, parseInline } from 'marked';
    import { user, config, eventDetail } from '$lib/stores';
    import { register, unregister } from '$lib/actions';
    import FlagIcon from './FlagIcon.svelte';
    import HandleBadge from './HandleBadge.svelte';
    import CalendarAvatar from './CalendarAvatar.svelte';

    countries.registerLocale(enLocale);

    export let item;
    eventDetail.set(item);

    console.log(item.url)

    $: item = $eventDetail;
    $: countryName = item.placeCountry ? countries.getName(item.placeCountry, 'en') : ''
    $: userRegistered = $user && $user.events?.find(e => e.ref === item.id) ? true : false
</script>

<svelte:head>
    <title>{item.name} | {$config.sitename || $config.domain}</title> 
</svelte:head>

<div data-id="#{item.handle}" class="flex gap-8">
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
                    <a class="btn btn-secondary" href="/manage/event/{item.id}">Manage&nbsp;↗</a>
                </div>
            </div>
        {/if}

        {#if item.calendar}
            <div class="mt-6">
                <div class="flex gap-4 items-center">
                    <div class="w-10 h-10 aspect-square">
                        <CalendarAvatar calendar={item.calendar} size="40" />
                    </div>
                    <div>
                        <div class="text-sm">Presented by</div>
                        <div class="font-medium"><a href="{item.calendar.baseUrl}">{item.calendar.name}</a></div>
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

        {#if item.guestCountTotal}
            <div class="mt-6 border-t-0 border-l-0 border-r-0 border border-neutral pb-2">
                <div class="font-mono text-sm">{item.guestCountTotal} going</div>
            </div>
            <div class="mt-4 mb-8 text-sm text-neutral-content">
                {#if item.guests}
                    {item.guests.map(g => g.name).slice(0,2).join(', ')} and {item.guestCountTotal} others
                {:else}
                {/if}
            </div>
        {/if}

        <div class="mt-8 text-sm text-accent">
            <a href="#contact">Contact the host</a>
        </div>
    </div>
    <div class="w-full">
        <div class="mb-6">
            <h1 class="text-5xl font-semibold font-mono">{item.name}</h1>
            <HandleBadge {item} size="small" type="event" />
        </div>
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
                    <div class="text-sm text-neutral-content flex gap-2 items-center"><div>{item.placeCity}, {countryName}</div> <FlagIcon country={item.placeCountry} /></div>
                {/if}
            </div>
        </div>

        {#if item.registration?.enabled}
            <div class="mt-6 itembox no-padding w-full">
                {#if !userRegistered}
                    <div class="bg-neutral rounded-t-lg py-2 px-4 text-sm">Registration</div>
                {/if}
                <div class="py-4 px-4">
                    {#if userRegistered}
                        <div><img src={$user.img} alt={$user.name} class="w-10 rounded-full mb-2" /></div>
                        <div class="text-xl font-semibold font-mono">You’re In</div>
                        <div class="mt-1.5 text-neutral-content">A confirmation email has been sent to your email.</div>
                        <div class="mt-4 text-sm text-neutral-content">No longer able to attend? Notify the host by <button class="underline text-accent opacity-75 hover:opacity-100" on:click={unregister(item.id)}>canceling your registration</button>.</div>
                    {:else}
                        <div class="mb-4">Welcome! To join the event, please register below.</div>
                        {#if $user}
                            <div class="flex gap-2 mt-4 items-center">
                                <div><img src={$user.img} alt={$user.name} class="rounded-full w-5" /></div>
                                <div class="font-semibold">{$user.name} <!--span class="font-mono text-xs badge badge-neutral inline-block">{$user.did}</span--></div>
                            </div>
                        {/if}
                        <div class="mt-3">
                            <button class="btn w-full text-lg btn-neutral" on:click={register(item.id)}>Register</button>
                        </div>
                    {/if}
                </div>
            </div>
        {/if}

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