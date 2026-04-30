<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type Calendar } from '$lib/api.js';
	import Avatar from '$lib/components/Avatar.svelte';

	let owned = $state<Calendar[]>([]);
	let subscribed = $state<Calendar[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			const res = await api.calendars.list();
			owned = res.owned ?? [];
			subscribed = res.subscribed ?? [];
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	});
</script>

<main>
	<h1>Calendars</h1>

	{#if loading}
		<p class="muted">Loading…</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else}
		<section>
			<div class="section-header">
				<h2>My Calendars</h2>
				<a href="/calendars/create" class="btn-create">+ Create</a>
			</div>

			{#if owned.length === 0}
				<p class="muted">No calendars yet. <a href="/calendars/create">Create one.</a></p>
			{:else}
				<div class="calendar-grid">
					{#each owned as cal}
						<a href="/calendars/{cal.id}" class="calendar-card">
							<div class="card-avatar">
								<Avatar src={cal.avatar} did={cal.id} size={56} rounded={false} />
							</div>
							<div class="card-info">
								<span class="card-name">{cal.name}</span>
								<span class="card-subs">
									{cal.subscribers === 0 ? 'No Subscribers' : `${cal.subscribers} Subscriber${cal.subscribers === 1 ? '' : 's'}`}
								</span>
							</div>
						</a>
					{/each}
				</div>
			{/if}
		</section>

		<div class="divider"></div>

		<section>
			<h2>Subscribed Calendars</h2>

			{#if subscribed.length === 0}
				<div class="empty-subscribed">
					<p class="muted">No Subscriptions</p>
					<p class="muted-sm">You haven't subscribed to any calendars yet.</p>
				</div>
			{:else}
				<div class="subscribed-list">
					{#each subscribed as cal}
						<a href="/calendars/{cal.id}" class="subscribed-card">
							<div class="sub-left">
								<Avatar src={cal.avatar} did={cal.id} size={44} />
								<div class="sub-info">
									<span class="card-name">{cal.name}</span>
									{#if cal.description}
										<span class="card-desc">{cal.description}</span>
									{/if}
								</div>
							</div>
							<span class="card-subs">
								{cal.subscribers === 0 ? 'No Subscribers' : `${cal.subscribers} Subscriber${cal.subscribers === 1 ? '' : 's'}`}
							</span>
						</a>
					{/each}
				</div>
			{/if}
		</section>
	{/if}
</main>

<style>
	main {
		max-width: 900px;
		margin: 2.5rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 { font-size: 2rem; font-weight: 800; margin: 0 0 2rem; color: var(--text); }
	h2 { font-size: 1.1rem; font-weight: 700; margin: 0; color: var(--text); }

	section { margin-bottom: 2.5rem; }
	.section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1.25rem;
	}
	.btn-create {
		padding: 0.35rem 0.85rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--text);
		text-decoration: none;
		background: var(--bg-raised);
	}
	.btn-create:hover { background: var(--bg-hover); }

	.divider {
		height: 1px;
		background: var(--border-subtle);
		margin: 2rem 0;
	}

	/* Grid of owned calendars */
	.calendar-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 0.75rem;
	}
	.calendar-card {
		display: flex;
		flex-direction: column;
		gap: 0.85rem;
		padding: 1.25rem;
		border: 1px solid var(--border-card);
		border-radius: var(--radius-xl);
		background: var(--bg-card);
		text-decoration: none;
		transition: border-color 0.1s, background 0.1s;
	}
	.calendar-card:hover {
		border-color: var(--border-input);
		background: var(--bg-raised);
	}
	.card-info { display: flex; flex-direction: column; gap: 0.2rem; }
	.card-name { font-weight: 600; font-size: 0.95rem; color: var(--text); }
	.card-subs { font-size: 0.8rem; color: var(--text-muted); }

	/* Subscribed list */
	.empty-subscribed {
		padding: 2rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-xl);
		text-align: center;
	}
	.subscribed-list { display: flex; flex-direction: column; gap: 0; }
	.subscribed-card {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1rem 0;
		border-bottom: 1px solid var(--border-subtle);
		text-decoration: none;
		gap: 1rem;
	}
	.subscribed-card:last-child { border-bottom: none; }
	.sub-left { display: flex; align-items: center; gap: 0.85rem; }
	.sub-info { display: flex; flex-direction: column; gap: 0.15rem; }
	.card-desc { font-size: 0.8rem; color: var(--text-muted); }

	.muted { color: var(--text-muted); font-size: 0.95rem; }
	.muted-sm { color: var(--text-muted); font-size: 0.85rem; margin: 0.25rem 0 0; }
	.error { color: var(--text-error); }

	@media (max-width: 640px) {
		.calendar-grid { grid-template-columns: repeat(2, 1fr); }
	}
	@media (max-width: 400px) {
		.calendar-grid { grid-template-columns: 1fr; }
	}
</style>
