<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api, type CalendarDetail, type CalendarEvent } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import { intl } from '$lib/i18n.svelte.js';
	import Avatar from '$lib/components/Avatar.svelte';
	import EventCard from '$lib/components/EventCard.svelte';

	const id = $derived($page.params.id);

	let cal = $state<CalendarDetail | null>(null);
	let loading = $state(true);
	let error = $state('');
	let subscribing = $state(false);
	let hostProfiles = $state<Record<string, { displayName: string; avatar: string }>>({});
	let filter = $state<'upcoming' | 'past'>('upcoming');
	let calendarMonth = $state(new Date());
	let searchOpen = $state(false);

	onMount(async () => {
		try {
			cal = await api.calendars.get(id);
			const uniqueHosts = Array.from(
				new Set((cal.events ?? []).flatMap((ev) => ev.hosts ?? []))
			);
			if (uniqueHosts.length > 0) {
				const entries = await Promise.all(
					uniqueHosts.map(async (did) => {
						try {
							const user = await api.users.get(did);
							return [did, { displayName: user.display_name || did, avatar: user.avatar || '' }] as const;
						} catch {
							return [did, { displayName: did, avatar: '' }] as const;
						}
					})
				);
				hostProfiles = Object.fromEntries(entries);
			}
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	function isOwner() {
		if (!cal || !auth.user) return false;
		return cal.governance.owners.some(o => o.did === auth.user!.did);
	}

	async function toggleSubscribe() {
		if (!cal) return;
		subscribing = true;
		try {
			if (cal.subscribed) {
				await api.calendars.unsubscribe(id);
				cal = { ...cal, subscribed: false, subscribers: cal.subscribers - 1 };
			} else {
				await api.calendars.subscribe(id);
				cal = { ...cal, subscribed: true, subscribers: cal.subscribers + 1 };
			}
		} catch (e: any) {
			error = e.message;
		} finally {
			subscribing = false;
		}
	}


	function groupByDate(events: CalendarEvent[]) {
		const groups = new Map<string, CalendarEvent[]>();
		for (const ev of events) {
			const d = new Date(ev.starts_at);
			const weekday = d.toLocaleDateString(intl.dateLocale(), { weekday: 'long' });
			const monthDay = d.toLocaleDateString(intl.dateLocale(), { month: 'long', day: 'numeric' });
			const key = `${weekday}|${monthDay}`;
			if (!groups.has(key)) groups.set(key, []);
			groups.get(key)!.push(ev);
		}
		return groups;
	}

	function prevMonth() {
		calendarMonth = new Date(calendarMonth.getFullYear(), calendarMonth.getMonth() - 1, 1);
	}
	function nextMonth() {
		calendarMonth = new Date(calendarMonth.getFullYear(), calendarMonth.getMonth() + 1, 1);
	}
	function goToday() {
		calendarMonth = new Date();
	}

	function getCalendarDays(month: Date) {
		const year = month.getFullYear();
		const m = month.getMonth();
		const firstDay = new Date(year, m, 1).getDay();
		const daysInMonth = new Date(year, m + 1, 0).getDate();
		const days: (number | null)[] = [];
		for (let i = 0; i < firstDay; i++) days.push(null);
		for (let i = 1; i <= daysInMonth; i++) days.push(i);
		return days;
	}

	function eventDaysInMonth(events: CalendarEvent[], month: Date) {
		const set = new Set<number>();
		for (const ev of events) {
			const d = new Date(ev.starts_at);
			if (d.getFullYear() === month.getFullYear() && d.getMonth() === month.getMonth()) {
				set.add(d.getDate());
			}
		}
		return set;
	}

	const now = new Date();

	const filteredEvents = $derived(
		cal
			? cal.events.filter((ev) => {
					const t = new Date(ev.starts_at);
					return filter === 'upcoming' ? t >= now : t < now;
			  })
			: []
	);

	const grouped = $derived(groupByDate(filteredEvents));
	const calDays = $derived(getCalendarDays(calendarMonth));
	const eventDays = $derived(cal ? eventDaysInMonth(cal.events, calendarMonth) : new Set<number>());

	// Parse grouped key to extract short day label
	function parseDateKey(key: string): { monthDay: string; weekday: string } {
		const [weekday = '', monthDay = ''] = key.split('|');
		return { monthDay, weekday };
	}

	function subscriberLabel(count: number) {
		if (count === 0) return intl.t('calendars.noSubscribers');
		return intl.t(count === 1 ? 'calendars.subscriberCount.one' : 'calendars.subscriberCount.other', { count });
	}

	function weekdayLabels() {
		return intl.t('calendar.weekdays').split(',');
	}
</script>

<svelte:head>
	{#if cal}<title>{cal.name} — Evermeet</title>{/if}
</svelte:head>

{#if loading}
	<p class="status-msg muted">{intl.t('common.loading')}</p>
{:else if error}
	<p class="status-msg error">{error}</p>
{:else if cal}
	{#if cal.backdrop_url}
		<div
			class="backdrop"
			style={`background-image: url(${JSON.stringify(cal.backdrop_url)})`}
		></div>
	{/if}

	<div class="header-wrap" class:no-backdrop={!cal.backdrop_url}>
		<!-- Avatar overlapping backdrop -->
		<div class="avatar-row" class:with-backdrop={!!cal.backdrop_url}>
			<div class="cal-avatar" class:with-backdrop={!!cal.backdrop_url}>
				<div class="cal-avatar-inner">
					<Avatar src={cal.avatar} did={cal.id} size={80} rounded={false} />
				</div>
			</div>
			{#if auth.user && !isOwner()}
				<button
					type="button"
					class="btn-subscribe"
					class:subscribed={cal.subscribed}
					onclick={toggleSubscribe}
					disabled={subscribing}
				>
					{cal.subscribed ? intl.t('calendars.subscribedState') : intl.t('calendars.subscribe')}
				</button>
			{:else if isOwner()}
				<a href="/calendars/{id}/edit" class="btn-edit">{intl.t('calendars.edit')}</a>
			{/if}
		</div>

		<!-- Calendar info -->
		<div class="cal-info">
			<h1>{cal.name}</h1>
			{#if cal.description}
				<p class="cal-desc">{cal.description}</p>
			{/if}
			<div class="cal-meta">
				<span class="muted">{subscriberLabel(cal.subscribers)}</span>
			</div>
			{#if cal.website || (cal.links && cal.links.length > 0)}
				<div class="cal-links">
					{#if cal.website}
						<a href={cal.website} target="_blank" rel="noopener" class="cal-link" title={cal.website}>
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
						</a>
					{/if}
					{#each cal.links ?? [] as link}
						<a href={link.url} target="_blank" rel="noopener" class="cal-link" title={link.url}>
							{#if link.type === 'twitter'}
								<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-4.714-6.231-5.401 6.231H2.744l7.737-8.835L1.254 2.25H8.08l4.264 5.636zm-1.161 17.52h1.833L7.084 4.126H5.117z"/></svg>
							{:else if link.type === 'instagram'}
								<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="2" width="20" height="20" rx="5" ry="5"/><circle cx="12" cy="12" r="4"/><circle cx="17.5" cy="6.5" r="1" fill="currentColor" stroke="none"/></svg>
							{:else if link.type === 'youtube'}
								<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z"/></svg>
							{:else if link.type === 'tiktok'}
								<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M19.59 6.69a4.83 4.83 0 0 1-3.77-4.25V2h-3.45v13.67a2.89 2.89 0 0 1-2.88 2.5 2.89 2.89 0 0 1-2.89-2.89 2.89 2.89 0 0 1 2.89-2.89c.28 0 .54.04.79.1V9.01a6.33 6.33 0 0 0-.79-.05 6.34 6.34 0 0 0-6.34 6.34 6.34 6.34 0 0 0 6.34 6.34 6.34 6.34 0 0 0 6.33-6.34V8.69a8.18 8.18 0 0 0 4.78 1.52V6.77a4.85 4.85 0 0 1-1.01-.08z"/></svg>
							{:else if link.type === 'linkedin'}
								<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 0 1-2.063-2.065 2.064 2.064 0 1 1 2.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/></svg>
							{:else if link.type === 'bluesky'}
								<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M12 10.8c-1.087-2.114-4.046-6.053-6.798-7.995C2.566.944 1.561 1.266.902 1.565.139 1.908 0 3.08 0 3.768c0 .69.378 5.65.624 6.479.815 2.736 3.713 3.66 6.383 3.364.136-.02.275-.039.415-.056-.138.022-.276.04-.415.056-3.912.58-7.387 2.005-2.83 7.078 5.013 5.19 6.87-1.113 7.823-4.308.953 3.195 2.05 9.271 7.733 4.308 4.267-4.308 1.172-6.498-2.74-7.078a8.741 8.741 0 0 1-.415-.056c.14.017.279.036.415.056 2.67.297 5.568-.628 6.383-3.364.246-.828.624-5.79.624-6.478 0-.69-.139-1.861-.902-2.204-.659-.299-1.664-.62-4.3 1.24C16.046 4.748 13.087 8.687 12 10.8z"/></svg>
							{:else if link.type === 'facebook'}
								<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/></svg>
							{:else if link.type === 'nostr'}
								<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.477 2 2 6.477 2 12s4.477 10 10 10 10-4.477 10-10S17.523 2 12 2zm0 2c4.418 0 8 3.582 8 8s-3.582 8-8 8-8-3.582-8-8 3.582-8 8-8zm-1 4v4H7l5 6 5-6h-4V8h-2z"/></svg>
							{:else}
								<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
							{/if}
						</a>
					{/each}
				</div>
			{/if}
		</div>

		<div class="divider"></div>
	</div>

	<div class="page-wrap">
		<!-- Left: events -->
		<div class="left-col">
			<!-- Events header -->
			<div class="events-header">
				<h2>{intl.t('events.title')}</h2>
				<div class="events-toolbar">
					<button class="tool-btn" class:active={false} title={intl.t('events.cardView')}>
						<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/></svg>
					</button>
					<button class="tool-btn" title={intl.t('events.listView')}>
						<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
					</button>
					<button class="tool-btn" title={intl.t('nav.search')} onclick={() => searchOpen = !searchOpen}>
						<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
					</button>
				</div>
			</div>

			{#if cal.events.length === 0}
				<p class="muted">{intl.t('events.noneYet')}</p>
			{:else if filteredEvents.length === 0}
				<p class="muted">{intl.t('events.empty', { filter: intl.t(filter === 'upcoming' ? 'events.upcoming' : 'events.past').toLowerCase() })}</p>
			{:else}
				<div class="event-groups">
					{#each grouped as [key, evs]}
						{@const { monthDay, weekday } = parseDateKey(key)}
						<div class="date-group">
							<div class="date-label">
								<span class="dot"></span>
								<span class="date-month">{monthDay}</span>
								<span class="date-weekday">{weekday}</span>
							</div>
							<div class="event-cards">
								{#each evs as ev}
									<EventCard
										id={ev.id}
										title={ev.title}
										starts_at={ev.starts_at}
										location={ev.location}
										cover_url={ev.cover_url}
										hosts={ev.hosts}
										{hostProfiles}
									/>
								{/each}
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Right sidebar -->
		<aside class="right-col">
			<!-- Submit Event + RSS -->
			<div class="sidebar-actions">
				{#if auth.user}
					<a href="/events/create?calendar={id}" class="btn-submit">
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
						{intl.t('events.submit')}
					</a>
				{/if}
				<a
					href="/api/calendars/{id}/feed.ics"
					class="btn-rss"
					title={intl.t('events.subscribeFeed')}
					target="_blank"
					rel="noopener"
				>
					<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 11a9 9 0 0 1 9 9"/><path d="M4 4a16 16 0 0 1 16 16"/><circle cx="5" cy="19" r="1" fill="currentColor"/></svg>
				</a>
			</div>

			<!-- Mini calendar -->
			<div class="mini-cal">
				<div class="mini-cal-header">
					<span class="mini-cal-month">{calendarMonth.toLocaleDateString(intl.dateLocale(), { month: 'long' })}</span>
					<div class="mini-cal-nav">
						<button class="cal-nav-btn" onclick={prevMonth} title={intl.t('calendar.previousMonth')}>
							<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="15 18 9 12 15 6"/></svg>
						</button>
						<button class="cal-nav-btn cal-today-btn" onclick={goToday} title={intl.t('calendar.today')}>
							<svg width="6" height="6" viewBox="0 0 6 6"><circle cx="3" cy="3" r="3" fill="currentColor"/></svg>
						</button>
						<button class="cal-nav-btn" onclick={nextMonth} title={intl.t('calendar.nextMonth')}>
							<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"/></svg>
						</button>
					</div>
				</div>
				<div class="mini-cal-grid">
					{#each weekdayLabels() as d}
						<span class="cal-dow">{d}</span>
					{/each}
					{#each calDays as day}
						{#if day === null}
							<span></span>
						{:else}
							{@const isToday =
								day === now.getDate() &&
								calendarMonth.getMonth() === now.getMonth() &&
								calendarMonth.getFullYear() === now.getFullYear()}
							{@const hasEvent = eventDays.has(day)}
							<button
								class="cal-day"
								class:today={isToday}
								class:has-event={hasEvent}
							>{day}</button>
						{/if}
					{/each}
				</div>

				<!-- Upcoming / Past toggle -->
				<div class="filter-toggle">
					<button class="filter-btn" class:active={filter === 'upcoming'} onclick={() => filter = 'upcoming'}>
						{intl.t('events.upcoming')}
					</button>
					<button class="filter-btn" class:active={filter === 'past'} onclick={() => filter = 'past'}>
						{intl.t('events.past')}
					</button>
				</div>
			</div>

			<!-- World map -->
			<div class="world-map">
				<svg viewBox="0 0 1000 500" xmlns="http://www.w3.org/2000/svg" class="map-svg">
					<!-- Simplified world map paths -->
					<!-- North America -->
					<path d="M120 80 L200 60 L280 70 L310 90 L320 120 L300 150 L280 180 L260 200 L240 220 L220 240 L200 260 L180 270 L160 260 L140 240 L120 220 L100 200 L90 170 L95 140 L110 110 Z" class="landmass"/>
					<!-- Greenland -->
					<path d="M260 30 L300 20 L340 30 L350 50 L330 70 L300 80 L270 70 L255 50 Z" class="landmass"/>
					<!-- South America -->
					<path d="M210 280 L250 270 L280 290 L290 320 L280 360 L270 400 L250 430 L230 450 L210 440 L195 410 L185 370 L185 330 L195 300 Z" class="landmass"/>
					<!-- Europe -->
					<path d="M440 60 L480 50 L510 60 L520 80 L510 100 L490 110 L470 120 L450 115 L435 100 L430 80 Z" class="landmass"/>
					<!-- UK/Ireland -->
					<path d="M415 65 L425 58 L430 68 L425 78 L415 75 Z" class="landmass"/>
					<!-- Scandinavia -->
					<path d="M460 30 L490 25 L500 40 L490 55 L475 60 L460 55 L455 40 Z" class="landmass"/>
					<!-- Africa -->
					<path d="M440 130 L490 120 L520 130 L540 160 L545 200 L540 240 L530 280 L510 310 L490 330 L470 330 L450 310 L435 280 L425 240 L420 200 L422 160 Z" class="landmass"/>
					<!-- Russia/Asia -->
					<path d="M530 40 L600 30 L680 35 L750 45 L800 55 L830 70 L820 90 L790 100 L760 105 L720 100 L680 95 L640 100 L610 110 L580 105 L555 95 L535 80 L525 60 Z" class="landmass"/>
					<!-- Middle East -->
					<path d="M530 110 L570 105 L600 115 L610 135 L600 155 L570 160 L540 155 L525 135 Z" class="landmass"/>
					<!-- South Asia -->
					<path d="M600 120 L650 115 L680 125 L690 150 L680 175 L660 185 L635 180 L615 165 L605 145 Z" class="landmass"/>
					<!-- Southeast Asia -->
					<path d="M690 130 L740 120 L780 130 L800 150 L790 170 L760 175 L730 165 L705 150 Z" class="landmass"/>
					<!-- East Asia -->
					<path d="M730 70 L790 60 L840 65 L860 85 L850 110 L820 120 L785 115 L750 105 L725 90 Z" class="landmass"/>
					<!-- Japan -->
					<path d="M840 75 L855 70 L865 80 L855 95 L840 95 Z" class="landmass"/>
					<!-- Australia -->
					<path d="M720 290 L790 275 L840 280 L870 300 L875 335 L860 365 L830 380 L790 375 L755 355 L730 325 L715 300 Z" class="landmass"/>
					<!-- New Zealand -->
					<path d="M880 350 L895 340 L905 355 L895 370 L880 365 Z" class="landmass"/>
					<!-- Indonesia -->
					<path d="M730 225 L760 220 L785 225 L795 240 L780 252 L750 252 L730 242 Z" class="landmass"/>
					<path d="M800 230 L825 225 L840 235 L835 250 L810 252 L798 242 Z" class="landmass"/>
					<!-- Madagascar -->
					<path d="M550 280 L560 270 L570 280 L568 310 L555 315 L545 300 Z" class="landmass"/>
					<!-- Iceland -->
					<path d="M380 45 L400 38 L415 45 L412 58 L395 62 L378 56 Z" class="landmass"/>
					<!-- Central America / Caribbean -->
					<path d="M175 220 L200 215 L215 225 L210 240 L190 242 L172 232 Z" class="landmass"/>
				</svg>

				<!-- Event location dots (placeholder markers) -->
				{#if cal.events.some(ev => ev.location)}
					<div class="map-dots">
						<!-- Static dots in Europe area matching screenshot style -->
						<div class="map-dot" style="left: 48%; top: 42%"></div>
						<div class="map-dot" style="left: 50%; top: 45%"></div>
						<div class="map-dot" style="left: 72%; top: 52%"></div>
					</div>
				{/if}
			</div>
		</aside>
	</div>
{/if}

<style>
	.status-msg { max-width: var(--layout-page-wide); margin: 2rem auto; padding: 0 1.5rem; font-family: system-ui, sans-serif; }

	/* Backdrop */
	.backdrop {
		width: 100%;
		height: 240px;
		background-color: var(--bg-raised);
		background-size: cover;
		background-position: center;
		position: relative;
	}

	/* Full-width header section */
	.header-wrap {
		max-width: var(--layout-page-wide);
		margin: 0 auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	.header-wrap.no-backdrop {
		padding-top: 2.5rem;
	}

	/* Two-column layout below the divider */
	.page-wrap {
		max-width: var(--layout-page-wide);
		margin: 0 auto;
		padding: 0 1.5rem 3rem;
		font-family: system-ui, sans-serif;
		display: grid;
		grid-template-columns: 1fr 280px;
		gap: 2rem;
		align-items: start;
	}

	/* LEFT COLUMN */
	.left-col {
		min-width: 0;
	}

	.avatar-row {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		margin-top: -43px;
		margin-bottom: 1rem;
	}
	.avatar-row.with-backdrop {
		padding-top: 1.1rem;
	}
	.header-wrap.no-backdrop .avatar-row {
		margin-top: 0;
	}
	.cal-avatar {
		display: inline-block;
		padding: 2px;
		border-radius: var(--radius-xl);
		background: var(--bg);
		line-height: 0;
		position: relative;
		z-index: 1;
	}
	.cal-avatar.with-backdrop {
		background: color-mix(in srgb, var(--bg-card) 25%, transparent);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
	}
	.cal-avatar-inner {
		border-radius: calc(var(--radius-xl) - 2px);
		overflow: hidden;
		line-height: 0;
	}
	/* Profile only: frame is .cal-avatar; strip default avatar border so the 2px gutter shows */
	.cal-avatar-inner :global(.avatar-box) {
		border: none;
	}

	.btn-subscribe {
		padding: 0.5rem 1.25rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		background: var(--bg-raised);
		color: var(--text);
		transition: background 0.1s;
		margin-bottom: 4px;
	}
	.btn-subscribe:hover { background: var(--bg-hover); }
	.btn-subscribe.subscribed {
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border-color: transparent;
	}
	.btn-subscribe:disabled { opacity: 0.5; cursor: default; }

	.btn-edit {
		padding: 0.5rem 1.25rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text);
		text-decoration: none;
		background: var(--bg-raised);
		margin-bottom: 4px;
	}
	.btn-edit:hover { background: var(--bg-hover); }

	.cal-info { margin-bottom: 1.5rem; }
	.cal-info > h1 {
		font-size: 2rem;
		font-weight: 800;
		margin: 0 0 0.5rem;
		color: var(--text);
		line-height: 1.1;
	}
	.cal-desc { font-size: 0.95rem; color: var(--text-secondary); margin: 0 0 0.75rem; line-height: 1.6; }
	.cal-meta { display: flex; align-items: center; gap: 1rem; flex-wrap: wrap; margin-bottom: 0.75rem; }
	.cal-links {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-wrap: wrap;
		margin-bottom: 0.5rem;
	}
	.cal-link {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		color: var(--text-muted);
		text-decoration: none;
		transition: color 0.1s;
	}
	.cal-link:hover { color: var(--text); }

	.divider { height: 1px; background: var(--border-subtle); margin: 1.5rem 0; }

	/* Events header */
	.events-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1.25rem;
	}
	h2 { font-size: 1.2rem; font-weight: 700; margin: 0; color: var(--text); }
	.events-toolbar { display: flex; align-items: center; gap: 0.25rem; }
	.tool-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		background: var(--bg-raised);
		color: var(--text-muted);
		cursor: pointer;
		transition: background 0.1s, color 0.1s;
	}
	.tool-btn:hover { background: var(--bg-hover); color: var(--text); }
	.tool-btn.active { background: var(--bg-btn-primary); color: var(--text-btn-primary); border-color: transparent; }

	/* Event groups */
	.event-groups { display: flex; flex-direction: column; gap: 1.5rem; }

	.date-label {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		margin-bottom: 0.75rem;
		font-size: 0.875rem;
		font-weight: 600;
	}
	.dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--border-input);
		flex-shrink: 0;
	}
	.date-month { color: var(--text); }
	.date-weekday { color: var(--text-muted); font-weight: 400; }

	.event-cards {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		margin-left: 1.15rem;
		border-left: 1px solid var(--border-subtle);
		padding-left: 1.25rem;
	}

	/* RIGHT SIDEBAR */
	.right-col {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	/* Submit + RSS row */
	.sidebar-actions {
		display: flex;
		gap: 0.5rem;
	}
	.btn-submit {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.4rem;
		padding: 0.55rem 1rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text);
		text-decoration: none;
		background: var(--bg-raised);
		transition: background 0.1s;
	}
	.btn-submit:hover { background: var(--bg-hover); }
	.btn-rss {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		color: var(--text-muted);
		background: var(--bg-raised);
		text-decoration: none;
		transition: background 0.1s, color 0.1s;
		flex-shrink: 0;
	}
	.btn-rss:hover { background: var(--bg-hover); color: var(--text); }

	/* Mini calendar */
	.mini-cal {
		border: 1px solid var(--border-card);
		border-radius: var(--radius-xl);
		background: var(--bg-card);
		padding: 1rem;
	}
	.mini-cal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.75rem;
	}
	.mini-cal-month { font-size: 0.95rem; font-weight: 700; color: var(--text); }
	.mini-cal-nav { display: flex; align-items: center; gap: 0.2rem; }
	.cal-nav-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 26px;
		height: 26px;
		border: none;
		background: transparent;
		color: var(--text-muted);
		cursor: pointer;
		border-radius: var(--radius-sm);
		transition: background 0.1s, color 0.1s;
	}
	.cal-nav-btn:hover { background: var(--bg-hover); color: var(--text); }
	.cal-today-btn { color: var(--text-muted); }

	.mini-cal-grid {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		gap: 2px;
		text-align: center;
	}
	.cal-dow {
		font-size: 0.72rem;
		font-weight: 600;
		color: var(--text-muted);
		padding: 0.25rem 0;
	}
	.cal-day {
		font-size: 0.8rem;
		color: var(--text-secondary);
		padding: 0.3rem 0;
		border: none;
		background: transparent;
		cursor: pointer;
		border-radius: var(--radius-sm);
		position: relative;
		transition: background 0.1s;
	}
	.cal-day:hover { background: var(--bg-hover); color: var(--text); }
	.cal-day.today {
		font-weight: 700;
		color: var(--text);
	}
	.cal-day.has-event::after {
		content: '';
		position: absolute;
		bottom: 2px;
		left: 50%;
		transform: translateX(-50%);
		width: 4px;
		height: 4px;
		border-radius: 50%;
		background: var(--text-muted);
	}

	/* Upcoming / Past toggle */
	.filter-toggle {
		display: flex;
		margin-top: 0.75rem;
		border: 1px solid var(--border-input);
		border-radius: var(--radius-md);
		overflow: hidden;
	}
	.filter-btn {
		flex: 1;
		padding: 0.45rem 0;
		border: none;
		background: transparent;
		font-size: 0.85rem;
		font-weight: 500;
		color: var(--text-muted);
		cursor: pointer;
		transition: background 0.1s, color 0.1s;
	}
	.filter-btn.active {
		background: var(--bg-raised);
		color: var(--text);
		font-weight: 600;
	}
	.filter-btn:hover:not(.active) { background: var(--bg-hover); color: var(--text); }

	/* World map */
	.world-map {
		border: 1px solid var(--border-card);
		border-radius: var(--radius-xl);
		background: var(--bg-card);
		overflow: hidden;
		position: relative;
		aspect-ratio: 2 / 1;
	}
	.map-svg {
		width: 100%;
		height: 100%;
		display: block;
	}
	.landmass {
		fill: var(--border-subtle);
		opacity: 0.6;
	}
	.map-dots {
		position: absolute;
		inset: 0;
		pointer-events: none;
	}
	.map-dot {
		position: absolute;
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--text-muted);
		transform: translate(-50%, -50%);
		opacity: 0.8;
	}

	.muted { color: var(--text-muted); font-size: 0.9rem; }
	.error { color: var(--text-error); }

	@media (max-width: 700px) {
		.page-wrap {
			grid-template-columns: 1fr;
		}
		.right-col {
			order: -1;
		}
	}
</style>
