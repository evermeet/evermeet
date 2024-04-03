
import { apiCall } from '../../../../lib/api.js';

export function load({ params, fetch }) {
	return apiCall(fetch, 'event/' + params.id);
}