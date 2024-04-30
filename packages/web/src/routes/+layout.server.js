import { xrpcCall } from '$lib/api.js';
import { loadConfig } from '$lib/config.js';

export async function load({ fetch, cookies }) {

    const config = await loadConfig();
    let user = false;
    const sessionId = cookies.get('evermeet-session-id')
    if (sessionId) {
        try {
            const session = await xrpcCall({ fetch }, 'app.evermeet.auth.getSession', null, null, {
                headers: {
                    authorization: 'Bearer ' + sessionId
                }
            });
            //session.accessJwt = sessionId
            user = session.user
            user.token = sessionId

        } catch (e) {
            console.error(e)
        }
    }
    return {
        user,
        config
    }
}