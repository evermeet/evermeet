<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api, type Event } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import Avatar from '$lib/components/Avatar.svelte';

	let event = $state<Event | null>(null);
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

			if (isOrganizer()) {
				rsvps = await api.events.listRSVPs(id);
			}
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	function formatDate(iso: string) {
		return new Date(iso).toLocaleDateString('en', {
			weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
			hour: '2-digit', minute: '2-digit',
		});
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
		<div class="header">
			<h1>{event.title}</h1>
			{#if isOrganizer()}
				<a href="/events/{id}/edit" class="edit-btn">Edit</a>
			{/if}
		</div>

		<div class="meta">
			<span>{formatDate(event.starts_at)}</span>
			{#if event.ends_at}
				<span> – {formatDate(event.ends_at)}</span>
			{/if}
		</div>

		{#if event.location}
			<p class="location">📍 {event.location.name}{event.location.address ? `, ${event.location.address}` : ''}</p>
		{/if}

		{#if event.description}
			<p class="description">{event.description}</p>
		{/if}

		<div class="rsvp-section">
			{#if isOrganizer()}
				<h3>RSVPs ({rsvps.length})</h3>
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
			{:else}
				<div class="rsvp-block">
					{#if event.rsvp.count !== undefined}
						<span class="count">{event.rsvp.count}{event.rsvp.limit ? `/${event.rsvp.limit}` : ''} attending</span>
					{/if}
				</div>

				{#if rsvpSent}
					<div class="success-box">
						<h3>RSVP Sent!</h3>
						<p>Your encrypted RSVP has been delivered to the organizer.</p>
					</div>
				{:else if auth.user}
					<form class="rsvp-form" onsubmit={submitRSVP}>
						<h3>RSVP to this event</h3>
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
					<a href="/auth/login" class="rsvp-btn">Sign in to RSVP</a>
				{/if}
			{/if}
		</div>

		{#if event.tags?.length}
			<div class="tags">
				{#each event.tags as tag}
					<span class="tag">{tag}</span>
				{/each}
			</div>
		{/if}

		<div class="organizer-box">
			<Avatar did={event.organizer} size={32} />
			<span class="muted">Organized by <a href="/u/{event.organizer}">{event.organizer.slice(0, 16)}…</a></span>
		</div>
	{/if}
</main>

<style>
	main {
		max-width: 680px;
		margin: 2rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	.header { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
	h1 { font-size: 2rem; font-weight: 700; margin: 0 0 0.5rem; }
	.edit-btn {
		padding: 0.4rem 0.8rem;
		border: 1px solid #d0d0d0;
		border-radius: 6px;
		text-decoration: none;
		font-size: 0.875rem;
		color: #333;
		white-space: nowrap;
	}
	.meta { color: #555; font-size: 0.95rem; margin-bottom: 0.5rem; }
	.location { color: #555; margin: 0.25rem 0 1rem; }
	.description { margin: 1.5rem 0; line-height: 1.6; white-space: pre-wrap; }
	.rsvp-section { margin: 2rem 0; padding-top: 2rem; border-top: 1px solid #eee; }
	.rsvp-section h3 { font-size: 1.25rem; margin-bottom: 1rem; }
	.rsvp-block { display: flex; align-items: center; gap: 1rem; margin: 1rem 0; }
	.count { color: #555; font-size: 0.9rem; }
	.rsvp-btn {
		padding: 0.6rem 1.25rem;
		background: #1a1a1a;
		color: #fff;
		border: none;
		border-radius: 6px;
		font-size: 0.95rem;
		font-weight: 600;
		cursor: pointer;
		text-decoration: none;
	}
	.rsvp-btn:disabled { opacity: 0.5; }
	
	.rsvp-form { display: flex; flex-direction: column; gap: 0.75rem; background: #f9f9f9; padding: 1.5rem; border-radius: 8px; }
	.form-row { display: flex; gap: 0.75rem; }
	.form-row input { flex: 1; }
	input, textarea {
		padding: 0.6rem;
		border: 1px solid #d0d0d0;
		border-radius: 4px;
		font-size: 0.95rem;
	}
	textarea { min-height: 80px; resize: vertical; }
	
	.rsvp-list { display: flex; flex-direction: column; gap: 1rem; }
	.rsvp-item { padding: 1rem; border: 1px solid #eee; border-radius: 6px; }
	.rsvp-head { display: flex; justify-content: space-between; margin-bottom: 0.5rem; }
	.rsvp-note { margin: 0.5rem 0; font-size: 0.95rem; color: #444; }
	.rsvp-meta { display: flex; justify-content: space-between; font-size: 0.8rem; color: #888; margin-top: 0.5rem; }
	
	.success-box { background: #e8f5e9; color: #2e7d32; padding: 1.5rem; border-radius: 8px; }
	.success-box h3 { margin-top: 0; }
	
	.tags { display: flex; flex-wrap: wrap; gap: 0.4rem; margin: 1rem 0; }
	.tag {
		padding: 0.2rem 0.6rem;
		background: #f0f0f0;
		border-radius: 4px;
		font-size: 0.8rem;
		color: #555;
	}
	.organizer-box {
		margin-top: 2rem;
		display: flex;
		align-items: center;
		gap: 0.75rem;
		font-size: 0.875rem;
	}
	.organizer-box a { color: inherit; text-decoration: none; font-weight: 500; }
	.organizer-box a:hover { color: #111; text-decoration: underline; }
	.muted { color: #999; }
	.error { color: #c00; }
</style>
