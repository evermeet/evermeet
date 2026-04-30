<script lang="ts">
	import Avatar from './Avatar.svelte';

	interface Props {
		id: string;
		title: string;
		starts_at: string;
		location?: { name: string; address?: string } | null;
		cover_url?: string | null;
		hosts?: string[];
		hostProfiles?: Record<string, { displayName: string; avatar: string }>;
	}

	let { id, title, starts_at, location, cover_url, hosts = [], hostProfiles = {} }: Props = $props();

	function formatTime(iso: string) {
		return new Date(iso).toLocaleTimeString('en', { hour: '2-digit', minute: '2-digit' });
	}
</script>

<a href="/events/{id}" class="event-card">
	<div class="event-main">
		<div class="event-time">{formatTime(starts_at)}</div>
		<div class="event-title">{title}</div>
		{#if hosts.length > 0}
			<div class="event-hosts">
				<div class="host-avatars">
					{#each hosts.slice(0, 4) as did}
						<Avatar src={hostProfiles[did]?.avatar} did={did} size={24} />
					{/each}
				</div>
				<span>By {hosts.map(did => hostProfiles[did]?.displayName || did).join(', ')}</span>
			</div>
		{/if}
		{#if location}
			<div class="event-location">
				<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>
				{location.name}
			</div>
		{/if}
	</div>
	{#if cover_url}
		<img src={cover_url} alt={title} class="event-cover" />
	{/if}
</a>

<style>
	.event-card {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		padding: 1rem 0.7rem 1rem 1.25rem;
		border: 1px solid var(--border-card);
		border-radius: var(--radius-xl);
		background: var(--bg-card);
		text-decoration: none;
		transition: border-color 0.1s, background 0.1s;
	}
	.event-card:hover { border-color: var(--border-input); background: var(--bg-raised); }

	.event-main { display: flex; flex-direction: column; gap: 0.5rem; }
	.event-time { font-size: 1rem; color: var(--text-muted); }
	.event-title { font-size: 1.3rem; font-weight: 600; color: var(--text); }

	.event-hosts {
		display: flex;
		align-items: center;
		gap: 0.45rem;
		font-size: 1rem;
		color: var(--text-muted);
	}
	.host-avatars { display: flex; align-items: center; }
	.host-avatars :global(.avatar-box) { border: 2px solid var(--bg-card); }
	.host-avatars :global(.avatar-box:not(:first-child)) { margin-left: -8px; }

	.event-location {
		display: flex;
		align-items: center;
		gap: 0.3rem;
		font-size: 1rem;
		color: var(--text-muted);
	}

	.event-cover {
		width: 120px;
		height: 120px;
		object-fit: cover;
		border-radius: var(--radius-md);
		flex-shrink: 0;
	}
</style>
