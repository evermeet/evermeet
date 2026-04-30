<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type Event } from '$lib/api.js';
	import Avatar from '$lib/components/Avatar.svelte';

	let events = $state<Event[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			const res = await api.events.list();
			events = res.events ?? [];
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	function formatDate(iso: string) {
		return new Date(iso).toLocaleDateString('en', {
			weekday: 'short', month: 'short', day: 'numeric',
			hour: '2-digit', minute: '2-digit',
		});
	}
</script>

<main>
	<h1>Upcoming events</h1>

	{#if loading}
		<p class="muted">Loading…</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else if events.length === 0}
		<p class="muted">No events yet. <a href="/events/create">Create one.</a></p>
	{:else}
		<ul class="event-list">
			{#each events as event}
				<li>
					<a href="/events/{event.id}">
						<span class="date">{formatDate(event.starts_at)}</span>
						<span class="title">{event.title}</span>
						<span class="organizer">
							<Avatar did={event.organizer} size={20} />
						</span>
						{#if event.location}
							<span class="location">{event.location.name}</span>
						{:else}
							<span class="location"></span>
						{/if}
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</main>

<style>
	main {
		max-width: 680px;
		margin: 2rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 1.5rem; color: var(--text); }
	.event-list { list-style: none; padding: 0; margin: 0; }
	.event-list li { border-bottom: 1px solid var(--border-subtle); }
	.event-list a {
		display: grid;
		grid-template-columns: 160px 1fr auto auto;
		gap: 0.5rem;
		align-items: center;
		padding: 0.9rem 0;
		text-decoration: none;
		color: inherit;
	}
	.event-list a:hover .title { text-decoration: underline; }
	.organizer { display: flex; align-items: center; opacity: 0.7; }
	.date { font-size: 0.85rem; color: var(--text-subtle); white-space: nowrap; }
	.title { font-weight: 600; color: var(--text); }
	.location { font-size: 0.85rem; color: var(--text-muted); text-align: right; }
	.muted { color: var(--text-muted); }
	.error { color: var(--text-error); }
</style>
