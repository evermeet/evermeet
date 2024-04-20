
import { xrpcCall } from '../lib/api.js';

export async function load({ params, fetch, parent }) {
	return xrpcCall(fetch, 'app.evermeet.explore.getExplore');
}