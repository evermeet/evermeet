<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api, type AdminObjectItem, type AdminObjectList, type AdminObjectType } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { onMount } from 'svelte';

	const type = $derived($page.params.type as AdminObjectType);
	const limit = 50;

	let list = $state<AdminObjectList | null>(null);
	let loading = $state(true);
	let error = $state('');
	let offset = $state(0);

	onMount(async () => {
		await auth.load();
		if (!auth.user) {
			goto(`/auth/login?next=/admin/objects/${type}`);
			return;
		}
		if (!auth.user.is_admin) {
			goto('/');
			return;
		}
		await loadPage(0);
	});

	async function loadPage(nextOffset: number) {
		loading = true;
		error = '';
		try {
			list = await api.admin.objectsByType(type, limit, nextOffset);
			offset = nextOffset;
		} catch (err: any) {
			error = err.message ?? 'Failed to load objects';
		} finally {
			loading = false;
		}
	}

	function itemMeta(item: AdminObjectItem) {
		const parts: string[] = [];
		if (item.subtitle) parts.push(item.subtitle);
		if (item.type === 'blobs' && typeof item.meta?.size === 'number') parts.unshift(formatBytes(item.meta.size));
		if (item.meta?.visibility) parts.push(String(item.meta.visibility));
		if (item.meta?.uploaded_by) parts.push(`uploaded by ${item.meta.uploaded_by}`);
		return parts.join(' · ');
	}

	function formatBytes(size: number) {
		if (size < 1024) return `${size} B`;
		if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
		return `${(size / 1024 / 1024).toFixed(1)} MB`;
	}

	const canPrev = $derived(offset > 0);
	const canNext = $derived(!!list && offset + limit < list.count);
</script>

<main>
	<div class="page-header">
		<div>
			<p class="eyebrow">Admin</p>
			<h1>{list?.label ?? type}</h1>
			<p class="muted">Paginated database object listing.</p>
		</div>
	</div>
	<AdminNav active="objects" />

	<div class="back-row">
		<a href="/admin/objects">← Back to Objects</a>
	</div>

	{#if loading && !list}
		<p class="muted">Loading objects...</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else if list}
		<section class="panel">
			<div class="list-header">
				<p class="muted">Showing {offset + 1}-{Math.min(offset + limit, list.count)} of {list.count}</p>
				<div class="pager">
					<button disabled={!canPrev || loading} onclick={() => loadPage(Math.max(0, offset - limit))}>Previous</button>
					<button disabled={!canNext || loading} onclick={() => loadPage(offset + limit)}>Next</button>
				</div>
			</div>

			{#if list.items.length === 0}
				<p class="muted">No objects found.</p>
			{:else}
				<div class="object-list">
					{#each list.items as item}
						<div class="object-row">
							<div>
								<strong>{item.label}</strong>
								<span class="mono">{item.id}</span>
							</div>
							<div class="row-meta">
								{#if itemMeta(item)}
									<span>{itemMeta(item)}</span>
								{/if}
								{#if item.updated_at}
									<time>{new Date(item.updated_at).toLocaleString()}</time>
								{/if}
							</div>
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
		text-transform: capitalize;
	}

	.muted {
		color: var(--text-secondary);
	}

	.error {
		color: var(--text-error);
	}

	.back-row {
		margin-bottom: 1rem;
	}

	.back-row a {
		color: var(--text-accent);
		font-weight: 700;
		text-decoration: none;
	}

	.panel {
		padding: 1.25rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-lg);
		background: var(--bg-card);
	}

	.list-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.pager {
		display: flex;
		gap: 0.5rem;
	}

	button {
		padding: 0.5rem 0.85rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		background: var(--bg-btn-secondary);
		color: var(--text-btn-secondary);
		font-weight: 700;
		cursor: pointer;
	}

	button:disabled {
		opacity: 0.5;
		cursor: default;
	}

	.object-list {
		display: flex;
		flex-direction: column;
		border-top: 1px solid var(--border-subtle);
	}

	.object-row {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.9rem 0;
		border-bottom: 1px solid var(--border-subtle);
	}

	.object-row strong,
	.object-row span,
	.object-row time {
		display: block;
	}

	.mono {
		margin-top: 0.2rem;
		color: var(--text-muted);
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 0.8rem;
		overflow-wrap: anywhere;
	}

	.row-meta {
		flex: 0 0 260px;
		color: var(--text-muted);
		font-size: 0.85rem;
		text-align: right;
	}
</style>
