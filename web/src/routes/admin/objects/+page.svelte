<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, type AdminObjectSummary } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { onMount } from 'svelte';

	let objects = $state<AdminObjectSummary[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		await auth.load();
		if (!auth.user) {
			goto('/auth/login?next=/admin/objects');
			return;
		}
		if (!auth.user.is_admin) {
			goto('/');
			return;
		}
		try {
			const res = await api.admin.objects();
			objects = res.objects ?? [];
		} catch (err: any) {
			error = err.message ?? 'Failed to load objects';
		} finally {
			loading = false;
		}
	});

	function formatBytes(size: number) {
		if (size < 1024) return `${size} B`;
		if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
		return `${(size / 1024 / 1024).toFixed(1)} MB`;
	}
</script>

<main>
	<div class="page-header">
		<div>
			<p class="eyebrow">Admin</p>
			<h1>Objects</h1>
			<p class="muted">Overview of database objects, counts, and approximate stored bytes by type.</p>
		</div>
	</div>
	<AdminNav active="objects" />

	{#if loading}
		<p class="muted">Loading objects...</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else}
		<div class="objects-list">
			{#each objects as summary}
				<a class="object-row" href={`/admin/objects/${summary.type}`}>
					<div class="object-header">
						<h2>{summary.label}</h2>
					</div>
					<div class="metric-row">
						<div>
							<span class="metric-label">Count</span>
							<strong>{summary.count}</strong>
						</div>
						<div>
							<span class="metric-label">Bytes</span>
							<strong>{formatBytes(summary.bytes)}</strong>
						</div>
						<div>
							<span class="metric-label">Hosted here</span>
							<strong>{summary.hosted_here}</strong>
						</div>
					</div>
					<span class="view-all">View all →</span>
				</a>
			{/each}
		</div>
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
		margin-bottom: 0;
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
		margin: 0;
		color: var(--text);
		font-size: 1.15rem;
	}

	.muted {
		color: var(--text-secondary);
	}

	.error {
		color: var(--text-error);
	}

	.objects-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.object-row {
		display: grid;
		grid-template-columns: minmax(220px, 1.2fr) minmax(340px, 1fr) auto;
		align-items: center;
		gap: 1.5rem;
		padding: 1rem 1.25rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-lg);
		background: var(--bg-card);
		color: inherit;
		text-decoration: none;
		transition: border-color 0.15s, transform 0.15s;
	}

	.object-row:hover {
		border-color: var(--text-accent);
		transform: translateX(2px);
	}

	.object-header {
		display: flex;
		align-items: center;
		min-width: 0;
	}

	.view-all {
		color: var(--text-accent);
		white-space: nowrap;
		font-weight: 700;
	}

	.metric-row {
		display: grid;
		grid-template-columns: repeat(3, minmax(90px, 1fr));
		gap: 2rem;
		margin-bottom: 0;
	}

	.metric-label {
		display: block;
		color: var(--text-muted);
		font-size: 0.8rem;
		font-weight: 700;
	}

	.metric-row strong {
		display: block;
		margin-top: 0.25rem;
		color: var(--text);
		font-size: 1.45rem;
	}

	@media (max-width: 760px) {
		.object-row {
			grid-template-columns: 1fr;
			gap: 1rem;
		}

		.metric-row {
			padding-top: 0.75rem;
			border-top: 1px solid var(--border-subtle);
		}

		.view-all {
			justify-self: start;
		}
	}
</style>
