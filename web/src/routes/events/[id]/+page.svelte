<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api, type CalendarDetail, type Event as EvermeetEvent, type EventAttendee, type MyRSVPStatus } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import { intl } from '$lib/i18n.svelte.js';
	import Avatar from '$lib/components/Avatar.svelte';
	import { marked } from 'marked';

	let event = $state<EvermeetEvent | null>(null);
	let founding = $state<any>(null);
let calendar = $state<CalendarDetail | null>(null);
let currentHash = $state('');
	let hosts = $state<{ did: string; displayName: string; avatar: string }[]>([]);
	let attendees = $state<EventAttendee[]>([]);
	let showAttendees = $state(false);
	let loading = $state(true);
	let error = $state('');

	let rsvps = $state<any[]>([]);
	let rsvpSent = $state(false);
	let myRsvpStatus = $state<MyRSVPStatus | null>(null);
	let rsvpName = $state('');
	let rsvpEmail = $state('');
	let rsvpNote = $state('');
	let showRsvpDetails = $state(false);
	let rsvpSubmitting = $state(false);

	const id = $derived($page.params.id);

	onMount(async () => {
		try {
			const res = await api.events.get(id);
			event = res.state;
			event.id = res.id;
			founding = res.founded;
			currentHash = res.hash;

			if (event.calendar) {
				try {
					calendar = await api.calendars.get(event.calendar);
				} catch (e) {
					calendar = null;
				}
			}

			const hostDIDs = Array.from(
				new Set([
					...(Array.isArray(event.governance?.owners)
						? event.governance.owners
								.map((o: any) => o?.did)
								.filter((v: unknown): v is string => typeof v === 'string' && v.length > 0)
						: []),
					...(event.organizer ? [event.organizer] : [])
				])
			);
			const hostProfiles = await Promise.all(
				hostDIDs.map(async (did) => {
					try {
						const profile = await api.users.get(did);
						return {
							did,
							displayName: profile.display_name || did,
							avatar: profile.avatar || ''
						};
					} catch {
						return { did, displayName: did, avatar: '' };
					}
				})
			);
			hosts = hostProfiles;
			if (event.rsvp?.visible !== false) {
				const attendeeRes = await api.events.attendees(id);
				attendees = attendeeRes.attendees ?? [];
			}

			await waitForAuth();
			if (auth.user) {
				myRsvpStatus = await api.events.myRSVPStatus(id);
			}
			if (isOrganizer()) {
				rsvps = await api.events.listRSVPs(id);
			}
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	function formatDateBlock(iso: string) {
		const d = new Date(iso);
		return {
			weekday: d.toLocaleDateString(intl.dateLocale(), { weekday: 'long' }),
			date: d.toLocaleDateString(intl.dateLocale(), { month: 'long', day: 'numeric' }),
			month: d.toLocaleDateString(intl.dateLocale(), { month: 'short' }).toUpperCase(),
			day: d.getDate(),
			time: d.toLocaleTimeString(intl.dateLocale(), { hour: '2-digit', minute: '2-digit' }),
		};
	}

	function isOrganizer() {
		return auth.user?.did === event?.organizer;
	}

	function canShowRSVP() {
		return isOrganizer() || event?.rsvp?.visible !== false;
	}

	async function waitForAuth() {
		if (!auth.loading) return;
		await new Promise<void>((resolve) => {
			const timer = setInterval(() => {
				if (!auth.loading) {
					clearInterval(timer);
					resolve();
				}
			}, 20);
		});
	}

	function rsvpStatusText(status?: string) {
		if (status === 'confirmed') return intl.t('events.rsvpAccepted');
		if (status === 'rejected') return intl.t('events.rsvpRejected');
		if (status === 'waitlisted') return intl.t('events.rsvpWaitlisted');
		if (status === 'cancelled') return intl.t('events.rsvpCancelled');
		return intl.t('events.rsvpPending');
	}

	async function submitRSVP(e?: SubmitEvent) {
		e?.preventDefault();
		rsvpSubmitting = true;
		error = '';
		try {
			const hadRsvp = myRsvpStatus?.has_rsvp === true;
			const res = await api.events.rsvp(id, {
				name: rsvpName || auth.user?.display_name || '',
				email: rsvpEmail,
				note: rsvpNote
			});
			myRsvpStatus = { has_rsvp: true, status: res.status, received_at: res.received_at };
			if (!hadRsvp && res.status === 'confirmed' && event?.rsvp) {
				event = { ...event, rsvp: { ...event.rsvp, count: event.rsvp.count + 1 } };
			}
			if (isOrganizer()) {
				rsvps = await api.events.listRSVPs(id);
			}
			if (event?.rsvp?.visible !== false) {
				const attendeeRes = await api.events.attendees(id);
				attendees = attendeeRes.attendees ?? [];
				if (event?.rsvp) {
					event = { ...event, rsvp: { ...event.rsvp, count: attendeeRes.count } };
				}
			}
			rsvpSent = true;
		} catch (e: any) {
			error = e.message;
		} finally {
			rsvpSubmitting = false;
		}
	}

	function attendeeSummary() {
		const names = attendees.map((attendee) => attendee.display_name || attendee.did);
		if (names.length === 0) return '';
		if (names.length === 1) return intl.t('events.attendeeSummaryOne', { name: names[0] });
		if (names.length === 2) return intl.t('events.attendeeSummaryTwo', { first: names[0], second: names[1] });
		return intl.t('events.attendeeSummaryMany', {
			names: `${names[0]}, ${names[1]}`,
			count: names.length - 2
		});
	}
</script>

<main>
	{#if loading}
		<p class="muted">{intl.t('common.loading')}</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else if event}
		{@const start = formatDateBlock(event.starts_at)}
		{@const end = event.ends_at ? formatDateBlock(event.ends_at) : null}

		<div class="layout">

			<!-- Left column -->
			<aside class="left-col">
				<!-- Cover image / placeholder -->
				<div class="cover">
					{#if event.cover_url}
						<img src={event.cover_url} alt={event.title} />
					{:else}
						<div class="cover-placeholder">
							<span>{event.title}</span>
						</div>
					{/if}
				</div>

				{#if calendar}
					<div class="side-section presenter-section">
						<p class="side-label">{intl.t('events.presentedBy')}</p>
						<div class="presenter-row">
							<a href="/calendars/{calendar.id}" class="presenter-avatar">
								<Avatar src={calendar.avatar} did={calendar.id} size={36} rounded={false} />
							</a>
							<div class="presenter-meta">
								<a href="/calendars/{calendar.id}" class="presenter-name">{calendar.name} ›</a>
								{#if calendar.description}
									<p class="presenter-description">{calendar.description}</p>
								{/if}
								{#if calendar.website || (calendar.links && calendar.links.length > 0)}
									<div class="presenter-links">
										{#if calendar.website}
											<a href={calendar.website} target="_blank" rel="noopener" class="presenter-icon-link" title={calendar.website}>
												<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
											</a>
										{/if}
										{#each calendar.links ?? [] as link}
											<a href={link.url} target="_blank" rel="noopener" class="presenter-icon-link" title={link.url}>
												{#if link.type === 'twitter'}
													<svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor"><path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-4.714-6.231-5.401 6.231H2.744l7.737-8.835L1.254 2.25H8.08l4.264 5.636zm-1.161 17.52h1.833L7.084 4.126H5.117z"/></svg>
												{:else if link.type === 'instagram'}
													<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="2" width="20" height="20" rx="5" ry="5"/><circle cx="12" cy="12" r="4"/><circle cx="17.5" cy="6.5" r="1" fill="currentColor" stroke="none"/></svg>
												{:else if link.type === 'facebook'}
													<svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor"><path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/></svg>
												{:else if link.type === 'youtube'}
													<svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor"><path d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z"/></svg>
												{:else if link.type === 'tiktok'}
													<svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor"><path d="M19.59 6.69a4.83 4.83 0 0 1-3.77-4.25V2h-3.45v13.67a2.89 2.89 0 0 1-2.88 2.5 2.89 2.89 0 0 1-2.89-2.89 2.89 2.89 0 0 1 2.89-2.89c.28 0 .54.04.79.1V9.01a6.33 6.33 0 0 0-.79-.05 6.34 6.34 0 0 0-6.34 6.34 6.34 6.34 0 0 0 6.34 6.34 6.34 6.34 0 0 0 6.33-6.34V8.69a8.18 8.18 0 0 0 4.78 1.52V6.77a4.85 4.85 0 0 1-1.01-.08z"/></svg>
												{:else if link.type === 'linkedin'}
													<svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor"><path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 0 1-2.063-2.065 2.064 2.064 0 1 1 2.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/></svg>
												{:else if link.type === 'bluesky'}
													<svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor"><path d="M12 10.8c-1.087-2.114-4.046-6.053-6.798-7.995C2.566.944 1.561 1.266.902 1.565.139 1.908 0 3.08 0 3.768c0 .69.378 5.65.624 6.479.815 2.736 3.713 3.66 6.383 3.364.136-.02.275-.039.415-.056-.138.022-.276.04-.415.056-3.912.58-7.387 2.005-2.83 7.078 5.013 5.19 6.87-1.113 7.823-4.308.953 3.195 2.05 9.271 7.733 4.308 4.267-4.308 1.172-6.498-2.74-7.078a8.741 8.741 0 0 1-.415-.056c.14.017.279.036.415.056 2.67.297 5.568-.628 6.383-3.364.246-.828.624-5.79.624-6.478 0-.69-.139-1.861-.902-2.204-.659-.299-1.664-.62-4.3 1.24C16.046 4.748 13.087 8.687 12 10.8z"/></svg>
												{:else if link.type === 'nostr'}
													<svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.477 2 2 6.477 2 12s4.477 10 10 10 10-4.477 10-10S17.523 2 12 2zm0 2c4.418 0 8 3.582 8 8s-3.582 8-8 8-8-3.582-8-8 3.582-8 8-8zm-1 4v4H7l5 6 5-6h-4V8h-2z"/></svg>
												{:else}
													<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
												{/if}
											</a>
										{/each}
									</div>
								{/if}
							</div>
						</div>
					</div>
				{/if}

				<!-- Hosted By -->
				<div class="side-section">
					<p class="side-label">{intl.t('events.hostedBy')}</p>
					{#if hosts.length > 0}
						<div class="host-list">
							{#each hosts as host}
								<div class="host-row">
									<Avatar src={host.avatar} did={host.did} size={36} />
									<a href="/u/{host.did}" class="host-name">{host.displayName}</a>
								</div>
							{/each}
						</div>
					{:else}
						<div class="host-row">
							<Avatar did={event.organizer} size={36} />
							<a href="/u/{event.organizer}" class="host-name">
								{event.organizer.slice(0, 20)}…
							</a>
						</div>
					{/if}
				</div>

				<!-- Going -->
				{#if canShowRSVP() && event.rsvp?.count !== undefined}
					<div class="side-section">
						<p class="side-label">{intl.t('events.going', { count: event.rsvp.count, limit: event.rsvp.limit ? `/${event.rsvp.limit}` : '' })}</p>
						{#if attendees.length > 0}
							<button type="button" class="attendee-summary" onclick={() => (showAttendees = true)}>
								<div class="avatar-stack" aria-hidden="true">
									{#each attendees.slice(0, 6) as attendee}
										<Avatar src={attendee.avatar || ''} did={attendee.did} size={32} />
									{/each}
								</div>
								<span>{attendeeSummary()}</span>
							</button>
						{/if}
					</div>
				{/if}

				<!-- Tags -->
				{#if event.tags?.length}
					<div class="side-section">
						<div class="tags">
							{#each event.tags as tag}
								<span class="tag"># {tag}</span>
							{/each}
						</div>
					</div>
				{/if}

				<div class="side-section">
					<p class="side-label">{intl.t('events.revision')}</p>
					<code class="revision-id">{currentHash ? `${currentHash.slice(0, 16)}…` : 'n/a'}</code>
					<a class="revision-link" href="/events/{id}/history">{intl.t('events.viewEditHistory')}</a>
				</div>
			</aside>

			<!-- Right column -->
			<div class="right-col">
				<div class="right-top">
					<h1>{event.title}</h1>
					{#if isOrganizer()}
						<a href="/events/{id}/edit" class="edit-btn">{intl.t('events.edit')}</a>
					{/if}
				</div>

				<!-- Date -->
				<div class="detail-row">
					<div class="date-badge">
						<span class="date-badge-month">{start.month}</span>
						<span class="date-badge-day">{start.day}</span>
					</div>
					<div class="detail-text">
						<strong>{start.weekday}, {start.date}</strong>
						<span class="muted">
							{start.time}{end ? ` – ${end.time}` : ''}
						</span>
					</div>
				</div>

				<!-- Location -->
				{#if event.location}
					<div class="detail-row">
						<div class="detail-icon">
							<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>
						</div>
						<div class="detail-text">
							<strong>{event.location.name}</strong>
							{#if event.location.address}
								<span class="muted">{event.location.address}</span>
							{/if}
						</div>
					</div>
				{/if}

				<!-- RSVP panel -->
				{#if canShowRSVP() || myRsvpStatus?.has_rsvp}
					<div class="rsvp-panel">
						{#if isOrganizer()}
							<p class="panel-label">{intl.t('events.register')}</p>
							{#if auth.user && (myRsvpStatus?.has_rsvp || rsvpSent)}
								<div class="success-box">
									<strong>{rsvpStatusText(myRsvpStatus?.status)}</strong>
									<p>{intl.t('events.rsvpPrivateStatus')}</p>
								</div>
							{:else if auth.user}
								<button type="button" class="rsvp-btn" onclick={() => submitRSVP()} disabled={rsvpSubmitting}>
									{rsvpSubmitting ? intl.t('auth.sending') : intl.t('events.confirmRsvp')}
								</button>
								<button type="button" class="details-toggle" onclick={() => showRsvpDetails = !showRsvpDetails}>
									{showRsvpDetails ? intl.t('events.hideRsvpDetails') : intl.t('events.addRsvpDetails')}
								</button>
								{#if showRsvpDetails}
									<form class="rsvp-form" onsubmit={submitRSVP}>
										<div class="form-row">
											<input type="text" bind:value={rsvpName} placeholder={intl.t('events.yourName')} />
											<input type="email" bind:value={rsvpEmail} placeholder={intl.t('common.email')} />
										</div>
										<textarea bind:value={rsvpNote} placeholder={intl.t('events.noteToOrganizer')}></textarea>
										<button type="submit" class="rsvp-btn secondary" disabled={rsvpSubmitting}>
											{rsvpSubmitting ? intl.t('auth.sending') : intl.t('events.confirmRsvpWithDetails')}
										</button>
									</form>
								{/if}
							{/if}
							<div class="rsvp-section-divider"></div>
							<p class="panel-label">{intl.t('events.rsvps', { count: rsvps.length })}</p>
							{#if rsvps.length === 0}
								<p class="muted">{intl.t('events.noRsvps')}</p>
							{:else}
								<div class="rsvp-list">
									{#each rsvps as env}
										<div class="rsvp-item">
											<div class="rsvp-head">
												<strong>{env.rsvp.name}</strong>
												<span class="muted">{env.rsvp.email}</span>
											</div>
											{#if env.rsvp.note}
												<p class="rsvp-note">{env.rsvp.note}</p>
											{/if}
											<div class="rsvp-meta">
												<code>{env.sender_did.slice(0, 16)}…</code>
												<span>{new Date(env.received_at).toLocaleDateString(intl.dateLocale())}</span>
											</div>
										</div>
									{/each}
								</div>
							{/if}
						{:else if auth.user && (myRsvpStatus?.has_rsvp || rsvpSent)}
							<div class="success-box">
								<strong>{rsvpStatusText(myRsvpStatus?.status)}</strong>
								<p>{intl.t('events.rsvpPrivateStatus')}</p>
							</div>
						{:else if auth.user}
							<p class="panel-label">{intl.t('events.register')}</p>
							<button type="button" class="rsvp-btn" onclick={() => submitRSVP()} disabled={rsvpSubmitting}>
								{rsvpSubmitting ? intl.t('auth.sending') : intl.t('events.confirmRsvp')}
							</button>
							<button type="button" class="details-toggle" onclick={() => showRsvpDetails = !showRsvpDetails}>
								{showRsvpDetails ? intl.t('events.hideRsvpDetails') : intl.t('events.addRsvpDetails')}
							</button>
							{#if showRsvpDetails}
								<form class="rsvp-form" onsubmit={submitRSVP}>
									<div class="form-row">
										<input type="text" bind:value={rsvpName} placeholder={intl.t('events.yourName')} />
										<input type="email" bind:value={rsvpEmail} placeholder={intl.t('common.email')} />
									</div>
									<textarea bind:value={rsvpNote} placeholder={intl.t('events.noteToOrganizer')}></textarea>
									<button type="submit" class="rsvp-btn secondary" disabled={rsvpSubmitting}>
										{rsvpSubmitting ? intl.t('auth.sending') : intl.t('events.confirmRsvpWithDetails')}
									</button>
								</form>
							{/if}
						{:else}
							<p class="panel-label">{intl.t('events.register')}</p>
							<a href="/auth/login" class="rsvp-btn">{intl.t('events.signInToRsvp')}</a>
						{/if}
					</div>
				{/if}

				<!-- About -->
				{#if event.description}
					<div class="about-section">
						<p class="about-label">{intl.t('events.about')}</p>
						<div class="description">{@html marked(event.description)}</div>
					</div>
				{/if}

				{#if founding?.instance_id}
					<p class="muted instance-hint">{intl.t('events.homeInstance', { instance: founding.instance_id })}</p>
				{/if}
			</div>
		</div>

		{#if showAttendees}
			<div class="modal-backdrop" role="presentation" onclick={(e) => {
				if (e.target === e.currentTarget) showAttendees = false;
			}}>
				<div class="attendee-modal" role="dialog" aria-modal="true" aria-labelledby="attendees-title" tabindex="-1">
					<div class="modal-header">
						<h2 id="attendees-title">{intl.t('events.attendeesTitle', { count: attendees.length })}</h2>
						<button type="button" class="modal-close" aria-label={intl.t('common.close')} onclick={() => (showAttendees = false)}>×</button>
					</div>
					<div class="attendee-list">
						{#each attendees as attendee}
							<a href="/u/{attendee.did}" class="attendee-row">
								<Avatar src={attendee.avatar || ''} did={attendee.did} size={40} />
								<span>{attendee.display_name || attendee.did}</span>
							</a>
						{/each}
					</div>
				</div>
			</div>
		{/if}
	{/if}
</main>

<style>
	main {
		max-width: 900px;
		margin: 2.5rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}

	.layout {
		display: grid;
		grid-template-columns: 300px 1fr;
		gap: 3rem;
		align-items: start;
	}

	/* ── Left column ── */
	.left-col {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.cover {
		width: 100%;
		aspect-ratio: 1 / 1;
		border-radius: var(--radius-xl);
		overflow: hidden;
		background: var(--bg-raised);
	}
	.cover img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.cover-placeholder {
		width: 100%;
		height: 100%;
		display: flex;
		align-items: flex-end;
		padding: 1.25rem;
		background: linear-gradient(135deg, var(--bg-raised) 0%, var(--bg-subtle) 100%);
	}
	.cover-placeholder span {
		font-size: 1.5rem;
		font-weight: 800;
		color: var(--text);
		line-height: 1.2;
		text-transform: uppercase;
		letter-spacing: -0.02em;
	}

	.side-section {
		border-top: 1px solid var(--border-subtle);
		padding-top: 1.25rem;
	}
	.side-label {
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin: 0 0 0.75rem;
	}
	.host-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.host-list {
		display: flex;
		flex-direction: column;
		gap: 0.55rem;
	}
	.host-name {
		font-weight: 600;
		font-size: 0.9rem;
		color: var(--text);
		text-decoration: none;
	}
	.host-name:hover { text-decoration: underline; }
	.attendee-summary {
		border: none;
		background: transparent;
		color: var(--text);
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		align-items: flex-start;
		text-align: left;
		cursor: pointer;
		font: inherit;
	}
	.attendee-summary:hover span {
		text-decoration: underline;
	}
	.avatar-stack {
		display: flex;
		align-items: center;
		padding-left: 0.35rem;
	}
	.avatar-stack :global(.avatar-box) {
		margin-left: -0.35rem;
		border: 2px solid var(--bg);
	}
	.presenter-section { border-top: none; padding-top: 0; }
	.presenter-row {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
	}
	.presenter-avatar {
		flex-shrink: 0;
		display: inline-flex;
	}
	.presenter-meta {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	.presenter-name {
		font-weight: 700;
		font-size: 0.95rem;
		color: var(--text);
		text-decoration: none;
	}
	.presenter-name:hover { text-decoration: underline; }
	.presenter-description {
		margin: 0.25rem 0 0;
		font-size: 0.85rem;
		color: var(--text-muted);
		line-height: 1.35;
	}
	.presenter-links {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		flex-wrap: wrap;
		margin-top: 0.25rem;
	}
	.presenter-icon-link {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 26px;
		height: 26px;
		color: var(--text-muted);
		text-decoration: none;
		transition: color 0.1s;
	}
	.presenter-icon-link:hover { color: var(--text); }

	.tags { display: flex; flex-wrap: wrap; gap: 0.4rem; }
	.tag {
		padding: 0.25rem 0.7rem;
		border: 1px solid var(--border);
		border-radius: var(--radius-full);
		font-size: 0.8rem;
		color: var(--text-secondary);
	}
	.revision-id {
		display: inline-block;
		font-size: 0.78rem;
		color: var(--text-secondary);
		word-break: break-all;
		margin-bottom: 0.5rem;
	}
	.revision-link {
		display: inline-block;
		font-size: 0.85rem;
		color: var(--text-accent);
		text-decoration: none;
	}
	.revision-link:hover { text-decoration: underline; }

	/* ── Right column ── */
	.right-col {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.right-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}
	h1 {
		font-size: 2.25rem;
		font-weight: 800;
		margin: 0;
		line-height: 1.1;
		color: var(--text);
	}
	.edit-btn {
		flex-shrink: 0;
		padding: 0.35rem 0.75rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		text-decoration: none;
		font-size: 0.8rem;
		color: var(--text);
		background: var(--bg);
		margin-top: 0.4rem;
	}
	.edit-btn:hover { background: var(--bg-hover); }

	/* Date / location rows */
	.detail-row {
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	.date-badge {
		width: 44px;
		height: 44px;
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		background: var(--bg-card);
	}
	.date-badge-month {
		font-size: 0.6rem;
		font-weight: 700;
		text-transform: uppercase;
		color: var(--text-muted);
		line-height: 1;
	}
	.date-badge-day {
		font-size: 1.1rem;
		font-weight: 800;
		color: var(--text);
		line-height: 1;
	}
	.detail-icon {
		width: 44px;
		height: 44px;
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		color: var(--text-secondary);
		background: var(--bg-card);
	}
	.detail-text {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
		font-size: 0.95rem;
	}
	.detail-text strong { color: var(--text); font-weight: 600; }

	/* RSVP panel */
	.rsvp-panel {
		background: var(--bg-subtle);
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-xl);
		padding: 1.25rem 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.panel-label {
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin: 0;
	}
	.rsvp-btn {
		width: 100%;
		padding: 0.75rem;
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		text-decoration: none;
		text-align: center;
		display: block;
	}
	.rsvp-btn:disabled { opacity: 0.5; }
	.rsvp-btn.secondary {
		background: var(--bg-btn-secondary);
		color: var(--text-btn-secondary);
		border: 1px solid var(--border-input);
	}
	.details-toggle {
		border: none;
		background: transparent;
		color: var(--text-link);
		font: inherit;
		font-size: 0.9rem;
		font-weight: 600;
		cursor: pointer;
		padding: 0;
		text-align: left;
	}

	.rsvp-form { display: flex; flex-direction: column; gap: 0.6rem; }
	.form-row { display: flex; gap: 0.6rem; }
	.form-row input { flex: 1; }
	input, textarea {
		padding: 0.6rem 0.75rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		font-size: 0.9rem;
		background: var(--bg-input);
		color: var(--text);
		font-family: inherit;
		width: 100%;
	}
	input:focus, textarea:focus { outline: none; border-color: var(--border-input-focus); }
	textarea { min-height: 70px; resize: vertical; }

	.rsvp-list { display: flex; flex-direction: column; gap: 0.75rem; }
	.rsvp-section-divider {
		border-top: 1px solid var(--border-subtle);
		margin: 0.25rem 0;
	}
	.rsvp-item {
		padding: 0.9rem 1rem;
		border: 1px solid var(--border-card);
		border-radius: var(--radius-md);
		background: var(--bg-card);
	}
	.rsvp-head { display: flex; justify-content: space-between; margin-bottom: 0.4rem; color: var(--text); }
	.rsvp-note { margin: 0.4rem 0; font-size: 0.9rem; color: var(--text-secondary); }
	.rsvp-meta { display: flex; justify-content: space-between; font-size: 0.78rem; color: var(--text-muted); margin-top: 0.4rem; }

	.success-box {
		background: var(--bg-success);
		color: var(--text-success);
		padding: 1rem 1.25rem;
		border-radius: var(--radius-md);
	}
	.success-box p { margin: 0.25rem 0 0; font-size: 0.9rem; }

	.modal-backdrop {
		position: fixed;
		inset: 0;
		z-index: 50;
		background: rgba(0, 0, 0, 0.55);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1.5rem;
	}
	.attendee-modal {
		width: min(440px, 100%);
		max-height: min(640px, 85vh);
		overflow: hidden;
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-card);
		background: var(--bg-card);
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.35);
		display: flex;
		flex-direction: column;
	}
	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 1rem 1.25rem;
		border-bottom: 1px solid var(--border-subtle);
	}
	.modal-header h2 {
		margin: 0;
		font-size: 1.1rem;
		color: var(--text);
	}
	.modal-close {
		border: none;
		background: transparent;
		color: var(--text-muted);
		font-size: 1.6rem;
		line-height: 1;
		cursor: pointer;
	}
	.attendee-list {
		overflow: auto;
		padding: 0.5rem;
	}
	.attendee-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.7rem 0.75rem;
		border-radius: var(--radius-md);
		color: var(--text);
		text-decoration: none;
		font-weight: 600;
	}
	.attendee-row:hover {
		background: var(--bg-hover);
	}

	/* About section */
	.about-section {
		border-top: 1px solid var(--border-subtle);
		padding-top: 1.5rem;
	}
	.about-label {
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin: 0 0 0.75rem;
	}
	.description {
		line-height: 1.7;
		color: var(--text);
		font-size: 0.95rem;
		margin: 0;
	}
	.description :global(h1),
	.description :global(h2),
	.description :global(h3) {
		margin: 1.25rem 0 0.4rem;
		font-weight: 700;
		color: var(--text);
	}
	.description :global(p) { margin: 0.6rem 0; }
	.description :global(ul),
	.description :global(ol) { padding-left: 1.5rem; margin: 0.6rem 0; }
	.description :global(li) { margin: 0.25rem 0; }
	.description :global(a) { color: var(--text-accent); }
	.description :global(code) {
		background: var(--bg-code);
		color: var(--text-code);
		padding: 0.1em 0.35em;
		border-radius: var(--radius-sm);
		font-size: 0.88em;
	}
	.description :global(pre) {
		background: var(--bg-code);
		padding: 0.75rem 1rem;
		border-radius: var(--radius-md);
		overflow-x: auto;
		margin: 0.75rem 0;
	}
	.description :global(blockquote) {
		border-left: 3px solid var(--border);
		margin: 0.75rem 0;
		padding: 0.25rem 0 0.25rem 1rem;
		color: var(--text-secondary);
	}
	.description :global(hr) {
		border: none;
		border-top: 1px solid var(--border-subtle);
		margin: 1rem 0;
	}
	.description :global(img) {
		max-width: 100%;
		height: auto;
		border-radius: var(--radius-md);
		display: block;
	}

	.instance-hint { font-size: 0.8rem; margin: 0; }

	.muted { color: var(--text-muted); }
	.error { color: var(--text-error); }

	@media (max-width: 640px) {
		.layout { grid-template-columns: 1fr; gap: 2rem; }
		h1 { font-size: 1.75rem; }
	}
</style>
