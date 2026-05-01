<script lang="ts">
	import { onMount } from 'svelte';
	import { intl } from '$lib/i18n.svelte.js';

	let status = $state<'verifying' | 'approved' | 'error'>('verifying');
	let error = $state('');

	onMount(async () => {
		const params = new URLSearchParams(window.location.search);
		if (params.get('approved') === '1') {
			status = 'approved';
			return;
		}
		const token = params.get('token');
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
		<p>{intl.t('auth.lookingUp')}</p>
	{:else if status === 'approved'}
		<h1>{intl.t('auth.magicLinkApproved')}</h1>
		<p>{intl.t('auth.magicLinkApprovedHelp')}</p>
	{:else}
		<p class="error">{error}</p>
		<a href="/auth/login">Try again</a>
	{/if}
</main>

<style>
	main {
		max-width: var(--layout-page-narrow);
		margin: 4rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	.error { color: #c00; }
</style>
