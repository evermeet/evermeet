<script lang="ts">
	import { api } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import { bufferToBase64, recursiveBase64ToBuffer } from '$lib/webauthn.js';

	let registering = $state(false);
	let error = $state('');
	let success = $state('');

	async function registerPasskey() {
		registering = true;
		error = '';
		success = '';
		try {
			const { data: options, session } = await api.auth.passkey.registerStart();
			
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

			await api.auth.passkey.registerFinish(finishData, session);
			success = 'Passkey registered successfully!';
		} catch (err: any) {
			console.error(err);
			error = err.message;
		} finally {
			registering = false;
		}
	}
</script>

<main>
	<h1>Settings</h1>

	<section>
		<h2>Passkeys</h2>
		<p class="muted">Add a passkey to sign in without waiting for email links.</p>
		
		{#if success}
			<p class="success">{success}</p>
		{/if}
		{#if error}
			<p class="error">{error}</p>
		{/if}

		<button onclick={registerPasskey} disabled={registering}>
			{registering ? 'Registering…' : 'Register new Passkey'}
		</button>
	</section>

	<section>
		<h2>Your Identity</h2>
		<p><strong>DID:</strong> <code>{auth.user?.did}</code></p>
		<p class="muted">This is your self-sovereign identity. It is permanent across all Evermeet instances.</p>
	</section>
</main>

<style>
	main {
		max-width: 600px;
		margin: 2rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 2rem; }
	section {
		padding: 1.5rem;
		border: 1px solid #eee;
		border-radius: 8px;
		margin-bottom: 1.5rem;
	}
	h2 { font-size: 1.1rem; margin-top: 0; }
	p { font-size: 0.95rem; line-height: 1.5; }
	.muted { color: #666; }
	
	button {
		margin-top: 1rem;
		padding: 0.6rem 1.2rem;
		background: #1a1a1a;
		color: #fff;
		border: none;
		border-radius: 6px;
		font-weight: 600;
		cursor: pointer;
	}
	button:disabled { opacity: 0.5; cursor: default; }
	
	.error { color: #c00; font-size: 0.9rem; margin: 1rem 0; }
	.success { color: #080; font-size: 0.9rem; margin: 1rem 0; }
	code { font-size: 0.8rem; background: #f4f4f4; padding: 0.2rem 0.4rem; border-radius: 4px; word-break: break-all; }
</style>
