import { browser } from '$app/environment'
import { env } from '$env/dynamic/public'
import { Client } from '@atproto/xrpc'
//import { session as origSession } from '$lib/stores'
import { getContext } from 'svelte'

let xrpcClient;

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

export async function xrpcCall({ fetch, user }, id, params, data, opts={}) {
    const apiHost = env[browser ? "VITE_BACKEND_URL_PUBLIC" : "VITE_BACKEND_URL"] || ""
    if (!xrpcClient) {
        xrpcClient = await createXrpcClient(fetch)
    }
    const headers = {}
    if (opts.token) {
        headers.Authorization = 'Bearer ' + opts.token
    } else if (user?.token) {
        headers.Authorization = 'Bearer ' + user.token
    }
    if (opts.mimeType) {
        headers['Content-Type'] = opts.mimeType
    }
    const resp = await xrpcClient.service(apiHost + "/xrpc").call(id, params, data, { headers, ...opts })
    return resp.data
}

export async function blobUpload(args, { body, mimeType }) {
    return xrpcCall(args, 'app.evermeet.object.uploadBlob', null, body, { mimeType })
}

export function imgBlobUrl (did, cid, size) {
    return `http://localhost:3002/width_${size*2},format_avif/${blobUrl(did, cid)}`
}

export function blobUrl (did, cid) {
    return `http://localhost:3000/xrpc/app.evermeet.object.getBlob?did=${did}&cid=${cid}`
}