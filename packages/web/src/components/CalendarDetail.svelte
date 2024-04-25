<script>
    import EventList from './EventList.svelte';
    import { parse } from 'marked';
    import { user } from '$lib/stores';
    import { page } from '$app/stores';
    import { calendarSubscribe, calendarUnsubscribe } from '$lib/actions';
    import { config } from '$lib/stores';
    import { CheckCircle, LockClosed } from 'svelte-heros-v2';
    import HandleBadge from './HandleBadge.svelte';
    import CalendarAvatar from './CalendarAvatar.svelte';
    import { imgBlobUrl } from '$lib/api';
    
    export let item;

    $: subscribed = $user?.calendarSubscriptions?.find(sc => sc.ref === item.did)
    $: managed = $user ? item.managers?.find(mi => mi.ref === $user.did) : false
    $: backdropImg = (item.headerBlob && imgBlobUrl(item.did, item.headerBlob, 1000)) || item.backdropImg
</script>

<svelte:head>
    <title>{item.name}</title> 
</svelte:head>

{#if backdropImg}
    <div class="page-extra-wide relative">
        <div class="">
            <img
                alt={item.name}
                class="lg:rounded-2xl w-full object-cover max-h-[308px] aspect-[3.5/1] bg-base-300" 
                src={backdropImg} />
        </div>
    </div>
{/if}

<div class="page-wide {backdropImg ? "-mt-12" : ""} z-0">
    <div class="flex items-end mb-2">
        <div class="{backdropImg ? 'mb-6' : 'mt-10'} grow z-10">
            <CalendarAvatar calendar={item} size="88" className="{backdropImg ? 'border-neutral/50 border-4' : ''}" />
        </div>
            {#if $user}
                <div>
                    {#if item.$userContext?.isManager}
                        <a href="/manage/calendar/{item.id}" class="btn btn-accent">Manage</a>
                    {:else}
                        {#if subscribed}
                            <button class="btn btn-neutral" on:click={calendarUnsubscribe(item.id)}>Subscribed</button>
                        {:else}
                            <button class="btn btn-secondary" on:click={calendarSubscribe(item.id)}>Subscribe</button>
                        {/if}
                    {/if}
                </div>
            {:else}
                <div>
                    <a href="/login?next={encodeURIComponent($page.url)}" class="btn btn-secondary">Subscribe</a>
                </div>
            {/if}

    </div>
    <h1 class="text-4xl font-medium {item.backdropImg ? '' : 'mt-6'}">{item.name}</h1>
    <HandleBadge {item} />

    {#if item.description}
        <div class="mt-3 text-base-content/75">{@html parse(item.description)}</div>
    {/if}
</div>

<div class="h-1 border border-neutral border-b-0 border-l-0 border-r-0 mt-8">
</div>

<div class="page-wide">
    <h2 class="text-2xl font-medium mt-6">Events</h2>

    <div class="mt-6">
        <EventList events={item.events} />
    </div>
</div>