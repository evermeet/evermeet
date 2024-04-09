import { browser } from '$app/environment'; 
import { env } from '$env/dynamic/public';

export async function apiCall (fetch, path, opts) {
    const apiHost = env[browser ? "VITE_BACKEND_URL_PUBLIC" : "VITE_BACKEND_URL"] || ""
    const targetUrl = apiHost + "/api/" + path;
    const resp = await fetch(targetUrl, opts)
    return resp.json()
}