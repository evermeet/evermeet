<script lang="ts">
	import { api } from '$lib/api.js';

	let email = $state('');
	let sent = $state(false);
	let error = $state('');
	let submitting = $state(false);

	async function submit(e: Event) {
		e.preventDefault();
		if (!email) return;
		submitting = true;
		error = '';
		try {
			await api.auth.requestMagicLink(email);
			sent = true;
		} catch (err: any) {
			error = err.message;
		} finally {
			submitting = false;
		}
	}
</script>

<main>
	{#if sent}
		<h1>Check your email</h1>
		<p>A sign-in link was sent to <strong>{email}</strong>.</p>
		<p class="muted">The link expires in 15 minutes.</p>
	{:else}
		<h1>Sign in</h1>
		<p class="muted">We'll email you a sign-in link — no password needed.</p>

		<form onsubmit={submit}>
			<label for="email">Email</label>
			<input
				id="email"
				type="email"
				bind:value={email}
				placeholder="you@example.com"
				autocomplete="email"
				required
			/>
			{#if error}<p class="error">{error}</p>{/if}
			<button type="submit" disabled={submitting}>
				{submitting ? 'Sending…' : 'Send sign-in link'}
			</button>
		</form>
	{/if}
</main>

<style>
	main {
		max-width: 400px;
		margin: 4rem auto;
		padding: 0 1.5rem;
		font-family: system-ui, sans-serif;
	}
	h1 { font-size: 1.5rem; font-weight: 700; margin-bottom: 0.5rem; }
	p { color: #555; margin: 0.25rem 0 1rem; }
	.muted { color: #999; }
	form { display: flex; flex-direction: column; gap: 0.5rem; }
	label { font-size: 0.875rem; font-weight: 500; }
	input {
		padding: 0.6rem 0.75rem;
		border: 1px solid #d0d0d0;
		border-radius: 6px;
		font-size: 1rem;
		outline: none;
	}
	input:focus { border-color: #1a1a1a; }
	button {
		margin-top: 0.5rem;
		padding: 0.7rem;
		background: #1a1a1a;
		color: #fff;
		border: none;
		border-radius: 6px;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
	}
	button:disabled { opacity: 0.5; cursor: default; }
	.error { color: #c00; font-size: 0.875rem; }
</style>
