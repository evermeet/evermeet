<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
import { api, type Calendar } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import ImageUpload from '$lib/components/ImageUpload.svelte';

	const id = $page.params.id;

	let title = $state('');
	let description = $state('');
	let cover_url = $state('');
	let starts_at = $state('');
	let ends_at = $state('');
	let locationName = $state('');
	let visibility = $state<'public' | 'unlisted' | 'private'>('public');
	let rsvpLimit = $state(0);
let calendars = $state<Calendar[]>([]);
let calendarId = $state('');
let owners = $state<string[]>([]);

	let loading = $state(true);
	let saving = $state(false);
let deleting = $state(false);
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
			cover_url = e.cover_url || '';
			starts_at = formatForInput(e.starts_at);
			ends_at = e.ends_at ? formatForInput(e.ends_at) : '';
			locationName = e.location?.name || '';
			visibility = e.visibility;
			rsvpLimit = e.rsvp?.limit || 0;
			calendarId = e.calendar ?? '';
			owners = (e.governance?.owners ?? []).map((o: any) => o.did).filter(Boolean);
			if (owners.length === 0 && auth.user?.did) {
				owners = [auth.user.did];
			}

			const calRes = await api.calendars.list();
			calendars = calRes.owned ?? [];
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
			const cleanedOwners = owners.map((o) => o.trim()).filter(Boolean);
			if (cleanedOwners.length === 0) {
				throw new Error('At least one host is required');
			}
			await api.events.update(id, {
				title,
				description,
				cover_url: cover_url || undefined,
				starts_at: new Date(starts_at).toISOString(),
				ends_at: ends_at ? new Date(ends_at).toISOString() : undefined,
				location: locationName ? { name: locationName } : undefined,
				calendar_id: calendarId,
				owners: cleanedOwners,
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

	async function handleDelete() {
		const confirmed = window.confirm('Delete this event permanently? This cannot be undone.');
		if (!confirmed) return;
		deleting = true;
		error = '';
		try {
			await api.events.delete(id);
			goto('/');
		} catch (e: any) {
			error = e.message;
		} finally {
			deleting = false;
		}
	}

	function addOwner() {
		owners = [...owners, ''];
	}

	function removeOwner(index: number) {
		if (owners.length <= 1) return;
		owners = owners.filter((_, i) => i !== index);
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
				<textarea id="description" bind:value={description} placeholder="Tell us about it… (Markdown supported)"></textarea>
			</div>

			<div class="field">
				<span class="field-label">Cover Image (optional)</span>
				<ImageUpload bind:value={cover_url} />
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

			<div class="field migrate-section">
				<label for="calendar_id">Migrate event</label>
				<select id="calendar_id" bind:value={calendarId}>
					<option value="">Personal event (no calendar)</option>
					{#each calendars as cal}
						<option value={cal.id}>{cal.name}</option>
					{/each}
				</select>
				<p class="muted">Only calendars where you are an owner are available.</p>
			</div>

			<div class="field hosts-section">
				<div class="hosts-header">
					<p class="hosts-title">Hosts (owners)</p>
					<button type="button" class="btn-add-owner" onclick={addOwner}>+ Add host</button>
				</div>
				<div class="owner-list">
					{#each owners as owner, i}
						<div class="owner-row">
							<input type="text" bind:value={owners[i]} placeholder="did:em:..." />
							<button type="button" class="btn-remove-owner" onclick={() => removeOwner(i)} disabled={owners.length <= 1}>
								Remove
							</button>
						</div>
					{/each}
				</div>
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

			<div class="danger-zone">
				<p class="danger-title">Delete event</p>
				<p class="muted">This permanently removes the event and its RSVP data.</p>
				<button type="button" class="delete-btn" onclick={handleDelete} disabled={deleting}>
					{deleting ? 'Deleting…' : 'Delete Event'}
				</button>
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
	label, .field-label { font-size: 0.9rem; font-weight: 600; color: var(--text-label); }

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
	.migrate-section {
		padding: 0.8rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-md);
		background: var(--bg-subtle);
	}
	.hosts-section {
		padding: 0.8rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-md);
		background: var(--bg-subtle);
	}
	.hosts-header { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; }
	.hosts-title {
		margin: 0;
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--text-label);
	}
	.owner-list { display: flex; flex-direction: column; gap: 0.5rem; }
	.owner-row { display: flex; gap: 0.5rem; }
	.btn-add-owner, .btn-remove-owner {
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		background: var(--bg-raised);
		color: var(--text);
		padding: 0.35rem 0.6rem;
		font-size: 0.8rem;
		cursor: pointer;
	}
	.btn-add-owner:hover, .btn-remove-owner:hover { background: var(--bg-hover); }
	.danger-zone {
		margin-top: 1rem;
		padding-top: 1rem;
		border-top: 1px solid var(--border-subtle);
	}
	.danger-title {
		margin: 0 0 0.25rem;
		font-weight: 700;
		color: var(--text);
	}
	.delete-btn {
		margin-top: 0.5rem;
		background: var(--bg-btn-danger, #b42318);
		color: var(--text-btn-primary, #fff);
		border: none;
		padding: 0.6rem 1rem;
		border-radius: var(--radius-md);
		font-weight: 600;
		cursor: pointer;
	}
	.delete-btn:disabled { opacity: 0.6; cursor: not-allowed; }

	@media (max-width: 480px) {
		.grid { grid-template-columns: 1fr; }
	}
</style>
