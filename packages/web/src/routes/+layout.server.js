import { apiCall } from '$lib/api.js';

export async function load({ fetch, cookies }) {

    const sessionId = cookies.get('evermeet-session-id');
    let user = null;
    if (sessionId) {
        const resp = await apiCall(fetch, 'me', {
            headers: {
                'evermeet-session-id': sessionId
            }
        })
        user = resp.user;
    }

    return {
        sessionId,
        user,
    }
}