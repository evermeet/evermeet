<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api, type CalendarDetail } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import Avatar from '$lib/components/Avatar.svelte';

	const id = $derived($page.params.id);

	let cal = $state<CalendarDetail | null>(null);
	let loading = $state(true);
	let error = $state('');
	let subscribing = $state(false);
let hostProfiles = $state<Record<string, { displayName: string; avatar: string }>>({});

	onMount(async () => {
		try {
			cal = await api.calendars.get(id);
			const uniqueHosts = Array.from(
				new Set((cal.events ?? []).flatMap((ev) => ev.hosts ?? []))
			);
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
	});

	function isOwner() {
		if (!cal || !auth.user) return false;
		return cal.governance.owners.some(o => o.did === auth.user!.did);
	}

	async function toggleSubscribe() {
		if (!cal) return;
		subscribing = true;
		try {
			if (cal.subscribed) {
				await api.calendars.unsubscribe(id);
				cal = { ...cal, subscribed: false, subscribers: cal.subscribers - 1 };
			} else {
				await api.calendars.subscribe(id);
				cal = { ...cal, subscribed: true, subscribers: cal.subscribers + 1 };
			}
		} catch (e: any) {
			error = e.message;
		} finally {
			subscribing = false;
		}
	}

	function formatEventDate(iso: string) {
		const d = new Date(iso);
		return {
			weekday: d.toLocaleDateString('en', { weekday: 'short' }),
			date: d.toLocaleDateString('en', { month: 'short', day: 'numeric' }),
			time: d.toLocaleTimeString('en', { hour: '2-digit', minute: '2-digit' }),
		};
	}

	function groupByDate(events: CalendarDetail['events']) {
		const groups = new Map<string, typeof events>();
		for (const ev of events) {
			const key = new Date(ev.starts_at).toLocaleDateString('en', {
				weekday: 'long', month: 'long', day: 'numeric',
			});
			if (!groups.has(key)) groups.set(key, []);
			groups.get(key)!.push(ev);
		}
		return groups;
	}

	const grouped = $derived(cal ? groupByDate(cal.events) : new Map());
</script>

<svelte:head>
	{#if cal}<title>{cal.name} — Evermeet</title>{/if}
</svelte:head>

{#if loading}
	<p class="loading muted">Loading…</p>
{:else if error}
	<p class="loading error">{error}</p>
{:else if cal}
	<!-- Backdrop -->
	<div class="backdrop" style={cal.backdrop_url ? `background-image: url('${cal.backdrop_url}')` : ''}>
		{#if !cal.backdrop_url}
			<div class="backdrop-gradient"></div>
		{/if}
	</div>

	<main>
		<!-- Avatar overlapping the backdrop -->
		<div class="avatar-row">
			<div class="cal-avatar">
				<Avatar src={cal.avatar} did={cal.id} size={80} rounded={false} />
			</div>
			{#if auth.user && !isOwner()}
				<button
					class="btn-subscribe"
					class:subscribed={cal.subscribed}
					onclick={toggleSubscribe}
					disabled={subscribing}
				>
					{cal.subscribed ? 'Subscribed ✓' : 'Subscribe'}
				</button>
			{:else if isOwner()}
				<a href="/calendars/{id}/edit" class="btn-edit">Edit Calendar</a>
			{/if}
		</div>

		<!-- Calendar info -->
		<div class="cal-info">
			<h1>{cal.name}</h1>
			{#if cal.description}
				<p class="cal-desc">{cal.description}</p>
			{/if}
			<div class="cal-meta">
				{#if cal.website}
					<a href={cal.website} target="_blank" rel="noopener" class="meta-link">
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
						{cal.website.replace(/^https?:\/\//, '')}
					</a>
				{/if}
				<span class="muted">
					{cal.subscribers} subscriber{cal.subscribers === 1 ? '' : 's'}
				</span>
			</div>
		</div>

		<div class="divider"></div>

		<!-- Events -->
		<div class="events-section">
			<h2>Events</h2>

			{#if cal.events.length === 0}
				<p class="muted">No events yet.</p>
			{:else}
				<div class="event-groups">
					{#each grouped as [dateLabel, evs]}
						<div class="date-group">
							<div class="date-label">
								<span class="dot"></span>
								<span>{dateLabel}</span>
							</div>
							<div class="event-cards">
								{#each evs as ev}
									{@const d = formatEventDate(ev.starts_at)}
									<a href="/events/{ev.id}" class="event-card">
										<div class="event-main">
											<div class="event-time">{d.time}</div>
											<div class="event-title">{ev.title}</div>
											{#if ev.hosts && ev.hosts.length > 0}
												<div class="event-hosts">
													<div class="host-avatars">
														{#each ev.hosts.slice(0, 4) as did}
															<Avatar src={hostProfiles[did]?.avatar} did={did} size={24} />
														{/each}
													</div>
													<span>
														By {ev.hosts
															.map((did) => hostProfiles[did]?.displayName || did)
															.join(', ')}
													</span>
												</div>
											{/if}
											{#if ev.location}
												<div class="event-location">
													<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>
													{ev.location.name}
												</div>
											{/if}
										</div>
										{#if ev.cover_url}
											<img src={ev.cover_url} alt={ev.title} class="event-cover" />
										{/if}
									</a>
								{/each}
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</main>
{/if}

<style>
	.loading { max-width: 900px; margin: 2rem auto; padding: 0 1.5rem; font-family: system-ui, sans-serif; }

	/* Backdrop */
	.backdrop {
		width: 100%;
		height: 240px;
		background: var(--bg-raised);
		background-size: cover;
		background-position: center;
		position: relative;
	}
	.backdrop-gradient {
		position: absolute;
		inset: 0;
		background: linear-gradient(135deg, var(--bg-raised) 0%, var(--bg-subtle) 100%);
	}

	main {
		max-width: 900px;
		margin: 0 auto;
		padding: 0 1.5rem 3rem;
		font-family: system-ui, sans-serif;
	}

	/* Avatar row — pull avatar up so it straddles the backdrop bottom edge */
	.avatar-row {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		/* 80px avatar + 6px border = 86px total; pull up by half = 43px */
		margin-top: -43px;
		margin-bottom: 1rem;
	}
	.cal-avatar {
		border: 4px solid var(--bg);
		border-radius: var(--radius-xl);
		overflow: hidden;
		background: var(--bg-raised);
		/* Ensure it renders above the backdrop */
		position: relative;
		z-index: 1;
	}

	.btn-subscribe {
		padding: 0.5rem 1.25rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		background: var(--bg-raised);
		color: var(--text);
		transition: background 0.1s;
		margin-bottom: 4px;
	}
	.btn-subscribe:hover { background: var(--bg-hover); }
	.btn-subscribe.subscribed {
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border-color: transparent;
	}
	.btn-subscribe:disabled { opacity: 0.5; cursor: default; }

	.btn-edit {
		padding: 0.5rem 1.25rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text);
		text-decoration: none;
		background: var(--bg-raised);
		margin-bottom: 4px;
	}
	.btn-edit:hover { background: var(--bg-hover); }

	/* Calendar info */
	.cal-info { margin-bottom: 1.5rem; }
	h1 { font-size: 2rem; font-weight: 800; margin: 0 0 0.5rem; color: var(--text); line-height: 1.1; }
	.cal-desc { font-size: 0.95rem; color: var(--text-secondary); margin: 0 0 0.75rem; line-height: 1.6; }
	.cal-meta { display: flex; align-items: center; gap: 1rem; flex-wrap: wrap; }
	.meta-link {
		display: flex;
		align-items: center;
		gap: 0.35rem;
		font-size: 0.85rem;
		color: var(--text-secondary);
		text-decoration: none;
	}
	.meta-link:hover { color: var(--text); }

	.divider { height: 1px; background: var(--border-subtle); margin: 1.5rem 0; }

	/* Events section */
	.events-section h2 { font-size: 1.2rem; font-weight: 700; margin: 0 0 1.25rem; color: var(--text); }

	.event-groups { display: flex; flex-direction: column; gap: 1.5rem; }

	.date-group {}
	.date-label {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		margin-bottom: 0.75rem;
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text-muted);
	}
	.dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--border-input);
		flex-shrink: 0;
	}

	.event-cards { display: flex; flex-direction: column; gap: 0.5rem; margin-left: 1.15rem; border-left: 1px solid var(--border-subtle); padding-left: 1.25rem; }

	.event-card {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 1rem 1.25rem;
		border: 1px solid var(--border-card);
		border-radius: var(--radius-xl);
		background: var(--bg-card);
		text-decoration: none;
		transition: border-color 0.1s, background 0.1s;
	}
	.event-card:hover { border-color: var(--border-input); background: var(--bg-raised); }
	.event-main { display: flex; flex-direction: column; gap: 0.25rem; }
	.event-time { font-size: 0.92rem; color: var(--text-muted); }
	.event-title { font-size: 1.18rem; font-weight: 700; color: var(--text); }
	.event-hosts {
		display: flex;
		align-items: center;
		gap: 0.45rem;
		font-size: 0.95rem;
		color: var(--text-muted);
		margin-top: 0.2rem;
	}
	.host-avatars {
		display: flex;
		align-items: center;
	}
	.host-avatars :global(.avatar-box) {
		border: 2px solid var(--bg-card);
	}
	.host-avatars :global(.avatar-box:not(:first-child)) {
		margin-left: -8px;
	}
	.event-location {
		display: flex;
		align-items: center;
		gap: 0.3rem;
		font-size: 0.95rem;
		color: var(--text-muted);
		margin-top: 0.1rem;
	}
	.event-cover {
		width: 80px;
		height: 60px;
		object-fit: cover;
		border-radius: var(--radius-md);
		flex-shrink: 0;
	}

	.muted { color: var(--text-muted); font-size: 0.9rem; }
	.error { color: var(--text-error); }
</style>
