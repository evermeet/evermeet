<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, type AdminOverview, type AdminP2PPeer } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { onMount } from 'svelte';

	let overview = $state<AdminOverview | null>(null);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		await auth.load();
		if (!auth.user) {
			goto('/auth/login?next=/admin/p2p');
			return;
		}
		if (!auth.user.is_admin) {
			goto('/');
			return;
		}
		try {
			overview = await api.admin.overview();
		} catch (err: any) {
			error = err.message ?? 'Failed to load P2P status';
		} finally {
			loading = false;
		}
	});

	const p2pPeers = $derived(
		Array.isArray(overview?.p2p?.peers) ? (overview.p2p.peers as AdminP2PPeer[]) : []
	);
	const p2pAddresses = $derived(Array.isArray(overview?.p2p?.addresses) ? overview.p2p.addresses : []);
	const evermeetInstanceId = $derived(
		overview?.p2p?.evermeet_instance_id ?? overview?.instance_id ?? ''
	);
</script>

<main>
	<div class="page-header">
		<div>
			<p class="eyebrow">Admin</p>
			<h1>P2P</h1>
			<p class="muted">Libp2p node identity, listen addresses, and connected peers.</p>
		</div>
	</div>
	<AdminNav active="p2p" />

	{#if loading}
		<p class="muted">Loading P2P status...</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else if overview}
		<section class="panel">
			<h2>Status</h2>
			<div class="info-grid">
				<div class="info-row">
					<span class="label">Evermeet instance</span>
					<span class="value mono">{evermeetInstanceId || '—'}</span>
				</div>
				<div class="info-row">
					<span class="label">Libp2p Peer ID</span>
					<span class="value mono">{overview.p2p?.libp2p_peer_id ?? 'not initialized'}</span>
				</div>
				<div class="info-row">
					<span class="label">Connected peers</span>
					<span class="value">{p2pPeers.length}</span>
				</div>
			</div>
		</section>

		<section class="panel">
			<h2>Configuration</h2>
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
		</section>

		<section class="panel">
			<h2>Listening addresses</h2>
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
			<h2>Connected peers ({p2pPeers.length})</h2>
			{#if p2pPeers.length === 0}
				<p class="muted">No peers connected yet. mDNS is active — try starting another node on the same network.</p>
			{:else}
				<div class="peers-list">
					{#each p2pPeers as peer}
						<div class="peer-card">
							<div class="peer-info-grid">
								<div class="info-row">
									<span class="label">Evermeet instance</span>
									<span class="value mono">{peer.evermeet_instance_id || '—'}</span>
								</div>
								<div class="info-row">
									<span class="label">Libp2p Peer ID</span>
									<span class="value mono">{peer.libp2p_peer_id}</span>
								</div>
							</div>
							{#if peer.libp2p_fingerprint}
								<p class="peer-meta">
									<span class="meta-label">Libp2p key fingerprint</span>
									<code class="mono instance-id">{peer.libp2p_fingerprint}</code>
								</p>
							{/if}
							{#if peer.addresses?.length}
								<details>
									<summary>Show addresses ({peer.addresses.length})</summary>
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
		margin-bottom: 0.75rem;
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

	.muted {
		color: var(--text-secondary);
	}

	.error {
		color: var(--text-error);
	}

	.panel {
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-lg);
		background: var(--bg-card);
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

	.peer-info-grid {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		margin-bottom: 0.35rem;
	}

	.peer-meta {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.4rem 0.6rem;
		margin: 0 0 0.5rem;
		font-size: 0.85rem;
	}

	.meta-label {
		color: var(--text-muted);
		font-weight: 600;
	}

	.instance-id {
		color: var(--text);
		font-size: 0.85rem;
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
