<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';

	const id = $page.params.id;

	let title = $state('');
	let description = $state('');
	let starts_at = $state('');
	let ends_at = $state('');
	let locationName = $state('');
	let visibility = $state<'public' | 'unlisted' | 'private'>('public');
	let rsvpLimit = $state(0);

	let loading = $state(true);
	let saving = $state(false);
	let error = $state('');

	onMount(async () => {
		try {
			const res = await api.events.get(id);
			const e = res.state;

			if (auth.user?.did !== e.organizer) {
				error = 'You are not the organizer of this event';
				return;
			}

			title = e.title;
			description = e.description || '';
			starts_at = formatForInput(e.starts_at);
			ends_at = e.ends_at ? formatForInput(e.ends_at) : '';
			locationName = e.location?.name || '';
			visibility = e.visibility;
			rsvpLimit = e.rsvp?.limit || 0;
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	function formatForInput(iso: string) {
		const d = new Date(iso);
		const pad = (n: number) => n.toString().padStart(2, '0');
		return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		saving = true;
		error = '';

		try {
			await api.events.update(id, {
				title,
				description,
				starts_at: new Date(starts_at).toISOString(),
				ends_at: ends_at ? new Date(ends_at).toISOString() : undefined,
				location: locationName ? { name: locationName } : undefined,
				visibility,
				rsvp_limit: rsvpLimit > 0 ? rsvpLimit : undefined,
			});
			goto(`/events/${id}`);
		} catch (e: any) {
			error = e.message;
		} finally {
			saving = false;
		}
	}
</script>

<main>
	<h1>Edit Event</h1>

	{#if loading}
		<p class="muted">Loading event details…</p>
	{:else if error}
		<p class="error">{error}</p>
		<a href="/events/{id}" class="cancel">Back to event</a>
	{:else}
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

			<div class="actions">
				<button type="submit" disabled={saving}>
					{saving ? 'Saving…' : 'Save Changes'}
				</button>
				<a href="/events/{id}" class="cancel">Cancel</a>
			</div>
		</form>
	{/if}
</main>

<style>
	main {
		max-width: 600px;
		margin: 2rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 2rem; color: var(--text); }

	form { display: flex; flex-direction: column; gap: 1.5rem; }
	.field { display: flex; flex-direction: column; gap: 0.4rem; }
	label { font-size: 0.9rem; font-weight: 600; color: var(--text-label); }

	input, select, textarea {
		padding: 0.6rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		font-size: 1rem;
		font-family: inherit;
		background: var(--bg-input);
		color: var(--text);
	}
	input:focus, select:focus, textarea:focus {
		outline: none;
		border-color: var(--border-input-focus);
	}
	textarea { min-height: 100px; resize: vertical; }

	.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }

	.actions { display: flex; align-items: center; gap: 1.5rem; margin-top: 1rem; }

	button {
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border: none;
		padding: 0.7rem 1.5rem;
		border-radius: var(--radius-md);
		font-weight: 600;
		cursor: pointer;
		font-size: 1rem;
	}
	button:disabled { opacity: 0.5; cursor: not-allowed; }

	.cancel { text-decoration: none; color: var(--text-subtle); font-size: 0.9rem; }
	.cancel:hover { color: var(--text); }

	.error { color: var(--text-error); font-size: 0.9rem; margin: 0; }
	.muted { color: var(--text-muted); }

	@media (max-width: 480px) {
		.grid { grid-template-columns: 1fr; }
	}
</style>
