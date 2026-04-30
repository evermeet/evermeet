<script lang="ts">
	import { api } from '$lib/api.js';

	import { bufferToBase64, recursiveBase64ToBuffer } from '$lib/webauthn.js';

	let email = $state('');
	let sent = $state(false);
	let error = $state('');
	let submitting = $state(false);
	let passkeySupported = $state(false);

	import { onMount } from 'svelte';
	onMount(() => {
		passkeySupported = !!window.PublicKeyCredential;
	});

	async function submit(e: Event) {
		e.preventDefault();
		if (!email) return;
		submitting = true;
		error = '';
		try {
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
			error = err.name === 'NotAllowedError' ? 'Sign in cancelled.' : err.message;
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
			error = err.name === 'NotAllowedError' ? 'Signup cancelled.' : err.message;
		} finally {
			submitting = false;
		}
	}
</script>

<main>
	{#if sent}
		<h1>Check your email</h1>
		<p>A sign-in link was sent to <strong>{email}</strong>.</p>
		<p class="muted">The link expires in 15 minutes.</p>
	{:else}
		<h1>Sign in</h1>
		<p class="muted">We'll email you a sign-in link — no password needed.</p>

		<form onsubmit={submit}>
			<label for="email">Email</label>
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
				{submitting ? 'Sending…' : 'Send sign-in link'}
			</button>

			{#if passkeySupported}
				<div class="separator">or use passkey</div>
				<div class="passkey-btns">
					<button type="button" class="secondary" onclick={loginWithPasskey} disabled={submitting}>
						Sign in with Passkey
					</button>
					<button type="button" class="secondary" onclick={signupWithPasskey} disabled={submitting}>
						Create account with Passkey
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
