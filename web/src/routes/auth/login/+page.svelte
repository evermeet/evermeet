<script lang="ts">
	import { goto } from '$app/navigation';
	import { intl } from '$lib/i18n.svelte.js';

	declare global {
		interface Window {
			ethereum?: {
				request(args: { method: string; params?: unknown[] }): Promise<unknown>;
			};
		}
	}

	let email = $state('');
	let error = $state('');
	let submitting = $state(false);
	let ethereumSupported = $state(false);

	import { onMount } from 'svelte';
	onMount(() => {
		ethereumSupported = !!window.ethereum;
	});

	async function submit(e: Event) {
		e.preventDefault();
		if (!email) return;
		submitting = true;
		error = '';
		try {
			const eventId = new URLSearchParams(window.location.search).get('event_id') ?? '';
			const params = new URLSearchParams({ method: 'email', email });
			if (eventId) params.set('event_id', eventId);
			goto(`/auth/instance?${params.toString()}`);
		} catch (err: any) {
			error = err.message;
		} finally {
			submitting = false;
		}
	}

	function loginWithEthereum() {
		const eventId = new URLSearchParams(window.location.search).get('event_id') ?? '';
		const params = new URLSearchParams({ method: 'ethereum' });
		if (eventId) params.set('event_id', eventId);
		goto(`/auth/instance?${params.toString()}`);
	}
</script>

<main>
	<h1>{intl.t('auth.signIn')}</h1>
	<p class="muted">{intl.t('auth.emailLinkHelp')}</p>

	<form onsubmit={submit}>
		<label for="email">{intl.t('common.email')}</label>
		<input
			id="email"
			type="email"
			bind:value={email}
			placeholder="you@example.com"
			autocomplete="email"
			required
		/>
		{#if error}<p class="error">{error}</p>{/if}
		<button type="submit" disabled={submitting}>
			{submitting ? intl.t('auth.lookingUp') : intl.t('auth.findHomeServer')}
		</button>

		{#if ethereumSupported}
			<div class="separator">{intl.t('auth.orWallet')}</div>
			<button type="button" class="secondary" onclick={loginWithEthereum} disabled={submitting}>
				{intl.t('auth.signInEthereum')}
			</button>
		{/if}
	</form>
</main>

<style>
	main {
		max-width: 400px;
		margin: 4rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 0.5rem; color: var(--text); }
	p { color: var(--text-secondary); margin: 0.25rem 0 1rem; }
	.muted { color: var(--text-muted); }
	form { display: flex; flex-direction: column; gap: 0.5rem; }
	label { font-size: 0.875rem; font-weight: 500; color: var(--text); }
	input {
		padding: 0.6rem 0.75rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		font-size: 1rem;
		outline: none;
		background: var(--bg-input);
		color: var(--text);
	}
	input:focus { border-color: var(--border-input-focus); }
	button {
		margin-top: 0.5rem;
		padding: 0.7rem;
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
	}
	button:disabled { opacity: 0.5; cursor: default; }
	button.secondary {
		background: var(--bg-btn-secondary);
		color: var(--text-btn-secondary);
		border: 1px solid var(--border-input);
	}
	button.secondary:hover { background: var(--bg-hover); }
	.separator {
		text-align: center;
		font-size: 0.8rem;
		color: var(--text-muted);
		margin: 0.5rem 0;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.separator::before, .separator::after {
		content: "";
		flex: 1;
		height: 1px;
		background: var(--border-subtle);
	}
	.error { color: var(--text-error); font-size: 0.875rem; }
</style>
