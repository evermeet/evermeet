<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api, type CalendarDetail, type Event } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import Avatar from '$lib/components/Avatar.svelte';
	import { marked } from 'marked';

	let event = $state<any>(null);
	let founding = $state<any>(null);
let calendar = $state<CalendarDetail | null>(null);
let currentHash = $state('');
	let loading = $state(true);
	let error = $state('');

	let rsvps = $state<any[]>([]);
	let rsvpSent = $state(false);
	let rsvpName = $state('');
	let rsvpEmail = $state('');
	let rsvpNote = $state('');
	let rsvpSubmitting = $state(false);

	const id = $derived($page.params.id);

	onMount(async () => {
		try {
			const res = await api.events.get(id);
			event = res.state;
			event.id = res.id;
			founding = res.founded;
			currentHash = res.hash;

			if (event.calendar) {
				try {
					calendar = await api.calendars.get(event.calendar);
				} catch (e) {
					calendar = null;
				}
			}

			if (isOrganizer()) {
				rsvps = await api.events.listRSVPs(id);
			}
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	function formatDateBlock(iso: string) {
		const d = new Date(iso);
		return {
			weekday: d.toLocaleDateString('en', { weekday: 'long' }),
			date: d.toLocaleDateString('en', { month: 'long', day: 'numeric' }),
			month: d.toLocaleDateString('en', { month: 'short' }).toUpperCase(),
			day: d.getDate(),
			time: d.toLocaleTimeString('en', { hour: '2-digit', minute: '2-digit' }),
		};
	}

	function isOrganizer() {
		return auth.user?.did === event?.organizer;
	}

	async function submitRSVP(e: Event) {
		e.preventDefault();
		rsvpSubmitting = true;
		error = '';
		try {
			await api.events.rsvp(id, { name: rsvpName, email: rsvpEmail, note: rsvpNote });
			rsvpSent = true;
		} catch (e: any) {
			error = e.message;
		} finally {
			rsvpSubmitting = false;
		}
	}
</script>

<main>
	{#if loading}
		<p class="muted">Loading…</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else if event}
		{@const start = formatDateBlock(event.starts_at)}
		{@const end = event.ends_at ? formatDateBlock(event.ends_at) : null}

		<div class="layout">

			<!-- Left column -->
			<aside class="left-col">
				<!-- Cover image / placeholder -->
				<div class="cover">
					{#if event.cover_url}
						<img src={event.cover_url} alt={event.title} />
					{:else}
						<div class="cover-placeholder">
							<span>{event.title}</span>
						</div>
					{/if}
				</div>

				{#if calendar}
					<div class="side-section presenter-section">
						<p class="side-label">Presented By</p>
						<div class="presenter-row">
							<a href="/calendars/{calendar.id}" class="presenter-avatar">
								<Avatar src={calendar.avatar} did={calendar.id} size={36} rounded={false} />
							</a>
							<div class="presenter-meta">
								<a href="/calendars/{calendar.id}" class="presenter-name">{calendar.name}</a>
								{#if calendar.description}
									<p class="presenter-description">{calendar.description}</p>
								{/if}
								{#if calendar.website}
									<a href={calendar.website} target="_blank" rel="noreferrer" class="presenter-link">{calendar.website}</a>
								{/if}
							</div>
						</div>
					</div>
				{/if}

				<!-- Hosted By -->
				<div class="side-section">
					<p class="side-label">Hosted By</p>
					<div class="host-row">
						<Avatar did={event.organizer} size={36} />
						<a href="/u/{event.organizer}" class="host-name">
							{event.organizer.slice(0, 20)}…
						</a>
					</div>
				</div>

				<!-- Going -->
				{#if event.rsvp?.count !== undefined}
					<div class="side-section">
						<p class="side-label">{event.rsvp.count}{event.rsvp.limit ? `/${event.rsvp.limit}` : ''} Going</p>
					</div>
				{/if}

				<!-- Tags -->
				{#if event.tags?.length}
					<div class="side-section">
						<div class="tags">
							{#each event.tags as tag}
								<span class="tag"># {tag}</span>
							{/each}
						</div>
					</div>
				{/if}

				<div class="side-section">
					<p class="side-label">Revision</p>
					<code class="revision-id">{currentHash ? `${currentHash.slice(0, 16)}…` : 'n/a'}</code>
					<a class="revision-link" href="/events/{id}/history">View edit history</a>
				</div>
			</aside>

			<!-- Right column -->
			<div class="right-col">
				<div class="right-top">
					<h1>{event.title}</h1>
					{#if isOrganizer()}
						<a href="/events/{id}/edit" class="edit-btn">Edit</a>
					{/if}
				</div>

				<!-- Date -->
				<div class="detail-row">
					<div class="date-badge">
						<span class="date-badge-month">{start.month}</span>
						<span class="date-badge-day">{start.day}</span>
					</div>
					<div class="detail-text">
						<strong>{start.weekday}, {start.date}</strong>
						<span class="muted">
							{start.time}{end ? ` – ${end.time}` : ''}
						</span>
					</div>
				</div>

				<!-- Location -->
				{#if event.location}
					<div class="detail-row">
						<div class="detail-icon">
							<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>
						</div>
						<div class="detail-text">
							<strong>{event.location.name}</strong>
							{#if event.location.address}
								<span class="muted">{event.location.address}</span>
							{/if}
						</div>
					</div>
				{/if}

				<!-- RSVP panel -->
				<div class="rsvp-panel">
					{#if isOrganizer()}
						<p class="panel-label">RSVPs ({rsvps.length})</p>
						{#if rsvps.length === 0}
							<p class="muted">No RSVPs yet.</p>
						{:else}
							<div class="rsvp-list">
								{#each rsvps as env}
									<div class="rsvp-item">
										<div class="rsvp-head">
											<strong>{env.rsvp.name}</strong>
											<span class="muted">{env.rsvp.email}</span>
										</div>
										{#if env.rsvp.note}
											<p class="rsvp-note">{env.rsvp.note}</p>
										{/if}
										<div class="rsvp-meta">
											<code>{env.sender_did.slice(0, 16)}…</code>
											<span>{new Date(env.received_at).toLocaleDateString()}</span>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					{:else if rsvpSent}
						<div class="success-box">
							<strong>RSVP Sent!</strong>
							<p>Your encrypted RSVP has been delivered to the organizer.</p>
						</div>
					{:else if auth.user}
						<p class="panel-label">Register</p>
						<form class="rsvp-form" onsubmit={submitRSVP}>
							<div class="form-row">
								<input type="text" bind:value={rsvpName} placeholder="Your Name" required />
								<input type="email" bind:value={rsvpEmail} placeholder="Email" required />
							</div>
							<textarea bind:value={rsvpNote} placeholder="Note to organizer (optional)"></textarea>
							<button type="submit" class="rsvp-btn" disabled={rsvpSubmitting}>
								{rsvpSubmitting ? 'Sending…' : 'Confirm RSVP'}
							</button>
						</form>
					{:else}
						<p class="panel-label">Register</p>
						<a href="/auth/login" class="rsvp-btn">Sign in to RSVP</a>
					{/if}
				</div>

				<!-- About -->
				{#if event.description}
					<div class="about-section">
						<p class="about-label">About Event</p>
						<div class="description">{@html marked(event.description)}</div>
					</div>
				{/if}

				{#if founding?.instance_id}
					<p class="muted instance-hint">Home: {founding.instance_id}</p>
				{/if}
			</div>
		</div>
	{/if}
</main>

<style>
	main {
		max-width: 900px;
		margin: 2.5rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}

	.layout {
		display: grid;
		grid-template-columns: 300px 1fr;
		gap: 3rem;
		align-items: start;
	}

	/* ── Left column ── */
	.left-col {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.cover {
		width: 100%;
		aspect-ratio: 1 / 1;
		border-radius: var(--radius-xl);
		overflow: hidden;
		background: var(--bg-raised);
	}
	.cover img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.cover-placeholder {
		width: 100%;
		height: 100%;
		display: flex;
		align-items: flex-end;
		padding: 1.25rem;
		background: linear-gradient(135deg, var(--bg-raised) 0%, var(--bg-subtle) 100%);
	}
	.cover-placeholder span {
		font-size: 1.5rem;
		font-weight: 800;
		color: var(--text);
		line-height: 1.2;
		text-transform: uppercase;
		letter-spacing: -0.02em;
	}

	.side-section {
		border-top: 1px solid var(--border-subtle);
		padding-top: 1.25rem;
	}
	.side-label {
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin: 0 0 0.75rem;
	}
	.host-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.host-name {
		font-weight: 600;
		font-size: 0.9rem;
		color: var(--text);
		text-decoration: none;
	}
	.host-name:hover { text-decoration: underline; }
	.presenter-section { border-top: none; padding-top: 0; }
	.presenter-row {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
	}
	.presenter-avatar {
		flex-shrink: 0;
		display: inline-flex;
	}
	.presenter-meta {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	.presenter-name {
		font-weight: 700;
		font-size: 0.95rem;
		color: var(--text);
		text-decoration: none;
	}
	.presenter-name:hover { text-decoration: underline; }
	.presenter-description {
		margin: 0;
		font-size: 0.85rem;
		color: var(--text-muted);
		line-height: 1.35;
	}
	.presenter-link {
		font-size: 0.82rem;
		color: var(--text-accent);
		text-decoration: none;
		word-break: break-word;
	}
	.presenter-link:hover { text-decoration: underline; }

	.tags { display: flex; flex-wrap: wrap; gap: 0.4rem; }
	.tag {
		padding: 0.25rem 0.7rem;
		border: 1px solid var(--border);
		border-radius: var(--radius-full);
		font-size: 0.8rem;
		color: var(--text-secondary);
	}
	.revision-id {
		display: inline-block;
		font-size: 0.78rem;
		color: var(--text-secondary);
		word-break: break-all;
		margin-bottom: 0.5rem;
	}
	.revision-link {
		display: inline-block;
		font-size: 0.85rem;
		color: var(--text-accent);
		text-decoration: none;
	}
	.revision-link:hover { text-decoration: underline; }

	/* ── Right column ── */
	.right-col {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.right-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}
	h1 {
		font-size: 2.25rem;
		font-weight: 800;
		margin: 0;
		line-height: 1.1;
		color: var(--text);
	}
	.edit-btn {
		flex-shrink: 0;
		padding: 0.35rem 0.75rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		text-decoration: none;
		font-size: 0.8rem;
		color: var(--text);
		background: var(--bg);
		margin-top: 0.4rem;
	}
	.edit-btn:hover { background: var(--bg-hover); }

	/* Date / location rows */
	.detail-row {
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	.date-badge {
		width: 44px;
		height: 44px;
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		background: var(--bg-card);
	}
	.date-badge-month {
		font-size: 0.6rem;
		font-weight: 700;
		text-transform: uppercase;
		color: var(--text-muted);
		line-height: 1;
	}
	.date-badge-day {
		font-size: 1.1rem;
		font-weight: 800;
		color: var(--text);
		line-height: 1;
	}
	.detail-icon {
		width: 44px;
		height: 44px;
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		color: var(--text-secondary);
		background: var(--bg-card);
	}
	.detail-text {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
		font-size: 0.95rem;
	}
	.detail-text strong { color: var(--text); font-weight: 600; }

	/* RSVP panel */
	.rsvp-panel {
		background: var(--bg-subtle);
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-xl);
		padding: 1.25rem 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.panel-label {
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin: 0;
	}
	.rsvp-btn {
		width: 100%;
		padding: 0.75rem;
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		text-decoration: none;
		text-align: center;
		display: block;
	}
	.rsvp-btn:disabled { opacity: 0.5; }

	.rsvp-form { display: flex; flex-direction: column; gap: 0.6rem; }
	.form-row { display: flex; gap: 0.6rem; }
	.form-row input { flex: 1; }
	input, textarea {
		padding: 0.6rem 0.75rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		font-size: 0.9rem;
		background: var(--bg-input);
		color: var(--text);
		font-family: inherit;
		width: 100%;
	}
	input:focus, textarea:focus { outline: none; border-color: var(--border-input-focus); }
	textarea { min-height: 70px; resize: vertical; }

	.rsvp-list { display: flex; flex-direction: column; gap: 0.75rem; }
	.rsvp-item {
		padding: 0.9rem 1rem;
		border: 1px solid var(--border-card);
		border-radius: var(--radius-md);
		background: var(--bg-card);
	}
	.rsvp-head { display: flex; justify-content: space-between; margin-bottom: 0.4rem; color: var(--text); }
	.rsvp-note { margin: 0.4rem 0; font-size: 0.9rem; color: var(--text-secondary); }
	.rsvp-meta { display: flex; justify-content: space-between; font-size: 0.78rem; color: var(--text-muted); margin-top: 0.4rem; }

	.success-box {
		background: var(--bg-success);
		color: var(--text-success);
		padding: 1rem 1.25rem;
		border-radius: var(--radius-md);
	}
	.success-box p { margin: 0.25rem 0 0; font-size: 0.9rem; }

	/* About section */
	.about-section {
		border-top: 1px solid var(--border-subtle);
		padding-top: 1.5rem;
	}
	.about-label {
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin: 0 0 0.75rem;
	}
	.description {
		line-height: 1.7;
		color: var(--text);
		font-size: 0.95rem;
		margin: 0;
	}
	.description :global(h1),
	.description :global(h2),
	.description :global(h3) {
		margin: 1.25rem 0 0.4rem;
		font-weight: 700;
		color: var(--text);
	}
	.description :global(p) { margin: 0.6rem 0; }
	.description :global(ul),
	.description :global(ol) { padding-left: 1.5rem; margin: 0.6rem 0; }
	.description :global(li) { margin: 0.25rem 0; }
	.description :global(a) { color: var(--text-accent); }
	.description :global(code) {
		background: var(--bg-code);
		color: var(--text-code);
		padding: 0.1em 0.35em;
		border-radius: var(--radius-sm);
		font-size: 0.88em;
	}
	.description :global(pre) {
		background: var(--bg-code);
		padding: 0.75rem 1rem;
		border-radius: var(--radius-md);
		overflow-x: auto;
		margin: 0.75rem 0;
	}
	.description :global(blockquote) {
		border-left: 3px solid var(--border);
		margin: 0.75rem 0;
		padding: 0.25rem 0 0.25rem 1rem;
		color: var(--text-secondary);
	}
	.description :global(hr) {
		border: none;
		border-top: 1px solid var(--border-subtle);
		margin: 1rem 0;
	}
	.description :global(img) {
		max-width: 100%;
		height: auto;
		border-radius: var(--radius-md);
		display: block;
	}

	.instance-hint { font-size: 0.8rem; margin: 0; }

	.muted { color: var(--text-muted); }
	.error { color: var(--text-error); }

	@media (max-width: 640px) {
		.layout { grid-template-columns: 1fr; gap: 2rem; }
		h1 { font-size: 1.75rem; }
	}
</style>
