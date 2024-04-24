<script>
    import { writable } from 'svelte/store'
    import { goto } from '$app/navigation';

    import { config } from '$lib/stores'
    import { xrpcCall, blobUpload } from '$lib/api';
    import { stringToSlug }from '$lib/utils';
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

    let avatar, avatarData, avatarInput;

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

    async function submitForm () {
        console.log(valid)
        if (!valid) {
            return false
        }
        let avatarBlob;
        if (avatar) {
            const blob = await blobUpload(fetch, avatarData)
            avatarBlob = { $cid: blob?.blob.cid }
        }
        loading = true;
        const data = {
            ...normalized,
            avatarBlob
        }
        console.log(data)
        const cal = await xrpcCall(fetch, 'app.evermeet.calendar.createCalendar', null, data)
        if (cal.error) {
            loading = false
            return false
        }
        goto(cal.baseUrl)
    }

    function onAvatarSelected (e) {
        const img = e.target.files[0];
        const r = new FileReader();
        r.readAsArrayBuffer(img);
        r.onload = re => {
            avatarData = {
                body: re.target.result,
                size: img.size,
                mimeType: img.type,
                name: img.name,
            }
        }
        const reader = new FileReader();
        reader.readAsDataURL(img);
        reader.onload = re => {
            avatar = re.target.result
        }
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
                <div class="-mt-12 ml-4" on:click={avatarInput.click()}>
                    <CalendarAvatar calendar={{}} size="88" key={handle} data={avatar} className="border-neutral/50 border-4" />
                    <input type="file" class="hidden" accept=".jpg, .jpeg, .png, .webp, .avif" on:change={(e)=>onAvatarSelected(e)} bind:this={avatarInput} >
                </div>
                <div class="grow ml-5 mt-5">
                    {#if $visibility !== 'private'}
                        <div class="font-mono text-neutral-content/50">@{handle}</div>
                    {/if}
                </div>
                <div class="mr-4 mt-4">
                    <VisibilitySelector bind={visibility} type="calendar" disabled={loading} />
                </div>
            </div>
            <div class="py-5 px-4">
                <div class="">
                    <input type="text" class="input text-xl w-full" disabled={loading} placeholder="Calendar Name" bind:value={$item.name} />
                </div>
                <div class="mt-4">
                    <input type="text" class="input w-full" disabled={loading} placeholder="Add a short description." bind:value={$item.description} />
                </div>
            </div>
        </div>
        {#if $visibility !== 'private'}
            <div class="itembox mt-4">
                <label class="text-sm mb-2 block text-neutral-content" for="handle">Handle</label>
                <div class="mt-2 mb-2 relative">
                    <input id="identifier" type="text" disabled={loading} placeholder={handlePlaceholder} class="input input-bordered w-full" bind:value={$item.localHandle} />
                    <span class="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-500 text-sm">.{$config.domain}</span>
                </div>
            </div>
        {/if}
        <div class="mt-4">
            <button type="submit" class="btn btn-primary min-w-48 {(!valid || loading) && 'btn-disabled'}">
                {#if !loading}
                    Create Calendar
                {:else}
                    <span class="loading loading-infinity loading-lg"></span>
                {/if}
            </button>
        </div>
    </form>
</div>