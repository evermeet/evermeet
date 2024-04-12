import { apiCall } from '$lib/api.js';
import { loadConfig } from '$lib/config.js';

export async function load({ fetch, cookies }) {

    let loggedIn = false;
    if (cookies.get('evermeet-session-id')) {
        loggedIn = await apiCall(fetch, 'me');
    }
    const config = await loadConfig();

    return {
        user: loggedIn?.user,
        config
    }
}