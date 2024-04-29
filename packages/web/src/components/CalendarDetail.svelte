<script>
    import EventList from './EventList.svelte';
    import { parse } from 'marked';
    import { user } from '$lib/stores';
    import { page } from '$app/stores';
    import { calendarSubscribe, calendarUnsubscribe } from '$lib/actions';
    import { config } from '$lib/stores';
    import { CheckCircle, LockClosed, CalendarDays, QueueList, ChatBubbleLeft, Clipboard, Inbox, Users, PresentationChartBar, ArrowsPointingOut, ArrowsPointingIn } from 'svelte-heros-v2';
    import HandleBadge from './HandleBadge.svelte';
    import CalendarAvatar from './CalendarAvatar.svelte';
    import EventBox from './EventBox.svelte';
    import Refs from './Refs.svelte';
    import Chat from './Chat.svelte';
    import ChatRoomSelect from './ChatRoomSelect.svelte';
    import { imgBlobUrl } from '$lib/api';
    
    export let item
    export let selectedTab
    export let params = {}

    let tabs = [
        { id: null, name: 'Events', ico: CalendarDays, bubble: item.events.length },
        { id: 'concepts', name: 'Drafts', ico: Inbox, bubble: item.concepts.length, bubbleAccent: true },
        { id: 'talks', name: 'Talks', ico: PresentationChartBar },
        { id: 'contributors', name: 'People', ico: Users },
        //{ id: 'about', name: 'About' },
        { id: 'feed', name: 'Feed', ico: QueueList },
    ]

    if (item.rooms && item.rooms.length > 0) {
        tabs.push({ id: 'chat', name: 'Chat', ico: ChatBubbleLeft })
    }

    $: subscribed = $user?.calendarSubscriptions?.find(sc => sc.ref === item.did)
    $: managed = $user ? item.managers?.find(mi => mi.ref === $user.did) : false

    $: isFullPage = selectedTab === null && !params.expand
    $: backdropImg = isFullPage && ((item.headerBlob && imgBlobUrl(item.did, item.headerBlob, 1000)) || item.backdropImg)

    $: currentRoom = typeof(params.room) === 'string' ? params.room : 'general'
    $: currentRoomObj = item.rooms.find(r => r.slug === currentRoom)
    
    function changeTab () {
    }
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
    <div class="flex items-end mb-2 gap-4">
        <div class="{backdropImg ? 'mb-6' : (!isFullPage ? '' : 'mt-10')} {isFullPage ? 'grow' : ''} z-10">
            <CalendarAvatar calendar={item} size={isFullPage ? 88 : 60} className="{backdropImg ? 'border-neutral/50 border-4' : ''}" />
        </div>
        {#if !isFullPage}
            <div class="grow">
                <h1 class="text-3xl font-medium">{item.name}</h1>
                <HandleBadge {item} margin="mt-1.5" size="small" />
            </div>
        {/if}
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
    {#if isFullPage}
        <h1 class="text-4xl font-medium {item.backdropImg ? '' : 'mt-6'}">{item.name}</h1>
        <HandleBadge {item} />

        {#if item.description}
            <div class="mt-3 text-base-content/75">{@html parse(item.description)}</div>
        {/if}
        {#if item.refs}
            <div class="mt-4">
                <Refs {item} />
            </div>
        {/if}
    {/if}
</div>

<div class="page-wide">
    <div>
        <div class="xtabs text-left mt-6 text-sm">
            {#each tabs as tab}
                <a href="{item.baseUrl}/{tab.id ? tab.id : ''}" class="{selectedTab === tab.id ? 'active' : ''} flex items-center gap-1.5" on:click={changeTab(tab.id)}>
                    {#if tab.ico}
                        <svelte:component this={tab.ico} size="18" class="outline-none" tabindex="-1" />
                    {/if}
                    {tab.name}
                    {#if tab.bubble}
                        <div class="badge badge-sm {tab.bubbleAccent ? 'badge-accent' : 'badge-neutral'}">{tab.bubble}</div>
                    {/if}
                </a>
            {/each}
            <div class="grow"></div>
            {#if selectedTab === null}
                <div class="">
                    {#if params.expand}
                        <a href="{item.baseUrl}/{selectedTab || ''}" class="flex items-center gap-1.5 opacity-50 text-xs p-2 hover:opacity-100"><ArrowsPointingIn size="16" class="" /> Show info</a>
                    {:else}    
                        <a href="{item.baseUrl}/{selectedTab || ''}?expand" class="flex items-center gap-1.5 opacity-50 text-xs p-2 hover:opacity-100"><ArrowsPointingOut size="16" class="" /> Expand</a>
                    {/if}
                </div>
            {/if}
        </div>
    </div>
</div>

<div class="h-1 border border-neutral border-b-0 border-l-0 border-r-0">
</div>

{#if selectedTab === "feed"}
    <div class="page-wide">
        <h2 class="text-2xl font-medium mt-6">Feed</h2>

    </div>

{:else if selectedTab === 'concepts'}

    <div class="page-wide">
        <h2 class="text-2xl font-medium mt-6">Event Concepts</h2>

        <div class="mt-6">
            {#each item.concepts as event}
                <EventBox item={event} />
            {/each}
        </div>
    </div>

{:else if selectedTab === 'contributors'}

    <div class="page-wide">
        <h2 class="text-2xl font-medium mt-6">Speakers</h2>

        <div class="mt-6">
            Contributors
        </div>
    </div>

{:else if selectedTab === 'chat'}

    <div class="page-wide">
    </div>
    <div class="page-wide">
        <ChatRoomSelect rooms={item.rooms} baseUrl={item.baseUrl} {currentRoom} />
        <Chat {item} room={currentRoomObj} chatData={item._chat} />
    </div>

{:else if selectedTab === null}
    <div class="page-wide">
        <h2 class="text-2xl font-medium mt-6">Planned Events</h2>

        <div class="mt-6">
            <EventList events={item.events} />
        </div>
    </div>
{/if}