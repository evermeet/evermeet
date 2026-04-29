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

export const api = {
	auth: {
		me: () => request<{ did: string; display_name: string; avatar: string; bio: string }>('/api/auth/me'),
		requestMagicLink: (email: string) =>
			request<{ status: string }>('/api/auth/magic-link', {
				method: 'POST',
				body: JSON.stringify({ email }),
			}),
		logout: () => request<{ status: string }>('/api/auth/logout', { method: 'POST' }),
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
	},
};

export interface Event {
	id: string;
	organizer: string;
	title: string;
	description?: string;
	starts_at: string;
	ends_at?: string;
	location?: { name: string; address?: string; lat?: number; lon?: number };
	visibility: 'public' | 'unlisted' | 'private';
	rsvp: { limit: number; count: number; deadline?: string; approval: string };
	tags?: string[];
	updated_at: string;
}

export interface CreateEventInput {
	title: string;
	description?: string;
	starts_at: string;
	ends_at?: string;
	location?: { name: string; address?: string };
	visibility?: string;
	rsvp_limit?: number;
	tags?: string[];
}
