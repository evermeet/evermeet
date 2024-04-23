
import { xrpcCall } from '../../lib/api.js';

export async function load({ params, fetch, parent }) {

    const data = await parent()
	return {
        calendars: await xrpcCall(fetch, 'app.evermeet.calendar.getUserCalendars', null, null, { token: data.session?.accessJwt })
    }
}