<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { api, type Calendar, type ImportEventPreview } from '$lib/api.js';

	let title = $state('');
	let description = $state('');
	let cover_url = $state('');
	let starts_at = $state('');
	let ends_at = $state('');
	let locationName = $state('');
	let calendars = $state<Calendar[]>([]);
	let calendarId = $state('');
	let visibility = $state<'public' | 'unlisted' | 'private'>('public');
	let rsvpLimit = $state(0);
	let importOpen = $state(false);
	let importUrl = $state('');
	let importFetching = $state(false);
	let importPreview = $state<ImportEventPreview | null>(null);

	let loading = $state(false);
	let error = $state('');

	onMount(async () => {
		try {
			const res = await api.calendars.list();
			calendars = res.owned ?? [];
		} catch (e: any) {
			error = e.message;
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		loading = true;
		error = '';

		try {
			const res = await api.events.create({
				title,
				description,
				cover_url: cover_url || undefined,
				starts_at: new Date(starts_at).toISOString(),
				calendar_id: calendarId,
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

	async function fetchImportPreview() {
		if (!importUrl.trim()) return;
		importFetching = true;
		error = '';
		importPreview = null;
		try {
			importPreview = await api.events.importPreview(importUrl.trim());
		} catch (e: any) {
			error = e.message;
		} finally {
			importFetching = false;
		}
	}

	function applyImportToForm() {
		if (!importPreview) return;
		title = importPreview.title || title;
		description = importPreview.description || description;
		cover_url = importPreview.cover_url || cover_url;
		locationName = importPreview.location_name || locationName;
		if (importPreview.starts_at) {
			starts_at = formatForInput(importPreview.starts_at);
		}
		if (importPreview.ends_at) {
			ends_at = formatForInput(importPreview.ends_at);
		}
		importOpen = false;
	}

	function formatForInput(iso: string): string {
		const d = new Date(iso);
		if (Number.isNaN(d.getTime())) return '';
		const pad = (n: number) => n.toString().padStart(2, '0');
		return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
	}
</script>

<main>
	<div class="header-row">
		<h1>Create Event</h1>
		<button type="button" class="btn-import-toggle" onclick={() => (importOpen = !importOpen)}>
			{importOpen ? 'Close Import' : 'Import events'}
		</button>
	</div>

	{#if importOpen}
		<div class="import-panel">
			<div class="field">
				<label for="import_url">Import URL</label>
				<div class="import-row">
					<input id="import_url" type="url" bind:value={importUrl} placeholder="https://lu.ma/..." />
					<button type="button" onclick={fetchImportPreview} disabled={importFetching}>
						{importFetching ? 'Fetching…' : 'Fetch'}
					</button>
				</div>
				<p class="muted">Currently supported: `luma.com`</p>
			</div>
			{#if importPreview}
				<div class="import-preview">
					<p class="preview-label">Preview ({importPreview.provider})</p>
					<p class="preview-title">{importPreview.title}</p>
					<p class="muted">{new Date(importPreview.starts_at).toLocaleString()}</p>
					{#if importPreview.location_name}
						<p class="muted">{importPreview.location_name}{importPreview.location_address ? `, ${importPreview.location_address}` : ''}</p>
					{/if}
					<button type="button" onclick={applyImportToForm}>
						Use in form
					</button>
				</div>
			{/if}
		</div>
	{/if}

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
			<label for="cover_url">Cover Image URL (optional)</label>
			<input type="url" id="cover_url" bind:value={cover_url} placeholder="https://…" />
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
				<label for="calendar">Calendar</label>
				<select id="calendar" bind:value={calendarId}>
					<option value="">Personal event (no calendar)</option>
					{#each calendars as cal}
						<option value={cal.id}>{cal.name}</option>
					{/each}
				</select>
			</div>
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
	h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 2rem; color: var(--text); }
	.header-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1rem;
	}
	.header-row h1 { margin: 0; }
	.btn-import-toggle {
		border: 1px solid var(--border-input);
		background: var(--bg-raised);
		color: var(--text);
		border-radius: var(--radius-md);
		padding: 0.45rem 0.85rem;
		font-size: 0.85rem;
		font-weight: 600;
		cursor: pointer;
	}
	.btn-import-toggle:hover { background: var(--bg-hover); }

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
	.import-panel {
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-lg);
		background: var(--bg-subtle);
		padding: 0.9rem 1rem;
		margin-bottom: 1.2rem;
		display: flex;
		flex-direction: column;
		gap: 0.9rem;
	}
	.import-row {
		display: flex;
		gap: 0.6rem;
	}
	.import-row input { flex: 1; }
	.import-preview {
		border: 1px solid var(--border-card);
		background: var(--bg-card);
		border-radius: var(--radius-md);
		padding: 0.8rem;
	}
	.preview-label {
		margin: 0 0 0.3rem;
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--text-muted);
		text-transform: uppercase;
	}
	.preview-title {
		margin: 0 0 0.25rem;
		font-size: 1.05rem;
		font-weight: 700;
		color: var(--text);
	}

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

	@media (max-width: 480px) {
		.grid { grid-template-columns: 1fr; }
	}
</style>
