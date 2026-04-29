//#region src/lib/api.ts
var base = "";
async function request(path, init) {
	const res = await fetch(base + path, {
		credentials: "include",
		headers: {
			"Content-Type": "application/json",
			...init?.headers
		},
		...init
	});
	if (!res.ok) {
		const err = await res.json().catch(() => ({ error: res.statusText }));
		throw new Error(err.error ?? res.statusText);
	}
	return res.json();
}
var api = {
	auth: {
		me: () => request("/api/auth/me"),
		requestMagicLink: (email) => request("/api/auth/magic-link", {
			method: "POST",
			body: JSON.stringify({ email })
		}),
		logout: () => request("/api/auth/logout", { method: "POST" })
	},
	events: {
		list: (limit = 20, offset = 0) => request(`/api/events?limit=${limit}&offset=${offset}`),
		get: (id) => request(`/api/events/${id}`),
		create: (data) => request("/api/events", {
			method: "POST",
			body: JSON.stringify(data)
		}),
		update: (id, data) => request(`/api/events/${id}`, {
			method: "PUT",
			body: JSON.stringify(data)
		})
	}
};
//#endregion
export { api as t };
