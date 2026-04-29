<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	let status = $state<'verifying' | 'error'>('verifying');
	let error = $state('');

	onMount(async () => {
		const token = new URLSearchParams(window.location.search).get('token');
		if (!token) {
			status = 'error';
			error = 'Missing token.';
			return;
		}
		// The Go backend handles /api/auth/magic-link/verify and redirects to /.
		// This page is only reached if SvelteKit intercepts the redirect.
		// In SPA mode, navigate directly.
		window.location.href = `/api/auth/magic-link/verify?token=${token}`;
	});
</script>

<main>
	{#if status === 'verifying'}
		<p>Signing you in…</p>
	{:else}
		<p class="error">{error}</p>
		<a href="/auth/login">Try again</a>
	{/if}
</main>

<style>
	main {
		max-width: 400px;
		margin: 4rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	.error { color: #c00; }
</style>
