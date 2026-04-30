<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/auth.svelte.js';
	import { theme } from '$lib/theme.svelte.js';
	import Avatar from '$lib/components/Avatar.svelte';
	import '../app.css';

	let { children } = $props();

	let menuOpen = $state(false);
	let menuRef = $state<HTMLElement | null>(null);
	let clock = $state('');

	onMount(() => {
		auth.load();
		theme.load();
		updateClock();
		const t = setInterval(updateClock, 1000);
		return () => clearInterval(t);
	});

	function updateClock() {
		const now = new Date();
		const time = now.toLocaleTimeString('en', { hour: '2-digit', minute: '2-digit', hour12: false });
		const tz = now.toLocaleDateString('en', { timeZoneName: 'short' }).split(', ')[1] ?? '';
		clock = `${time} ${tz}`;
	}

	function toggleMenu() { menuOpen = !menuOpen; }

	function closeMenu() { menuOpen = false; }

	function handleLogout() {
		closeMenu();
		auth.logout();
	}

	function handleClickOutside(e: MouseEvent) {
		if (menuRef && !menuRef.contains(e.target as Node)) closeMenu();
	}
</script>

<svelte:window onclick={handleClickOutside} />

<nav>
	<!-- Left: logo -->
	<a href="/" class="logo" aria-label="Evermeet home">
		<svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
			<path d="M10 2L3 6v8l7 4 7-4V6L10 2z" fill="currentColor" opacity="0.9"/>
		</svg>
	</a>

	<!-- Center: main nav (only when logged in) -->
	{#if !auth.loading && auth.user}
		<div class="nav-center">
			<a href="/" class="nav-item">
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
				Events
			</a>
			<a href="/calendars" class="nav-item">
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/><circle cx="8" cy="14" r="1" fill="currentColor"/><circle cx="12" cy="14" r="1" fill="currentColor"/></svg>
				Calendars
			</a>
			<a href="/discover" class="nav-item">
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"/></svg>
				Discover
			</a>
		</div>
	{/if}

	<!-- Right -->
	<div class="nav-right">
		{#if clock}
			<span class="clock">{clock}</span>
		{/if}

		{#if !auth.loading}
			{#if auth.user}
				<a href="/events/create" class="btn-create">Create Event</a>
				<button class="icon-btn" title="Search" aria-label="Search">
					<svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
				</button>
				<button class="icon-btn" title="Notifications" aria-label="Notifications">
					<svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
				</button>
				<div class="avatar-menu" bind:this={menuRef}>
					<button class="avatar-btn" onclick={toggleMenu} aria-label="User menu" aria-expanded={menuOpen}>
						<Avatar src={auth.user.avatar} did={auth.user.did} size={32} />
					</button>
					{#if menuOpen}
						<div class="dropdown" role="menu">
							<div class="dropdown-header">
								<span class="dropdown-name">{auth.user.display_name || 'Anonymous'}</span>
								<span class="dropdown-did">{auth.user.did.slice(0, 20)}…</span>
							</div>
							<div class="dropdown-divider"></div>
							<a href="/u/{auth.user.did}" class="dropdown-item" onclick={closeMenu} role="menuitem">
								<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
								Profile
							</a>
							<a href="/settings" class="dropdown-item" onclick={closeMenu} role="menuitem">
								<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
								Settings
							</a>
							<div class="dropdown-divider"></div>
							<button class="dropdown-item dropdown-logout" onclick={handleLogout} role="menuitem">
								<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
								Sign out
							</button>
						</div>
					{/if}
				</div>
			{:else}
				<a href="/discover" class="nav-item">Explore events</a>
				<a href="/auth/login" class="btn-signin">Sign In</a>
			{/if}
		{/if}
	</div>
</nav>

<div class="content">
	{@render children()}
</div>

<footer>
	<div class="footer-links">
		<a href="/instance">Instance Status</a>
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
		padding: 0 1.5rem;
		height: 56px;
		border-bottom: 1px solid var(--border);
		font-family: system-ui, sans-serif;
		background: var(--bg);
		position: sticky;
		top: 0;
		z-index: 100;
	}

	/* Logo */
	.logo {
		display: flex;
		align-items: center;
		color: var(--text);
		text-decoration: none;
		flex-shrink: 0;
		width: 160px;
	}
	.logo:hover { opacity: 0.7; }

	/* Center nav */
	.nav-center {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		position: absolute;
		left: 50%;
		transform: translateX(-50%);
	}
	.nav-item {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		padding: 0.45rem 0.75rem;
		border-radius: var(--radius-md);
		color: var(--text-secondary);
		text-decoration: none;
		font-size: 0.875rem;
		font-weight: 500;
		transition: color 0.1s, background 0.1s;
	}
	.nav-item:hover {
		color: var(--text);
		background: var(--bg-hover);
	}
	.nav-item svg { flex-shrink: 0; opacity: 0.7; }

	/* Right side */
	.nav-right {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-shrink: 0;
		width: 160px;
		justify-content: flex-end;
	}

	.clock {
		font-size: 0.8rem;
		color: var(--text-muted);
		font-variant-numeric: tabular-nums;
		margin-right: 0.25rem;
		white-space: nowrap;
	}

	.btn-create {
		padding: 0.4rem 0.85rem;
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border-radius: var(--radius-md);
		font-size: 0.825rem;
		font-weight: 600;
		text-decoration: none;
		white-space: nowrap;
		transition: opacity 0.1s;
	}
	.btn-create:hover { opacity: 0.85; }

	.btn-signin {
		padding: 0.4rem 0.85rem;
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border-radius: var(--radius-md);
		font-size: 0.825rem;
		font-weight: 600;
		text-decoration: none;
		white-space: nowrap;
	}
	.btn-signin:hover { opacity: 0.85; }

	.icon-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border: none;
		background: none;
		cursor: pointer;
		color: var(--text-secondary);
		border-radius: var(--radius-md);
		padding: 0;
		transition: color 0.1s, background 0.1s;
	}
	.icon-btn:hover { color: var(--text); background: var(--bg-hover); }

	/* Avatar + dropdown */
	.avatar-menu {
		position: relative;
	}
	.avatar-btn {
		display: flex;
		align-items: center;
		border: none;
		background: none;
		cursor: pointer;
		padding: 0;
		border-radius: 50%;
		transition: opacity 0.1s;
	}
	.avatar-btn:hover { opacity: 0.8; }

	.dropdown {
		position: absolute;
		top: calc(100% + 8px);
		right: 0;
		min-width: 200px;
		background: var(--bg-raised);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		box-shadow: 0 8px 24px rgba(0,0,0,0.12);
		overflow: hidden;
		z-index: 200;
	}
	.dropdown-header {
		padding: 0.75rem 1rem;
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
	}
	.dropdown-name {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text);
	}
	.dropdown-did {
		font-size: 0.72rem;
		color: var(--text-muted);
		font-family: monospace;
	}
	.dropdown-divider {
		height: 1px;
		background: var(--border-subtle);
		margin: 0;
	}
	.dropdown-item {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.6rem 1rem;
		font-size: 0.875rem;
		color: var(--text);
		text-decoration: none;
		background: none;
		border: none;
		width: 100%;
		text-align: left;
		cursor: pointer;
		transition: background 0.1s;
	}
	.dropdown-item:hover { background: var(--bg-hover); }
	.dropdown-item svg { opacity: 0.6; flex-shrink: 0; }
	.dropdown-logout { color: var(--text-error); }
	.dropdown-logout svg { opacity: 0.7; }

	.content {
		min-height: calc(100vh - 180px);
	}

	footer {
		margin-top: 4rem;
		padding: 3rem 1.5rem;
		border-top: 1px solid var(--border);
		text-align: center;
		font-family: system-ui, sans-serif;
		background: var(--bg);
	}
	.footer-links {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 0.5rem;
		font-size: 0.875rem;
		color: var(--text-subtle);
	}
	.footer-links a {
		color: var(--text-subtle);
		text-decoration: none;
	}
	.footer-links a:hover { color: var(--text); }
	.version {
		font-family: monospace;
		color: var(--text-muted);
	}
	.muted {
		font-size: 0.8rem;
		color: var(--text-faint);
		margin: 0;
	}
</style>
