
<script>
    import { UserPlus } from 'svelte-heros-v2';
    import { pkg } from '../../lib/config.js';
    import { apiCall } from '../../lib/api.js';
    import { writable } from 'svelte/store';
    import { goto } from '$app/navigation';
    import { user, config } from '$lib/stores';
    import Cookies from 'js-cookie';

    const email = writable('');
    let validEmail = false;
    let isProcessing = false;

    email.subscribe((x) => {
        validEmail = Boolean(x.match(/^[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/i));
    })

    async function submitLogin() {
        if (!validEmail) {
            return false
        }
        isProcessing = true;
        const resp = await apiCall(fetch, 'login', {
            method: 'post',
            body: JSON.stringify({
                email: $email
            }),
            headers: {
                'content-type': 'application/json'
            },
        })
        isProcessing = false;
        if (!resp.user) {
            return false;
        }

        Cookies.set('evermeet-session-id', resp.sessionId, { expires: 1000, secure: true, sameSite: 'Lax' })
        user.set(resp.user)
        goto('/events');
    }

</script>

<div class="w-[26rem] m-auto my-24 itembox p-6 shadow-xl">
    <UserPlus size="50" />
    <div class="text-2xl mt-4">Welcome to <span class="font-semibold font-mono">{$config.sitename || $config.domain}</span>!</div>
    <div class="mt-2 text-neutral-content">Please sign in or sign up below.</div>

    <form class="mt-6" on:submit|preventDefault={submitLogin}>
        <div>
            <label for="email" class="">Handle / DID / Email</label>
        </div>
        <div class="mt-2">
            <input id="email" type="text" placeholder="you@example.org" class="input input-bordered {isProcessing ? 'input-disabled' : ''} w-full" bind:value={$email} />
        </div>
        <div class="mt-4">
            <button class="btn btn-primary {!validEmail || isProcessing ? 'btn-disabled' : ''} w-full">
                {#if !isProcessing}
                    Login
                {:else}
                    <span class="loading loading-infinity loading-lg"></span>
                {/if}
            </button>
        </div>
    </form>
</div>