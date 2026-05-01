<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Avatar from '$lib/components/Avatar.svelte';
	import { api } from '$lib/api.js';
	import type { ResolvedInstanceInfo } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import { intl } from '$lib/i18n.svelte.js';
	import { bufferToBase64, recursiveBase64ToBuffer } from '$lib/webauthn.js';

	declare global {
		interface Window {
			ethereum?: {
				request(args: { method: string; params?: unknown[] }): Promise<unknown>;
			};
		}
	}

	type Method = 'email' | 'ethereum';

	let method = $state<Method>('email');
	let email = $state('');
	let eventId = $state('');
	let next = $state('');
	let autoEmail = $state(false);
	let autoEmailRequested = false;
	let address = $state('');
	let chainId = $state('');
	let identity = $state('');
	let discoveredHome = $state('');
	let instance = $state<ResolvedInstanceInfo | null>(null);
	let delegateURL = $state('');
	let homeIsCurrentInstance = $state(false);
	let passkeyAvailable = $state(false);
	let noHomeFound = $state(false);
	let sent = $state(false);
	let magicLinkPollToken = '';
	let magicLinkPollTimer: ReturnType<typeof setInterval> | null = null;
	let loading = $state(true);
	let submitting = $state(false);
	let error = $state('');

	onMount(() => {
		const params = new URLSearchParams(window.location.search);
		method = params.get('method') === 'ethereum' ? 'ethereum' : 'email';
		email = params.get('email') ?? '';
		eventId = params.get('event_id') ?? '';
		next = sanitizeNext(params.get('next') ?? '');
		autoEmail = params.get('auto') === '1';

		if (method === 'email') {
			identity = email;
			if (!email) {
				error = 'Email is required.';
				loading = false;
				return;
			}
			resolveEmail();
			return;
		}
		connectAndResolveEthereum();
	});

	onDestroy(() => {
		stopMagicLinkPolling();
	});

	function sanitizeNext(value: string) {
		return value.startsWith('/') && !value.startsWith('//') ? value : '';
	}

	function afterSignInPath() {
		return next || (eventId ? `/events/${eventId}` : '/');
	}

	function isDelegationRequest() {
		return next.startsWith('/auth/delegate?');
	}

	function delegationReturnHost() {
		if (!isDelegationRequest()) return '';
		try {
			const query = next.split('?')[1] ?? '';
			const returnTo = new URLSearchParams(query).get('return_to') ?? '';
			return returnTo ? new URL(returnTo).host : '';
		} catch {
			return '';
		}
	}

	async function resolveEmail() {
		await resolveHome({ type: 'email', email, event_id: eventId });
	}

	async function connectAndResolveEthereum() {
		loading = true;
		error = '';
		try {
			const wallet = window.ethereum;
			if (!wallet) throw new Error(intl.t('auth.noWallet'));
			const accounts = await wallet.request({ method: 'eth_requestAccounts' }) as string[];
			address = accounts[0] ?? '';
			if (!address) throw new Error(intl.t('auth.noAccount'));
			const chainHex = await wallet.request({ method: 'eth_chainId' }) as string;
			chainId = String(Number.parseInt(chainHex, 16));
			identity = intl.t('auth.ethereumIdentity', { address, chainId });
			await resolveHome({ type: 'ethereum', chain_id: chainId, address, event_id: eventId });
		} catch (err: any) {
			error = err.code === 4001 ? intl.t('auth.walletCancelled') : err.message;
		} finally {
			loading = false;
		}
	}

	async function resolveHome(input: Parameters<typeof api.auth.resolveHome>[0]) {
		loading = true;
		error = '';
		discoveredHome = '';
		instance = null;
		delegateURL = '';
		homeIsCurrentInstance = false;
		passkeyAvailable = false;
		noHomeFound = false;
		try {
			const resolved = await api.auth.resolveHome(input);
			discoveredHome = resolved.home_instance_url;
			instance = resolved.instance ?? null;
			delegateURL = resolved.delegate_url;
			homeIsCurrentInstance = resolved.home_instance_url.replace(/\/$/, '') === window.location.origin.replace(/\/$/, '');
			passkeyAvailable = !!resolved.auth_methods?.passkey;
			if (autoEmail && !isDelegationRequest() && method === 'email' && homeIsCurrentInstance && !passkeyAvailable && !autoEmailRequested) {
				autoEmailRequested = true;
				await requestLocalMagicLink();
			}
		} catch (err: any) {
			if (String(err.message).includes('not found')) {
				noHomeFound = true;
				return;
			}
			error = err.message;
		} finally {
			loading = false;
		}
	}

	async function continueOnThisInstance() {
		if (method === 'email') {
			await requestLocalMagicLink();
			return;
		}
		await signInWithEthereumHere();
	}

	async function requestLocalMagicLink() {
		if (!email) return;
		submitting = true;
		error = '';
		try {
			const res = await api.auth.requestMagicLink(email);
			magicLinkPollToken = res.poll_token;
			sent = true;
			startMagicLinkPolling();
		} catch (err: any) {
			error = err.message;
		} finally {
			submitting = false;
		}
	}

	function startMagicLinkPolling() {
		stopMagicLinkPolling();
		if (!magicLinkPollToken) return;
		magicLinkPollTimer = setInterval(checkMagicLinkStatus, 2000);
		checkMagicLinkStatus();
	}

	function stopMagicLinkPolling() {
		if (magicLinkPollTimer) {
			clearInterval(magicLinkPollTimer);
			magicLinkPollTimer = null;
		}
	}

	async function checkMagicLinkStatus() {
		if (!magicLinkPollToken) return;
		try {
			const res = await api.auth.magicLinkStatus(magicLinkPollToken);
			if (res.status !== 'signed_in') return;
			stopMagicLinkPolling();
			await auth.load();
			goto(afterSignInPath());
		} catch (err: any) {
			stopMagicLinkPolling();
			error = err.message;
		}
	}

	async function signInWithEthereumHere() {
		submitting = true;
		error = '';
		try {
			const wallet = window.ethereum;
			if (!wallet) throw new Error(intl.t('auth.noWallet'));
			if (!address || !chainId) {
				await connectAndResolveEthereum();
			}
			const { message } = await api.auth.siwe.start(address, chainId);
			const signature = await wallet.request({
				method: 'personal_sign',
				params: [message, address]
			}) as string;
			await api.auth.siwe.finish(message, signature);
			await auth.load();
			goto(afterSignInPath());
		} catch (err: any) {
			error = err.code === 4001 ? intl.t('auth.walletCancelled') : err.message;
		} finally {
			submitting = false;
		}
	}

	async function signInWithPasskeyHere() {
		submitting = true;
		error = '';
		try {
			const { data: options, session } = await api.auth.passkey.loginStart(email || undefined);
			const credential: any = await navigator.credentials.get({
				publicKey: recursiveBase64ToBuffer(options.publicKey)
			});

			const finishData = {
				id: credential.id,
				rawId: bufferToBase64(credential.rawId),
				type: credential.type,
				response: {
					authenticatorData: bufferToBase64(credential.response.authenticatorData),
					clientDataJSON: bufferToBase64(credential.response.clientDataJSON),
					signature: bufferToBase64(credential.response.signature),
					userHandle: credential.response.userHandle ? bufferToBase64(credential.response.userHandle) : null
				}
			};

			await api.auth.passkey.loginFinish(finishData, session);
			await auth.load();
			goto(afterSignInPath());
		} catch (err: any) {
			error = err.name === 'NotAllowedError' ? intl.t('auth.signInCancelled') : err.message;
		} finally {
			submitting = false;
		}
	}

	function instanceHost(url: string) {
		try {
			return new URL(url).host;
		} catch {
			return url.replace(/^https?:\/\//, '').replace(/\/$/, '');
		}
	}

	function shortValue(value: string, start = 12, end = 8) {
		if (!value || value.length <= start + end + 3) return value;
		return `${value.slice(0, start)}...${value.slice(-end)}`;
	}

	function instanceFingerprint(info: ResolvedInstanceInfo | null) {
		if (!info?.public_key) return '';
		return shortValue(info.public_key, 16, 12);
	}
</script>

<main>
	{#if sent}
		<h1>{intl.t('auth.checkEmail')}</h1>
		<p>{intl.t('auth.linkSentPrefix')} <strong>{email}</strong>.</p>
		<p class="muted">{intl.t('auth.linkExpires')}</p>
		<p class="muted">{intl.t('auth.magicLinkWaiting')}</p>
	{:else}
		<h1>{isDelegationRequest() ? intl.t('auth.approveSignInTitle') : intl.t('auth.chooseHomeTitle')}</h1>
		<p class="muted">
			{isDelegationRequest() ? intl.t('auth.approveSignInHelp') : intl.t('auth.chooseHomeHelp')}
		</p>

		{#if loading}
			<p class="muted">{method === 'ethereum' ? intl.t('auth.connectingWallet') : intl.t('auth.lookingUp')}</p>
		{/if}

		{#if error}
			<p class="error">{error}</p>
		{/if}

		{#if method === 'ethereum' && !loading && !address}
			<button type="button" onclick={connectAndResolveEthereum} disabled={submitting}>
				{intl.t('auth.connectWallet')}
			</button>
		{/if}

		{#if discoveredHome && isDelegationRequest()}
			<div class="consent-card">
				<div class="consent-requester">
					<Avatar did={delegationReturnHost() || next} size={56} rounded={false} />
					<div>
						<p class="eyebrow">{intl.t('auth.requestingInstance')}</p>
						<h2>{delegationReturnHost() || intl.t('auth.unknownInstance')}</h2>
					</div>
				</div>
				<div class="consent-detail">
					<span>{intl.t('auth.signInAs')}</span>
					<strong>{identity}</strong>
				</div>
				<p class="muted">{intl.t('auth.approveSignInDescription')}</p>
				<div class="actions" class:two-up={method === 'email' && passkeyAvailable}>
					<button type="button" onclick={continueOnThisInstance} disabled={submitting}>
						{submitting ? intl.t('auth.sending') : intl.t('auth.continueEmailLink')}
					</button>
					{#if method === 'email' && passkeyAvailable}
						<button type="button" class="secondary" onclick={signInWithPasskeyHere} disabled={submitting}>
							{submitting ? intl.t('auth.sending') : intl.t('auth.continuePasskey')}
						</button>
					{/if}
				</div>
			</div>
		{:else if discoveredHome}
			<div class="instance-card">
				<div class="instance-summary">
					<Avatar did={instance?.public_key ?? discoveredHome} size={64} rounded={false} />
					<div>
						<p class="eyebrow">
							{homeIsCurrentInstance ? intl.t('auth.instanceLocal') : intl.t('auth.instanceRemote')}
						</p>
						<h2>{instanceHost(discoveredHome)}</h2>
						<p class="muted">{intl.t('auth.foundHome', { identity })}</p>
					</div>
				</div>
				<div class="instance-details">
					<div>
						<span>{intl.t('auth.instanceHost')}</span>
						<strong>{discoveredHome}</strong>
					</div>
					{#if instance?.id}
						<div>
							<span>{intl.t('auth.instanceId')}</span>
							<code title={instance.id}>{shortValue(instance.id)}</code>
						</div>
					{/if}
					{#if instance?.public_key}
						<div>
							<span>{intl.t('auth.instancePublicKey')}</span>
							<code title={instance.public_key}>{instanceFingerprint(instance)}</code>
						</div>
					{/if}
				</div>
				{#if instance?.verified}
					<p class="verified">{intl.t('auth.instanceVerified')}</p>
				{/if}
				{#if homeIsCurrentInstance}
					<div class="actions" class:two-up={method === 'email' && passkeyAvailable}>
						<button type="button" onclick={continueOnThisInstance} disabled={submitting}>
							{submitting ? intl.t('auth.sending') : (method === 'email' && passkeyAvailable ? intl.t('auth.continueEmailLink') : intl.t('auth.continueLocal'))}
						</button>
						{#if method === 'email' && passkeyAvailable}
							<button type="button" class="secondary" onclick={signInWithPasskeyHere} disabled={submitting}>
								{submitting ? intl.t('auth.sending') : intl.t('auth.continuePasskey')}
							</button>
						{/if}
					</div>
				{:else}
					<div class="actions two-up">
						<a class="button" href={delegateURL}>{intl.t('auth.continueRemote')}</a>
						<button type="button" class="secondary" onclick={continueOnThisInstance} disabled={submitting}>
							{submitting ? intl.t('auth.sending') : intl.t('auth.registerLocal')}
						</button>
					</div>
				{/if}
			</div>
		{/if}

		{#if noHomeFound}
			<div class="instance-card">
				<p>{intl.t('auth.noHomeFound', { identity })}</p>
				<button type="button" onclick={continueOnThisInstance} disabled={submitting}>
					{submitting ? intl.t('auth.sending') : intl.t('auth.registerLocal')}
				</button>
			</div>
		{/if}
	{/if}
</main>

<style>
	main {
		max-width: 720px;
		margin: 4rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 0.5rem; color: var(--text); }
	p { color: var(--text-secondary); margin: 0.25rem 0 1rem; }
	.muted { color: var(--text-muted); }
	button {
		padding: 0.7rem;
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		width: 100%;
	}
	button:disabled { opacity: 0.5; cursor: default; }
	button.secondary {
		background: var(--bg-btn-secondary);
		color: var(--text-btn-secondary);
		border: 1px solid var(--border-input);
	}
	a.button {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0.7rem;
		background: var(--bg-btn-primary);
		color: var(--text-btn-primary);
		border-radius: var(--radius-md);
		text-align: center;
		text-decoration: none;
		font-weight: 600;
	}
	.instance-card {
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-xl);
		padding: 1.25rem;
		margin-top: 1rem;
		background: var(--bg-card);
	}
	.instance-card p { margin-bottom: 0.5rem; }
	.consent-card {
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-xl);
		padding: 1.25rem;
		margin-top: 1rem;
		background: var(--bg-card);
	}
	.consent-requester {
		display: flex;
		gap: 1rem;
		align-items: center;
		margin-bottom: 1rem;
	}
	.consent-requester h2 {
		margin: 0.1rem 0 0;
		color: var(--text);
		font-size: 1.25rem;
		word-break: break-word;
	}
	.consent-requester p { margin: 0; }
	.consent-detail {
		display: grid;
		gap: 0.15rem;
		padding: 0.85rem;
		margin-bottom: 0.75rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-md);
		background: var(--bg);
	}
	.consent-detail span {
		color: var(--text-muted);
		font-size: 0.75rem;
		font-weight: 600;
	}
	.consent-detail strong {
		color: var(--text);
		word-break: break-word;
	}
	.instance-summary {
		display: flex;
		gap: 1rem;
		align-items: flex-start;
		margin-bottom: 1.25rem;
	}
	.instance-summary h2 {
		margin: 0.1rem 0 0.2rem;
		color: var(--text);
		font-size: 1.15rem;
		word-break: break-word;
	}
	.instance-summary p { margin: 0; }
	.eyebrow {
		color: var(--text-muted);
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}
	.instance-details {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.6rem;
		margin: 0.75rem 0;
	}
	.instance-details div:first-child {
		grid-column: 1 / -1;
	}
	.instance-details div {
		display: grid;
		gap: 0.15rem;
		padding: 0.75rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-md);
		background: var(--bg);
	}
	.instance-details span {
		color: var(--text-muted);
		font-size: 0.75rem;
		font-weight: 600;
	}
	.instance-details strong,
	.instance-details code {
		color: var(--text);
		word-break: break-word;
	}
	.instance-details code {
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 0.8rem;
	}
	.actions {
		display: grid;
		gap: 0.6rem;
		margin-top: 0.75rem;
	}
	.actions.two-up {
		grid-template-columns: 1fr 1fr;
	}
	.verified {
		display: inline-flex;
		margin-bottom: 0.75rem;
		padding: 0.25rem 0.55rem;
		border-radius: 999px;
		background: var(--bg-success);
		color: var(--text-success);
		font-size: 0.8rem;
		font-weight: 600;
	}
	.error { color: var(--text-error); font-size: 0.875rem; }
	@media (max-width: 640px) {
		main {
			max-width: 420px;
		}
		.instance-details,
		.actions.two-up {
			grid-template-columns: 1fr;
		}
	}
</style>
