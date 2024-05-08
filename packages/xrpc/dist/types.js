"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.XRPCInvalidResponseError = exports.XRPCError = exports.XRPCResponse = exports.ResponseTypeStrings = exports.ResponseTypeNames = exports.ResponseType = exports.errorResponseBody = void 0;
const zod_1 = require("zod");
exports.errorResponseBody = zod_1.z.object({
    error: zod_1.z.string().optional(),
    message: zod_1.z.string().optional(),
});
var ResponseType;
(function (ResponseType) {
    ResponseType[ResponseType["Unknown"] = 1] = "Unknown";
    ResponseType[ResponseType["InvalidResponse"] = 2] = "InvalidResponse";
    ResponseType[ResponseType["Success"] = 200] = "Success";
    ResponseType[ResponseType["InvalidRequest"] = 400] = "InvalidRequest";
    ResponseType[ResponseType["AuthRequired"] = 401] = "AuthRequired";
    ResponseType[ResponseType["Forbidden"] = 403] = "Forbidden";
    ResponseType[ResponseType["XRPCNotSupported"] = 404] = "XRPCNotSupported";
    ResponseType[ResponseType["PayloadTooLarge"] = 413] = "PayloadTooLarge";
    ResponseType[ResponseType["RateLimitExceeded"] = 429] = "RateLimitExceeded";
    ResponseType[ResponseType["InternalServerError"] = 500] = "InternalServerError";
    ResponseType[ResponseType["MethodNotImplemented"] = 501] = "MethodNotImplemented";
    ResponseType[ResponseType["UpstreamFailure"] = 502] = "UpstreamFailure";
    ResponseType[ResponseType["NotEnoughResources"] = 503] = "NotEnoughResources";
    ResponseType[ResponseType["UpstreamTimeout"] = 504] = "UpstreamTimeout";
})(ResponseType || (exports.ResponseType = ResponseType = {}));
exports.ResponseTypeNames = {
    [ResponseType.InvalidResponse]: 'InvalidResponse',
    [ResponseType.Success]: 'Success',
    [ResponseType.InvalidRequest]: 'InvalidRequest',
    [ResponseType.AuthRequired]: 'AuthenticationRequired',
    [ResponseType.Forbidden]: 'Forbidden',
    [ResponseType.XRPCNotSupported]: 'XRPCNotSupported',
    [ResponseType.PayloadTooLarge]: 'PayloadTooLarge',
    [ResponseType.RateLimitExceeded]: 'RateLimitExceeded',
    [ResponseType.InternalServerError]: 'InternalServerError',
    [ResponseType.MethodNotImplemented]: 'MethodNotImplemented',
    [ResponseType.UpstreamFailure]: 'UpstreamFailure',
    [ResponseType.NotEnoughResources]: 'NotEnoughResources',
    [ResponseType.UpstreamTimeout]: 'UpstreamTimeout',
};
exports.ResponseTypeStrings = {
    [ResponseType.InvalidResponse]: 'Invalid Response',
    [ResponseType.Success]: 'Success',
    [ResponseType.InvalidRequest]: 'Invalid Request',
    [ResponseType.AuthRequired]: 'Authentication Required',
    [ResponseType.Forbidden]: 'Forbidden',
    [ResponseType.XRPCNotSupported]: 'XRPC Not Supported',
    [ResponseType.PayloadTooLarge]: 'Payload Too Large',
    [ResponseType.RateLimitExceeded]: 'Rate Limit Exceeded',
    [ResponseType.InternalServerError]: 'Internal Server Error',
    [ResponseType.MethodNotImplemented]: 'Method Not Implemented',
    [ResponseType.UpstreamFailure]: 'Upstream Failure',
    [ResponseType.NotEnoughResources]: 'Not Enough Resources',
    [ResponseType.UpstreamTimeout]: 'Upstream Timeout',
};
class XRPCResponse {
    constructor(data, headers) {
        Object.defineProperty(this, "data", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: data
        });
        Object.defineProperty(this, "headers", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: headers
        });
        Object.defineProperty(this, "success", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: true
        });
    }
}
exports.XRPCResponse = XRPCResponse;
class XRPCError extends Error {
    constructor(status, error, message, headers) {
        super(message || error || exports.ResponseTypeStrings[status]);
        Object.defineProperty(this, "status", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: status
        });
        Object.defineProperty(this, "error", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: error
        });
        Object.defineProperty(this, "success", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: false
        });
        Object.defineProperty(this, "headers", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: void 0
        });
        if (!this.error) {
            this.error = exports.ResponseTypeNames[status];
        }
        this.headers = headers;
    }
}
exports.XRPCError = XRPCError;
class XRPCInvalidResponseError extends XRPCError {
    constructor(lexiconNsid, validationError, responseBody) {
        super(ResponseType.InvalidResponse, exports.ResponseTypeStrings[ResponseType.InvalidResponse], `The server gave an invalid response and may be out of date.`);
        Object.defineProperty(this, "lexiconNsid", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: lexiconNsid
        });
        Object.defineProperty(this, "validationError", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: validationError
        });
        Object.defineProperty(this, "responseBody", {
            enumerable: true,
            configurable: true,
            writable: true,
            value: responseBody
        });
    }
}
exports.XRPCInvalidResponseError = XRPCInvalidResponseError;
//# sourceMappingURL=types.js.map