import { browser } from '$app/environment'; 

export async function apiCall (fetch, path, opts) {
    const resp = await fetch("http://localhost:3000/api/" + path, opts)
    return resp.json()
}