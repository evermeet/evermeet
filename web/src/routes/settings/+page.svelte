<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api.js';
	import type { AuthMethods } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import { theme, themes } from '$lib/theme.svelte.js';
	import { bufferToBase64, recursiveBase64ToBuffer } from '$lib/webauthn.js';
	import Avatar from '$lib/components/Avatar.svelte';
	import ImageUpload from '$lib/components/ImageUpload.svelte';

	let registering = $state(false);
	let savingProfile = $state(false);
	let profileError = $state('');
	let profileSuccess = $state('');
	let passkeyError = $state('');
	let passkeySuccess = $state('');
	let methods = $state<AuthMethods | null>(null);
	let methodsError = $state('');
	let loadingMethods = $state(true);
	let methodsRequested = $state(false);

	let displayName = $state('');
	let bio = $state('');
	let avatar = $state('');
	let populated = $state(false);

	$effect(() => {
		if (!populated && auth.user) {
			displayName = auth.user.display_name || '';
			bio = auth.user.bio || '';
			avatar = auth.user.avatar || '';
			populated = true;
		}
	});

	onMount(() => {
		if (auth.user?.is_local) loadAuthMethods();
	});

	$effect(() => {
		if (auth.user && !auth.user.is_local) {
			loadingMethods = false;
			methodsRequested = true;
		}
		if (auth.user?.is_local && !methodsRequested) {
			loadAuthMethods();
		}
	});

	async function loadAuthMethods() {
		if (!auth.user?.is_local) return;
		methodsRequested = true;
		loadingMethods = true;
		methodsError = '';
		try {
			methods = await api.auth.methods();
		} catch (err: any) {
			methodsError = err.message;
		} finally {
			loadingMethods = false;
		}
	}

	async function updateProfile(e: Event) {
		e.preventDefault();
		savingProfile = true;
		profileError = '';
		profileSuccess = '';
		try {
			await api.auth.updateProfile({
				display_name: displayName,
				bio: bio,
				avatar: avatar
			});
			profileSuccess = 'Profile updated successfully!';
			await auth.load();
		} catch (err: any) {
			profileError = err.message;
		} finally {
			savingProfile = false;
		}
	}

	async function registerPasskey() {
		registering = true;
		passkeyError = '';
		passkeySuccess = '';
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
			passkeySuccess = 'Passkey registered successfully!';
			await loadAuthMethods();
		} catch (err: any) {
			console.error(err);
			passkeyError = err.message;
		} finally {
			registering = false;
		}
	}

	function formatDate(value: string) {
		if (!value) return '';
		return new Date(value).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function shortID(value: string) {
		if (!value || value.length <= 20) return value;
		return `${value.slice(0, 12)}...${value.slice(-8)}`;
	}

	function homeInstanceHost() {
		const url = auth.user?.home_instance_url ?? '';
		if (!url) return '';
		try {
			return new URL(url).host;
		} catch {
			return url;
		}
	}
</script>

<main>
	<h1>Settings</h1>

	{#if auth.user && !auth.user.is_local}
		<section>
			<h2>Remote account</h2>
			<div class="remote-profile">
				<Avatar src={auth.user.avatar} did={auth.user.did} size={72} />
				<div>
					<h3>{auth.user.display_name || 'Anonymous'}</h3>
					<p class="muted">{auth.user.bio || 'No profile bio cached on this instance.'}</p>
				</div>
			</div>
			<div class="method-row">
				<div>
					<strong>Managed on home instance</strong>
					<p>{homeInstanceHost()}</p>
				</div>
				{#if auth.user.home_instance_url}
					<a class="button-link" href={auth.user.home_instance_url}>Open home instance</a>
				{/if}
			</div>
			<p class="muted">This instance can use your signed-in session, but profile and authentication methods are managed by your home instance.</p>
		</section>
	{:else}
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
			{#if profileSuccess}
				<p class="success">{profileSuccess}</p>
			{/if}
			{#if profileError}
				<p class="error">{profileError}</p>
			{/if}
			<button type="submit" disabled={savingProfile}>
				{savingProfile ? 'Saving…' : 'Save Profile'}
			</button>
		</form>
	</section>
	{/if}

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

	{#if !auth.user || auth.user.is_local}
	<section>
		<h2>Authentication methods</h2>
		<p class="muted">These are the ways you can sign in to this Evermeet identity.</p>

		{#if loadingMethods}
			<p class="muted">Loading authentication methods…</p>
		{:else if methodsError}
			<p class="error">{methodsError}</p>
		{:else if methods}
			<div class="auth-methods">
				<div class="method-row">
					<div>
						<strong>Email magic link</strong>
						<p>{methods.email.linked ? methods.email.address : 'No email linked'}</p>
					</div>
					<span class="badge" class:muted-badge={!methods.email.linked}>
						{methods.email.linked ? (methods.email.verified ? 'Verified' : 'Linked') : 'Not linked'}
					</span>
				</div>

				<div class="method-row">
					<div>
						<strong>Ethereum backup login</strong>
						<p>
							{methods.ethereum.linked
								? `${methods.ethereum.address} on chain ${methods.ethereum.chain_id}`
								: 'No wallet linked'}
						</p>
					</div>
					<span class="badge" class:muted-badge={!methods.ethereum.linked}>
						{methods.ethereum.linked ? 'Linked' : 'Not linked'}
					</span>
				</div>

				<div class="method-block">
					<div class="method-heading">
						<div>
							<strong>Passkeys</strong>
							<p>{methods.passkeys.length} registered</p>
						</div>
					</div>
					{#if methods.passkeys.length > 0}
						<div class="passkey-list">
							{#each methods.passkeys as passkey}
								<div class="passkey-card">
									<div>
										<strong title={passkey.id}>{shortID(passkey.id)}</strong>
										<p>Created {formatDate(passkey.created_at)}</p>
									</div>
									<div class="passkey-meta">
										<span>Counter {passkey.counter}</span>
										{#if passkey.backup_eligible}
											<span>{passkey.backup_state ? 'Synced backup' : 'Backup eligible'}</span>
										{/if}
										{#if passkey.user_verified}
											<span>User verified</span>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					{:else}
						<p class="muted">No passkeys registered yet.</p>
					{/if}
				</div>
			</div>
		{/if}

		<h3>Add passkey</h3>
		<p class="muted">Add a passkey to sign in without waiting for email links on this instance.</p>

		{#if passkeySuccess}
			<p class="success">{passkeySuccess}</p>
		{/if}
		{#if passkeyError}
			<p class="error">{passkeyError}</p>
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
	{:else}
	<section>
		<h2>Your Identity</h2>
		<p><strong>DID:</strong> <code>{auth.user.did}</code></p>
		<p><strong>Session type:</strong> Remote delegated session</p>
		<p class="muted">Log out here to end only this instance session. Your home-instance session is separate.</p>
	</section>
	{/if}
</main>

<style>
	main {
		max-width: var(--layout-page-narrow);
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
	h3 { font-size: 0.95rem; margin: 1.25rem 0 0.25rem; color: var(--text); }
	p { font-size: 0.95rem; line-height: 1.5; color: var(--text); }
	.muted { color: var(--text-subtle); }
	.remote-profile {
		display: flex;
		gap: 1rem;
		align-items: center;
		margin-bottom: 1rem;
	}
	.remote-profile h3 {
		margin-top: 0;
		font-size: 1rem;
	}

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
	.button-link {
		flex-shrink: 0;
		padding: 0.45rem 0.7rem;
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border-radius: var(--radius-md);
		font-size: 0.85rem;
		font-weight: 600;
		text-decoration: none;
	}

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

	.auth-methods {
		display: grid;
		gap: 0.75rem;
		margin-top: 1rem;
	}
	.method-row,
	.method-block {
		padding: 1rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-lg);
		background: var(--bg);
	}
	.method-row {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}
	.method-row p,
	.method-block p,
	.passkey-card p {
		margin: 0.2rem 0 0;
		color: var(--text-subtle);
		word-break: break-word;
	}
	.method-heading {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
	}
	.badge {
		flex-shrink: 0;
		padding: 0.2rem 0.5rem;
		border-radius: var(--radius-full);
		background: var(--bg-success);
		color: var(--text-success);
		font-size: 0.75rem;
		font-weight: 700;
	}
	.muted-badge {
		background: var(--bg-raised);
		color: var(--text-muted);
	}
	.passkey-list {
		display: grid;
		gap: 0.5rem;
		margin-top: 0.75rem;
	}
	.passkey-card {
		display: grid;
		gap: 0.5rem;
		padding: 0.75rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-md);
		background: var(--bg-subtle);
	}
	.passkey-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 0.4rem;
	}
	.passkey-meta span {
		padding: 0.15rem 0.45rem;
		border-radius: var(--radius-full);
		background: var(--bg-raised);
		color: var(--text-subtle);
		font-size: 0.75rem;
	}

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
