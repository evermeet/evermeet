
import { apiCall } from '../../lib/api.js';

export async function load({ params, fetch, parent }) {
	return apiCall(fetch, 'explore');
}