<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, type AdminAccount, type AdminRole } from '$lib/api.js';
	import { auth } from '$lib/auth.svelte.js';
	import AdminNav from '$lib/components/AdminNav.svelte';
	import { onMount } from 'svelte';

	let admins = $state<AdminAccount[]>([]);
	let loading = $state(true);
	let error = $state('');
	let saving = $state(false);
	let success = $state('');

	let myRole = $state<AdminRole | ''>('');
	let email = $state('');
	let role = $state<AdminRole>('admin');

	const canManage = $derived(myRole === 'owner');

	onMount(async () => {
		await auth.load();
		if (!auth.user) {
			goto('/auth/login?next=/admin/admins');
			return;
		}
		if (!auth.user.is_admin) {
			goto('/');
			return;
		}
		await loadAdmins();
	});

	async function loadAdmins() {
		loading = true;
		error = '';
		try {
			const res = await api.admin.admins();
			admins = res.admins ?? [];
			myRole = res.my_role;
		} catch (err: any) {
			error = err.message ?? 'Failed to load admins';
		} finally {
			loading = false;
		}
	}

	async function addAdmin() {
		if (!email.trim()) return;
		saving = true;
		error = '';
		success = '';
		try {
			await api.admin.createAdmin({ email: email.trim(), role });
			success = 'Admin added.';
			email = '';
			role = 'admin';
			await loadAdmins();
		} catch (err: any) {
			error = err.message ?? 'Failed to add admin';
		} finally {
			saving = false;
		}
	}

	async function updateRole(did: string, nextRole: AdminRole) {
		saving = true;
		error = '';
		success = '';
		try {
			await api.admin.setAdminRole(did, nextRole);
			success = 'Role updated.';
			await loadAdmins();
		} catch (err: any) {
			error = err.message ?? 'Failed to update role';
		} finally {
			saving = false;
		}
	}
</script>

<main>
	<div class="page-header">
		<div>
			<p class="eyebrow">Admin</p>
			<h1>Admins</h1>
			<p class="muted">Manage admin accounts and roles for this instance.</p>
		</div>
	</div>
	<AdminNav active="admins" />

	{#if loading}
		<p class="muted">Loading admins...</p>
	{:else}
		{#if error}<p class="error">{error}</p>{/if}
		{#if success}<p class="success">{success}</p>{/if}

		<section class="panel">
			<h2>Add Admin</h2>
			{#if canManage}
				<div class="form-row">
					<input
						type="email"
						placeholder="user@example.com"
						bind:value={email}
						disabled={saving}
					/>
					<select bind:value={role} disabled={saving}>
						<option value="admin">admin</option>
						<option value="owner">owner</option>
					</select>
					<button onclick={addAdmin} disabled={saving || !email.trim()}>
						{saving ? 'Saving...' : 'Add admin'}
					</button>
				</div>
			{:else}
				<p class="muted">Only owners can add new admins.</p>
			{/if}
		</section>

		<section class="panel">
			<h2>Current Admins</h2>
			<div class="list">
				{#each admins as admin}
					<div class="row">
						<div class="identity">
							<strong>{admin.display_name}</strong>
							<code>{admin.did}</code>
							{#if admin.endpoint}<span class="muted">{admin.endpoint}</span>{/if}
						</div>
						<div class="actions">
							{#if canManage}
								<select
									value={admin.role}
									disabled={saving}
									onchange={(e) => updateRole(admin.did, (e.currentTarget as HTMLSelectElement).value as AdminRole)}
								>
									<option value="owner">owner</option>
									<option value="admin">admin</option>
								</select>
							{:else}
								<span class="role-badge">{admin.role}</span>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</section>
	{/if}
</main>

<style>
	main {
		max-width: 980px;
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
	h1 { margin: 0; }
	.muted { color: var(--text-secondary); }
	.error { color: var(--text-error); }
	.success { color: var(--text-success); }
	.panel {
		margin-top: 1rem;
		padding: 1rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-lg);
		background: var(--bg-card);
	}
	.form-row {
		display: grid;
		grid-template-columns: 1fr auto auto;
		gap: 0.6rem;
	}
	input, select, button {
		padding: 0.55rem 0.7rem;
		border-radius: var(--radius-md);
		border: 1px solid var(--border-input);
		background: var(--bg-input);
		color: var(--text);
	}
	button {
		background: var(--bg-raised);
		color: var(--text);
		border-color: var(--border-input);
		cursor: pointer;
	}
	.list { display: grid; gap: 0.5rem; }
	.row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem;
		border: 1px solid var(--border-subtle);
		border-radius: var(--radius-md);
	}
	.identity { display: grid; gap: 0.2rem; min-width: 0; }
	code {
		font-size: 0.75rem;
		color: var(--text-muted);
		overflow-wrap: anywhere;
	}
	.role-badge {
		padding: 0.2rem 0.55rem;
		border-radius: 999px;
		border: 1px solid var(--border-subtle);
		text-transform: lowercase;
	}
	@media (max-width: 760px) {
		.form-row { grid-template-columns: 1fr; }
		.row { flex-direction: column; align-items: flex-start; }
	}
</style>
