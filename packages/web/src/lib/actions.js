
import { user, eventDetail } from '$lib/stores';
import { apiCall } from '$lib/api';

export function calendarSubscribe (id) {
    return async () => {
        const resp = await apiCall(fetch, 'me/calendarSubscribe', {}, { id });
        user.update(u => Object.assign(u, { subscribedCalendars: resp.subscribedCalendars }))
    }
}

export function calendarUnsubscribe (id) {
    return async () => {
        const resp = await apiCall(fetch, 'me/calendarUnsubscribe', {}, { id });
        user.update(u => Object.assign(u, { subscribedCalendars: resp.subscribedCalendars }))
    }
}

export function register (id) {
    return async () => {
        const resp = await apiCall(fetch, 'me/register', {}, { id })
        if (resp) {
            user.update(u => Object.assign(u, { events: resp.events }))
            eventDetail.update(u => {
                u.guestsNative = resp.event.guestsNative
                u.guestCountTotal = resp.event.guestCountTotal
                return u
            })
        }
    }
}

export function unregister (id) {
    return async () => {
        const resp = await apiCall(fetch, 'me/unregister', {}, { id })
        if (resp) {
            user.update(u => Object.assign(u, { events: resp.events }))
            eventDetail.update(u => {
                u.guestsNative = resp.event.guestsNative
                u.guestCountTotal = resp.event.guestCountTotal
                return u
            })
        }
    }
}