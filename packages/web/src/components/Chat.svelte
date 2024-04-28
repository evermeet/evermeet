<script>
    import { user, socket } from '$lib/stores';
    import { afterUpdate, onDestroy, onMount } from 'svelte';
    import { LockClosed, PaperAirplane } from 'svelte-heros-v2';
    //import SvelteMarkdown from 'svelte-markdown'

    import UserAvatar from './UserAvatar.svelte';
    import { writable } from 'svelte/store';


    export let room;
    export let item;

    let input
    let isLoading = false

    onMount(async () => {
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

    afterUpdate(() => {
        input.focus()
    })

    function submitMessage () {
        //isLoading = true
        messages.update(arr => {
            arr.unshift({ txt: input.value })
            return arr
        })
        input.value = ''
    }

    function onKeyDown(e) {
        const char = String.fromCharCode(e.keyCode)
        console.log(input)
        if (document.activeElement !== input && char.match(/^[\w\W]$/)) {
            input.focus()
            return false;
        }
        return true;
    }

    const msgs = [
        {
            txt: '111 Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Integer rutrum, orci vestibulum ullamcorper ultricies, lacus quam ultricies odio, vitae placerat pede sem sit amet enim.'
        },
        {
            txt: 'Aliquam id dolor. Duis bibendum, lectus ut viverra rhoncus, dolor nunc faucibus libero, eget facilisis enim ipsum id lacus. Proin pede metus, vulputate nec, fermentum fringilla, vehicula vitae, justo. Mauris dictum facilisis augue. Pellentesque pretium lectus id turpis. In rutrum. Etiam quis quam. Mauris dolor felis, sagittis at, luctus sed, aliquam non, tellus. Nullam sapien sem, ornare ac, nonummy non, lobortis a enim. Ut tempus purus at lorem.'
        },
        {
            txt: 'Fusce aliquam vestibulum ipsum.'
        },
        {
            txt: 'https://web3privacy.info'
        },
        {
            txt: 'Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Praesent lobortis dui nec ante.'
        },
        {
            txt: 'Nunc sed turpis. Sed sagittis, nisl in dictum ultricies, nisi libero ultricies odio, ut vehicula magna dia sed magna. Morbi pellentesque libero sit amet ante. Maecenas tellus. Etiam vel tortor sodales tellus ultricies commodo. Suspendisse potenti.',
            username: 'very long name of some person maybe',
        },
        {
            txt: 'Aenean fermentum, elit eget tincidunt condimentum, eros ipsum rutrum orci, sagittis tempus lacus enim ac dui. Donec non enim in turpis pulvinar facilisis.'
        },
        {
            txt: 'Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Integer rutrum, orci vestibulum ullamcorper ultricies, lacus quam ultricies odio, vitae placerat pede sem sit amet enim.'
        },
        {
            txt: 'Aliquam id dolor. Duis bibendum, lectus ut viverra rhoncus, dolor nunc faucibus libero, eget facilisis enim ipsum id lacus. Proin pede metus, vulputate nec, fermentum fringilla, vehicula vitae, justo. Mauris dictum facilisis augue. Pellentesque pretium lectus id turpis. In rutrum. Etiam quis quam. Mauris dolor felis, sagittis at, luctus sed, aliquam non, tellus. Nullam sapien sem, ornare ac, nonummy non, lobortis a enim. Ut tempus purus at lorem.'
        },
        {
            txt: 'Fusce aliquam vestibulum ipsum.'
        },
        {
            txt: 'Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Praesent lobortis dui nec ante.'
        },
        {
            txt: 'Nunc sed turpis. Sed sagittis, nisl in dictum ultricies, nisi libero ultricies odio, ut vehicula magna dia sed magna. Morbi pellentesque libero sit amet ante. Maecenas tellus. Etiam vel tortor sodales tellus ultricies commodo. Suspendisse potenti.'
        },
        {
            txt: 'Aenean fermentum, elit eget tincidunt condimentum, eros ipsum rutrum orci, sagittis tempus lacus enim ac dui. Donec non enim in turpis pulvinar facilisis.'
        }
    ]

    const roomMsgs = {
        general: msgs,
        help: msgs.slice(3,10),
        assistant: msgs.slice(7,10),
        speakers: [],
        team: [],
    }

    const rooms = [
        { id: 'general' },
        { id: 'help', badge: 5 },
        { id: 'assistant' },
        { id: 'speakers', private: true },
        { id: 'team', private: true },
    ]

    $: messages = writable(roomMsgs[room])

</script>

<svelte:window on:keydown={onKeyDown} />

<div class="flex gap-10 items-center">
    <h2 class="text-2xl font-medium">Chat</h2>
    <ul class="mt-2 menu menu-horizontal menu-sm bg-base-200 rounded-box">
        {#each rooms as r}
            <li>
                <a href="{item.baseUrl}/chat{r.id !== 'general' ? '?room='+r.id : ''}" class:active={room===r.id}>
                    #{r.id}
                    {#if r.private}
                        <LockClosed size="14" class="mr-0 outline-none" tabindex="-1" />
                    {/if}
                </a>
            </li>
        {/each}
    </ul> 
</div>

<div class="mt-6 itembox no-padding xchat relative">
    <div class="absolute w-full h-20 bg-gradient-to-b from-base-300 to-transparent z-10"></div>
    <div class="h-[58vh] overflow-x-scroll flex flex-col-reverse p-4 w-full">
        {#each $messages as m}
            <div class="flex gap-4 w-full mt-3">
                <div class="w-52 shrink-0 flex justify-end">
                    <div>
                        <div class="flex gap-1.5 items-center mb-auto mt-1">
                            <div class="font-bold break-words text-sm text-right">{m.username || $user.name}</div>
                            <div class="shrink-0"><UserAvatar user={$user} size="18" /></div>
                        </div>
                        <div class="text-right text-xs opacity-50">12:34</div>
                    </div>
                </div>
                <div class="grow opacity-90 prose prose-md max-w-none"><!--SvelteMarkdown source={m.txt} /-->{m.txt}</div>
            </div>
        {/each}
    </div>
    <form class="pb-1.5 px-1.5 flex items-center relative" on:submit|preventDefault={submitMessage}>
        <div class="rounded-l-lg bg-base-100 py-3 pl-3 pr-3">
            <UserAvatar user={$user} />
        </div>
        <input id="xchat-input" type="text" class="input w-full !outline-none !border-0 pl-0 ml-0 rounded-l-none" bind:this={input} autocomplete="off" placeholder="Send a message..." />
        <button type="submit" class="absolute right-4 btn btn-sm btn-neutral" disabled={isLoading}><PaperAirplane size="16" /></button>
    </form>
</div>