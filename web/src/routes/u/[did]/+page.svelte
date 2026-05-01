<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api.js';
	import Avatar from '$lib/components/Avatar.svelte';

	let user = $state<any>(null);
	let loading = $state(true);
	let error = $state('');

	const did = $derived($page.params.did);

	onMount(async () => {
		try {
			user = await api.users.get(did);
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	});
</script>

<main>
	{#if loading}
		<p class="muted">Loading profile…</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else if user}
		<div class="profile-card">
			<div class="avatar-wrapper">
				<Avatar src={user.avatar} did={user.did} size={120} />
			</div>

			<h1>{user.display_name || 'Anonymous User'}</h1>
			<code class="did">{user.did}</code>

			{#if user.bio}
				<p class="bio">{user.bio}</p>
			{/if}

			<div class="meta">
				<span class="muted">Joined {new Date(user.updated_at).toLocaleDateString()}</span>
				{#if user.instance_id}
					<span class="muted">•</span>
					<span class="muted">Home: {user.instance_id}</span>
				{/if}
			</div>
		</div>
	{/if}
</main>

<style>
	main {
		max-width: var(--layout-page-narrow);
		margin: 3rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	.profile-card {
		text-align: center;
		padding: 2.5rem;
		background: var(--bg-card);
		border: 1px solid var(--border-card);
		border-radius: var(--radius-2xl);
		box-shadow: var(--shadow-card);
	}
	.avatar-wrapper {
		margin-bottom: 1.5rem;
		display: flex;
		justify-content: center;
	}
	h1 { font-size: 1.75rem; font-weight: 800; margin: 0 0 0.5rem; color: var(--text); }
	.did {
		display: inline-block;
		font-size: 0.8rem;
		background: var(--bg-code);
		color: var(--text-subtle);
		padding: 0.3rem 0.6rem;
		border-radius: var(--radius-sm);
		margin-bottom: 1.5rem;
		max-width: 100%;
		word-break: break-all;
	}
	.bio {
		font-size: 1.1rem;
		line-height: 1.6;
		color: var(--text-secondary);
		margin: 1.5rem 0;
		white-space: pre-wrap;
	}
	.meta {
		margin-top: 2rem;
		padding-top: 1.5rem;
		border-top: 1px solid var(--border-subtle);
		font-size: 0.85rem;
		display: flex;
		justify-content: center;
		gap: 0.5rem;
	}
	.muted { color: var(--text-muted); }
	.error { color: var(--text-error); }
</style>
