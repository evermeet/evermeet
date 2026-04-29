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
			</div>
		</div>
	{/if}
</main>

<style>
	main {
		max-width: 600px;
		margin: 3rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	.profile-card {
		text-align: center;
		padding: 2.5rem;
		background: #fff;
		border: 1px solid #eee;
		border-radius: 12px;
		box-shadow: 0 4px 12px rgba(0,0,0,0.03);
	}
	.avatar-wrapper {
		margin-bottom: 1.5rem;
		display: flex;
		justify-content: center;
	}
	h1 { font-size: 1.75rem; font-weight: 800; margin: 0 0 0.5rem; }
	.did {
		display: inline-block;
		font-size: 0.8rem;
		background: #f4f4f4;
		padding: 0.3rem 0.6rem;
		border-radius: 4px;
		color: #666;
		margin-bottom: 1.5rem;
		max-width: 100%;
		word-break: break-all;
	}
	.bio {
		font-size: 1.1rem;
		line-height: 1.6;
		color: #444;
		margin: 1.5rem 0;
		white-space: pre-wrap;
	}
	.meta {
		margin-top: 2rem;
		padding-top: 1.5rem;
		border-top: 1px solid #eee;
		font-size: 0.85rem;
	}
	.muted { color: #999; }
	.error { color: #c00; }
</style>
