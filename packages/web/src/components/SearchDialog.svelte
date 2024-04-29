<script>

    import { createDialog, createCombobox, melt } from '@melt-ui/svelte'
    import {  MagnifyingGlass } from 'svelte-heros-v2';

    import { fade } from 'svelte/transition';
    import { writable } from 'svelte/store';
    import MiniSearch from 'minisearch'
    import { xrpcCall } from '$lib/api';
    import { onMount } from 'svelte';
    import { goto } from '$app/navigation';
    import CalendarAvatar from './CalendarAvatar.svelte';
    import { browser } from '$app/environment';
    import { searchItemsBase } from '$lib/search';
    
    let miniSearch;
    let miniSearchLoading = false;
    let data = searchItemsBase()
    const q = writable('')
    let suggest = ''
    let results = []

    onMount(async () => {
        miniSearch = new MiniSearch({
            fields: ['name', 'handle', 'keywords'],
            storeFields: ['id'],
            searchOptions: {
                boost: { name: 2 },
                fuzzy: 0.2
            }
        })
        miniSearch.addAll(data)
        const calData = await xrpcCall(fetch, 'app.evermeet.explore.getExplore')
        const calendars = calData.calendars.map(x => {
            return {
                id: x.id,
                name: x.name,
                handle: x.handle,
                avatarBlob: x.avatarBlob,
                type: 'calendar',
                baseUrl: x.baseUrl,
                did: x.did,
                img: x.img,
            }
        })
        data.push(...calendars)
        miniSearch.addAll(calendars)
    })

    const {
        elements: { trigger, portalled, overlay, content, title, description, close },
        states: { open }
    } = createDialog({
        closeOnOutsideClick: false
    })


    function onSelectedChange (x) {
        const id = x.next.value

        const item = data.find(i => i.id === id)
        if (item) {
            open.set(false)

            if (item.baseUrl.match(/^http/)) {
                window.location = item.baseUrl
            } else {
                console.log('Going to:', item.baseUrl)
                goto(item.baseUrl, { invalidateAll: true })
            }
        }
    }

    const {
        elements: { trigger: triggerCombo, menu, input, option, label },
        states: { open: comboOpen, inputValue, touchedInput, selected, highlighted, highlightedItem },
        helpers: { isSelected },
    } = createCombobox({
        forceVisible: true,
        onSelectedChange,
        //closeOnOutsideClick: false, 
        //defaultOpen: true,
        //defaultSelected: { value: 'create-calendar' },
        positioning: {
            placement: 'bottom',
            offset: {
                mainAxis: 0,
                crossAxis: -15,
            },
            gutter: 0
            //strategy: 'fixed'
        },
    });

    comboOpen.subscribe((x) => {
        /*if (!x) {
            //open.set(x)
        }*/
        open.set(x)
    })

    //highlightedItem.subscribe(console.log)

    inputValue.subscribe((str) => {
        if (!str) {
            results = data
        } else if (miniSearch) {
            const suggest = miniSearch.autoSuggest(str, { fuzzy: 0.2 })
            let out = []
            for (const s of suggest) {
                for (const ss of miniSearch.search(s.suggestion)) {
                    if (out.find(s => s.id === ss.id)) {
                        continue
                    }
                    out.push(ss)
                }
            }
            results = out
            //console.log({ results: miniSearch.search(str), suggest: results, data })
        }
        if (browser) {
            setTimeout(() => {
                highlightedItem.set(document.querySelector('.dropdown-content li:first-child'))
            }, 10)
        }
    })

    function toggleSearchDialog () {
            //open.update(v => !v)
            comboOpen.update(v => !v)
            inputValue.set('')
            setTimeout(() => {
                highlightedItem.set(document.querySelector('.dropdown-content li:first-child'))
            }, 0)
            //comboOpen.update(v => !v)
            //console.log({ $open, $comboOpen })
    }

    function onKeyDown(x, event) {
        if (x.code === 'KeyK' && x.metaKey === true) {
            x.preventDefault()
            toggleSearchDialog()
        }
    }

</script>

<svelte:window on:keydown={onKeyDown} />

<button class="tooltip tooltip-bottom" data-tip="Search ⎯ ⌘K" on:click={() => toggleSearchDialog()}>
    <div class="w-8 h-8 rounded-full aspect-square border-[0.4em] border-transparent hover:border-neutral hover:bg-neutral cursor-pointer flex items-center justify-center">
        <MagnifyingGlass size="20" tabindex="-1" class="outline-none" />
    </div>
</button>

{#if $comboOpen}
    <div use:melt={$portalled}>
        <div use:melt={$overlay}
            class="fixed inset-0 z-50 bg-black/30 backdrop-blur-sm"
        />
        <div use:melt={$content}
            class="fixed left-1/2 top-[10vh] z-50 max-h-[85vh] w-[90vw]
            max-w-[600px] -translate-x-1/2 rounded-md bg-base-100
            shadow"
        >
            <div class="flex items-center gap-2 px-4 py-0.5">
                <MagnifyingGlass class="opacity-50" />
                <input type="input"
                    class="input input-ghost px-1.5 py-0 w-full focus:outline-none focus:border-none text-lg grow"
                    placeholder="Search Evermeet" 
                    on:click|preventDefault={(p) => {
                        setTimeout(() => {
                            console.log('.', $comboOpen)
                            if(!$comboOpen) {
                                //comboOpen.set(true)
                            }
                        }, 0)
                    }}
                    use:melt={$input}

                />
                <div class="hidden">
                    <h2 use:melt={$title}>Dialog Title</h2>
                    <p use:melt={$description}>Dialog description</p>
                </div>
            </div>
            {#if results}
                <ul id="search-menu-content" class="dropdown-content popup-menu rounded-t-none bg-base-100 w-[90vw] max-w-[600px] max-h-[73vh] overflow-scroll border-t border-neutral/50 rounded-t-0" use:melt={$menu}>
                    {#each results.map(r => data.find(d => d.id === r.id)) as item}
                        <li class="py-2 px-2 cursor-pointer" use:melt={$option({ value: item.id, label: item.name })}>
                            <div class="flex items-center gap-3">
                                {#if item.type === 'calendar'}
                                    <CalendarAvatar calendar={item} size="38" />
                                    <div>
                                        <div>{item.name} {#if $isSelected(item.id)} (selected){/if}</div>
                                        <div class="text-xs opacity-50">@{item.handle}</div>
                                    </div>
                                {:else}
                                    <div class="w-[40px] h-[40px]">
                                        {#if item.icon}
                                            <svelte:component this={item.icon} size="38px" class="p-2 bg-base-300/75 rounded opacity-75" />
                                        {/if}
                                    </div>
                                    <div>
                                        <div>{item.name} {#if $isSelected(item.id)} (selected){/if}</div>
                                        {#if item.description}
                                            <div class="opacity-50 text-xs">{item.description}</div>
                                        {/if}
                                    </div>
                                {/if}
                            </div>
                        </li>
                    {/each}
                </ul>
            {/if}
        </div>
    </div>
{/if}