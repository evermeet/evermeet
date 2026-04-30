<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api.js';

	let name = $state('');
	let description = $state('');
	let avatar = $state('');
	let backdrop_url = $state('');
	let website = $state('');

	let loading = $state(false);
	let error = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		loading = true;
		error = '';
		try {
			const res = await api.calendars.create({ name, description, avatar, backdrop_url, website });
			goto(`/calendars/${res.id}`);
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}
</script>

<main>
	<h1>Create Calendar</h1>

	<form onsubmit={handleSubmit}>
		<div class="field">
			<label for="name">Name</label>
			<input type="text" id="name" bind:value={name} required placeholder="My Community Events" />
		</div>

		<div class="field">
			<label for="description">Description</label>
			<textarea id="description" bind:value={description} placeholder="What is this calendar about?"></textarea>
		</div>

		<div class="field">
			<label for="avatar">Avatar URL (optional)</label>
			<input type="url" id="avatar" bind:value={avatar} placeholder="https://…" />
		</div>

		<div class="field">
			<label for="backdrop_url">Backdrop Image URL (optional)</label>
			<input type="url" id="backdrop_url" bind:value={backdrop_url} placeholder="https://…" />
		</div>

		<div class="field">
			<label for="website">Website (optional)</label>
			<input type="url" id="website" bind:value={website} placeholder="https://…" />
		</div>

		{#if error}
			<p class="error">{error}</p>
		{/if}

		<div class="actions">
			<button type="submit" disabled={loading}>
				{loading ? 'Creating…' : 'Create Calendar'}
			</button>
			<a href="/calendars" class="cancel">Cancel</a>
		</div>
	</form>
</main>

<style>
	main {
		max-width: 560px;
		margin: 2.5rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 2rem; color: var(--text); }

	form { display: flex; flex-direction: column; gap: 1.25rem; }
	.field { display: flex; flex-direction: column; gap: 0.4rem; }
	label { font-size: 0.9rem; font-weight: 600; color: var(--text-label); }

	input, textarea {
		padding: 0.6rem 0.75rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		font-size: 0.95rem;
		font-family: inherit;
		background: var(--bg-input);
		color: var(--text);
	}
	input:focus, textarea:focus { outline: none; border-color: var(--border-input-focus); }
	textarea { min-height: 90px; resize: vertical; }

	.actions { display: flex; align-items: center; gap: 1.5rem; margin-top: 0.5rem; }
	button {
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border: none;
		padding: 0.7rem 1.5rem;
		border-radius: var(--radius-md);
		font-weight: 600;
		cursor: pointer;
		font-size: 0.95rem;
	}
	button:disabled { opacity: 0.5; cursor: not-allowed; }
	.cancel { text-decoration: none; color: var(--text-subtle); font-size: 0.9rem; }
	.cancel:hover { color: var(--text); }
	.error { color: var(--text-error); font-size: 0.9rem; margin: 0; }
</style>
