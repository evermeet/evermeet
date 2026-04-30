<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api.js';

	let status = $state<any>(null);
	let error = $state<string | null>(null);
	let loading = $state(true);

	async function fetchStatus() {
		try {
			status = await api.instance.status();
		} catch (e) {
			error = String(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		fetchStatus();
		const interval = setInterval(fetchStatus, 5000);
		return () => clearInterval(interval);
	});
</script>

<svelte:head>
	<title>Instance Status — Evermeet</title>
</svelte:head>

<main>
	<h1>Instance Status</h1>

	{#if loading && !status}
		<p class="muted">Loading…</p>
	{:else if error}
		<div class="error">{error}</div>
	{:else if status}

		<section>
			<h2>Overview</h2>
			<div class="card info-grid">
				<div class="info-row">
					<span class="label">Instance ID</span>
					<span class="value mono">{status.instance_id}</span>
				</div>
				<div class="info-row">
					<span class="label">Version</span>
					<span class="value">
						<span class="badge">{status.version}</span>
					</span>
				</div>
				<div class="info-row">
					<span class="label">Uptime</span>
					<span class="value"><strong>{status.uptime}</strong></span>
				</div>
				<div class="info-row">
					<span class="label">Started At</span>
					<span class="value mono">{new Date(status.started_at).toLocaleString()}</span>
				</div>
				<div class="info-row">
					<span class="label">P2P Peer ID</span>
					<span class="value mono">{status.p2p.id}</span>
				</div>
				<div class="info-row">
					<span class="label">Connected Peers</span>
					<span class="value">{status.p2p.peers.length}</span>
				</div>
			</div>
		</section>

		<section>
			<h2>Active Configuration</h2>
			<div class="card">
				<div class="config-group">
					<h3>Node</h3>
					<div class="info-grid">
						{#each Object.entries(status.config.node) as [key, value]}
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
							<span class="value mono">{status.config.p2p.listen_port}</span>
						</div>
						<div class="info-row">
							<span class="label">bootstrap_peers</span>
							<span class="value mono">
								{#if status.config.p2p.bootstrap_peers?.length}
									{status.config.p2p.bootstrap_peers.join(', ')}
								{:else}
									<em class="muted">none (mDNS only)</em>
								{/if}
							</span>
						</div>
					</div>
				</div>
			</div>
		</section>

		<section>
			<h2>P2P Network Node</h2>
			<div class="card">
				<p><span class="label">Listening Addresses</span></p>
				<ul class="addr-list">
					{#each status.p2p.addresses as addr}
						<li class="mono">{addr}</li>
					{/each}
				</ul>
			</div>
		</section>

		<section>
			<h2>Connected Peers ({status.p2p.peers.length})</h2>
			{#if status.p2p.peers.length === 0}
				<p class="muted">No peers connected yet. mDNS is active — try starting another node on the same network.</p>
			{:else}
				<div class="peers-list">
					{#each status.p2p.peers as peer}
						<div class="card peer-card">
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
		max-width: 820px;
		margin: 2rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 {
		font-size: 1.75rem;
		font-weight: 800;
		margin-bottom: 2rem;
		color: var(--text);
	}
	h2 {
		font-size: 1rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-heading-section);
		margin: 2rem 0 0.75rem;
	}
	h3 {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--text-heading-sub);
		text-transform: uppercase;
		letter-spacing: 0.04em;
		margin: 1rem 0 0.5rem;
	}
	.card {
		background: var(--bg-card);
		border: 1px solid var(--border-card);
		border-radius: var(--radius-xl);
		padding: 1.25rem 1.5rem;
		margin-bottom: 0.5rem;
	}
	.peer-card {
		border-left: 3px solid var(--border-peer);
		margin-bottom: 0.5rem;
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
		font-size: 0.8rem;
		color: var(--text-heading-section);
		min-width: 140px;
		flex-shrink: 0;
		font-weight: 500;
	}
	.value {
		font-size: 0.9rem;
		color: var(--text-value);
		word-break: break-all;
	}
	.mono {
		font-family: monospace;
		font-size: 0.85rem;
		color: var(--text-mono);
	}
	.peer-id {
		word-break: break-all;
		margin: 0 0 0.5rem;
	}
	.badge {
		display: inline-block;
		background: var(--bg-badge);
		color: var(--text-badge);
		font-size: 0.75rem;
		padding: 0.15rem 0.5rem;
		border-radius: var(--radius-full);
		font-family: monospace;
	}
	.addr-list {
		list-style: none;
		padding: 0;
		margin: 0.5rem 0 0;
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	.addr-list li {
		font-size: 0.8rem;
		color: var(--text-addr);
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
	details summary {
		cursor: pointer;
		color: var(--text-accent);
		font-size: 0.8rem;
		margin-top: 0.25rem;
	}
	.peers-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.muted { color: var(--text-muted); }
	.error {
		color: var(--text-error);
		background: var(--bg-error);
		padding: 1rem;
		border-radius: var(--radius-lg);
	}
</style>
