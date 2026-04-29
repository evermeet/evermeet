<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api, type Event } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';

	let event = $state<Event | null>(null);
	let loading = $state(true);
	let error = $state('');

	const id = $derived($page.params.id);

	onMount(async () => {
		try {
			const res = await api.events.get(id);
			event = res.state;
			event.id = res.id;
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

		<div class="rsvp-block">
			{#if event.rsvp.count !== undefined}
				<span class="count">{event.rsvp.count}{event.rsvp.limit ? `/${event.rsvp.limit}` : ''} attending</span>
			{/if}
			{#if auth.user && !isOrganizer()}
				<button class="rsvp-btn">RSVP</button>
			{:else if !auth.user}
				<a href="/auth/login" class="rsvp-btn">Sign in to RSVP</a>
			{/if}
		</div>

		{#if event.tags?.length}
			<div class="tags">
				{#each event.tags as tag}
					<span class="tag">{tag}</span>
				{/each}
			</div>
		{/if}

		<p class="organizer muted">Organized by <code>{event.organizer}</code></p>
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
	.rsvp-block { display: flex; align-items: center; gap: 1rem; margin: 1.5rem 0; }
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
	.tags { display: flex; flex-wrap: wrap; gap: 0.4rem; margin: 1rem 0; }
	.tag {
		padding: 0.2rem 0.6rem;
		background: #f0f0f0;
		border-radius: 4px;
		font-size: 0.8rem;
		color: #555;
	}
	.organizer { margin-top: 2rem; font-size: 0.8rem; }
	.muted { color: #999; }
	.error { color: #c00; }
</style>
