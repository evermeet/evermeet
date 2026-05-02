<script lang="ts">
	let { active = 'overview' }: { active?: 'overview' | 'objects' | 'email' | 'admins' | 'p2p' | 'config' } = $props();
</script>

<nav class="admin-nav" aria-label="Admin sections">
	<a href="/admin" class:active={active === 'overview'}>Overview</a>
	<a href="/admin/p2p" class:active={active === 'p2p'}>Network</a>
	<a href="/admin/objects" class:active={active === 'objects'}>Objects</a>
	<a href="/admin/email" class:active={active === 'email'}>Email</a>
	<a href="/admin/admins" class:active={active === 'admins'}>Admins</a>
	<a href="/admin/config" class:active={active === 'config'}>Config</a>
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
