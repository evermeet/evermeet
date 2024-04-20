<script>
    import { user } from '$lib/stores'
    import { calendarSubscribe, calendarUnsubscribe } from '$lib/actions';

    export let item;
    export let preview;

    const c = item;

    $: subscribed = $user?.subscribedCalendars.find(sc => sc.ref === c.id || sc.ref === c.slug)
    $: managed = $user ? item.managers?.find(mi => mi.ref === $user.did) : false
</script>

<a href={c.baseUrl}>
    <div class="itembox itembox-hover h-full">
        <div class="flex">
            <div class="w-12 h-12 mb-2 grow">
                <img
                    class="{c.personal ? "rounded-full" : "rounded-lg"} w-12 h-12"
                    src={c.img} />
            </div>
            {#if $user && !c.personal && !managed && !$user.calendarsManage.find(cm => cm.ref === item.id || cm.ref === item.slug )}
                {#if subscribed}
                    <button class="btn btn-sm" on:click|preventDefault={calendarUnsubscribe(c.id)}>Subscribed</button>
                {:else}
                    <button class="btn btn-sm btn-secondary opacity-50" on:click|preventDefault={calendarSubscribe(c.id)}>Subscribe</button>
                {/if}
            {/if}
        </div>
        <div class="text-lg font-semibold">{c.name}</div>
        {#if c._remote}
            <div class="badge badge-neutral font-mono text-xs my-2">{c._remote}</div>
        {/if}
        {#if preview}
            <div class="mt-1 text-sm text-neutral-content">{c.description}</div>
        {:else}
            <div class="text-sm mt-1 text-neutral-content">{c.subs} subscribers</div>
            <div class="mt-4 text-sm text-neutral-content">
                {#if c.personal}
                    Personal
                {:else}
                    {c.managers?.length || 0} admin
                {/if}
            </div>
        {/if}
    </div>
</a>