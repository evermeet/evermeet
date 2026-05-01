<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api, type EventRevision } from '$lib/api.js';

	const id = $derived($page.params.id);

	let revisions = $state<EventRevision[]>([]);
	let loading = $state(true);
	let error = $state('');
	let selectedIndex = $state<number | null>(null);

	onMount(async () => {
		try {
			const res = await api.events.history(id);
			revisions = res.revisions ?? [];
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	function shortHash(hash: string) {
		return `${hash.slice(0, 16)}…`;
	}

	function openDetails(index: number) {
		selectedIndex = index;
	}

	function closeDetails() {
		selectedIndex = null;
	}

	type DiffRow = { path: string; before: unknown; after: unknown };
	const ignoredDiffKeys = new Set(['prev', 'updated_at', 'sigs']);

	function isObject(v: unknown): v is Record<string, unknown> {
		return typeof v === 'object' && v !== null && !Array.isArray(v);
	}

	function diffObjects(before: unknown, after: unknown, prefix = ''): DiffRow[] {
		if (JSON.stringify(before) === JSON.stringify(after)) return [];

		if (!isObject(before) || !isObject(after)) {
			return [{ path: prefix || '(root)', before, after }];
		}

		const keys = new Set([...Object.keys(before), ...Object.keys(after)]);
		const rows: DiffRow[] = [];
		for (const key of keys) {
			if (ignoredDiffKeys.has(key)) continue;
			const nextPath = prefix ? `${prefix}.${key}` : key;
			const b = before[key];
			const a = after[key];
			if (JSON.stringify(b) === JSON.stringify(a)) continue;
			if (isObject(b) && isObject(a)) {
				rows.push(...diffObjects(b, a, nextPath));
				continue;
			}
			rows.push({ path: nextPath, before: b, after: a });
		}
		return rows;
	}

	function revisionDiff(index: number): DiffRow[] {
		const current = revisions[index];
		const previous = revisions[index + 1];
		if (!current || !previous) return [];
		return diffObjects(previous.state, current.state);
	}

	function signatureCount(rev: EventRevision): number {
		const sigs = (rev.state as any)?.sigs;
		return Array.isArray(sigs) ? sigs.length : 0;
	}
</script>

<main>
	<h1>Event Edit History</h1>
	<p class="muted">Event: <a href="/events/{id}" class="event-link"><code>{id}</code></a></p>

	{#if loading}
		<p class="muted">Loading revisions…</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else if revisions.length === 0}
		<p class="muted">No revisions found.</p>
	{:else}
		<div class="table-wrap">
			<table>
				<thead>
					<tr>
						<th>Revision</th>
						<th>Updated</th>
						<th>Title</th>
						<th>Status</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{#each revisions as rev, i}
						<tr>
							<td><code>{shortHash(rev.hash)}</code></td>
							<td>{new Date(rev.created_at).toLocaleString()}</td>
							<td>{rev.state.title}</td>
							<td>{rev.is_current ? 'Current' : `Past · ${signatureCount(rev)} sig`}</td>
							<td>
								<button class="btn-link" type="button" onclick={() => openDetails(i)}>
									Show details
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}

	{#if selectedIndex !== null}
		{@const selected = revisions[selectedIndex]}
		<div class="modal-backdrop" role="button" tabindex="0" onclick={closeDetails} onkeydown={(e) => e.key === 'Escape' && closeDetails()}>
			<div class="modal" role="dialog" tabindex="-1" aria-modal="true" aria-label="Revision details" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()}>
				<div class="modal-head">
					<h2>Revision Details</h2>
					<button type="button" class="close-btn" onclick={closeDetails}>Close</button>
				</div>

				<p class="muted"><strong>Hash:</strong> <code>{selected.hash}</code></p>
				<p class="muted"><strong>Updated:</strong> {new Date(selected.created_at).toLocaleString()}</p>
				<p class="muted"><strong>Sigs:</strong> {signatureCount(selected)}</p>

				{#if revisionDiff(selectedIndex).length > 0}
					<p class="diff-title">JSON Diff vs previous revision</p>
					<div class="modal-table-wrap">
						<table class="diff-table">
							<thead>
								<tr>
									<th>Path</th>
									<th>Before</th>
									<th>After</th>
								</tr>
							</thead>
							<tbody>
								{#each revisionDiff(selectedIndex) as row}
									<tr>
										<td><code>{row.path}</code></td>
										<td><code>{JSON.stringify(row.before)}</code></td>
										<td><code>{JSON.stringify(row.after)}</code></td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{:else}
					<p class="muted">No previous revision to diff against.</p>
				{/if}

				<details>
					<summary>Raw state JSON</summary>
					<pre>{JSON.stringify(selected.state, null, 2)}</pre>
				</details>
			</div>
		</div>
	{/if}

	<a class="back" href="/events/{id}">Back to event</a>
</main>

<style>
	main {
		max-width: var(--layout-page-medium);
		margin: 2rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 {
		margin: 0 0 0.5rem;
		color: var(--text);
	}
	.table-wrap { margin-top: 1rem; overflow-x: auto; }
	table {
		width: 100%;
		border-collapse: collapse;
		border: 1px solid var(--border-card);
		border-radius: var(--radius-lg);
		overflow: hidden;
		background: var(--bg-card);
	}
	th, td {
		padding: 0.65rem 0.75rem;
		border-bottom: 1px solid var(--border-subtle);
		text-align: left;
		font-size: 0.9rem;
		vertical-align: top;
	}
	th { color: var(--text-muted); font-weight: 600; font-size: 0.8rem; text-transform: uppercase; }
	.btn-link {
		border: none;
		background: transparent;
		color: var(--text-accent);
		cursor: pointer;
		padding: 0;
	}
	.btn-link:hover { text-decoration: underline; }
	.diff-title { margin: 0.6rem 0 0.45rem; font-weight: 600; color: var(--text); }
	.diff-table th, .diff-table td {
		font-size: 0.8rem;
		padding: 0.45rem 0.55rem;
	}
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.65);
		display: flex;
		align-items: flex-start;
		justify-content: center;
		padding: 3rem 1rem;
		z-index: 50;
	}
	.modal {
		width: min(1100px, 100%);
		max-height: calc(100vh - 6rem);
		overflow: auto;
		background: var(--bg-card);
		border: 1px solid var(--border-card);
		border-radius: var(--radius-lg);
		padding: 1rem;
	}
	.modal-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
	}
	.modal-head h2 {
		margin: 0;
		font-size: 1.1rem;
		color: var(--text);
	}
	.close-btn {
		border: 1px solid var(--border-input);
		background: var(--bg-raised);
		color: var(--text);
		border-radius: var(--radius-md);
		padding: 0.35rem 0.7rem;
		cursor: pointer;
	}
	.close-btn:hover { background: var(--bg-hover); }
	.modal-table-wrap { overflow-x: auto; }
	details { margin-top: 0.65rem; }
	summary { cursor: pointer; color: var(--text); font-weight: 600; }
	pre {
		margin: 0.5rem 0 0;
		background: var(--bg-code);
		color: var(--text-code);
		padding: 0.6rem;
		border-radius: var(--radius-md);
		overflow-x: auto;
		font-size: 0.78rem;
	}
	.back {
		display: inline-block;
		margin-top: 1rem;
		color: var(--text-accent);
		text-decoration: none;
	}
	.back:hover { text-decoration: underline; }
	.muted { color: var(--text-muted); margin: 0.2rem 0; }
	.event-link {
		color: var(--text-accent);
		text-decoration: none;
	}
	.event-link:hover { text-decoration: underline; }
	.error { color: var(--text-error); }
</style>
