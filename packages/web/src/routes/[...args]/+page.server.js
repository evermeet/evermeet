
import { xrpcCall } from '$lib/api.js';

export async function load({ params, fetch, query }) {
	const [ id, tab ] = params.args.split('/')
	return {
		selectedTab: tab || null,
		query,
		id,
		result: await xrpcCall(fetch, 'app.evermeet.object.getProfile', { id })
	}
}