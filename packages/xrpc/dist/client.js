"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.defaultFetchHandler = exports.ServiceClient = exports.Client = void 0;
const lexicon_1 = require("@atproto/lexicon");
const util_1 = require("./util");
const types_1 = require("./types");
class Client {
    constructor() {
        Object.defineProperty(this, "fetch", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: defaultFetchHandler
        });
        Object.defineProperty(this, "lex", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: new lexicon_1.Lexicons()
        });
    }
    // method calls
    //
    async call(serviceUri, methodNsid, params, data, opts) {
        return this.service(serviceUri, fetch).call(methodNsid, params, data, opts);
    }
    service(serviceUri, fetch) {
        return new ServiceClient(this, serviceUri, fetch);
    }
    // schemas
    // =
    addLexicon(doc) {
        this.lex.add(doc);
    }
    addLexicons(docs) {
        for (const doc of docs) {
            this.addLexicon(doc);
        }
    }
    removeLexicon(uri) {
        this.lex.remove(uri);
    }
}
exports.Client = Client;
class ServiceClient {
    constructor(baseClient, serviceUri, fetch) {
        Object.defineProperty(this, "baseClient", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: void 0
        });
        Object.defineProperty(this, "uri", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: void 0
        });
        Object.defineProperty(this, "headers", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: {}
        });
        Object.defineProperty(this, "fetch", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: void 0
        });
        this.baseClient = baseClient;
        this.uri = typeof serviceUri === 'string' ? new URL(serviceUri) : serviceUri;
        this.fetch = fetch;
    }
    setHeader(key, value) {
        this.headers[key] = value;
    }
    unsetHeader(key) {
        delete this.headers[key];
    }
    async call(methodNsid, params, data, opts) {
        const def = this.baseClient.lex.getDefOrThrow(methodNsid);
        if (!def || (def.type !== 'query' && def.type !== 'procedure')) {
            throw new Error(`Invalid lexicon: ${methodNsid}. Must be a query or procedure.`);
        }
        const httpMethod = (0, util_1.getMethodSchemaHTTPMethod)(def);
        const httpUri = (0, util_1.constructMethodCallUri)(methodNsid, def, this.uri, params);
        const httpHeaders = (0, util_1.constructMethodCallHeaders)(def, data, {
            headers: {
                ...this.headers,
                ...opts?.headers,
            },
            encoding: opts?.encoding,
        });
        const res = await this.baseClient.fetch(httpUri, httpMethod, httpHeaders, data, this.fetch || fetch);
        const resCode = (0, util_1.httpResponseCodeToEnum)(res.status);
        if (resCode === types_1.ResponseType.Success) {
            try {
                this.baseClient.lex.assertValidXrpcOutput(methodNsid, res.body);
            }
            catch (e) {
                if (e instanceof lexicon_1.ValidationError) {
                    throw new types_1.XRPCInvalidResponseError(methodNsid, e, res.body);
                }
                else {
                    throw e;
                }
            }
            return new types_1.XRPCResponse(res.body, res.headers);
        }
        else {
            if (res.body && isErrorResponseBody(res.body)) {
                throw new types_1.XRPCError(resCode, res.body.error, res.body.message, res.headers);
            }
            else {
                throw new types_1.XRPCError(resCode);
            }
        }
    }
}
exports.ServiceClient = ServiceClient;
async function defaultFetchHandler(httpUri, httpMethod, httpHeaders, httpReqBody, fetch) {
    try {
        // The duplex field is now required for streaming bodies, but not yet reflected
        // anywhere in docs or types. See whatwg/fetch#1438, nodejs/node#46221.
        const headers = (0, util_1.normalizeHeaders)(httpHeaders);
        const reqInit = {
            method: httpMethod,
            headers,
            body: (0, util_1.encodeMethodCallBody)(headers, httpReqBody),
            duplex: 'half',
        };
        const res = await fetch(httpUri, reqInit);
        const resBody = await res.arrayBuffer();
        return {
            status: res.status,
            headers: Object.fromEntries(res.headers.entries()),
            body: (0, util_1.httpResponseBodyParse)(res.headers.get('content-type'), resBody),
        };
    }
    catch (e) {
        throw new types_1.XRPCError(types_1.ResponseType.Unknown, String(e));
    }
}
exports.defaultFetchHandler = defaultFetchHandler;
function isErrorResponseBody(v) {
    return types_1.errorResponseBody.safeParse(v).success;
}
//# sourceMappingURL=client.js.map