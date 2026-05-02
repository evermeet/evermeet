<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, type AdminEmailConfig } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { onMount } from 'svelte';

	let cfg = $state<AdminEmailConfig | null>(null);
	let loading = $state(true);
	let error = $state('');
	let testTo = $state('');
	let testing = $state(false);
	let testMsg = $state('');
	let testError = $state('');

	onMount(async () => {
		await auth.load();
		if (!auth.user) {
			goto('/auth/login?next=/admin/email');
			return;
		}
		if (!auth.user.is_admin) {
			goto('/');
			return;
		}
		await loadConfig();
	});

	async function loadConfig() {
		loading = true;
		error = '';
		try {
			cfg = await api.admin.email();
		} catch (err: any) {
			error = err.message ?? 'Failed to load email settings';
		} finally {
			loading = false;
		}
	}

	async function sendTest() {
		const to = testTo.trim();
		if (!to) return;
		testing = true;
		testMsg = '';
		testError = '';
		try {
			await api.admin.emailTest({ to });
			testMsg = 'Test message sent.';
		} catch (err: any) {
			testError = err.message ?? 'Test send failed';
		} finally {
			testing = false;
		}
	}
</script>

<main>
	<div class="page-header">
		<div>
			<p class="eyebrow">Admin</p>
			<h1>Email</h1>
			<p class="muted">Outbound mail from <code>evermeet.toml</code> — SMTP, or sendmail when SMTP is unset.</p>
		</div>
	</div>
	<AdminNav active="email" />

	{#if loading}
		<p class="muted">Loading...</p>
	{:else if error && !cfg}
		<p class="error">{error}</p>
	{:else if cfg}
		<section class="panel">
			<h2>Mail configuration</h2>
			<dl class="kv">
				<dt>Sending enabled</dt>
				<dd>
					{#if cfg.sending_enabled}
						<span class="ok">Yes</span>
					{:else}
						<span class="warn">No</span>
						<span class="muted small">
							— set <code>smtp_host</code>, or install sendmail and set <code>node.base_url</code> (or
							<code>from</code>), then restart.
						</span>
					{/if}
				</dd>
				<dt>Transport</dt>
				<dd>
					<code>{cfg.transport}</code>
					{#if cfg.transport === 'sendmail' && cfg.sendmail_path}
						<span class="muted small">(<code>{cfg.sendmail_path}</code>)</span>
					{/if}
				</dd>
				<dt>SMTP host</dt>
				<dd><code>{cfg.smtp_host || '—'}</code></dd>
				<dt>SMTP port</dt>
				<dd><code>{cfg.smtp_port || '—'}</code></dd>
				<dt>SMTP user</dt>
				<dd><code>{cfg.smtp_user || '—'}</code></dd>
				<dt>Password</dt>
				<dd>{cfg.password_set ? 'Set (hidden)' : 'Not set'}</dd>
				<dt>From (config)</dt>
				<dd><code>{cfg.from || '—'}</code></dd>
				<dt>From (used)</dt>
				<dd><code>{cfg.effective_from || '—'}</code></dd>
			</dl>
		</section>

		<section class="panel">
			<h2>Send test email</h2>
			<p class="muted">Sends a short HTML message with subject “Evermeet mail test”.</p>
			<div class="form-row">
				<input
					type="email"
					placeholder="recipient@example.com"
					bind:value={testTo}
					disabled={testing || !cfg.sending_enabled}
				/>
				<button onclick={sendTest} disabled={testing || !cfg.sending_enabled || !testTo.trim()}>
					{testing ? 'Sending…' : 'Send test'}
				</button>
			</div>
			{#if testError}<p class="feedback error">{testError}</p>{/if}
			{#if testMsg}<p class="feedback success">{testMsg}</p>{/if}
		</section>
	{/if}
</main>

<style>
	main {
		max-width: var(--layout-page-wide);
		margin: 2.5rem auto;
		padding: 0 1.5rem 4rem;
		font-family: system-ui, sans-serif;
	}

	.page-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		margin-bottom: 0.75rem;
	}

	.eyebrow {
		margin: 0 0 0.35rem;
		color: var(--text-muted);
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}
	h1 {
		margin: 0;
	}
	h2 {
		margin: 0 0 0.75rem;
		font-size: 1.05rem;
	}
	.muted {
		color: var(--text-secondary);
	}
	.small {
		font-size: 0.85rem;
	}
	.error {
		color: var(--text-error);
	}
	.success {
		color: var(--text-success);
	}
	.ok {
		color: var(--text-success);
		font-weight: 600;
	}
	.warn {
		color: var(--text-warning, #b45309);
		font-weight: 600;
	}
	.panel {
		margin-top: 1rem;
		padding: 1rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-lg);
		background: var(--bg-card);
	}
	.kv {
		display: grid;
		grid-template-columns: 10rem 1fr;
		gap: 0.5rem 1rem;
		margin: 0;
	}
	dt {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.9rem;
	}
	dd {
		margin: 0;
		min-width: 0;
	}
	code {
		font-size: 0.85rem;
		overflow-wrap: anywhere;
	}
	.form-row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.6rem;
		align-items: center;
		margin-top: 0.75rem;
	}
	input,
	button {
		padding: 0.55rem 0.7rem;
		border-radius: var(--radius-md);
		border: 1px solid var(--border-input);
		background: var(--bg-input);
		color: var(--text);
	}
	input {
		flex: 1;
		min-width: 12rem;
	}
	button {
		background: var(--bg-raised);
		cursor: pointer;
	}
	button:disabled {
		opacity: 0.55;
		cursor: not-allowed;
	}
	.feedback {
		margin: 0.75rem 0 0;
		font-size: 0.95rem;
	}
</style>
