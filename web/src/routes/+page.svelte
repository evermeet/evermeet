<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api, type CalendarEvent } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import EventCard from '$lib/components/EventCard.svelte';

	let events = $state<CalendarEvent[]>([]);
	let loading = $state(true);
	let error = $state('');
	let filter = $state<'upcoming' | 'past'>('upcoming');
	let hostProfiles = $state<Record<string, { displayName: string; avatar: string }>>({});

	onMount(() => {
		const unwatch = $effect.root(() => {
			$effect(() => {
				if (!auth.loading && !auth.user) goto('/discover');
			});
		});

		async function load() {
			// wait until auth is resolved
			await new Promise<void>(resolve => {
				if (!auth.loading) { resolve(); return; }
				const iv = setInterval(() => { if (!auth.loading) { clearInterval(iv); resolve(); } }, 50);
			});
			if (!auth.user) return;

			try {
				const res = await api.events.list(100);
				events = (res.events ?? []).map((ev: any) => ({
					id: ev.id,
					title: ev.title,
					starts_at: ev.starts_at,
					ends_at: ev.ends_at,
					location: ev.location,
					cover_url: ev.cover_url,
					hosts: ev.governance?.owners?.map((o: any) => o.did) ?? [],
				}));

				const uniqueHosts = Array.from(new Set(events.flatMap(ev => ev.hosts ?? [])));
				if (uniqueHosts.length > 0) {
					const entries = await Promise.all(
						uniqueHosts.map(async (did) => {
							try {
								const user = await api.users.get(did);
								return [did, { displayName: user.display_name || did, avatar: user.avatar || '' }] as const;
							} catch {
								return [did, { displayName: did, avatar: '' }] as const;
							}
						})
					);
					hostProfiles = Object.fromEntries(entries);
				}
			} catch (e: any) {
				error = e.message;
			} finally {
				loading = false;
			}
		}

		load();
		return () => unwatch();
	});

	const now = new Date();

	const filtered = $derived(
		events
			.filter(ev => {
				const t = new Date(ev.starts_at);
				return filter === 'upcoming' ? t >= now : t < now;
			})
			.sort((a, b) => {
				const diff = new Date(a.starts_at).getTime() - new Date(b.starts_at).getTime();
				return filter === 'past' ? -diff : diff;
			})
	);

	function groupByDate(evs: CalendarEvent[]) {
		const groups = new Map<string, CalendarEvent[]>();
		for (const ev of evs) {
			const d = new Date(ev.starts_at);
			const key = d.toLocaleDateString('en', { month: 'long', day: 'numeric', weekday: 'long' });
			if (!groups.has(key)) groups.set(key, []);
			groups.get(key)!.push(ev);
		}
		return groups;
	}

	const grouped = $derived(groupByDate(filtered));

	function parseDateKey(key: string) {
		const parts = key.split(',');
		const weekday = parts[0]?.trim() ?? '';
		const monthDay = parts[1]?.trim() ?? '';
		return { monthDay, weekday };
	}
</script>

<svelte:head>
	<title>Events — Evermeet</title>
</svelte:head>

<main>
	<div class="page-header">
		<h1>Events</h1>
		<div class="filter-toggle">
			<button class="filter-btn" class:active={filter === 'upcoming'} onclick={() => filter = 'upcoming'}>Upcoming</button>
			<button class="filter-btn" class:active={filter === 'past'} onclick={() => filter = 'past'}>Past</button>
		</div>
	</div>

	{#if loading}
		<p class="muted">Loading…</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else if filtered.length === 0}
		<p class="muted">No {filter} events.</p>
	{:else}
		<div class="event-groups">
			{#each grouped as [key, evs]}
				{@const { monthDay, weekday } = parseDateKey(key)}
				<div class="date-group">
					<div class="date-label">
						<div class="date-label-text">
							<span class="date-month">{monthDay}</span>
							<span class="date-weekday">{weekday}</span>
						</div>
					</div>
					<span class="dot"></span>
					<div class="event-cards">
						{#each evs as ev}
							<EventCard
								id={ev.id}
								title={ev.title}
								starts_at={ev.starts_at}
								location={ev.location}
								cover_url={ev.cover_url}
								hosts={ev.hosts}
								{hostProfiles}
							/>
						{/each}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</main>

<style>
	main {
		max-width: 780px;
		margin: 2.5rem auto;
		padding: 0 1.5rem 3rem;
		font-family: system-ui, sans-serif;
	}

	.page-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 2rem;
	}
	h1 { font-size: 2rem; font-weight: 800; margin: 0; color: var(--text); }

	.filter-toggle {
		display: flex;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		overflow: hidden;
	}
	.filter-btn {
		padding: 0.4rem 1rem;
		border: none;
		background: transparent;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text-muted);
		cursor: pointer;
		transition: background 0.1s, color 0.1s;
	}
	.filter-btn.active {
		background: var(--bg-raised);
		color: var(--text);
		font-weight: 600;
	}
	.filter-btn:hover:not(.active) { background: var(--bg-hover); color: var(--text); }

	.event-groups { display: flex; flex-direction: column; gap: 0; }

	.date-group {
		display: grid;
		grid-template-columns: 120px 1fr;
		gap: 0 1.5rem;
		margin-bottom: 1.5rem;
		position: relative;
	}

	.date-label {
		display: flex;
		align-items: flex-start;
		justify-content: flex-end;
		padding-top: 1rem;
	}
	.date-label-text {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 0.1rem;
	}
	.date-month {
		font-size: 0.9rem;
		font-weight: 700;
		color: var(--text);
		white-space: nowrap;
	}
	.date-weekday {
		font-size: 0.85rem;
		color: var(--text-muted);
		white-space: nowrap;
	}
	.dot {
		position: absolute;
		left: calc(120px + 1.5rem / 2);
		top: calc(1rem + 3px);
		transform: translateX(-50%);
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--border-input);
	}

	.event-cards {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		border-left: 1px solid var(--border-subtle);
		padding-left: 1.5rem;
	}

	.muted { color: var(--text-muted); font-size: 0.9rem; }
	.error { color: var(--text-error); }
</style>
