import { browser } from '$app/environment'; 
import { env } from '$env/dynamic/public';

export async function apiCall (fetch, path, opts = {}, data=false) {
    const apiHost = env[browser ? "VITE_BACKEND_URL_PUBLIC" : "VITE_BACKEND_URL"] || ""
    const targetUrl = apiHost + "/api/" + path;
    if (data) {
        opts.method = 'POST'
        opts.headers = {
            'content-type': 'application/json'
        }
        opts.body = JSON.stringify(data)
    }
    const resp = await fetch(targetUrl, Object.assign(opts, { credentials: 'include' }))
    return resp.json()
}