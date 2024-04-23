import { browser } from '$app/environment'
import { env } from '$env/dynamic/public'
import { Client } from '@atproto/xrpc'
import { session as origSession } from '$lib/stores'

let xrpcClient;
let session;

origSession.subscribe(cs => {
    session = cs
})

async function createXrpcClient (fetch) {
    const client = new Client()
    const lexicons = await internalApiCall(fetch, '_lexicons')
    client.addLexicons(lexicons)
    return client
}

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

export async function xrpcCall(fetch, id, params, data, opts={}) {
    const apiHost = env[browser ? "VITE_BACKEND_URL_PUBLIC" : "VITE_BACKEND_URL"] || ""
    if (!xrpcClient) {
        xrpcClient = await createXrpcClient(fetch)
    }
    const headers = {}
    if (opts.token) {
        headers.authorization = 'Bearer ' + opts.token
    } else if (session?.accessJwt) {
        headers.authorization = 'Bearer ' + session.accessJwt
    }

    const resp = await xrpcClient.service(apiHost + "/xrpc").call(id, params, data, { headers, ...opts })
    return resp.data
}