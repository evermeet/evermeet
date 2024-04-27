
<script>
    import { onMount } from "svelte"
    import { xrpcCall } from "../../lib/api";
    import EventBox from "../../components/EventBox.svelte";
    import { fly } from 'svelte/transition';

    let url = ''
    let isLoading
    let results
    let selected = []
    let error

    onMount(() => {
        document.getElementById('url').focus()
    })

    async function formSubmit () {
        results = null
        error = null
        isLoading = true
        let res
        try {
            res = await xrpcCall(fetch, 'app.evermeet.event.importEvents', null, { url })
        } catch (e) {
            error = e.message
        }
        if (res) {
            console.log(results, res)
            results = res
        }
        isLoading = false
    }

    function submitExample (u) {
        if (isLoading) {
            return
        }
        url = u
        formSubmit()
    }
</script>

<div class="page-wide">
    <h1 class="heading1">Import Event(s)</h1>
    <form class="flex gap-2 w-full" on:submit={formSubmit}>
        <div class="grow">
            <input id="url" type="input" class="input input-bordered w-full" disabled={isLoading} placeholder="External URL of event or calendar" bind:value={url} />
        </div>
        <div>
            <button type="submit" class="btn btn-primary" disabled={isLoading}>
                {#if isLoading}
                    <span class="loading loading-infinity"></span>
                {:else}
                    Load
                {/if}
            </button>
        </div>
    </form>
    <div class="mt-2 text-xs opacity-50">
        Examples: 
            <a href="#" class="underline" on:click={() => submitExample('https://lu.ma/uncloud')}>lu.ma (event)</a>,
            <a href="#" class="underline" on:click={() => submitExample('https://lu.ma/berlin-blockchain-week')}>lu.ma (calendar)</a>,
            <a href="#" class="underline" on:click={() => submitExample('https://www.meetup.com/paralelnipolis/events/300284599/')}>meetup.com</a>
    </div>
    {#if error}
        <div class="mt-4 text-error">Error: {error}</div>
    {/if}
    {#if results}
        <div id={results.url} class="mt-8">
            <h2 class="heading2">Events found ({results.inspect?.events?.length})</h2>
            <div>
                {#each results.inspect?.events as e}
                    <div class="flex gap-4 items-center" id={e.remoteId}>
                        <div><input id={e.name + '-input'} type="checkbox" class="checkbox" bind:group={selected} value={e.name} /></div>
                        <div class="grow cursor-pointer" on:click={document.getElementById(e.name + '-input').click()}><EventBox item={e} virtual="true" /></div>
                    </div>
                    <!--div class="mt-4 pre code font-mono text-sm">
                        <pre>{JSON.stringify(results, null, 2)}</pre>
                    </div-->
                {/each}
            </div>
            <div>
                <button class="btn btn-primary" disabled={selected.length === 0}>Import {selected.length} Events</button>
            </div>
        </div>
    {/if}
</div>