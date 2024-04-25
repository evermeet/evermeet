
<script>
    import { User } from 'svelte-heros-v2';
    import Cookies from 'js-cookie'
    import { writable } from 'svelte/store';
    import { goto } from '$app/navigation';

    import { pkg } from '../../lib/config.js';
    import { xrpcCall } from '../../lib/api.js';
    import { user, config, session } from '$lib/stores';
    import InstanceSelector from '../../components/InstanceSelector.svelte';
    import { onMount } from 'svelte';

    onMount(() => {
        document.querySelector("#identifier").focus()
    })

    const credentials = writable({ visibility: 'public' })
    const instance = writable($config.domain)
    let isProcessing = false;
    $: localInstance = $instance === $config.domain
    $: showPassword = localInstance && $credentials.identifier?.length >= 3
    $: validCredentials = $credentials.identifier && (!localInstance || (localInstance && $credentials.password?.length >= 3))

    credentials.subscribe((x) => {
        const domain = x.identifier ? x.identifier?.split('.').slice(1).join('.').toLowerCase() : null
        if (domain) {
            if (instances.find(i => i.domain === domain)) {
                instance.set(domain)
                credentials.update(c => {
                    c.identifier = c.identifier.split('.').slice(0,1).join('.')
                    return c
                })
            } else {
                instance.set('Other Instance')
            }
        }
    })

    async function submitLogin() {
        if (!validCredentials) {
            return false
        }
        isProcessing = true;
        const ident = $credentials.identifier
        const localIdent = !ident.match(/\./)
        const normalizedQuery = {
            identifier: ident.match(/\./) ? ident : (ident + '.' + $config.domain),
            password: $credentials.password,
        }
        if (!localIdent) {
            const resp = await xrpcCall(fetch, 'app.evermeet.identity.resolveHandle', { handle: ident })
            console.log(resp)
            isProcessing = false
            return;
        }

        let s;
        try {
            s = await xrpcCall(fetch, 'app.evermeet.auth.createSession', null, normalizedQuery)
        } catch (e) {}

        isProcessing = false;
        if (!s.did) {
            return false;
        }

        Cookies.set('evermeet-session-id', s.accessJwt)
        session.set(s)

        goto('/events');

        return {}
    }

    const instances = [
        { domain: $config.domain, local: true },
        { domain: 'utxo.events' },
        { domain: 'events.gwei.cz' },
        { domain: 'jednadvacet.org'},
        { domain: 'dev.evermeet.app' },
        { domain: 'Other Instance' },
    ]

</script>

<div class="w-[26rem] m-auto my-24 itembox p-6 shadow-xl">
    <User size="50" />
    <div class="text-2xl mt-4">Welcome!</div>
    <div class="mt-2 text-base-content/75">Please sign in below or <a href="/register" class="underline hover:no-underline">sign up</a>.</div>

    <form class="mt-6" on:submit|preventDefault={submitLogin}>

        <div class="flex items-center">
            <label for="identifier" class="grow">Handle / DID</label>
            <div>
                <InstanceSelector {instances} bind={instance} />
            </div>
        </div>
        <div class="mt-2 mb-2 relative">
            <input id="identifier" type="text" placeholder="your-handle{$instance === 'Other Instance' ? '.my-instance.com' : ''}" class="input input-bordered {isProcessing ? 'input-disabled' : ''} w-full" bind:value={$credentials.identifier} />
            <span class="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-500 text-sm">{$instance === 'Other Instance' ? '' : '.' + $instance}</span>
        </div>
        {#if showPassword}
            <label for="password" class="">Password</label>
            <div class="mt-2">
                <input id="password" type="password" placeholder="Password" class="input input-bordered {isProcessing ? 'input-disabled' : ''} w-full" bind:value={$credentials.password} />
            </div>
        {/if}
        <div class="mt-4">
            <button class="btn btn-primary {!validCredentials || isProcessing ? 'btn-disabled' : ''} w-full">
                {#if !isProcessing}
                    Login {#if $instance != $config.domain}<span class="font-mono">on {$instance}</span>{/if}
                {:else}
                    <span class="loading loading-infinity loading-lg"></span>
                {/if}
            </button>
        </div>
    </form>
</div>
