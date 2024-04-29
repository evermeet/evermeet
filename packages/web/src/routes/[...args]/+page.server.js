
import { xrpcCall } from '$lib/api.js';

export async function load({ params, fetch, query, url }) {
	const [ id, tab ] = params.args.split('/')
	const result = await xrpcCall(fetch, 'app.evermeet.object.getProfile', { id })

	if (tab === 'chat' && result.item && result.item.rooms) {
		const roomQuery = url.searchParams.get('room')
		const roomName = typeof(roomQuery) === 'string' ? roomQuery : 'general'
		const roomId = result.item.rooms.find(r => r.slug === roomName)?.id
		if (roomId) {
			result.item._chat = (await xrpcCall(fetch, 'app.evermeet.chat.getMessages', { room: roomId })).reverse()
		}
	}

	return {
		selectedTab: tab || null,
		query,
		id,
		result,
	}
}