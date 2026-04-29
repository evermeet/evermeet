import { api } from './api.js';

interface AuthUser {
	did: string;
	display_name: string;
	avatar: string;
	bio: string;
}

function createAuthStore() {
	let user = $state<AuthUser | null>(null);
	let loading = $state(true);

	async function load() {
		try {
			user = await api.auth.me();
		} catch {
			user = null;
		} finally {
			loading = false;
		}
	}

	async function logout() {
		await api.auth.logout();
		user = null;
	}

	return {
		get user() { return user; },
		get loading() { return loading; },
		load,
		logout,
	};
}

export const auth = createAuthStore();
