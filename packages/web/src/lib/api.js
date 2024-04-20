import { browser } from '$app/environment'; 
import { env } from '$env/dynamic/public';
import { Client } from '@atproto/xrpc'

let xrpcClient;

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

export async function internalApiCall(fetch, id) {
    const apiHost = env[browser ? "VITE_BACKEND_URL_PUBLIC" : "VITE_BACKEND_URL"] || ""
    const resp = await fetch(apiHost + '/xrpc/' + id)
    return resp.json()
}

export async function xrpcCall(fetch, ...args) {
    const apiHost = env[browser ? "VITE_BACKEND_URL_PUBLIC" : "VITE_BACKEND_URL"] || ""
    if (!xrpcClient) {
        xrpcClient = new Client()
        const lexicons = await internalApiCall(fetch, '_lexicons')
        xrpcClient.addLexicons(lexicons)
    }
    console.log(args)
    const resp = await xrpcClient.service(apiHost + "/xrpc").call(...args)
    return resp.data
}