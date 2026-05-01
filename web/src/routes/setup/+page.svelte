<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import { onMount } from 'svelte';

	let token = $state('');
	let email = $state('');
	let displayName = $state('');
	let loading = $state(true);
	let submitting = $state(false);
	let error = $state('');

	onMount(async () => {
		try {
			const status = await api.setup.status();
			if (!status.required) {
				goto('/');
				return;
			}
		} catch (err: any) {
			error = err.message ?? 'Unable to check setup status';
		} finally {
			loading = false;
		}
	});

	async function completeSetup(e: Event) {
		e.preventDefault();
		error = '';
		submitting = true;
		try {
			await api.setup.complete({
				token,
				email,
				display_name: displayName || undefined,
			});
			await auth.load();
			if (!auth.user) {
				throw new Error('Admin account was created, but sign-in did not complete. Please sign in.');
			}
			goto('/');
		} catch (err: any) {
			error = err.message ?? 'Setup failed';
		} finally {
			submitting = false;
		}
	}
</script>

<main>
	<section class="card">
		<p class="eyebrow">First-Time Setup</p>
		<h1>Welcome to your Evermeet instance!</h1>
		<p class="muted">
			This instance does not have an admin account yet. Paste the one-time admin token from the server console to create the first admin account.
		</p>

		{#if loading}
			<p class="muted">Checking setup status...</p>
		{:else}
			<form onsubmit={completeSetup}>
				<label for="token">Admin token</label>
				<input
					id="token"
					type="password"
					bind:value={token}
					placeholder="Paste setup token"
					autocomplete="one-time-code"
					required
				/>

				<label for="email">Admin email</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					placeholder="you@example.com"
					autocomplete="email"
					required
				/>

				<label for="name">Display name</label>
				<input
					id="name"
					type="text"
					bind:value={displayName}
					placeholder="Your name"
					autocomplete="name"
				/>

				{#if error}
					<p class="error">{error}</p>
				{/if}

				<button type="submit" disabled={submitting}>
					{submitting ? 'Creating admin account...' : 'Create admin account'}
				</button>
			</form>
		{/if}
	</section>
</main>

<style>
	main {
		min-height: calc(100vh - 180px);
		display: grid;
		place-items: center;
		padding: 2rem 1.5rem;
	}

	.card {
		width: min(100%, 520px);
		padding: 2rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-xl);
		background: var(--bg-card);
		box-shadow: var(--shadow-md);
	}

	.eyebrow {
		margin: 0 0 0.5rem;
		color: var(--text-muted);
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	h1 {
		margin: 0 0 0.75rem;
		color: var(--text);
		font-size: 1.75rem;
	}

	.muted {
		color: var(--text-secondary);
		line-height: 1.5;
	}

	form {
		display: flex;
		flex-direction: column;
		gap: 0.65rem;
		margin-top: 1.5rem;
	}

	label {
		color: var(--text);
		font-size: 0.875rem;
		font-weight: 600;
	}

	input {
		padding: 0.75rem 0.85rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		background: var(--bg-input);
		color: var(--text);
		font-size: 1rem;
	}

	input:focus {
		border-color: var(--border-input-focus);
		outline: none;
	}

	button {
		margin-top: 0.5rem;
		padding: 0.8rem;
		border: none;
		border-radius: var(--radius-md);
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		font-size: 1rem;
		font-weight: 700;
		cursor: pointer;
	}

	button:disabled {
		opacity: 0.6;
		cursor: default;
	}

	.error {
		margin: 0.35rem 0 0;
		color: var(--text-error);
		font-size: 0.9rem;
	}
</style>
