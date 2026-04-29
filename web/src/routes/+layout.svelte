<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/auth.svelte.js';
	import Avatar from '$lib/components/Avatar.svelte';

	let { children } = $props();

	onMount(() => auth.load());
</script>

<nav>
	<a href="/">Evermeet</a>
	<div class="nav-right">
		{#if !auth.loading}
			{#if auth.user}
				<a href="/events/create">+ New event</a>
				<a href="/settings">Settings</a>
				<a href="/u/{auth.user.did}" class="user-profile">
					<Avatar src={auth.user.avatar} did={auth.user.did} size={28} />
					<span class="did">{auth.user.display_name || auth.user.did.slice(0, 10) + '…'}</span>
				</a>
				<button onclick={() => auth.logout()}>Sign out</button>
			{:else}
				<a href="/auth/login">Sign in</a>
			{/if}
		{/if}
	</div>
</nav>

<div class="content">
	{@render children()}
</div>

<footer>
	<div class="footer-links">
		<a href="/node">Node Status</a>
		<span>•</span>
		<span class="version">v0.1.0-alpha</span>
	</div>
	<p class="muted">Evermeet — Decentralized Event Platform</p>
</footer>

<style>
	nav {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.75rem 1.5rem;
		border-bottom: 1px solid #e5e5e5;
		font-family: system-ui, sans-serif;
	}
	nav a {
		color: inherit;
		text-decoration: none;
		font-weight: 600;
	}
	.nav-right {
		display: flex;
		align-items: center;
		gap: 1rem;
		font-size: 0.9rem;
	}
	.nav-right a {
		font-weight: 400;
		color: #555;
	}
	.nav-right a:hover { color: #111; }
	.user-profile {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		text-decoration: none;
		color: inherit;
		padding: 0.2rem 0.5rem;
		border-radius: 6px;
	}
	.user-profile:hover {
		background: #f4f4f4;
	}
	.did { color: #888; font-size: 0.8rem; font-weight: 500; }
	button {
		border: none;
		background: none;
		cursor: pointer;
		color: #555;
		font-size: 0.9rem;
		padding: 0;
	}
	button:hover { color: #111; }

	.content {
		min-height: calc(100vh - 180px);
	}

	footer {
		margin-top: 4rem;
		padding: 3rem 1.5rem;
		border-top: 1px solid #e5e5e5;
		text-align: center;
		font-family: system-ui, sans-serif;
	}
	.footer-links {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 0.5rem;
		font-size: 0.875rem;
		color: #666;
	}
	.footer-links a {
		color: #666;
		text-decoration: none;
	}
	.footer-links a:hover { color: #111; }
	.version {
		font-family: monospace;
		color: #999;
	}
	.muted {
		font-size: 0.8rem;
		color: #aaa;
		margin: 0;
	}
</style>
