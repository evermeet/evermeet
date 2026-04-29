<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/auth.svelte.js';

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
				<span class="did">{auth.user.display_name || auth.user.did.slice(0, 20) + '…'}</span>
				<button onclick={() => auth.logout()}>Sign out</button>
			{:else}
				<a href="/auth/login">Sign in</a>
			{/if}
		{/if}
	</div>
</nav>

{@render children()}

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
	.did { color: #888; font-size: 0.8rem; }
	button {
		border: none;
		background: none;
		cursor: pointer;
		color: #555;
		font-size: 0.9rem;
		padding: 0;
	}
	button:hover { color: #111; }
</style>
