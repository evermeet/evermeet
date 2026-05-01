<script lang="ts">
	import { api } from '$lib/api.js';
	import { intl } from '$lib/i18n.svelte.js';

	import { bufferToBase64, recursiveBase64ToBuffer } from '$lib/webauthn.js';

	let email = $state('');
	let sent = $state(false);
	let error = $state('');
	let submitting = $state(false);
	let passkeySupported = $state(false);
	let discoveredHome = $state('');
	let delegateURL = $state('');

	import { onMount } from 'svelte';
	onMount(() => {
		passkeySupported = !!window.PublicKeyCredential;
	});

	async function submit(e: Event) {
		e.preventDefault();
		if (!email) return;
		submitting = true;
		error = '';
		discoveredHome = '';
		delegateURL = '';
		try {
			const eventId = new URLSearchParams(window.location.search).get('event_id') ?? '';
			try {
				const resolved = await api.auth.resolveHome(email, eventId);
				if (resolved.home_instance_url.replace(/\/$/, '') !== window.location.origin.replace(/\/$/, '')) {
					discoveredHome = resolved.home_instance_url;
					delegateURL = resolved.delegate_url;
					return;
				}
			} catch (resolveErr: any) {
				if (!String(resolveErr.message).includes('not found')) {
					throw resolveErr;
				}
			}
			await api.auth.requestMagicLink(email);
			sent = true;
		} catch (err: any) {
			error = err.message;
		} finally {
			submitting = false;
		}
	}

	async function loginWithPasskey() {
		submitting = true;
		error = '';
		try {
			const { data: options, session } = await api.auth.passkey.loginStart(email || undefined);
			const credential: any = await navigator.credentials.get({
				publicKey: recursiveBase64ToBuffer(options.publicKey)
			});

			const finishData = {
				id: credential.id,
				rawId: bufferToBase64(credential.rawId),
				type: credential.type,
				response: {
					authenticatorData: bufferToBase64(credential.response.authenticatorData),
					clientDataJSON: bufferToBase64(credential.response.clientDataJSON),
					signature: bufferToBase64(credential.response.signature),
					userHandle: credential.response.userHandle ? bufferToBase64(credential.response.userHandle) : null
				}
			};

			await api.auth.passkey.loginFinish(finishData, session);
			window.location.href = '/';
		} catch (err: any) {
			console.error(err);
			error = err.name === 'NotAllowedError' ? intl.t('auth.signInCancelled') : err.message;
		} finally {
			submitting = false;
		}
	}

	async function signupWithPasskey() {
		submitting = true;
		error = '';
		try {
			const { data: options, session } = await api.auth.passkey.signupStart();
			const credential: any = await navigator.credentials.create({
				publicKey: recursiveBase64ToBuffer(options.publicKey)
			});

			const finishData = {
				id: credential.id,
				rawId: bufferToBase64(credential.rawId),
				type: credential.type,
				response: {
					attestationObject: bufferToBase64(credential.response.attestationObject),
					clientDataJSON: bufferToBase64(credential.response.clientDataJSON),
				}
			};

			await api.auth.passkey.signupFinish(finishData, session);
			window.location.href = '/';
		} catch (err: any) {
			console.error(err);
			error = err.name === 'NotAllowedError' ? intl.t('auth.signupCancelled') : err.message;
		} finally {
			submitting = false;
		}
	}
</script>

<main>
	{#if sent}
		<h1>{intl.t('auth.checkEmail')}</h1>
		<p>{intl.t('auth.linkSentPrefix')} <strong>{email}</strong>.</p>
		<p class="muted">{intl.t('auth.linkExpires')}</p>
	{:else}
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
			{#if discoveredHome}
				<div class="discovered">
					<p>Your home instance is <strong>{discoveredHome}</strong>.</p>
					<a class="button" href={delegateURL}>Continue there to sign in</a>
				</div>
			{/if}
			<button type="submit" disabled={submitting}>
				{submitting ? intl.t('auth.sending') : intl.t('auth.sendLink')}
			</button>

			{#if passkeySupported}
				<div class="separator">{intl.t('auth.orPasskey')}</div>
				<div class="passkey-btns">
					<button type="button" class="secondary" onclick={loginWithPasskey} disabled={submitting}>
						{intl.t('auth.signInPasskey')}
					</button>
					<button type="button" class="secondary" onclick={signupWithPasskey} disabled={submitting}>
						{intl.t('auth.createPasskey')}
					</button>
				</div>
			{/if}
		</form>
	{/if}
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
	a.button {
		display: block;
		margin-top: 0.5rem;
		padding: 0.7rem;
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border-radius: var(--radius-md);
		text-align: center;
		text-decoration: none;
		font-weight: 600;
	}
	.discovered {
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-md);
		padding: 0.75rem;
	}
	.discovered p { margin-bottom: 0.5rem; }
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
	.passkey-btns { display: flex; flex-direction: column; gap: 0.5rem; }
	.error { color: var(--text-error); font-size: 0.875rem; }
</style>
