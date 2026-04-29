<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api.js';

	let title = $state('');
	let description = $state('');
	let starts_at = $state('');
	let ends_at = $state('');
	let locationName = $state('');
	let visibility = $state<'public' | 'unlisted' | 'private'>('public');
	let rsvpLimit = $state(0);

	let loading = $state(false);
	let error = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		loading = true;
		error = '';

		try {
			const res = await api.events.create({
				title,
				description,
				starts_at: new Date(starts_at).toISOString(),
				ends_at: ends_at ? new Date(ends_at).toISOString() : undefined,
				location: locationName ? { name: locationName } : undefined,
				visibility,
				rsvp_limit: rsvpLimit > 0 ? rsvpLimit : undefined,
			});
			goto(`/events/${res.id}`);
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	}
</script>

<main>
	<h1>Create Event</h1>

	<form onsubmit={handleSubmit}>
		<div class="field">
			<label for="title">Title</label>
			<input type="text" id="title" bind:value={title} required placeholder="My Awesome Meetup" />
		</div>

		<div class="field">
			<label for="description">Description</label>
			<textarea id="description" bind:value={description} placeholder="Tell us about it…"></textarea>
		</div>

		<div class="grid">
			<div class="field">
				<label for="starts_at">Starts At</label>
				<input type="datetime-local" id="starts_at" bind:value={starts_at} required />
			</div>
			<div class="field">
				<label for="ends_at">Ends At (optional)</label>
				<input type="datetime-local" id="ends_at" bind:value={ends_at} />
			</div>
		</div>

		<div class="field">
			<label for="location">Location Name</label>
			<input type="text" id="location" bind:value={locationName} placeholder="The Coffee Shop / Zoom" />
		</div>

		<div class="grid">
			<div class="field">
				<label for="visibility">Visibility</label>
				<select id="visibility" bind:value={visibility}>
					<option value="public">Public (on homepage)</option>
					<option value="unlisted">Unlisted (direct link only)</option>
					<option value="private">Private (Phase 8 feature)</option>
				</select>
			</div>
			<div class="field">
				<label for="rsvp_limit">RSVP Limit (0 for no limit)</label>
				<input type="number" id="rsvp_limit" bind:value={rsvpLimit} min="0" />
			</div>
		</div>

		{#if error}
			<p class="error">{error}</p>
		{/if}

		<div class="actions">
			<button type="submit" disabled={loading}>
				{loading ? 'Creating…' : 'Create Event'}
			</button>
			<a href="/" class="cancel">Cancel</a>
		</div>
	</form>
</main>

<style>
	main {
		max-width: 600px;
		margin: 2rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 2rem; }
	
	form { display: flex; flex-direction: column; gap: 1.5rem; }
	.field { display: flex; flex-direction: column; gap: 0.4rem; }
	label { font-size: 0.9rem; font-weight: 600; color: #444; }
	
	input, select, textarea {
		padding: 0.6rem;
		border: 1px solid #ddd;
		border-radius: 6px;
		font-size: 1rem;
		font-family: inherit;
	}
	textarea { min-height: 100px; resize: vertical; }
	
	.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
	
	.actions { display: flex; align-items: center; gap: 1.5rem; margin-top: 1rem; }
	
	button {
		background: #111;
		color: #fff;
		border: none;
		padding: 0.7rem 1.5rem;
		border-radius: 6px;
		font-weight: 600;
		cursor: pointer;
	}
	button:disabled { opacity: 0.5; cursor: not-allowed; }
	
	.cancel { text-decoration: none; color: #666; font-size: 0.9rem; }
	.cancel:hover { color: #111; }
	
	.error { color: #c00; font-size: 0.9rem; margin: 0; }

	@media (max-width: 480px) {
		.grid { grid-template-columns: 1fr; }
	}
</style>
