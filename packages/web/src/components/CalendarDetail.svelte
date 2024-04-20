<script>
    import EventList from './EventList.svelte';
    import { parse } from 'marked';
    import { user } from '$lib/stores';
    import { page } from '$app/stores';
    import { calendarSubscribe, calendarUnsubscribe } from '$lib/actions';
    import { config } from '$lib/stores';
    import { CheckCircle, LockClosed } from 'svelte-heros-v2';
    import HandleBadge from './HandleBadge.svelte';
    
    export let item;

    $: subscribed = $user?.subscribedCalendars.find(sc => sc.ref === item.id || sc.ref === item.slug)
    $: managed = $user ? item.managers?.find(mi => mi.ref === $user.did) : false
</script>

<svelte:head>
    <title>{item.name}</title> 
</svelte:head>

{#if item.backdropImg}
    <div class="page-extra-wide relative -z-10">
        <div class="">
            <img
                class="lg:rounded-2xl w-full object-cover max-h-[308px]" 
                src={item.backdropImg} />
        </div>
    </div>
{/if}

<div class="page-wide {item.backdropImg ? "-mt-12" : ""} z-1">
    <div class="flex items-end mb-2">
        <div class="w-24 h-24 {item.backdropImg ? 'mb-6' : 'mt-10'} grow">
            <img
                class="{item.personal ? "rounded-full" : "rounded-lg"} w-24 h-24 {item.backdropImg ? 'border border-neutral border-4' : ''}"
                src={item.img} />
        </div>
            {#if $user}
                <div>
                    {#if managed}
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
        <div class="mt-3 text-neutral-content">{@html parse(item.description)}</div>
    {/if}
</div>

<div class="h-1 border border-neutral border-b-0 border-l-0 border-r-0 mt-8">
</div>

<div class="page-wide">
    <h2 class="text-2xl font-medium mt-6">Events</h2>

    <div class="mt-6">
        <EventList events={item.events || []} />
    </div>
</div>