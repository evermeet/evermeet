<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api, type CalendarLink, type CalendarLinkType } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import ImageUpload from '$lib/components/ImageUpload.svelte';

	const id = $page.params.id;

	const LINK_TYPES: CalendarLinkType[] = ['twitter', 'instagram', 'facebook', 'youtube', 'tiktok', 'linkedin', 'bluesky', 'nostr'];

	let name = $state('');
	let description = $state('');
	let avatar = $state('');
	let backdrop_url = $state('');
	let website = $state('');
	let links = $state<CalendarLink[]>([]);
	let owners = $state<string[]>([]);

	let loading = $state(true);
	let saving = $state(false);
	let error = $state('');

	onMount(async () => {
		try {
			await new Promise<void>(resolve => {
				if (!auth.loading) { resolve(); return; }
				const iv = setInterval(() => { if (!auth.loading) { clearInterval(iv); resolve(); } }, 20);
			});
			const cal = await api.calendars.get(id);
			const isOwner = cal.governance.owners.some(o => o.did === auth.user?.did);
			if (!isOwner) {
				error = 'You are not an owner of this calendar';
				loading = false;
				return;
			}
			name = cal.name;
			description = cal.description ?? '';
			avatar = cal.avatar ?? '';
			backdrop_url = cal.backdrop_url ?? '';
			website = cal.website ?? '';
			links = cal.links ?? [];
			owners = (cal.governance.owners ?? []).map((o) => o.did);
			if (owners.length === 0 && auth.user?.did) {
				owners = [auth.user.did];
			}
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		saving = true;
		error = '';
		try {
			const cleanedOwners = owners.map((o) => o.trim()).filter(Boolean);
			if (cleanedOwners.length === 0) {
				throw new Error('At least one owner is required');
			}
			await api.calendars.update(id, {
				name,
				description,
				avatar,
				backdrop_url,
				website,
				links: links.filter(l => l.url.trim()),
				owners: cleanedOwners
			});
			goto(`/calendars/${id}`);
		} catch (e: any) {
			error = e.message;
		} finally {
			saving = false;
		}
	}

	function addLink() {
		const used = new Set(links.map(l => l.type));
		const next = LINK_TYPES.find(t => !used.has(t));
		if (next) links = [...links, { type: next, url: '' }];
	}

	function removeLink(index: number) {
		links = links.filter((_, i) => i !== index);
	}

	function addOwner() {
		owners = [...owners, ''];
	}

	function removeOwner(index: number) {
		if (owners.length <= 1) return;
		owners = owners.filter((_, i) => i !== index);
	}
</script>

<main>
	<h1>Edit Calendar</h1>

	{#if loading}
		<p class="muted">Loading…</p>
	{:else if error && !name}
		<p class="error">{error}</p>
		<a href="/calendars/{id}" class="cancel">Back</a>
	{:else}
		<form onsubmit={handleSubmit}>
			<div class="field">
				<label for="name">Name</label>
				<input type="text" id="name" bind:value={name} required />
			</div>

			<div class="field">
				<label for="description">Description</label>
				<textarea id="description" bind:value={description}></textarea>
			</div>

			<div class="field">
				<span class="field-label">Avatar</span>
				<ImageUpload bind:value={avatar} square={true} previewSize={120} />
			</div>

			<div class="field">
				<span class="field-label">Backdrop Image</span>
				<ImageUpload bind:value={backdrop_url} previewSize={160} />
			</div>

			<div class="field">
				<label for="website">Website</label>
				<input type="url" id="website" bind:value={website} placeholder="https://…" />
			</div>

			<div class="field">
				<div class="owners-header">
					<p class="owners-title">Social Links</p>
					<button type="button" class="btn-add-owner" onclick={addLink} disabled={links.length >= LINK_TYPES.length}>+ Add link</button>
				</div>
				<div class="owner-list">
					{#each links as link, i}
						<div class="owner-row">
							<select bind:value={links[i].type} class="link-type-select">
								{#each LINK_TYPES as t}
									<option value={t} disabled={t !== link.type && links.some(l => l.type === t)}>{t}</option>
								{/each}
							</select>
							<input type="url" bind:value={links[i].url} placeholder="https://…" />
							<button type="button" class="btn-remove-owner" onclick={() => removeLink(i)}>Remove</button>
						</div>
					{/each}
				</div>
			</div>

			<div class="field">
				<div class="owners-header">
					<p class="owners-title">Owners (DIDs)</p>
					<button type="button" class="btn-add-owner" onclick={addOwner}>+ Add owner</button>
				</div>
				<div class="owner-list">
					{#each owners as owner, i}
						<div class="owner-row">
							<input type="text" bind:value={owners[i]} placeholder="did:em:..." />
							<button type="button" class="btn-remove-owner" onclick={() => removeOwner(i)} disabled={owners.length <= 1}>
								Remove
							</button>
						</div>
					{/each}
				</div>
			</div>

			{#if error}
				<p class="error">{error}</p>
			{/if}

			<div class="actions">
				<button type="submit" disabled={saving}>
					{saving ? 'Saving…' : 'Save Changes'}
				</button>
				<a href="/calendars/{id}" class="cancel">Cancel</a>
			</div>
		</form>
	{/if}
</main>

<style>
	main {
		max-width: 560px;
		margin: 2.5rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 2rem; color: var(--text); }

	form { display: flex; flex-direction: column; gap: 1.25rem; }
	.field { display: flex; flex-direction: column; gap: 0.4rem; }
	label, .field-label { font-size: 0.9rem; font-weight: 600; color: var(--text-label); }

	input, textarea, select {
		padding: 0.6rem 0.75rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		font-size: 0.95rem;
		font-family: inherit;
		background: var(--bg-input);
		color: var(--text);
	}
	input:focus, textarea:focus, select:focus { outline: none; border-color: var(--border-input-focus); }
	.link-type-select { width: 110px; flex-shrink: 0; text-transform: capitalize; cursor: pointer; }
	textarea { min-height: 90px; resize: vertical; }

	.actions { display: flex; align-items: center; gap: 1.5rem; margin-top: 0.5rem; }
	.owners-header { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; }
	.owners-title {
		margin: 0;
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--text-label);
	}
	.owner-list { display: flex; flex-direction: column; gap: 0.5rem; }
	.owner-row { display: flex; gap: 0.5rem; }
	.btn-add-owner, .btn-remove-owner {
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		background: var(--bg-raised);
		color: var(--text);
		padding: 0.35rem 0.6rem;
		font-size: 0.8rem;
		cursor: pointer;
	}
	.btn-add-owner:hover, .btn-remove-owner:hover { background: var(--bg-hover); }
	button {
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border: none;
		padding: 0.7rem 1.5rem;
		border-radius: var(--radius-md);
		font-weight: 600;
		cursor: pointer;
		font-size: 0.95rem;
	}
	button:disabled { opacity: 0.5; cursor: not-allowed; }
	.cancel { text-decoration: none; color: var(--text-subtle); font-size: 0.9rem; }
	.cancel:hover { color: var(--text); }
	.muted { color: var(--text-muted); }
	.error { color: var(--text-error); font-size: 0.9rem; margin: 0; }
</style>
