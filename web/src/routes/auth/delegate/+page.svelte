<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Avatar from '$lib/components/Avatar.svelte';
	import { api, type ForeignInstanceSig, type SignedDelegationToken } from '$lib/api.js';
	import { intl } from '$lib/i18n.svelte.js';

	let status = $state<'loading' | 'consent' | 'posting' | 'error'>('loading');
	let error = $state('');
	let returnTo = '';
	let nonce = '';
	let eventId = '';
	let method = '';
	let email = '';
	let foreignSig: ForeignInstanceSig | null = null;
	let currentUser = $state<{ did: string; display_name: string; avatar: string; bio: string } | null>(null);

	onMount(async () => {
		const params = new URLSearchParams(window.location.search);
		returnTo = params.get('return_to') ?? '';
		nonce = params.get('nonce') ?? '';
		eventId = params.get('event_id') ?? '';
		method = params.get('method') ?? '';
		email = params.get('email') ?? '';
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
			currentUser = await api.auth.me();
			status = 'consent';
		} catch {
			redirectToVerification();
		}
	});

	function redirectToVerification() {
		const next = `${window.location.pathname}${window.location.search}`;
		const target = new URLSearchParams({ next, auto: '1' });
		if (method === 'email' && email) {
			target.set('method', 'email');
			target.set('email', email);
			if (eventId) target.set('event_id', eventId);
			goto(`/auth/instance?${target.toString()}`, { replaceState: true });
			return;
		}
		goto(`/auth/login?${target.toString()}`, { replaceState: true });
	}

	async function approveSignIn() {
		if (!foreignSig) return;
		status = 'posting';
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
				redirectToVerification();
				return;
			}
			status = 'error';
			error = err.message;
		}
	}

	function requesterHost() {
		try {
			return new URL(returnTo).host;
		} catch {
			return returnTo || intl.t('auth.unknownInstance');
		}
	}

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
	{:else if status === 'consent'}
		<h1>{intl.t('auth.approveSignInTitle')}</h1>
		<p class="muted">{intl.t('auth.approveSignInHelp')}</p>
		<div class="consent-card">
			<div class="requester">
				<Avatar did={requesterHost()} size={56} rounded={false} />
				<div>
					<p class="eyebrow">{intl.t('auth.requestingInstance')}</p>
					<h2>{requesterHost()}</h2>
				</div>
			</div>
			<div class="identity">
				<span>{intl.t('auth.signInAs')}</span>
				<strong>{currentUser?.display_name || currentUser?.did}</strong>
				{#if currentUser?.display_name}<code>{currentUser.did}</code>{/if}
			</div>
			<p class="muted">{intl.t('auth.approveSignInDescription')}</p>
			<button type="button" onclick={approveSignIn}>{intl.t('auth.allowSignIn')}</button>
		</div>
	{:else if status === 'posting'}
		<h1>Returning to event</h1>
		<p>Signing you in on the event instance…</p>
	{:else}
		<h1>Could not authorize</h1>
		<p class="error">{error}</p>
	{/if}
</main>

<style>
	main {
		max-width: var(--layout-page-narrow);
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
	button {
		width: 100%;
		padding: 0.7rem;
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
	}
	.muted {
		color: var(--text-muted);
	}
	.consent-card {
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-xl);
		padding: 1.25rem;
		margin-top: 1rem;
		background: var(--bg-card);
	}
	.requester {
		display: flex;
		gap: 1rem;
		align-items: center;
		margin-bottom: 1rem;
	}
	.requester h2 {
		margin: 0.1rem 0 0;
		color: var(--text);
		font-size: 1.25rem;
		word-break: break-word;
	}
	.requester p {
		margin: 0;
	}
	.eyebrow {
		color: var(--text-muted);
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}
	.identity {
		display: grid;
		gap: 0.2rem;
		padding: 0.85rem;
		margin-bottom: 0.75rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-md);
		background: var(--bg);
	}
	.identity span {
		color: var(--text-muted);
		font-size: 0.75rem;
		font-weight: 600;
	}
	.identity strong {
		color: var(--text);
		word-break: break-word;
	}
	code {
		color: var(--text-muted);
		font-size: 0.75rem;
		word-break: break-all;
	}
	.error {
		color: var(--text-error);
	}
</style>
