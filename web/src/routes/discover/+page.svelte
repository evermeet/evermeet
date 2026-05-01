<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type DiscoverCalendar } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import { intl } from '$lib/i18n.svelte.js';
	import Avatar from '$lib/components/Avatar.svelte';

	let calendars = $state<DiscoverCalendar[]>([]);
	let loading = $state(true);
	let error = $state('');
	let pending = $state<Record<string, boolean>>({});

	onMount(async () => {
		try {
			const res = await api.calendars.discover();
			calendars = res.calendars ?? [];
		} catch (e: any) {
			error = e.message;
		} finally {
			loading = false;
		}
	});

	async function toggleSubscribe(cal: DiscoverCalendar) {
		if (!auth.user) {
			window.location.href = '/auth/login';
			return;
		}
		if (pending[cal.id]) return;
		pending = { ...pending, [cal.id]: true };
		try {
			if (cal.subscribed) {
				await api.calendars.unsubscribe(cal.id);
				cal.subscribed = false;
				cal.subscribers = Math.max(0, cal.subscribers - 1);
			} else {
				await api.calendars.subscribe(cal.id);
				cal.subscribed = true;
				cal.subscribers += 1;
			}
			calendars = [...calendars];
		} catch (e: any) {
			error = e.message;
		} finally {
			pending = { ...pending, [cal.id]: false };
		}
	}
</script>

<main>
	<h1>{intl.t('discover.title')}</h1>
	<h2>{intl.t('discover.featured')}</h2>

	{#if loading}
		<p class="muted">{intl.t('common.loading')}</p>
	{:else if error}
		<p class="error">{error}</p>
	{:else if calendars.length === 0}
		<p class="muted">{intl.t('discover.empty')}</p>
	{:else}
		<div class="calendar-grid">
			{#each calendars as cal}
				<div class="calendar-card">
					<div class="card-top">
						<a href="/calendars/{cal.id}" class="card-avatar">
							<Avatar src={cal.avatar} did={cal.id} size={48} rounded={false} />
						</a>
						<button
							type="button"
							class="btn-subscribe"
							class:subscribed={cal.subscribed}
							onclick={() => toggleSubscribe(cal)}
							disabled={pending[cal.id]}
						>
							{cal.subscribed ? intl.t('discover.subscribed') : intl.t('discover.subscribe')}
						</button>
					</div>
					<a href="/calendars/{cal.id}" class="card-name">{cal.name}</a>
					{#if cal.description}
						<p class="card-desc">{cal.description}</p>
					{/if}
					<p class="card-subs">
						{cal.subscribers === 0
							? intl.t('discover.noSubscribers')
							: intl.t(cal.subscribers === 1 ? 'discover.subscriberCount.one' : 'discover.subscriberCount.other', { count: cal.subscribers })}
					</p>
				</div>
			{/each}
		</div>
	{/if}
</main>

<style>
	main {
		max-width: var(--layout-page-medium);
		margin: 2.5rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 { font-size: 1.9rem; font-weight: 800; margin: 0; color: var(--text); }
	h2 { font-size: 1.1rem; margin: 1.5rem 0 1rem; color: var(--text); }

	.calendar-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 0.85rem;
	}
	.calendar-card {
		padding: 1rem;
		border: 1px solid var(--border-card);
		border-radius: var(--radius-xl);
		background: var(--bg-card);
		display: flex;
		flex-direction: column;
		gap: 0.55rem;
	}
	.card-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
	}
	.card-avatar {
		display: inline-flex;
		align-items: center;
	}
	.card-name {
		color: var(--text);
		text-decoration: none;
		font-weight: 600;
	}
	.card-name:hover { text-decoration: underline; }
	.card-desc {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.88rem;
		line-height: 1.35;
	}
	.card-subs {
		margin: 0;
		font-size: 0.8rem;
		color: var(--text-subtle);
	}

	.btn-subscribe {
		border: 1px solid var(--border-input);
		background: var(--bg-raised);
		color: var(--text);
		border-radius: 999px;
		padding: 0.28rem 0.7rem;
		font-size: 0.8rem;
		font-weight: 600;
		cursor: pointer;
	}
	.btn-subscribe:hover { background: var(--bg-hover); }
	.btn-subscribe.subscribed {
		background: var(--bg-subtle);
		color: var(--text-muted);
	}
	.btn-subscribe:disabled {
		opacity: 0.6;
		cursor: default;
	}

	.muted { color: var(--text-muted); }
	.error { color: var(--text-error); }

	@media (max-width: 760px) {
		.calendar-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
	}
	@media (max-width: 480px) {
		.calendar-grid { grid-template-columns: 1fr; }
	}
</style>
