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
	const headers = new Headers(init?.headers);
	headers.set('Content-Type', 'application/json');
	if (sessionToken) headers.set('X-WebAuthn-Session', sessionToken);

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
	blobs: {
		upload: async (file: File): Promise<{ hash: string; url: string }> => {
			const form = new FormData();
			form.append('file', file);
			const res = await fetch('/api/blobs', {
				method: 'POST',
				credentials: 'include',
				body: form,
			});
			if (!res.ok) {
				const err = await res.json().catch(() => ({ error: res.statusText }));
				throw new Error(err.error ?? res.statusText);
			}
			return res.json();
		},
	},
	auth: {
		me: () => request<AuthUser>('/api/auth/me'),
		methods: () => request<AuthMethods>('/api/auth/methods'),
		requestMagicLink: (email: string) =>
			request<{ status: string; poll_token: string }>('/api/auth/magic-link', {
				method: 'POST',
				body: JSON.stringify({ email }),
			}),
		magicLinkStatus: (pollToken: string) =>
			request<{ status: 'pending' | 'signed_in' }>('/api/auth/magic-link/status', {
				method: 'POST',
				body: JSON.stringify({ poll_token: pollToken }),
			}),
		updateProfile: (data: { display_name: string; avatar: string; bio: string }) =>
			request<{ status: string }>('/api/auth/profile', {
				method: 'PUT',
				body: JSON.stringify(data),
			}),
		resolveHome: (input: ResolveHomeRequest | string, eventId = '') => {
			const body = typeof input === 'string'
				? { type: 'email', email: input, event_id: eventId }
				: input;
			return request<ResolveHomeResponse>('/api/auth/resolve-home', {
				method: 'POST',
				body: JSON.stringify(body),
			});
		},
		delegate: (data: DelegateRequest) =>
			request<SignedDelegationToken>('/api/auth/delegate', {
				method: 'POST',
				body: JSON.stringify(data),
			}),
		siwe: {
			start: (address: string, chainId: string) =>
				request<SIWEStartResponse>('/api/auth/siwe/start', {
					method: 'POST',
					body: JSON.stringify({ address, chain_id: chainId }),
				}),
			finish: (message: string, signature: string) =>
				request<{ status: string }>('/api/auth/siwe/finish', {
					method: 'POST',
					body: JSON.stringify({ message, signature }),
				}),
		},
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
		history: (id: string) =>
			request<{ id: string; revisions: EventRevision[] }>(`/api/events/${id}/history`),
		create: (data: CreateEventInput) =>
			request<{ id: string; hash: string; state: Event }>('/api/events', {
				method: 'POST',
				body: JSON.stringify(data),
			}),
		importPreview: (url: string) =>
			request<ImportEventPreview>('/api/events/import/preview', {
				method: 'POST',
				body: JSON.stringify({ url }),
			}),
		update: (id: string, data: CreateEventInput) =>
			request<{ id: string; hash: string; state: Event }>(`/api/events/${id}`, {
				method: 'PUT',
				body: JSON.stringify(data),
			}),
		delete: (id: string) =>
			request<{ status: string }>(`/api/events/${id}`, {
				method: 'DELETE',
			}),
		attendees: (id: string) =>
			request<{ attendees: EventAttendee[]; count: number }>(`/api/events/${id}/attendees`),
		listRSVPs: (id: string) => request<any[]>(`/api/events/${id}/rsvps`),
		myRSVPStatus: (id: string) =>
			request<MyRSVPStatus>(`/api/events/${id}/rsvp/status`),
		rsvp: (id: string, data: { name?: string; email?: string; note?: string }) =>
			request<{ status: string; received_at?: string }>(`/api/events/${id}/rsvp`, {
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
	calendar?: string;
	title: string;
	description?: string;
	cover_url?: string;
	starts_at: string;
	ends_at?: string;
	location?: { name: string; address?: string; lat?: number; lon?: number };
	visibility: 'public' | 'unlisted' | 'private';
	rsvp: { limit: number; count: number; deadline?: string; approval: string; visible?: boolean };
	tags?: string[];
	updated_at: string;
}

export interface EventRevision {
	hash: string;
	prev?: string;
	is_current: boolean;
	created_at: string;
	state: Event;
}

export interface MyRSVPStatus {
	has_rsvp: boolean;
	status?: 'pending' | 'confirmed' | 'rejected' | 'waitlisted' | 'cancelled' | string;
	received_at?: string;
}

export interface EventAttendee {
	did: string;
	display_name: string;
	avatar?: string;
}

export type CalendarLinkType = 'website' | 'twitter' | 'instagram' | 'youtube' | 'tiktok' | 'linkedin' | 'bluesky' | 'nostr' | 'facebook';

export interface CalendarLink {
	type: CalendarLinkType;
	url: string;
}

export interface Calendar {
	id: string;
	name: string;
	description?: string;
	avatar?: string;
	backdrop_url?: string;
	website?: string;
	links?: CalendarLink[];
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
	hosts?: string[];
}

export interface CalendarInput {
	name: string;
	description?: string;
	avatar?: string;
	backdrop_url?: string;
	website?: string;
	links?: CalendarLink[];
	owners?: string[];
}

export interface CreateEventInput {
	title: string;
	description?: string;
	cover_url?: string;
	starts_at: string;
	calendar_id?: string;
	owners?: string[];
	ends_at?: string;
	location?: { name: string; address?: string };
	visibility?: string;
	rsvp_limit?: number;
	rsvp_visible?: boolean;
	tags?: string[];
}

export interface ImportEventPreview {
	provider: string;
	source_url: string;
	title: string;
	description?: string;
	cover_url?: string;
	starts_at: string;
	ends_at?: string;
	location_name?: string;
	location_address?: string;
}

export interface ForeignInstanceSig {
	return_to: string;
	nonce: string;
	event_id: string;
	issued_at: number;
	sig: string;
}

export interface ResolveHomeResponse {
	home_instance_url: string;
	nonce: string;
	return_to: string;
	foreign_sig: ForeignInstanceSig;
	delegate_url: string;
	instance?: ResolvedInstanceInfo;
	auth_methods?: ResolvedAuthMethods;
}

export type ResolveHomeRequest =
	| { type: 'email'; email: string; event_id?: string }
	| { type: 'ethereum'; chain_id: string; address: string; event_id?: string };

export interface ResolvedInstanceInfo {
	id: string;
	public_key: string;
	verified: boolean;
}

export interface ResolvedAuthMethods {
	passkey: boolean;
}

export interface DelegateRequest {
	return_to: string;
	nonce: string;
	event_id: string;
	foreign_sig: ForeignInstanceSig;
}

export interface DelegationToken {
	sub: string;
	aud: string;
	iat: number;
	exp: number;
	nonce: string;
	event_id: string;
}

export interface SignedDelegationToken {
	token: DelegationToken;
	sig: string;
	public_key: string;
	home_instance_url: string;
}

export interface SIWEStartResponse {
	message: string;
	nonce: string;
}

export interface AuthUser {
	did: string;
	display_name: string;
	avatar: string;
	bio: string;
	is_local: boolean;
	auth_kind: 'local' | 'remote';
	home_instance_url: string;
}

export interface AuthMethods {
	email: {
		address: string;
		verified: boolean;
		linked: boolean;
	};
	ethereum: {
		chain_id: string;
		address: string;
		linked: boolean;
	};
	passkeys: AuthPasskey[];
}

export interface AuthPasskey {
	id: string;
	attestation_type: string;
	counter: number;
	backup_eligible: boolean;
	backup_state: boolean;
	user_verified: boolean;
	user_present: boolean;
	created_at: string;
}
