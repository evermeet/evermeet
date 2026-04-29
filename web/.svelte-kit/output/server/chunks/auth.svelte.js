import "./dev.js";
import { t as api } from "./api.js";
//#region src/lib/auth.svelte.ts
function createAuthStore() {
	let user = null;
	let loading = true;
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
		get user() {
			return user;
		},
		get loading() {
			return loading;
		},
		load,
		logout
	};
}
var auth = createAuthStore();
//#endregion
export { auth as t };
