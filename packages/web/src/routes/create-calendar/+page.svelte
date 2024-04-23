<script>
    import { writable } from 'svelte/store'
    import { config } from '$lib/stores'
    import { goto } from '$app/navigation';

    import { xrpcCall } from '../../lib/api';
    import CalendarAvatar from '../../components/CalendarAvatar.svelte';
    import VisibilitySelector from '../../components/VisibilitySelector.svelte';

    const item = writable({ visibility: 'public' })
    const visibility = writable($item.visibility)

    visibility.subscribe(v => {
        $item.visibility = v
    })

    let loading = false;
    let valid = false;
    let normalized = {};

    //$: handle = ($item.handle || ($item.name && stringToSlug($item.name)) || '%') + '.' + $config.domain
    $: handlePlaceholder = $item.name ? stringToSlug($item.name) : 'your-calendar'
    $: handle = normalized.handle

    item.subscribe(i => {
        if (!i) {
            return
        }
        const n = {
            name: i.name,
            description: i.description,
            visibility: i.visibility,
        }

        if (!n.handle) {
            n.handle = (i.localHandle || (n.name && stringToSlug(n.name)) || '%') + '.' + $config.domain
        }
        if (n.visibility === 'private') {
            delete n.handle
        }
        normalized = n

        if (!n.name || n.name.length < 3) {
            valid = false
            return
        }
        if (n.handle && !n.handle.match(new RegExp(`^[a-z0-9\-]{3,}\.${$config.domain}$`))) {
            valid = false
            return
        }
        valid = true
    })

    function stringToSlug(str) {
        return str
            .toLowerCase()     // Convert the string to lowercase letters
            .normalize('NFD')  // Normalize the string to decompose combined letters like "é" to "e"
            .replace(/[\u0300-\u036f]/g, '') // Remove diacritical marks
            .replace(/[^a-z0-9 -]/g, '') // Remove invalid characters
            .replace(/\s+/g, '-') // Replace spaces with -
            .replace(/-+/g, '-'); // Replace multiple - with single -
    }

    async function submitForm () {
        if (!valid) {
            return false
        }
        loading = true;
        const cal = await xrpcCall(fetch, 'app.evermeet.calendar.createCalendar', null, normalized)
        if (cal.error) {
            loading = false
            return false
        }
        goto(cal.baseUrl)
    }

</script>



<div class="page-wide">
    <h1 class="heading1">Create Calendar</h1>
    <form on:submit|preventDefault={submitForm}>
        <div class="itembox no-padding">
            <div class="-z-10">
                <div class="w-full h-auto bg-neutral/50 aspect-[12/2] rounded-t-md"></div>
            </div>
            <div class="w-full flex">
                <div class="-mt-12 ml-4 w-24 h-24 rounded-lg border border-neutral/50 border-4">
                    <CalendarAvatar calendar={{}} size="88" key={handle} />
                </div>
                    <div class="grow ml-5 mt-5">
                        {#if $visibility !== 'private'}
                            <div class="font-mono text-neutral-content/50">@{handle}</div>
                        {/if}
                    </div>
                <div class="mr-4 mt-4">
                    <VisibilitySelector bind={visibility} type="calendar" />
                </div>
            </div>
            <div class="py-5 px-4">
                <div class="">
                    <input type="text" class="input text-xl w-full" placeholder="Calendar Name" bind:value={$item.name} />
                </div>
                <div class="mt-4">
                    <input type="text" class="input w-full" placeholder="Add a short description." bind:value={$item.description} />
                </div>
            </div>
        </div>
        {#if $visibility !== 'private'}
            <div class="itembox mt-4">
                <label class="text-sm mb-2 block text-neutral-content" for="handle">Handle</label>
                <div class="mt-2 mb-2 relative">
                    <input id="identifier" type="text" placeholder={handlePlaceholder} class="input input-bordered w-full" bind:value={$item.localHandle} />
                    <span class="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-500 text-sm">.{$config.domain}</span>
                </div>
            </div>
        {/if}
        <div class="mt-4">
            <button type="submit" class="btn btn-primary {!valid && 'btn-disabled'}">Create Calendar</button>
        </div>
    </form>
</div>