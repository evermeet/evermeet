
import { xrpcCall } from '../../lib/api.js';

export function load({ params, fetch }) {
	return xrpcCall(fetch, 'app.evermeet.object.getProfile', { id: params.id });
}