import { xrpcCall } from '$lib/api.js';
import { loadConfig } from '$lib/config.js';

export async function load({ fetch, cookies }) {

    const config = await loadConfig();
    let session = false;
    const sessionId = cookies.get('evermeet-session-id')
    if (sessionId) {

        console.log(sessionId)
        try {
            session = await xrpcCall(fetch, 'app.evermeet.auth.getSession', null, null, {
                headers: {
                    authorization: 'Bearer ' + sessionId
                }
            });
            session.accessJwt = sessionId
        } catch (e) {
            console.error(e)
        }

    }
    return {
        session,
        config
    }
}