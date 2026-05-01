<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	interface InstancePeer {
		evermeet_instance_id?: string;
		libp2p_peer_id: string;
		libp2p_fingerprint?: string;
		addresses: string[];
	}

	interface InstanceP2P {
		evermeet_instance_id: string;
		libp2p_peer_id: string;
		addresses: string[];
		peers: InstancePeer[];
	}

	interface InstanceStatus {
		instance_id: string;
		p2p: InstanceP2P;
	}

	let status = $state<InstanceStatus | null>(null);
	let error = $state<string | null>(null);
	let loading = $state(true);

	async function fetchStatus() {
		try {
			status = await api.node.status();
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

<div class="admin-container">
	<h1>Node Monitoring</h1>

	{#if loading && !status}
		<p>Loading node status...</p>
	{:else if error}
		<div class="error">{error}</div>
	{:else if status}
		<section>
			<h2>Local instance</h2>
			<div class="card">
				<p><strong>Instance id:</strong> <span class="id">{status.instance_id}</span></p>
				<p><strong>Libp2p Peer ID:</strong> <span class="id">{status.p2p.libp2p_peer_id}</span></p>
				<p><strong>Listening addresses:</strong></p>
				<ul>
					{#each status.p2p.addresses as addr}
						<li>{addr}</li>
					{/each}
				</ul>
			</div>
		</section>

		<section>
			<h2>Connected peers ({status.p2p.peers.length})</h2>
			{#if status.p2p.peers.length === 0}
				<p class="empty">No peers connected yet. mDNS is active, try starting another node nearby.</p>
			{:else}
				<div class="peers-list">
					{#each status.p2p.peers as peer}
						<div class="card peer-card">
							<p><strong>Instance id:</strong> <span class="id">{peer.evermeet_instance_id || '—'}</span></p>
							<p><strong>Libp2p Peer ID:</strong> <span class="id muted-sm">{peer.libp2p_peer_id}</span></p>
							<details>
								<summary>Show addresses ({peer.addresses.length})</summary>
								<ul>
									{#each peer.addresses as addr}
										<li>{addr}</li>
									{/each}
								</ul>
							</details>
						</div>
					{/each}
				</div>
			{/if}
		</section>
	{/if}
</div>

<style>
	.admin-container {
		max-width: 800px;
		margin: 2rem auto;
		padding: 0 1rem;
		font-family: system-ui, sans-serif;
	}
	h1 {
		font-size: 1.5rem;
		margin-bottom: 2rem;
	}
	h2 {
		font-size: 1.1rem;
		color: #444;
		margin-top: 2rem;
	}
	.card {
		background: #f9f9f9;
		border: 1px solid #eee;
		border-radius: 8px;
		padding: 1rem;
		margin-bottom: 1rem;
	}
	.id {
		font-family: monospace;
		background: #eee;
		padding: 0.1rem 0.3rem;
		border-radius: 4px;
		word-break: break-all;
	}
	.muted-sm {
		font-size: 0.85rem;
		color: #555;
	}
	ul {
		list-style: none;
		padding: 0;
		margin: 0.5rem 0;
		font-size: 0.85rem;
		color: #666;
	}
	li {
		margin-bottom: 0.25rem;
	}
	.error {
		color: #d32f2f;
		background: #ffebee;
		padding: 1rem;
		border-radius: 8px;
	}
	.empty {
		color: #888;
		font-style: italic;
	}
	summary {
		cursor: pointer;
		color: #3b82f6;
		font-size: 0.85rem;
		margin-top: 0.5rem;
	}
	.peer-card {
		border-left: 4px solid #3b82f6;
	}
</style>
