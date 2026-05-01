<script lang="ts">
	import { api } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import { theme, themes } from '$lib/theme.svelte.js';
	import { bufferToBase64, recursiveBase64ToBuffer } from '$lib/webauthn.js';
	import ImageUpload from '$lib/components/ImageUpload.svelte';

	let registering = $state(false);
	let savingProfile = $state(false);
	let error = $state('');
	let success = $state('');

	let displayName = $state(auth.user?.display_name || '');
	let bio = $state(auth.user?.bio || '');
	let avatar = $state(auth.user?.avatar || '');

	async function updateProfile(e: Event) {
		e.preventDefault();
		savingProfile = true;
		error = '';
		success = '';
		try {
			await api.auth.updateProfile({
				display_name: displayName,
				bio: bio,
				avatar: avatar
			});
			success = 'Profile updated successfully!';
			await auth.load();
		} catch (err: any) {
			error = err.message;
		} finally {
			savingProfile = false;
		}
	}

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
		<h2>Profile</h2>
		<form onsubmit={updateProfile}>
			<div class="field">
				<label for="display_name">Display Name</label>
				<input type="text" id="display_name" bind:value={displayName} placeholder="Your Name" />
			</div>
			<div class="field">
				<label for="bio">Bio</label>
				<textarea id="bio" bind:value={bio} placeholder="Tell us about yourself…"></textarea>
			</div>
			<div class="field">
				<span class="field-label">Avatar</span>
				<ImageUpload bind:value={avatar} rounded={true} previewSize={120} />
			</div>
			<button type="submit" disabled={savingProfile}>
				{savingProfile ? 'Saving…' : 'Save Profile'}
			</button>
		</form>
	</section>

	<section>
		<h2>Appearance</h2>
		<p class="muted">Choose a theme for the interface.</p>
		<div class="theme-grid">
			{#each themes as t}
				<button
					type="button"
					class="theme-btn"
					class:active={theme.current === t.id}
					onclick={() => theme.apply(t.id)}
				>
					<span class="theme-name">{t.name}</span>
					<span class="theme-desc">{t.description}</span>
				</button>
			{/each}
		</div>
	</section>

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
	h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 2rem; color: var(--text); }
	section {
		padding: 1.5rem;
		border: 1px solid var(--border-card);
		border-radius: var(--radius-lg);
		margin-bottom: 1.5rem;
		background: var(--bg-card);
	}
	h2 { font-size: 1.1rem; margin-top: 0; color: var(--text); }
	p { font-size: 0.95rem; line-height: 1.5; color: var(--text); }
	.muted { color: var(--text-subtle); }

	form { display: flex; flex-direction: column; gap: 1rem; }
	.field { display: flex; flex-direction: column; gap: 0.3rem; }
	label, .field-label { font-size: 0.85rem; font-weight: 600; color: var(--text-label); }
	input, textarea {
		padding: 0.6rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		font-size: 0.95rem;
		font-family: inherit;
		background: var(--bg-input);
		color: var(--text);
	}
	input:focus, textarea:focus {
		outline: none;
		border-color: var(--border-input-focus);
	}
	textarea { min-height: 80px; resize: vertical; }

	button {
		margin-top: 1rem;
		padding: 0.6rem 1.2rem;
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border: none;
		border-radius: var(--radius-md);
		font-weight: 600;
		cursor: pointer;
		font-size: 0.95rem;
	}
	button:disabled { opacity: 0.5; cursor: default; }

	.theme-grid {
		display: flex;
		gap: 0.75rem;
		flex-wrap: wrap;
		margin-top: 0.75rem;
	}
	.theme-btn {
		margin-top: 0;
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 0.2rem;
		padding: 0.75rem 1rem;
		background: var(--bg);
		color: var(--text);
		border: 2px solid var(--border-input);
		border-radius: var(--radius-lg);
		cursor: pointer;
		font-size: 0.875rem;
		transition: border-color 0.1s;
		min-width: 120px;
	}
	.theme-btn:hover { border-color: var(--border-input-focus); }
	.theme-btn.active { border-color: var(--border-input-focus); background: var(--bg-raised); }
	.theme-name { font-weight: 600; }
	.theme-desc { font-size: 0.75rem; color: var(--text-muted); font-weight: 400; }

	.error { color: var(--text-error); font-size: 0.9rem; margin: 1rem 0; }
	.success { color: var(--text-success); font-size: 0.9rem; margin: 1rem 0; }
	code {
		font-size: 0.8rem;
		background: var(--bg-code);
		color: var(--text-code);
		padding: 0.2rem 0.4rem;
		border-radius: var(--radius-sm);
		word-break: break-all;
	}
</style>
