<script>
    import { writable } from 'svelte/store'
    import { goto } from '$app/navigation';
    import { Photo, Trash } from 'svelte-heros-v2';

    import { config } from '$lib/stores'
    import { xrpcCall, blobUpload } from '$lib/api';
    import { stringToSlug }from '$lib/utils';
    import CalendarAvatar from '../../components/CalendarAvatar.svelte';
    import VisibilitySelector from '../../components/VisibilitySelector.svelte';
    import { getContext, onMount } from 'svelte';
    import { t } from 'svelte-i18n-lingui';

    const user = getContext('user')
    const item = writable({ visibility: 'public' })
    const visibility = writable($item.visibility)

    let loading = $state(false);
    let valid = $state(false);
    let normalized = $state(normalize($item));
    let handle = $derived(normalized.handle)

    let files = $state({ avatar: {}, header: {} })
    let handlePlaceholder = $derived($item.name ? stringToSlug($item.name) : 'your-calendar')

    onMount(() => {
        console.log('x')
    })

    visibility.subscribe(v => {
        $item.visibility = v
    })

    function normalize(i) {
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
        return n
    }

    item.subscribe(i => {
        if (!i) {
            return
        }
        const n = normalize(i)
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
        let avatarBlob, headerBlob;
        if (files.avatar.data) {
            const blob = await blobUpload({ user }, files.avatar.data)
            avatarBlob = { $cid: blob?.blob.cid }
        }
        if (files.header.data) {
            const blob = await blobUpload({ user }, files.header.data)
            headerBlob = { $cid: blob?.blob.cid }
        }

        loading = true;
        const data = {
            ...normalized,
            avatarBlob,
            headerBlob,
        }
        console.log(data)
        const cal = await xrpcCall({ user }, 'app.evermeet.calendar.createCalendar', null, data)
        if (cal.error) {
            loading = false
            return false
        }
        goto(cal.baseUrl)
    }

    function onFileSelected(key) {
        return function (e) {
            const img = e.target.files[0];
            const r = new FileReader();
            r.readAsArrayBuffer(img);
            r.onload = re => {
                item.update(i => {
                    files[key].data = {
                        body: re.target.result,
                        size: img.size,
                        mimeType: img.type,
                        name: img.name,
                    }
                    return i
                })
            }
            const reader = new FileReader();
            reader.readAsDataURL(img);
            reader.onload = re => {
                files[key].dataUrl = re.target.result
            }
        }
    }
    function onFileRemoved(key) {
        return function (e) {
            files[key].dataUrl = null
            files[key].data = null
        }
    }


</script>



<div class="page-wide">
    <h1 class="heading1">{$t`Create Calendar`}</h1>
    <form on:submit|preventDefault={submitForm}>
        <div class="itembox no-padding">
            <div class="-z-10">
                <div class="group w-full h-auto bg-neutral/50 aspect-[3.5/1] rounded-t-md relative rounde-t-lg overflow-hidden">
                    {#if files.header?.dataUrl}
                        <img
                            class="w-full h-full object-cover"
                            alt="Header image"
                            src={files.header?.dataUrl} />
                    {:else}
                        <div class="w-full h-full bg-[#dddddd]/50 flex items-center justify-center font-bold text-neutral/50 leading">
                            <div>
                                <div class="text-[4rem]">Header Picture</div>
                                <div class="text-[2rem]">1920x548 (3.5/1)*</div>
                                <div class="">*or something similar 😀</div>
                            </div>
                        </div>
                    {/if}
                    <button type="button" class="" on:click|preventDefault={() => files?.header?.input?.click()}>
                        <div class="hint opacity-0 group-hover:opacity-100 absolute w-full h-full bg-base-300/50 font-semibold top-0 left-0 transition-all duration-200 backdrop-blur">
                            <div class="flex py-2 flex-wrap h-full w-full gap-3 items-center justify-center">
                                <div><Photo class="outline-none" /></div>
                                <div>Choose header image</div>
                            </div>
                        </div>
                    </button>
                    {#if files.header?.dataUrl}
                        <button type="button" class="btn btn-accent absolute opacity-0 group-hover:opacity-100 top-4 right-6 z-50" on:click|preventDefault={onFileRemoved('header')}>
                            <Trash class="outline-none" />
                        </button>
                    {/if}
                    <input type="file" class="hidden" accept=".jpg, .jpeg, .png, .webp, .avif" on:change={onFileSelected('header')} bind:this={files.header.input} >
                </div>
            </div>
            <div class="w-full flex">
                <class class="group -mt-12 ml-4 relative w-[88px] h-[88px]">
                 
                    <div class="hint opacity-0 group-hover:opacity-100 absolute w-[80px] h-[80px] backdrop-blur-sm bg-base-300/50 text-xs font-semibold top-[4px] left-[4px] transition-all duration-200 flex justify-center items-center">
                        {#if files.avatar?.dataUrl}
                            <button type="button" class="btn btn-sm btn-accent absolute opacity-0 group-hover:opacity-100" on:click|preventDefault={onFileRemoved('avatar')}>
                                <Trash class="outline-none" size={18} />
                            </button>
                        {:else}
                            <button type="button" on:click|preventDefault={() => files?.avatar?.input?.click()} class="flex py-2 flex-wrap gap-1 h-full w-full items-center justify-center">
                                <div><Photo class="outline-none" /></div>
                                <div>Choose avatar</div>
                            </button>
                        {/if}
                    </div>

                    <CalendarAvatar calendar={{}} size="88" key={handle} data={files.avatar?.dataUrl} className="border-neutral/50 border-4" />
                    <input type="file" class="hidden" accept=".jpg, .jpeg, .png, .webp, .avif" on:change={onFileSelected('avatar')} bind:this={files.avatar.input} >
                </class>
                <div class="grow ml-5 mt-5">
                    {#if $visibility !== 'private'}
                        <div class="font-mono text-base-content/25">@{handle}</div>
                    {/if}
                </div>
                <div class="mr-4 mt-4">
                    <VisibilitySelector bind={visibility} type="calendar" disabled={loading} />
                </div>
            </div>
            <div class="py-5 px-4">
                <div class="">
                    <input type="text" class="input text-xl w-full" disabled={loading} placeholder={$t`Calendar Name`} bind:value={$item.name} />
                </div>
                <div class="mt-4">
                    <input type="text" class="input w-full" disabled={loading} placeholder="Add a short description." bind:value={$item.description} />
                </div>
            </div>
        </div>
        {#if $visibility !== 'private'}
            <div class="itembox mt-4">
                <label class="text-sm mb-2 block text-base-content/75" for="handle">Handle</label>
                <div class="mt-2 mb-2 relative">
                    <input id="identifier" type="text" disabled={loading} placeholder={handlePlaceholder} class="input input-bordered w-full" bind:value={$item.localHandle} />
                    <span class="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-500 text-sm">.{$config.domain}</span>
                </div>
            </div>
        {/if}
        <div class="mt-4">
            <button type="submit" class="btn btn-primary min-w-48 {(!valid || loading) && 'btn-disabled'}">
                {#if !loading}
                    {$t`Create Calendar`}
                {:else}
                    <span class="loading loading-infinity loading-lg"></span>
                {/if}
            </button>
        </div>
    </form>
</div>

<style>
    .avatar-upload:hover .avatar-upload-hint {
        display: block;
    }
</style>