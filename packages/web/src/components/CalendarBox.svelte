<script>
    import { user } from '$lib/stores'
    import { calendarSubscribe, calendarUnsubscribe } from '$lib/actions';
    import CalendarAvatar from './CalendarAvatar.svelte';

    export let item;
    export let preview = null;

    const c = item;

    $: subscribed = $user?.subscribedCalendars?.find(sc => sc.ref === c.id || sc.ref === c.slug)
    $: managed = $user ? item.managers?.find(mi => mi.ref === $user.did) : false
</script>

<a href={c.baseUrl}>
    <div class="itembox itembox-hover h-full group">
        <div class="flex">
            <div class="w-12 h-12 mb-2 grow">
                <CalendarAvatar calendar={item} size="48" />
            </div>
            {#if $user && !c.personal && !managed && !$user.calendarSubscriptions?.find(cm => cm.ref === item.did )}
                <div class="">
                    {#if subscribed}
                        <button class="btn btn-sm" on:click|preventDefault={calendarUnsubscribe(c.id)}>Subscribed</button>
                    {:else}
                        <button class="btn btn-sm btn-secondary opacity-35 hover:opacity-100 duration-200" on:click|preventDefault={calendarSubscribe(c.id)}>Subscribe</button>
                    {/if}
                </div>
            {/if}
        </div>
        <div class="text-lg font-semibold">{c.name}</div>
        {#if c._remote}
            <div class="badge badge-neutral font-mono text-xs my-2">{c._remote}</div>
        {/if}
        {#if preview}
            <div class="mt-1 text-sm text-base-content/75">{c.description}</div>
        {:else}
            <div class="text-sm mt-1 text-base-content/75">{c.subs} subscribers</div>
            <div class="mt-4 text-sm text-base-content/75">
                {#if c.personal}
                    Personal
                {:else}
                    {c.managers?.length || 0} admin
                {/if}
            </div>
        {/if}
    </div>
</a>