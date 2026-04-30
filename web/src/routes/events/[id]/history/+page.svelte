<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api, type EventRevision } from '$lib/api.js';

	const id = $derived($page.params.id);

	let revisions = $state<EventRevision[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			const res = await api.events.history(id);
			revisions = res.revisions ?? [];
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	function shortHash(hash: string) {
		return `${hash.slice(0, 16)}…`;
	}
</script>

<main>
	<h1>Event Edit History</h1>
	<p class="muted">Event: <code>{id}</code></p>

	{#if loading}
		<p class="muted">Loading revisions…</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else if revisions.length === 0}
		<p class="muted">No revisions found.</p>
	{:else}
		<div class="list">
			{#each revisions as rev}
				<div class="item">
					<div class="row">
						<strong>{shortHash(rev.hash)}</strong>
						{#if rev.is_current}
							<span class="badge">Current</span>
						{/if}
					</div>
					<p class="muted"><code>{rev.hash}</code></p>
					<p class="muted">Updated: {new Date(rev.created_at).toLocaleString()}</p>
					{#if rev.prev}
						<p class="muted">Prev: <code>{rev.prev}</code></p>
					{/if}
					<p>Title: {rev.state.title}</p>
				</div>
			{/each}
		</div>
	{/if}

	<a class="back" href="/events/{id}">Back to event</a>
</main>

<style>
	main {
		max-width: 760px;
		margin: 2rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 {
		margin: 0 0 0.5rem;
		color: var(--text);
	}
	.list {
		display: flex;
		flex-direction: column;
		gap: 0.9rem;
		margin-top: 1rem;
	}
	.item {
		border: 1px solid var(--border-card);
		border-radius: var(--radius-lg);
		background: var(--bg-card);
		padding: 0.9rem 1rem;
	}
	.row {
		display: flex;
		align-items: center;
		gap: 0.6rem;
	}
	.badge {
		font-size: 0.72rem;
		padding: 0.15rem 0.45rem;
		border-radius: var(--radius-full);
		background: var(--bg-subtle);
		color: var(--text-muted);
		border: 1px solid var(--border-subtle);
	}
	.back {
		display: inline-block;
		margin-top: 1rem;
		color: var(--text-accent);
		text-decoration: none;
	}
	.back:hover { text-decoration: underline; }
	.muted { color: var(--text-muted); margin: 0.2rem 0; }
	.error { color: var(--text-error); }
</style>
