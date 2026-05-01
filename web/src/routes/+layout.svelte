<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import { intl, localeNames, locales, type Locale } from '$lib/i18n.svelte.js';
	import { theme } from '$lib/theme.svelte.js';
	import Avatar from '$lib/components/Avatar.svelte';
	import '../app.css';

	let { children } = $props();

	let menuOpen = $state(false);
	let menuRef = $state<HTMLElement | null>(null);
	let clock = $state('');

	onMount(() => {
		intl.load();
		checkSetup();
		auth.load();
		theme.load();
		updateClock();
		const t = setInterval(updateClock, 1000);
		return () => clearInterval(t);
	});

	async function checkSetup() {
		try {
			const status = await api.setup.status();
			const path = window.location.pathname;
			if (status.required && path !== '/setup') {
				goto('/setup');
			} else if (!status.required && path === '/setup') {
				goto('/');
			}
		} catch {
			// If status cannot be read, let the current route surface the API error.
		}
	}

	function updateClock() {
		const now = new Date();
		const time = now.toLocaleTimeString(intl.dateLocale(), { hour: '2-digit', minute: '2-digit', hour12: false });
		const tz = now.toLocaleDateString(intl.dateLocale(), { timeZoneName: 'short' }).split(', ')[1] ?? '';
		clock = `${time} ${tz}`;
	}

	function toggleMenu() { menuOpen = !menuOpen; }

	function closeMenu() { menuOpen = false; }

	async function handleLogout() {
		closeMenu();
		await auth.logout();
		goto('/');
	}

	function handleClickOutside(e: MouseEvent) {
		if (menuRef && !menuRef.contains(e.target as Node)) closeMenu();
	}

	function changeLocale(e: Event) {
		intl.activate((e.target as HTMLSelectElement).value as Locale);
		updateClock();
	}

	function toggleColorScheme() {
		theme.apply(theme.current === 'luma' ? 'default' : 'luma');
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

	<!-- Center: main nav (only when logged in); flex middle so links never sit under avatar/dropdown -->
	{#if !auth.loading && auth.user}
		<div class="nav-mid">
			<div class="nav-center">
				<a href="/" class="nav-item">
					<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/></svg>
					{intl.t('nav.events')}
				</a>
				<a href="/calendars" class="nav-item">
					<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2"/><line x1="16" y1="2" x2="16" y2="6"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="3" y1="10" x2="21" y2="10"/><circle cx="8" cy="14" r="1" fill="currentColor"/><circle cx="12" cy="14" r="1" fill="currentColor"/></svg>
					{intl.t('nav.calendars')}
				</a>
				<a href="/discover" class="nav-item">
					<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"/></svg>
					{intl.t('nav.discover')}
				</a>
				{#if auth.user.is_admin}
					<a href="/admin" class="nav-item">
						<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
							<rect x="3" y="3" width="7" height="9" rx="1" />
							<rect x="14" y="3" width="7" height="5" rx="1" />
							<rect x="14" y="12" width="7" height="9" rx="1" />
							<rect x="3" y="16" width="7" height="5" rx="1" />
						</svg>
						Admin
					</a>
				{/if}
			</div>
		</div>
	{/if}

	<!-- Right -->
	<div class="nav-right">
		{#if clock}
			<span class="clock">{clock}</span>
		{/if}

		{#if !auth.loading}
			{#if auth.user}
				<a href="/events/create" class="nav-item">{intl.t('nav.createEvent')}</a>
				<button class="icon-btn" title={intl.t('nav.search')} aria-label={intl.t('nav.search')}>
					<svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
				</button>
				<button class="icon-btn" title={intl.t('nav.notifications')} aria-label={intl.t('nav.notifications')}>
					<svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
				</button>
				<div class="avatar-menu" bind:this={menuRef}>
					<button class="avatar-btn" onclick={toggleMenu} aria-label={intl.t('nav.userMenu')} aria-expanded={menuOpen}>
						<Avatar src={auth.user.avatar} did={auth.user.did} size={32} />
					</button>
					{#if menuOpen}
						<div class="dropdown" role="menu">
							<div class="dropdown-header">
								<span class="dropdown-name">{auth.user.display_name || intl.t('user.anonymous')}</span>
								<span class="dropdown-did">{auth.user.did.slice(0, 20)}…</span>
							</div>
							<div class="dropdown-divider"></div>
							<a href="/u/{auth.user.did}" class="dropdown-item" onclick={closeMenu} role="menuitem">
								<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
								{intl.t('nav.profile')}
							</a>
							<a href="/settings" class="dropdown-item" onclick={closeMenu} role="menuitem">
								<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>
								{intl.t('nav.settings')}
							</a>
							<div class="dropdown-divider"></div>
							<button class="dropdown-item dropdown-logout" onclick={handleLogout} role="menuitem">
								<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
								{intl.t('nav.signOut')}
							</button>
						</div>
					{/if}
				</div>
			{:else}
				<a href="/discover" class="nav-item">{intl.t('nav.exploreEvents')}</a>
				<a href="/auth/login" class="btn-signin">{intl.t('nav.signIn')}</a>
			{/if}
		{/if}
	</div>
</nav>

<div class="content">
	{@render children()}
</div>

<footer>
	<div class="footer-links">
		<span class="version">v0.1.0-alpha</span>
		<span>•</span>
		<a href="https://github.com/evermeet/evermeet" target="_blank" rel="noopener" class="footer-source">
			<svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0 1 12 6.844a9.59 9.59 0 0 1 2.504.337c1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.02 10.02 0 0 0 22 12.017C22 6.484 17.522 2 12 2z"/></svg>
			GitHub
		</a>
		<span>•</span>
		<a href="https://radicle.network/nodes/rosa.radicle.network/rad%3Az2t188nFbqetZnwj1tfe4ZBD1Ky3Z" target="_blank" rel="noopener" class="footer-source">
			<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 44 44" fill="none" class="footer-radicle-icon" aria-hidden="true">
				<g shape-rendering="crispEdges">
					<rect x="8" y="0" width="4" height="4" fill="#5555FF" />
					<rect x="32" y="0" width="4" height="4" fill="#5555FF" />
					<rect x="12" y="4" width="4" height="4" fill="#5555FF" />
					<rect x="28" y="4" width="4" height="4" fill="#5555FF" />
					<rect x="12" y="8" width="4" height="4" fill="#5555FF" />
					<rect x="16" y="8" width="4" height="4" fill="#3333DD" />
					<rect x="20" y="8" width="4" height="4" fill="#5555FF" />
					<rect x="24" y="8" width="4" height="4" fill="#3333DD" />
					<rect x="28" y="8" width="4" height="4" fill="#5555FF" />
					<rect x="8" y="12" width="4" height="4" fill="#5555FF" />
					<rect x="12" y="12" width="4" height="4" fill="#5555FF" />
					<rect x="16" y="12" width="4" height="4" fill="#5555FF" />
					<rect x="20" y="12" width="4" height="4" fill="#5555FF" />
					<rect x="24" y="12" width="4" height="4" fill="#5555FF" />
					<rect x="28" y="12" width="4" height="4" fill="#5555FF" />
					<rect x="32" y="12" width="4" height="4" fill="#5555FF" />
					<rect x="4" y="16" width="4" height="4" fill="#5555FF" />
					<rect x="8" y="16" width="4" height="4" fill="#5555FF" />
					<rect x="12" y="16" width="4" height="4" fill="#F4F4F4" />
					<rect x="16" y="16" width="4" height="4" fill="#F4F4F4" />
					<rect x="20" y="16" width="4" height="4" fill="#5555FF" />
					<rect x="24" y="16" width="4" height="4" fill="#5555FF" />
					<rect x="28" y="16" width="4" height="4" fill="#F4F4F4" />
					<rect x="32" y="16" width="4" height="4" fill="#F4F4F4" />
					<rect x="36" y="16" width="4" height="4" fill="#5555FF" />
					<rect x="4" y="20" width="4" height="4" fill="#5555FF" />
					<rect x="8" y="20" width="4" height="4" fill="#5555FF" />
					<rect x="12" y="20" width="4" height="4" fill="#F4F4F4" />
					<rect x="16" y="20" width="4" height="4" fill="#FF55FF" />
					<rect x="20" y="20" width="4" height="4" fill="#5555FF" />
					<rect x="24" y="20" width="4" height="4" fill="#5555FF" />
					<rect x="28" y="20" width="4" height="4" fill="#F4F4F4" />
					<rect x="32" y="20" width="4" height="4" fill="#FF55FF" />
					<rect x="36" y="20" width="4" height="4" fill="#5555FF" />
					<rect x="0" y="24" width="4" height="4" fill="#3333DD" />
					<rect x="4" y="24" width="4" height="4" fill="#3333DD" />
					<rect x="8" y="24" width="4" height="4" fill="#3333DD" />
					<rect x="12" y="24" width="4" height="4" fill="#3333DD" />
					<rect x="16" y="24" width="4" height="4" fill="#3333DD" />
					<rect x="20" y="24" width="4" height="4" fill="#3333DD" />
					<rect x="24" y="24" width="4" height="4" fill="#3333DD" />
					<rect x="28" y="24" width="4" height="4" fill="#3333DD" />
					<rect x="32" y="24" width="4" height="4" fill="#3333DD" />
					<rect x="36" y="24" width="4" height="4" fill="#3333DD" />
					<rect x="40" y="24" width="4" height="4" fill="#3333DD" />
					<rect x="8" y="28" width="4" height="4" fill="#3333DD" />
					<rect x="16" y="28" width="4" height="4" fill="#5555FF" />
					<rect x="24" y="28" width="4" height="4" fill="#5555FF" />
					<rect x="32" y="28" width="4" height="4" fill="#3333DD" />
					<rect x="8" y="32" width="4" height="4" fill="#3333DD" />
					<rect x="16" y="32" width="4" height="4" fill="#5555FF" />
					<rect x="24" y="32" width="4" height="4" fill="#5555FF" />
					<rect x="32" y="32" width="4" height="4" fill="#3333DD" />
					<rect x="16" y="36" width="4" height="4" fill="#5555FF" />
					<rect x="24" y="36" width="4" height="4" fill="#5555FF" />
					<rect x="12" y="40" width="4" height="4" fill="#5555FF" />
					<rect x="16" y="40" width="4" height="4" fill="#5555FF" />
					<rect x="24" y="40" width="4" height="4" fill="#5555FF" />
					<rect x="28" y="40" width="4" height="4" fill="#5555FF" />
				</g>
			</svg>
			Radicle
		</a>
		<span>•</span>
		<button
			type="button"
			class="footer-theme-toggle"
			onclick={toggleColorScheme}
			aria-label={theme.current === 'luma' ? intl.t('footer.switchToLight') : intl.t('footer.switchToDark')}
		>
			{#if theme.current === 'luma'}
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
					<circle cx="12" cy="12" r="5"/>
					<path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
				</svg>
			{:else}
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
					<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
				</svg>
			{/if}
		</button>
		<span>•</span>
		<select class="locale-select" aria-label="Language" value={intl.locale} onchange={changeLocale}>
			{#each locales as locale}
				<option value={locale}>{localeNames[locale]}</option>
			{/each}
		</select>
	</div>
	<p class="muted">{intl.t('footer.tagline')}</p>
</footer>

<style>
	nav {
		display: flex;
		align-items: center;
		justify-content: space-between;
		position: relative;
		height: 56px;
		font-family: system-ui, sans-serif;
		background: var(--bg);
		padding: 0 1.5rem;
		gap: 0;
	}

	/* Logo */
	.logo {
		display: flex;
		align-items: center;
		color: var(--text);
		text-decoration: none;
		flex-shrink: 0;
		position: relative;
		z-index: 1;
	}
	.logo:hover { opacity: 0.7; }

	/* Flex middle: grows between logo and right rail so center links are never covered by avatar/dropdown */
	.nav-mid {
		flex: 1;
		min-width: 0;
		display: flex;
		align-items: center;
		/* First link lines up with medium `main` text (same centering as margin: auto main + 1.5rem pad) */
		/* Include former translateX(-1rem) here — transform would shift hit-testing over the logo */
		padding-left: max(0px, calc((100vw - var(--layout-page-medium)) / 2 - 20px - 1rem));
	}
	.nav-center {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		min-width: 0;
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
		white-space: nowrap;
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
		position: relative;
	}

	.clock {
		font-size: 0.8rem;
		color: var(--text-muted);
		font-variant-numeric: tabular-nums;
		margin-right: 0.25rem;
		white-space: nowrap;
	}
	.locale-select {
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		background: var(--bg-input);
		color: var(--text-secondary);
		font-size: 0.8rem;
		padding: 0.25rem 0.45rem;
	}

	@media (max-width: 640px) {
		.clock { display: none; }
		.nav-mid {
			display: none;
		}
		nav { padding: 0 1rem; }
	}


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
		flex-shrink: 0;
		/* 32px avatar + 1px border each side */
		width: 34px;
		height: 34px;
		overflow: visible;
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
	.footer-source {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
	}
	.footer-radicle-icon {
		display: inline-block;
		flex-shrink: 0;
		vertical-align: text-bottom;
		image-rendering: pixelated;
		filter: grayscale(100%) contrast(170%);
		transition: filter 0.15s ease;
	}
	.footer-source:hover .footer-radicle-icon {
		filter: none;
	}
	.footer-theme-toggle {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 0.2rem;
		margin: 0;
		border: none;
		background: transparent;
		color: var(--text-subtle);
		border-radius: var(--radius-md);
		cursor: pointer;
		vertical-align: middle;
		transition: color 0.15s, background 0.15s;
	}
	.footer-theme-toggle:hover {
		color: var(--text);
		background: var(--bg-hover);
	}
	.footer-theme-toggle svg {
		display: block;
	}
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
