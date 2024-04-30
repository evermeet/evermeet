<script>
    import { formatDistanceToNow } from 'date-fns'
    import { writable } from 'svelte/store';
    import { LockClosed, PaperAirplane } from 'svelte-heros-v2';
    import { markdownToHTML } from '$lib/utils';

    import { socket } from '$lib/stores';
    import UserAvatar from './UserAvatar.svelte';
    import { xrpcCall } from '$lib/api'
    import { getContext } from 'svelte';

    const { room, item = null, chatData } = $props()

    let input
    let isLoading = false
    let error = null

    const user = getContext("user")
    const messages = $derived(chatData)

    $effect(() => {
        input.focus()
    })

    socket.subscribe((data) => {
        if (!data) {
            return null
        }
        const { sc, nc } = data
        const sub = nc.subscribe("chat");
        (async () => {
            for await (const m of sub) {
                const d = sc.decode(m.data)
                console.log(`[${sub.getProcessed()}]: ${d}`);

                messages.update(arr => {
                    arr.unshift({ txt: d })
                    return arr
                })
            }
            console.log("subscription closed");
        })();
    })

    async function submitMessage () {
        isLoading = true
        error = null
        let res;
        try {
            res = await xrpcCall({ fetch, user }, 'app.evermeet.chat.createMessage', null, {
                room: room.id,
                msg: input.value,
            })
        } catch(e) {
            error = e
            isLoading = false
            return false
        }
        messages.unshift(res)
        isLoading = false
        input.value = ''
    }

    function onKeyDown(e) {
        const isSearch = Boolean(window.document.getElementById('masterSearchDialog'))
        if (isSearch) {
            return false;
        }
        const char = String.fromCharCode(e.keyCode)
        //console.log(input)
        if (document.activeElement !== input && char.match(/^[\w\W]$/)) {
            input.focus()
            return false;
        }
        return true;
    }

</script>

<svelte:window on:keydown={onKeyDown} />

<div class="itembox no-padding xchat relative">
    <div class="absolute w-full h-20 bg-gradient-to-b from-base-300 to-transparent z-10"></div>
    <div class="h-[58vh] overflow-x-scroll flex flex-col-reverse py-4 w-full">
        {#each messages as m}
            <div class="flex gap-4 w-full py-1.5 px-2 hover:bg-base-100/50">
                <div class="w-52 shrink-0 flex justify-end">
                    <div class="">
                        <div class="flex gap-1.5 items-center mb-auto mt-1 justify-end">
                            <div class="font-bold break-words text-sm text-right">{m.author.name}</div>
                            <div class="shrink-0"><UserAvatar user={m.author} size="18" /></div>
                        </div>
                        <div class="text-right text-xs opacity-50">{formatDistanceToNow(new Date(m.createdOn))} ago</div>
                    </div>
                </div>
                <div class="grow opacity-90 prose prose-md max-w-none">{@html markdownToHTML(m.msg)}</div>
            </div>
        {/each}
    </div>
    {#if error}
        <div class="p-2 pt-0">
            <div class="text-error bg-error/10 py-2 px-3 rounded-lg">{error}</div>
        </div>
    {/if}
    {#if user}
        <form class="pb-1.5 px-1.5 flex items-center relative" on:submit|preventDefault={submitMessage}>
            <div class="rounded-l-lg bg-base-100 py-3 pl-3 pr-1.5" class:bg-base-200={isLoading}>
                <UserAvatar {user} />
            </div>
            <input id="xchat-input" type="text" class="input w-full !outline-none !border-0 pl-1.5 ml-0 rounded-l-none" bind:this={input} autocomplete="off" placeholder="Send a message..." disabled={isLoading} />
            <button type="submit" class="absolute right-4 btn btn-sm btn-neutral" disabled={isLoading}><PaperAirplane size="16" /></button>
        </form>
    {/if}
</div>

<style>
</style>