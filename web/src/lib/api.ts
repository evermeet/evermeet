const base = '';

async function request<T>(path: string, init?: RequestInit): Promise<T> {
	const res = await fetch(base + path, {
		credentials: 'include',
		headers: { 'Content-Type': 'application/json', ...init?.headers },
		...init,
	});
	if (!res.ok) {
		const err = await res.json().catch(() => ({ error: res.statusText }));
		throw new Error(err.error ?? res.statusText);
	}
	return res.json();
}

async function requestWithSession<T>(path: string, sessionToken?: string | null, init?: RequestInit): Promise<{ data: T; session: string }> {
	const headers: Record<string, string> = { 'Content-Type': 'application/json', ...init?.headers };
	if (sessionToken) headers['X-WebAuthn-Session'] = sessionToken;

	const res = await fetch(base + path, {
		credentials: 'include',
		headers,
		...init,
	});
	if (!res.ok) {
		const err = await res.json().catch(() => ({ error: res.statusText }));
		throw new Error(err.error ?? res.statusText);
	}
	return {
		data: await res.json(),
		session: res.headers.get('X-WebAuthn-Session') ?? sessionToken ?? '',
	};
}

export const api = {
	auth: {
		me: () => request<{ did: string; display_name: string; avatar: string; bio: string }>('/api/auth/me'),
		requestMagicLink: (email: string) =>
			request<{ status: string }>('/api/auth/magic-link', {
				method: 'POST',
				body: JSON.stringify({ email }),
			}),
		updateProfile: (data: { display_name: string; avatar: string; bio: string }) =>
			request<{ status: string }>('/api/auth/profile', {
				method: 'PUT',
				body: JSON.stringify(data),
			}),
		logout: () => request<{ status: string }>('/api/auth/logout', { method: 'POST' }),
		passkey: {
			signupStart: () => requestWithSession<any>('/api/auth/passkey/signup/start', null, { method: 'POST' }),
			signupFinish: (data: any, sessionToken: string) =>
				requestWithSession<any>('/api/auth/passkey/signup/finish', sessionToken, {
					method: 'POST',
					body: JSON.stringify(data),
				}),
			registerStart: () => requestWithSession<any>('/api/auth/passkey/register/start', null, { method: 'POST' }),
			registerFinish: (data: any, sessionToken: string) =>
				requestWithSession<any>('/api/auth/passkey/register/finish', sessionToken, {
					method: 'POST',
					body: JSON.stringify(data),
				}),
			loginStart: (email?: string) =>
				requestWithSession<any>('/api/auth/passkey/login/start', null, {
					method: 'POST',
					body: JSON.stringify({ email }),
				}),
			loginFinish: (data: any, sessionToken: string) =>
				requestWithSession<any>('/api/auth/passkey/login/finish', sessionToken, {
					method: 'POST',
					body: JSON.stringify(data),
				}),
		},
	},
	events: {
		list: (limit = 20, offset = 0) =>
			request<{ events: Event[]; limit: number; offset: number }>(
				`/api/events?limit=${limit}&offset=${offset}`
			),
		get: (id: string) =>
			request<{ id: string; state: Event; hash: string; founded: object }>(`/api/events/${id}`),
		create: (data: CreateEventInput) =>
			request<{ id: string; hash: string; state: Event }>('/api/events', {
				method: 'POST',
				body: JSON.stringify(data),
			}),
		update: (id: string, data: CreateEventInput) =>
			request<{ id: string; hash: string; state: Event }>(`/api/events/${id}`, {
				method: 'PUT',
				body: JSON.stringify(data),
			}),
		listRSVPs: (id: string) => request<any[]>(`/api/events/${id}/rsvps`),
		rsvp: (id: string, data: { name: string; email: string; note?: string }) =>
			request<{ status: string }>(`/api/events/${id}/rsvp`, {
				method: 'POST',
				body: JSON.stringify(data),
			}),
	},
	calendars: {
		list: () => request<{ owned: Calendar[]; subscribed: Calendar[] }>('/api/calendars'),
		discover: () => request<{ calendars: DiscoverCalendar[] }>('/api/calendars/discover'),
		get: (id: string) => request<CalendarDetail>(`/api/calendars/${id}`),
		create: (data: CalendarInput) =>
			request<{ id: string; state: Calendar }>('/api/calendars', {
				method: 'POST',
				body: JSON.stringify(data),
			}),
		update: (id: string, data: CalendarInput) =>
			request<{ id: string; state: Calendar }>(`/api/calendars/${id}`, {
				method: 'PUT',
				body: JSON.stringify(data),
			}),
		subscribe: (id: string) =>
			request<{ subscribed: boolean }>(`/api/calendars/${id}/subscribe`, { method: 'POST' }),
		unsubscribe: (id: string) =>
			request<{ subscribed: boolean }>(`/api/calendars/${id}/subscribe`, { method: 'DELETE' }),
	},
	node: {
		status: () => request<any>('/api/instance/status'),
	},
	instance: {
		status: () => request<any>('/api/instance/status'),
	},
	users: {
		get: (did: string) => request<{ did: string; display_name: string; avatar: string; bio: string; updated_at: string }>(`/api/users/${did}`),
	},
};

export interface Event {
	id: string;
	organizer: string;
	title: string;
	description?: string;
	cover_url?: string;
	starts_at: string;
	ends_at?: string;
	location?: { name: string; address?: string; lat?: number; lon?: number };
	visibility: 'public' | 'unlisted' | 'private';
	rsvp: { limit: number; count: number; deadline?: string; approval: string };
	tags?: string[];
	updated_at: string;
}

export interface Calendar {
	id: string;
	name: string;
	description?: string;
	avatar?: string;
	backdrop_url?: string;
	website?: string;
	subscribers: number;
}

export interface CalendarDetail extends Calendar {
	subscribed: boolean;
	governance: { threshold: number; owners: { did: string; role: string }[] };
	updated_at: string;
	events: CalendarEvent[];
}

export interface DiscoverCalendar extends Calendar {
	subscribed: boolean;
}

export interface CalendarEvent {
	id: string;
	title: string;
	starts_at: string;
	ends_at?: string;
	location?: { name: string; address?: string };
	cover_url?: string;
}

export interface CalendarInput {
	name: string;
	description?: string;
	avatar?: string;
	backdrop_url?: string;
	website?: string;
	owners?: string[];
}

export interface CreateEventInput {
	title: string;
	description?: string;
	cover_url?: string;
	starts_at: string;
	calendar_id?: string;
	ends_at?: string;
	location?: { name: string; address?: string };
	visibility?: string;
	rsvp_limit?: number;
	tags?: string[];
}
