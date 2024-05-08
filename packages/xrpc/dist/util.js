"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.httpResponseBodyParse = exports.httpResponseCodeToEnum = exports.encodeMethodCallBody = exports.constructMethodCallHeaders = exports.normalizeHeaders = exports.encodeQueryParam = exports.constructMethodCallUri = exports.getMethodSchemaHTTPMethod = void 0;
const lexicon_1 = require("@atproto/lexicon");
const types_1 = require("./types");
function getMethodSchemaHTTPMethod(schema) {
    if (schema.type === 'procedure') {
        return 'post';
    }
    return 'get';
}
exports.getMethodSchemaHTTPMethod = getMethodSchemaHTTPMethod;
function constructMethodCallUri(nsid, schema, serviceUri, params) {
    const uri = new URL(serviceUri);
    uri.pathname = `/xrpc/${nsid}`;
    // given parameters
    if (params) {
        for (const [key, value] of Object.entries(params)) {
            const paramSchema = schema.parameters?.properties?.[key];
            if (!paramSchema) {
                throw new Error(`Invalid query parameter: ${key}`);
            }
            if (value !== undefined) {
                if (paramSchema.type === 'array') {
                    const vals = [];
                    vals.concat(value).forEach((val) => {
                        uri.searchParams.append(key, encodeQueryParam(paramSchema.items.type, val));
                    });
                }
                else {
                    uri.searchParams.set(key, encodeQueryParam(paramSchema.type, value));
                }
            }
        }
    }
    return uri.toString();
}
exports.constructMethodCallUri = constructMethodCallUri;
function encodeQueryParam(type, value) {
    if (type === 'string' || type === 'unknown') {
        return String(value);
    }
    if (type === 'float') {
        return String(Number(value));
    }
    else if (type === 'integer') {
        return String(Number(value) | 0);
    }
    else if (type === 'boolean') {
        return value ? 'true' : 'false';
    }
    else if (type === 'datetime') {
        if (value instanceof Date) {
            return value.toISOString();
        }
        return String(value);
    }
    throw new Error(`Unsupported query param type: ${type}`);
}
exports.encodeQueryParam = encodeQueryParam;
function normalizeHeaders(headers) {
    const normalized = {};
    for (const [header, value] of Object.entries(headers)) {
        normalized[header.toLowerCase()] = value;
    }
    return normalized;
}
exports.normalizeHeaders = normalizeHeaders;
function constructMethodCallHeaders(schema, data, opts) {
    const headers = opts?.headers || {};
    if (schema.type === 'procedure') {
        if (opts?.encoding) {
            headers['Content-Type'] = opts.encoding;
        }
        if (data && typeof data === 'object') {
            if (!headers['Content-Type']) {
                headers['Content-Type'] = 'application/json';
            }
        }
    }
    return headers;
}
exports.constructMethodCallHeaders = constructMethodCallHeaders;
function encodeMethodCallBody(headers, data) {
    if (!headers['content-type'] || typeof data === 'undefined') {
        return undefined;
    }
    if (data instanceof ArrayBuffer) {
        return data;
    }
    if (headers['content-type'].startsWith('text/')) {
        return new TextEncoder().encode(data.toString());
    }
    if (headers['content-type'].startsWith('application/json')) {
        return new TextEncoder().encode((0, lexicon_1.stringifyLex)(data));
    }
    return data;
}
exports.encodeMethodCallBody = encodeMethodCallBody;
function httpResponseCodeToEnum(status) {
    let resCode;
    if (status in types_1.ResponseType) {
        resCode = status;
    }
    else if (status >= 100 && status < 200) {
        resCode = types_1.ResponseType.XRPCNotSupported;
    }
    else if (status >= 200 && status < 300) {
        resCode = types_1.ResponseType.Success;
    }
    else if (status >= 300 && status < 400) {
        resCode = types_1.ResponseType.XRPCNotSupported;
    }
    else if (status >= 400 && status < 500) {
        resCode = types_1.ResponseType.InvalidRequest;
    }
    else {
        resCode = types_1.ResponseType.InternalServerError;
    }
    return resCode;
}
exports.httpResponseCodeToEnum = httpResponseCodeToEnum;
function httpResponseBodyParse(mimeType, data) {
    if (mimeType) {
        if (mimeType.includes('application/json') && data?.byteLength) {
            try {
                const str = new TextDecoder().decode(data);
                return (0, lexicon_1.jsonStringToLex)(str);
            }
            catch (e) {
                throw new types_1.XRPCError(types_1.ResponseType.InvalidResponse, `Failed to parse response body: ${String(e)}`);
            }
        }
        if (mimeType.startsWith('text/') && data?.byteLength) {
            try {
                return new TextDecoder().decode(data);
            }
            catch (e) {
                throw new types_1.XRPCError(types_1.ResponseType.InvalidResponse, `Failed to parse response body: ${String(e)}`);
            }
        }
    }
    if (data instanceof ArrayBuffer) {
        return new Uint8Array(data);
    }
    return data;
}
exports.httpResponseBodyParse = httpResponseBodyParse;
//# sourceMappingURL=util.js.map