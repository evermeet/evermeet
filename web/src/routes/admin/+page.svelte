<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, type AdminOverview } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import { onMount } from 'svelte';

	let overview = $state<AdminOverview | null>(null);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		await auth.load();
		if (!auth.user) {
			goto('/auth/login?next=/admin');
			return;
		}
		if (!auth.user.is_admin) {
			goto('/');
			return;
		}
		try {
			overview = await api.admin.overview();
		} catch (err: any) {
			error = err.message ?? 'Failed to load admin overview';
		} finally {
			loading = false;
		}
	});

	const countCards = $derived(
		overview
			? [
					['Admins', overview.counts.admins],
					['Users', overview.counts.users],
					['Events', overview.counts.events],
					['Calendars', overview.counts.calendars],
					['Blobs', overview.counts.blobs],
				]
			: []
	);

	const p2pPeers = $derived(Array.isArray(overview?.p2p?.peers) ? overview.p2p.peers : []);
	const p2pAddresses = $derived(Array.isArray(overview?.p2p?.addresses) ? overview.p2p.addresses : []);
</script>

<main>
	<div class="page-header">
		<div>
			<p class="eyebrow">Admin</p>
			<h1>Instance Admin</h1>
			<p class="muted">Basic status and operational overview for this Evermeet instance.</p>
		</div>
	</div>

	{#if loading}
		<p class="muted">Loading admin overview...</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else if overview}
		<section class="stats-grid" aria-label="Instance counts">
			{#each countCards as [label, value]}
				<div class="stat-card">
					<span>{label}</span>
					<strong>{value}</strong>
				</div>
			{/each}
		</section>

		<section class="panel">
			<h2>Instance</h2>
			<div class="info-grid">
				<div class="info-row">
					<span class="label">Instance ID</span>
					<span class="value mono">{overview.instance_id}</span>
				</div>
				<div class="info-row">
					<span class="label">Base URL</span>
					<span class="value mono">{overview.base_url}</span>
				</div>
				<div class="info-row">
					<span class="label">Version</span>
					<span class="value"><span class="badge">{overview.version}</span></span>
				</div>
				<div class="info-row">
					<span class="label">Uptime</span>
					<span class="value"><strong>{overview.uptime}</strong></span>
				</div>
				<div class="info-row">
					<span class="label">Started At</span>
					<span class="value mono">{new Date(overview.started_at).toLocaleString()}</span>
				</div>
				<div class="info-row">
					<span class="label">P2P Peer ID</span>
					<span class="value mono">{overview.p2p?.id ?? 'not initialized'}</span>
				</div>
				<div class="info-row">
					<span class="label">Connected Peers</span>
					<span class="value">{p2pPeers.length}</span>
				</div>
			</div>
		</section>

		<section class="panel">
			<h2>Active Configuration</h2>
			<div class="config-group">
				<h3>Node</h3>
				<div class="info-grid">
					{#each Object.entries(overview.config.node) as [key, value]}
						<div class="info-row">
							<span class="label">{key}</span>
							<span class="value mono">{value}</span>
						</div>
					{/each}
				</div>
			</div>
			<div class="config-group">
				<h3>P2P</h3>
				<div class="info-grid">
					<div class="info-row">
						<span class="label">listen_port</span>
						<span class="value mono">{overview.config.p2p.listen_port}</span>
					</div>
					<div class="info-row">
						<span class="label">bootstrap_peers</span>
						<span class="value mono">
							{#if overview.config.p2p.bootstrap_peers?.length}
								{overview.config.p2p.bootstrap_peers.join(', ')}
							{:else}
								<em class="muted">none (mDNS only)</em>
							{/if}
						</span>
					</div>
				</div>
			</div>
		</section>

		<section class="panel">
			<h2>P2P Network Node</h2>
			<p><span class="label">Listening Addresses</span></p>
			{#if p2pAddresses.length}
				<ul class="addr-list">
					{#each p2pAddresses as addr}
						<li class="mono">{addr}</li>
					{/each}
				</ul>
			{:else}
				<p class="muted">No listening addresses reported.</p>
			{/if}
		</section>

		<section class="panel">
			<h2>Connected Peers ({p2pPeers.length})</h2>
			{#if p2pPeers.length === 0}
				<p class="muted">No peers connected yet. mDNS is active - try starting another node on the same network.</p>
			{:else}
				<div class="peers-list">
					{#each p2pPeers as peer}
						<div class="peer-card">
							<p class="mono peer-id">{peer.id}</p>
							{#if peer.addresses?.length}
								<details>
									<summary>Show Addresses ({peer.addresses.length})</summary>
									<ul class="addr-list">
										{#each peer.addresses as addr}
											<li class="mono">{addr}</li>
										{/each}
									</ul>
								</details>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</section>
	{/if}
</main>

<style>
	main {
		max-width: 980px;
		margin: 2.5rem auto;
		padding: 0 1.5rem 4rem;
		font-family: system-ui, sans-serif;
	}

	.page-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		margin-bottom: 1.5rem;
	}

	.eyebrow {
		margin: 0 0 0.35rem;
		color: var(--text-muted);
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	h1 {
		margin: 0;
		color: var(--text);
		font-size: 2rem;
	}

	h2 {
		margin: 0 0 1rem;
		color: var(--text);
		font-size: 1.1rem;
	}

	h3 {
		margin: 1rem 0 0.5rem;
		color: var(--text-heading-sub);
		font-size: 0.85rem;
		font-weight: 600;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}

	.muted {
		color: var(--text-secondary);
	}

	.error {
		color: var(--text-error);
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.stat-card,
	.panel {
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-lg);
		background: var(--bg-card);
	}

	.stat-card {
		padding: 1rem;
	}

	.stat-card span {
		display: block;
		color: var(--text-muted);
		font-size: 0.85rem;
		font-weight: 600;
	}

	.stat-card strong {
		display: block;
		margin-top: 0.35rem;
		color: var(--text);
		font-size: 1.8rem;
	}

	.panel {
		padding: 1.25rem;
		margin-top: 1rem;
	}

	.info-grid {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}

	.info-row {
		display: flex;
		gap: 1rem;
		align-items: baseline;
	}

	.label {
		color: var(--text-muted);
		font-size: 0.9rem;
		font-weight: 600;
		min-width: 140px;
		flex-shrink: 0;
	}

	.value {
		color: var(--text);
		word-break: break-all;
	}

	.mono {
		color: var(--text);
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 0.85rem;
	}

	.badge {
		display: inline-block;
		padding: 0.15rem 0.5rem;
		border-radius: var(--radius-full);
		background: var(--bg-badge);
		color: var(--text-badge);
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 0.75rem;
	}

	.config-group {
		border-top: 1px solid var(--border-separator);
		padding-top: 0.75rem;
		margin-top: 0.75rem;
	}

	.config-group:first-child {
		border-top: none;
		padding-top: 0;
		margin-top: 0;
	}

	.addr-list {
		list-style: none;
		padding: 0;
		margin: 0.5rem 0 0;
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.peer-card {
		border: 1px solid var(--border-subtle);
		border-left: 3px solid var(--border-peer);
		border-radius: var(--radius-md);
		padding: 1rem;
		background: var(--bg);
	}

	.peer-id {
		word-break: break-all;
		margin: 0 0 0.5rem;
	}

	.peers-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	details summary {
		cursor: pointer;
		color: var(--text-accent);
		font-size: 0.8rem;
		margin-top: 0.25rem;
	}
</style>
