<script lang="ts">
	import { intl } from '$lib/i18n.svelte.js';
	let { active = 'overview' }: { active?: 'overview' | 'objects' | 'email' | 'admins' | 'p2p' | 'config' } = $props();
</script>

<nav class="admin-nav" aria-label="Admin sections">
	<a href="/admin" class:active={active === 'overview'}>{intl.t('admin.nav.overview')}</a>
	<a href="/admin/p2p" class:active={active === 'p2p'}>{intl.t('admin.nav.network')}</a>
	<a href="/admin/objects" class:active={active === 'objects'}>{intl.t('admin.nav.objects')}</a>
	<a href="/admin/email" class:active={active === 'email'}>{intl.t('admin.nav.email')}</a>
	<a href="/admin/admins" class:active={active === 'admins'}>{intl.t('admin.nav.admins')}</a>
	<a href="/admin/config" class:active={active === 'config'}>{intl.t('admin.nav.config')}</a>
</nav>

<style>
	.admin-nav {
		display: flex;
		gap: 0.5rem;
		margin: 0 0 1.25rem;
		border-bottom: 1px solid var(--border-subtle);
	}

	a {
		padding: 0.7rem 0.9rem;
		color: var(--text-muted);
		text-decoration: none;
		font-size: 0.95rem;
		font-weight: 700;
		border-bottom: 2px solid transparent;
	}

	a:hover {
		color: var(--text);
	}

	a.active {
		color: var(--text);
		border-bottom-color: var(--text-accent);
	}
</style>
