<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type ForeignInstanceSig, type SignedDelegationToken } from '$lib/api.js';

	let status = $state<'loading' | 'posting' | 'login' | 'error'>('loading');
	let error = $state('');
	let returnTo = '';
	let nonce = '';
	let eventId = '';
	let foreignSig: ForeignInstanceSig | null = null;

	onMount(async () => {
		const params = new URLSearchParams(window.location.search);
		returnTo = params.get('return_to') ?? '';
		nonce = params.get('nonce') ?? '';
		eventId = params.get('event_id') ?? '';
		const rawForeignSig = params.get('foreign_sig') ?? '';

		if (!returnTo || !nonce || !rawForeignSig) {
			status = 'error';
			error = 'Missing delegation parameters.';
			return;
		}

		try {
			foreignSig = JSON.parse(rawForeignSig);
		} catch {
			status = 'error';
			error = 'Invalid delegation signature.';
			return;
		}

		try {
			const signed = await api.auth.delegate({
				return_to: returnTo,
				nonce,
				event_id: eventId,
				foreign_sig: foreignSig
			});
			postToForeignInstance(signed);
		} catch (err: any) {
			if (String(err.message).includes('authentication required')) {
				status = 'login';
				return;
			}
			status = 'error';
			error = err.message;
		}
	});

	function postToForeignInstance(signed: SignedDelegationToken) {
		status = 'posting';
		const form = document.createElement('form');
		form.method = 'POST';
		form.action = `${returnTo.replace(/\/$/, '')}/api/auth/delegate-verify`;
		form.style.display = 'none';

		const payload = document.createElement('input');
		payload.type = 'hidden';
		payload.name = 'payload';
		payload.value = JSON.stringify(signed);
		form.appendChild(payload);

		document.body.appendChild(form);
		form.submit();
	}
</script>

<main>
	{#if status === 'loading'}
		<h1>Authorizing sign-in</h1>
		<p>Preparing your delegation token…</p>
	{:else if status === 'posting'}
		<h1>Returning to event</h1>
		<p>Signing you in on the event instance…</p>
	{:else if status === 'login'}
		<h1>Sign in required</h1>
		<p>Please sign in on this home instance, then reopen the delegation link.</p>
		<a href="/auth/login">Sign in</a>
	{:else}
		<h1>Could not authorize</h1>
		<p class="error">{error}</p>
	{/if}
</main>

<style>
	main {
		max-width: 420px;
		margin: 4rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 {
		font-size: 1.5rem;
		font-weight: 700;
		margin-bottom: 0.5rem;
		color: var(--text);
	}
	p {
		color: var(--text-secondary);
	}
	a {
		color: var(--text-link);
	}
	.error {
		color: var(--text-error);
	}
</style>
